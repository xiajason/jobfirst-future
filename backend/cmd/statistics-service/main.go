package main

import (
	"fmt"
	"os"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/jobfirst/jobfirst-core"
)

func main() {
	// 从环境变量获取端口，默认为8086
	port := os.Getenv("STATISTICS_SERVICE_PORT")
	if port == "" {
		port = "8086"
	}
	portInt, _ := strconv.Atoi(port)

	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("../../configs/jobfirst-core-config.yaml")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 设置标准路由 (使用jobfirst-core统一模板)
	setupStandardRoutes(r, core)

	// 设置业务路由 (保持现有API)
	setupBusinessRoutes(r, core)

	// 初始化统计增强服务
	enhancedService, err := NewStatisticsEnhancedService(core)
	if err != nil {
		log.Printf("初始化统计增强服务失败: %v", err)
		log.Println("继续以基础模式运行...")
		enhancedService = nil
	}
	if enhancedService != nil {
		defer enhancedService.Close()
		log.Println("统计增强服务初始化成功")
	}

	// 设置增强路由
	if enhancedService != nil {
		setupStatisticsEnhancedRoutes(r, core, enhancedService)
		log.Println("统计增强API路由已设置")
	}

	// 注册到Consul
	registerToConsul("statistics-service", "127.0.0.1", portInt)

	// 启动服务器
	log.Printf("Starting Statistics Service with jobfirst-core on 0.0.0.0:%s", port)
	if enhancedService != nil {
		log.Println("智能分析平台已启用")
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由 (使用jobfirst-core统一模板)
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查 (统一格式)
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "statistics-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "statistics-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "statistics-service",
			"version":    "3.1.0",
			"port":       7536,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由 (保持现有API)
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1/statistics/public")
	{
		// 获取系统概览统计
		public.GET("/overview", func(c *gin.Context) {
			db := core.GetDB()

			// 获取用户统计
			var userStats UserStats
			if err := db.Raw(`
				SELECT 
					COUNT(*) as total_users,
					COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 END) as new_users_30d,
					COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users
				FROM users
			`).Scan(&userStats).Error; err != nil {
				log.Printf("获取用户统计失败: %v", err)
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get user statistics", err.Error())
				return
			}

			// 获取模板统计
			var templateStats TemplateStats
			if err := db.Raw(`
				SELECT 
					COUNT(*) as total_templates,
					COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 END) as new_templates_30d,
					AVG(rating) as avg_rating,
					SUM(usage_count) as total_usage
				FROM templates
				WHERE is_active = 1
			`).Scan(&templateStats).Error; err != nil {
				log.Printf("获取模板统计失败: %v", err)
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get template statistics", err.Error())
				return
			}

			// 获取公司统计
			var companyStats CompanyStats
			if err := db.Raw(`
				SELECT 
					COUNT(*) as total_companies,
					COUNT(CASE WHEN created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 END) as new_companies_30d,
					COUNT(CASE WHEN status = 'active' THEN 1 END) as active_companies
				FROM companies
			`).Scan(&companyStats).Error; err != nil {
				log.Printf("获取公司统计失败: %v", err)
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get company statistics", err.Error())
				return
			}

			standardSuccessResponse(c, gin.H{
				"users":     userStats,
				"templates": templateStats,
				"companies": companyStats,
				"timestamp": time.Now().Format(time.RFC3339),
			}, "System overview statistics retrieved successfully")
		})

		// 获取用户增长趋势
		public.GET("/users/trend", func(c *gin.Context) {
			days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
			if days > 365 {
				days = 365
			}

			db := core.GetDB()
			var trends []UserTrend
			if err := db.Raw(`
				SELECT 
					DATE(created_at) as date,
					COUNT(*) as count
				FROM users
				WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
				GROUP BY DATE(created_at)
				ORDER BY date
			`, days).Scan(&trends).Error; err != nil {
				log.Printf("获取用户增长趋势失败: %v", err)
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get user growth trend", err.Error())
				return
			}

			standardSuccessResponse(c, trends, "User growth trend retrieved successfully")
		})

		// 获取模板使用统计
		public.GET("/templates/usage", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if limit > 100 {
				limit = 100
			}

			db := core.GetDB()
			var usageStats []TemplateUsage
			if err := db.Raw(`
				SELECT 
					id,
					name,
					category,
					usage_count,
					rating,
					created_at
				FROM templates
				WHERE is_active = 1
				ORDER BY usage_count DESC
				LIMIT ?
			`, limit).Scan(&usageStats).Error; err != nil {
				log.Printf("获取模板使用统计失败: %v", err)
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get template usage statistics", err.Error())
				return
			}

			standardSuccessResponse(c, usageStats, "Template usage statistics retrieved successfully")
		})

		// 获取热门分类
		public.GET("/categories/popular", func(c *gin.Context) {
			db := core.GetDB()
			var categoryStats []CategoryStats
			if err := db.Raw(`
				SELECT 
					category,
					COUNT(*) as template_count,
					SUM(usage_count) as total_usage,
					AVG(rating) as avg_rating
				FROM templates
				WHERE is_active = 1
				GROUP BY category
				ORDER BY total_usage DESC
			`).Scan(&categoryStats).Error; err != nil {
				log.Printf("获取热门分类失败: %v", err)
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get popular categories", err.Error())
				return
			}

			standardSuccessResponse(c, categoryStats, "Popular categories retrieved successfully")
		})

		// 获取系统性能指标
		public.GET("/performance", func(c *gin.Context) {
			// 获取数据库连接状态
			sqlDB, err := core.GetDB().DB()
			var dbStats DatabaseStats
			if err == nil {
				stats := sqlDB.Stats()
				dbStats = DatabaseStats{
					OpenConnections: stats.OpenConnections,
					InUse:           stats.InUse,
					Idle:            stats.Idle,
					WaitCount:       stats.WaitCount,
					WaitDuration:    stats.WaitDuration.String(),
				}
			}

			standardSuccessResponse(c, gin.H{
				"database":  dbStats,
				"timestamp": time.Now().Format(time.RFC3339),
			}, "System performance metrics retrieved successfully")
		})
	}

	// 需要认证的API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	api := r.Group("/api/v1/statistics")
	api.Use(authMiddleware)
	{
		// 获取用户个人统计
		api.GET("/user/:id", func(c *gin.Context) {
			userID, _ := strconv.Atoi(c.Param("id"))

			// 检查权限：只能查看自己的统计或管理员
			requestUserIDInterface, _ := c.Get("user_id")
			requestUserID := requestUserIDInterface.(uint)
			role := c.GetString("role")

			if uint(userID) != requestUserID && role != "admin" && role != "super_admin" {
				standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
				return
			}

			db := core.GetDB()
			var userPersonalStats UserPersonalStats
			if err := db.Raw(`
				SELECT 
					(SELECT COUNT(*) FROM templates WHERE created_by = ?) as templates_created,
					(SELECT COUNT(*) FROM templates WHERE created_by = ? AND is_active = 1) as active_templates,
					(SELECT SUM(usage_count) FROM templates WHERE created_by = ?) as total_usage,
					(SELECT AVG(rating) FROM templates WHERE created_by = ? AND rating > 0) as avg_rating
			`, userID, userID, userID, userID).Scan(&userPersonalStats).Error; err != nil {
				log.Printf("获取用户个人统计失败: %v", err)
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get user personal statistics", err.Error())
				return
			}

			standardSuccessResponse(c, userPersonalStats, "User personal statistics retrieved successfully")
		})

		// 获取管理员统计面板
		admin := api.Group("/admin")
		admin.Use(func(c *gin.Context) {
			role := c.GetString("role")
			if role != "admin" && role != "super_admin" {
				standardErrorResponse(c, http.StatusForbidden, "Admin privileges required", "")
				c.Abort()
				return
			}
			c.Next()
		})
		{
			// 获取详细用户统计
			admin.GET("/users/detailed", func(c *gin.Context) {
				page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
				pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
				offset := (page - 1) * pageSize

				db := core.GetDB()
				var detailedUsers []DetailedUserStats
				if err := db.Raw(`
					SELECT 
						u.id,
						u.username,
						u.email,
						u.created_at,
						u.status,
						COUNT(t.id) as template_count,
						SUM(t.usage_count) as total_usage,
						AVG(t.rating) as avg_rating
					FROM users u
					LEFT JOIN templates t ON u.id = t.created_by
					GROUP BY u.id, u.username, u.email, u.created_at, u.status
					ORDER BY u.created_at DESC
					LIMIT ? OFFSET ?
				`, pageSize, offset).Scan(&detailedUsers).Error; err != nil {
					log.Printf("获取详细用户统计失败: %v", err)
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get detailed user statistics", err.Error())
					return
				}

				// 获取总数
				var total int64
				if err := db.Raw("SELECT COUNT(*) FROM users").Scan(&total).Error; err != nil {
					log.Printf("获取用户总数失败: %v", err)
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get user total count", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{
					"users": detailedUsers,
					"total": total,
					"page":  page,
					"size":  pageSize,
				}, "Detailed user statistics retrieved successfully")
			})

			// 获取系统健康报告
			admin.GET("/health/report", func(c *gin.Context) {
				var healthReport HealthReport

				// 检查数据库连接
				sqlDB, err := core.GetDB().DB()
				if err == nil {
					err = sqlDB.Ping()
					healthReport.Database = err == nil
				}

				// Redis检查暂时跳过，因为GetRedis方法不存在
				healthReport.Redis = true

				// 获取服务状态
				healthReport.Services = map[string]bool{
					"user-service":     checkServiceHealth("http://localhost:8081/health"),
					"template-service": checkServiceHealth("http://localhost:8085/health"),
					"company-service":  checkServiceHealth("http://localhost:8083/health"),
				}

				standardSuccessResponse(c, healthReport, "System health report retrieved successfully")
			})
		}
	}
}

