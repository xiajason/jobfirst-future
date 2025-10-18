package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	multidatabase "github.com/xiajason/zervi-basic/basic/backend/pkg/multi-database"
)

func main() {
	log.Println("启动多数据库管理服务...")

	// 加载配置
	config, err := multidatabase.LoadConfigFromFile("../../configs/multi-database-config.yaml")
	if err != nil {
		log.Printf("从文件加载配置失败，尝试从环境变量加载: %v", err)
		config = multidatabase.LoadConfigFromEnv()
	}

	// 打印配置信息用于调试
	log.Printf("PostgreSQL配置: Host=%s, Port=%d, User=%s, Database=%s",
		config.PostgreSQL.Host, config.PostgreSQL.Port, config.PostgreSQL.User, config.PostgreSQL.Database)

	// 创建多数据库管理器
	manager, err := multidatabase.NewMultiDatabaseManager(config)
	if err != nil {
		log.Fatalf("创建多数据库管理器失败: %v", err)
	}
	defer manager.Close()

	// 检查数据库健康状态
	if !manager.IsHealthy() {
		log.Println("警告: 部分数据库连接不健康")
	}

	// 创建同步服务
	syncService := multidatabase.NewSyncService(manager, 5) // 5个工作协程
	syncService.Start()
	defer syncService.Stop()

	// 创建一致性检查器
	consistencyConfig := &multidatabase.ConsistencyConfig{
		CheckInterval: 5 * 60 * time.Second, // 5分钟检查一次
		Timeout:       30 * time.Second,     // 30秒超时
		MaxRetries:    3,
		AutoRepair:    false, // 暂时不启用自动修复
		Rules: []multidatabase.ConsistencyRule{
			{
				ID:          "mysql_postgres_user_sync",
				Name:        "MySQL到PostgreSQL用户数据同步检查",
				Description: "检查用户数据在MySQL和PostgreSQL之间的一致性",
				Source:      multidatabase.DatabaseTypeMySQL,
				Target:      multidatabase.DatabaseTypePostgreSQL,
				Query:       "SELECT id, username, email, created_at FROM users WHERE deleted_at IS NULL",
				Enabled:     true,
			},
			{
				ID:          "mysql_neo4j_company_sync",
				Name:        "MySQL到Neo4j企业数据同步检查",
				Description: "检查企业数据在MySQL和Neo4j之间的一致性",
				Source:      multidatabase.DatabaseTypeMySQL,
				Target:      multidatabase.DatabaseTypeNeo4j,
				Query:       "MATCH (c:Company) RETURN c.id, c.name, c.industry, c.created_at",
				Enabled:     true,
			},
		},
	}

	consistencyChecker := multidatabase.NewConsistencyChecker(manager, consistencyConfig)

	// 创建事务管理器
	transactionConfig := &multidatabase.TransactionConfig{
		DefaultTimeout:        30, // 30秒默认超时
		MaxRetries:            3,
		RetryInterval:         1,     // 1秒重试间隔
		TwoPhaseCommitTimeout: 10,    // 10秒两阶段提交超时
		EnableDistributedLock: false, // 暂时不启用分布式锁
	}

	transactionManager := multidatabase.NewTransactionManager(manager, transactionConfig)

	// 创建API服务
	apiService := multidatabase.NewAPIService(manager, syncService, consistencyChecker, transactionManager)

	// 设置Gin路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置API路由
	apiService.SetupRoutes(r)

	// 添加健康检查端点
	r.GET("/health", func(c *gin.Context) {
		healthy := manager.IsHealthy()
		status := "healthy"
		if !healthy {
			status = "unhealthy"
		}
		c.JSON(200, gin.H{
			"status":    status,
			"service":   "multi-database-service",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// 启动一致性检查
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go consistencyChecker.Start(ctx)

	// 启动过期事务清理
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				transactionManager.CleanupExpiredTransactions()
			}
		}
	}()

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8607"
	}

	log.Printf("多数据库管理服务启动在端口: %s", port)
	log.Println("API端点:")
	log.Println("  GET  /health - 健康检查")
	log.Println("  GET  /api/v1/multi-database/health - 数据库健康检查")
	log.Println("  GET  /api/v1/multi-database/metrics - 数据库指标")
	log.Println("  POST /api/v1/multi-database/sync/task - 添加同步任务")
	log.Println("  GET  /api/v1/multi-database/sync/status - 同步状态")
	log.Println("  GET  /api/v1/multi-database/consistency/results - 一致性检查结果")
	log.Println("  POST /api/v1/multi-database/transaction/begin - 开始事务")
	log.Println("  POST /api/v1/multi-database/transaction/:id/commit - 提交事务")

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("收到关闭信号，正在优雅关闭...")

		// 停止服务
		syncService.Stop()
		manager.Close()

		log.Println("多数据库管理服务已关闭")
		os.Exit(0)
	}()

	// 启动HTTP服务器
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动HTTP服务器失败: %v", err)
	}
}
