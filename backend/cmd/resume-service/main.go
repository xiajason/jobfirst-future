package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/jobfirst/jobfirst-core"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 全局变量
var (
	db *gorm.DB
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "resume-service"
	}

	// 从环境变量获取端口，默认为8082
	port := os.Getenv("RESUME_SERVICE_PORT")
	if port == "" {
		port = "8082"
	}
	portInt, _ := strconv.Atoi(port)

	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("../../configs/jobfirst-core-config.yaml")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 初始化MySQL数据库连接
	if err := initMySQLDatabase(); err != nil {
		log.Fatalf("初始化MySQL数据库失败: %v", err)
	}

	// 初始化SQLite管理器
	basePath := "/Users/szjason72/jobfirst-future/user_data"
	if err := InitSecureSQLiteManager(basePath); err != nil {
		log.Fatalf("初始化SQLite管理器失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.Default()

	// 添加健康检查端点
	r.GET("/health", healthCheck)
	r.GET("/info", serviceInfo)

	// 添加简历上传端点 - 使用完整的业务逻辑
	r.POST("/api/v1/resume/resumes/upload", handleResumeUploadWithFullLogic)

	// 注册到Consul
	registerToConsul("resume-service", "127.0.0.1", portInt)

	// 启动服务
	log.Printf("Starting Resume Service with jobfirst-core on 0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Resume Service启动失败: %v", err)
	}
}

func registerToConsul(serviceName, serviceHost string, servicePort int) {
	// 创建Consul客户端
	config := api.DefaultConfig()
	config.Address = "localhost:8500"
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	// 服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Address: serviceHost,
		Port:    servicePort,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceHost, servicePort),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{"resume", "microservice"},
		Meta: map[string]string{
			"version":     "3.1.0",
			"environment": "production",
			"port":        "7532",
		},
	}

	// 注册服务
	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("服务 %s 已注册到Consul: %s:%d", serviceName, serviceHost, servicePort)
	}
}

// 健康检查端点
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "resume-service",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "3.1.0",
	})
}

// 服务信息端点
func serviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "resume-service",
		"version":   "3.1.0",
		"port":      7532,
		"status":    "running",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// initMySQLDatabase 初始化MySQL数据库连接
func initMySQLDatabase() error {
	// 从环境变量读取MySQL连接配置
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "jobfirst_future"
	}

	// MySQL连接配置
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接MySQL数据库失败: %v", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&ResumeMetadata{}, &ResumeParsingTask{}, &ResumeStructuredDataRecord{}); err != nil {
		return fmt.Errorf("迁移表结构失败: %v", err)
	}

	log.Println("✅ MySQL数据库连接成功")
	return nil
}

// handleResumeUploadWithFullLogic 使用完整业务逻辑处理简历上传
func handleResumeUploadWithFullLogic(c *gin.Context) {
	// 从请求头获取用户ID（从JWT token中解析）
	userID := uint(1) // 临时使用固定用户ID，实际应该从JWT token中解析

	// 使用现成的完整业务逻辑
	handleResumeDocumentUploadWithMinerU(c, db, userID)
}