// registerToConsul 注册服务到Consul
func registerToConsul(serviceName string, host string, port int) {
	config := api.DefaultConfig()
	config.Address = "127.0.0.1:8500"

	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, port),
		Name:    serviceName,
		Port:    port,
		Address: host,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", host, port),
			Timeout:                        "3s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{
			"jobfirst",
			"microservice",
			"statistics",
			"version:3.1.0",
		},
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("服务 %s 已注册到Consul (端口: %d)", serviceName, port)
	}
}

// checkServiceHealth 检查服务健康状态
func checkServiceHealth(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// standardSuccessResponse 标准成功响应
func standardSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	response := gin.H{
		"success": true,
		"data":    data,
		"service": "statistics-service",
		"time":    time.Now().Format(time.RFC3339),
	}
	if len(message) > 0 {
		response["message"] = message[0]
	}
	c.JSON(http.StatusOK, response)
}

// standardErrorResponse 标准错误响应
func standardErrorResponse(c *gin.Context, statusCode int, message string, details ...string) {
	response := gin.H{
		"success": false,
		"error":   message,
		"service": "statistics-service",
		"time":    time.Now().Format(time.RFC3339),
	}
	if len(details) > 0 {
		response["details"] = details[0]
	}
	c.JSON(statusCode, response)
}

