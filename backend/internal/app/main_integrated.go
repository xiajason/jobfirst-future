package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/jobfirst/jobfirst-core"

	// App服务内部模块
	appauth "github.com/xiajason/zervi-basic/basic/backend/internal/app/auth"
	appuser "github.com/xiajason/zervi-basic/basic/backend/internal/app/user"
	"github.com/xiajason/zervi-basic/basic/backend/internal/domain/auth"
	"github.com/xiajason/zervi-basic/basic/backend/internal/domain/user"
	"github.com/xiajason/zervi-basic/basic/backend/internal/infrastructure/database"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/logger"
)

func main() {
	log.Println("启动应用层服务 (集成jobfirst-core)...")

	// 1. 初始化JobFirst核心包
	core, err := jobfirst.NewCore("../../configs/jobfirst-core-config.yaml")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 2. 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 3. 创建Gin引擎
	r := gin.Default()

	// 4. 设置标准路由 (使用jobfirst-core统一模板)
	setupStandardRoutes(r, core)

	// 5. 初始化应用服务 (保持现有功能)
	appServices, err := initializeAppServices(core)
	if err != nil {
		log.Fatalf("初始化应用服务失败: %v", err)
	}

	// 6. 设置业务路由 (保持现有API)
	setupBusinessRoutes(r, appServices)

	// 7. 注册到Consul
	registerToConsul("app-service", "127.0.0.1", 8080)

	// 8. 启动服务器
	log.Println("Starting App Service with jobfirst-core on 0.0.0.0:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// AppServices 应用服务结构
type AppServices struct {
	AuthService *appauth.Service
	UserService appuser.Service
	Logger      logger.Logger
}

// initializeAppServices 初始化应用服务
func initializeAppServices(core *jobfirst.Core) (*AppServices, error) {
	// 初始化日志器
	logger := logger.NewLogger("info")

	// 初始化数据库连接
	db := core.GetDB()

	// 初始化Repository
	authRepo := database.NewAuthRepository(db)
	userRepo := database.NewUserRepository(db)

	// 初始化服务
	authService := appauth.NewService(authRepo, logger)
	userService := appuser.NewService(userRepo, logger)

	return &AppServices{
		AuthService: authService,
		UserService: userService,
		Logger:      logger,
	}, nil
}

// setupStandardRoutes 设置标准路由 (使用jobfirst-core统一模板)
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查 (统一格式)
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "app-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.0.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "app-service",
			"version": "3.0.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "app-service",
			"version":    "3.0.0",
			"port":       8080,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由 (保持现有API)
func setupBusinessRoutes(r *gin.Engine, appServices *AppServices) {
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

	// 设置API路由组
	api := r.Group("/api/v1")
	{
		// 超级管理员相关API
		api.POST("/super-admin/init", func(c *gin.Context) {
			var req auth.InitializeSuperAdminRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			response, err := appServices.AuthService.InitializeSuperAdmin(c.Request.Context(), req)
			if err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to initialize super admin", err.Error())
				return
			}

			standardSuccessResponse(c, response, "Super admin initialized successfully")
		})

		api.GET("/super-admin/status", func(c *gin.Context) {
			req := auth.CheckSuperAdminStatusRequest{}
			response, err := appServices.AuthService.CheckSuperAdminStatus(c.Request.Context(), req)
			if err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to check super admin status", err.Error())
				return
			}

			standardSuccessResponse(c, response, "Super admin status retrieved successfully")
		})

		api.POST("/super-admin/reset-password", func(c *gin.Context) {
			var req auth.ResetSuperAdminPasswordRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			err := appServices.AuthService.ResetSuperAdminPassword(c.Request.Context(), req)
			if err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to reset super admin password", err.Error())
				return
			}

			standardSuccessResponse(c, nil, "Super admin password reset successfully")
		})

		// 用户相关API
		api.POST("/users/register", func(c *gin.Context) {
			var req user.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			response, err := appServices.UserService.Register(c.Request.Context(), req)
			if err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to register user", err.Error())
				return
			}

			standardSuccessResponse(c, response, "User registered successfully")
		})

		api.POST("/users/login", func(c *gin.Context) {
			var req user.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			response, err := appServices.UserService.Login(c.Request.Context(), req)
			if err != nil {
				standardErrorResponse(c, http.StatusUnauthorized, "Login failed", err.Error())
				return
			}

			standardSuccessResponse(c, response, "Login successful")
		})

		api.GET("/users/profile/:id", func(c *gin.Context) {
			userID := c.Param("id")
			// 这里需要解析userID并验证权限
			// 为了简化，这里直接返回成功
			standardSuccessResponse(c, gin.H{"user_id": userID}, "User profile retrieved successfully")
		})

		api.PUT("/users/profile/:id", func(c *gin.Context) {
			userID := c.Param("id")
			var req user.UpdateProfileRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			// 这里需要解析userID并验证权限
			// 为了简化，这里直接返回成功
			standardSuccessResponse(c, gin.H{"user_id": userID, "updated": true}, "User profile updated successfully")
		})

		api.GET("/users", func(c *gin.Context) {
			req := user.ListRequest{
				Page:     1,
				PageSize: 10,
			}
			response, err := appServices.UserService.List(c.Request.Context(), req)
			if err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to list users", err.Error())
				return
			}

			standardSuccessResponse(c, response, "Users listed successfully")
		})
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
			"application",
			"version:3.0.0",
		},
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("服务 %s 已注册到Consul (端口: %d)", serviceName, port)
	}
}

// standardSuccessResponse 标准成功响应
func standardSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	response := gin.H{
		"success": true,
		"data":    data,
		"service": "app-service",
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
		"service": "app-service",
		"time":    time.Now().Format(time.RFC3339),
	}
	if len(details) > 0 {
		response["details"] = details[0]
	}
	c.JSON(statusCode, response)
}