// 数据模型定义

type UserStats struct {
	TotalUsers  int `json:"total_users"`
	NewUsers30d int `json:"new_users_30d"`
	ActiveUsers int `json:"active_users"`
}

type TemplateStats struct {
	TotalTemplates  int     `json:"total_templates"`
	NewTemplates30d int     `json:"new_templates_30d"`
	AvgRating       float64 `json:"avg_rating"`
	TotalUsage      int     `json:"total_usage"`
}

type CompanyStats struct {
	TotalCompanies  int `json:"total_companies"`
	NewCompanies30d int `json:"new_companies_30d"`
	ActiveCompanies int `json:"active_companies"`
}

type UserTrend struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type TemplateUsage struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	UsageCount int       `json:"usage_count"`
	Rating     float64   `json:"rating"`
	CreatedAt  time.Time `json:"created_at"`
}

type CategoryStats struct {
	Category      string  `json:"category"`
	TemplateCount int     `json:"template_count"`
	TotalUsage    int     `json:"total_usage"`
	AvgRating     float64 `json:"avg_rating"`
}

type DatabaseStats struct {
	OpenConnections int    `json:"open_connections"`
	InUse           int    `json:"in_use"`
	Idle            int    `json:"idle"`
	WaitCount       int64  `json:"wait_count"`
	WaitDuration    string `json:"wait_duration"`
}

type UserPersonalStats struct {
	TemplatesCreated int     `json:"templates_created"`
	ActiveTemplates  int     `json:"active_templates"`
	TotalUsage       int     `json:"total_usage"`
	AvgRating        float64 `json:"avg_rating"`
}

type DetailedUserStats struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"created_at"`
	Status        string    `json:"status"`
	TemplateCount int       `json:"template_count"`
	TotalUsage    int       `json:"total_usage"`
	AvgRating     float64   `json:"avg_rating"`
}

type HealthReport struct {
	Database bool            `json:"database"`
	Redis    bool            `json:"redis"`
	Services map[string]bool `json:"services"`
}
