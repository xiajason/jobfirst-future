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
	"github.com/jobfirst/jobfirst-core/auth"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "user-service"
	}

	// 从环境变量获取端口，默认为8081
	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "8081"
	}

	// 初始化JobFirst核心包
	configPath := os.Getenv("JOBFIRST_CONFIG_PATH")
	if configPath == "" {
		configPath = "../../configs/user-service-config.yaml"
	}

	core, err := jobfirst.NewCore(configPath)
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

	// 注册到Consul
	portInt, _ := strconv.Atoi(port)
	registerToConsul("user-service", "127.0.0.1", portInt)

	// 启动服务器
	log.Printf("Starting User Service with jobfirst-core on 0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由 (使用jobfirst-core统一模板)
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查 (统一格式)
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, health)
	})

	// 服务信息
	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "8081"
	}

	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "user-service",
			"version":    "3.1.0",
			"port":       port,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "user-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})
}

// setupBusinessRoutes 设置业务路由 (保持现有API)
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1")
	{
		// 用户注册
		public.POST("/auth/register", func(c *gin.Context) {
			var req auth.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			// 使用核心包的认证管理器
			user, err := core.AuthManager.Register(req)
			if err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "注册失败", err.Error())
				return
			}

			standardSuccessResponse(c, gin.H{
				"user_id":  user.ID,
				"username": user.Username,
			}, "注册成功")
		})

		// 用户登录
		public.POST("/auth/login", func(c *gin.Context) {
			var req auth.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
				return
			}

			// 使用核心包的认证管理器
			token, expiresAt, err := core.AuthManager.Login(req.Username, req.Password)
			if err != nil {
				standardErrorResponse(c, http.StatusUnauthorized, "登录失败", err.Error())
				return
			}

			standardSuccessResponse(c, gin.H{
				"token":      token,
				"expires_at": expiresAt.Format(time.RFC3339),
			}, "登录成功")
		})
	}

	// 需要认证的API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	api := r.Group("/api/v1")
	api.Use(authMiddleware)
	{
		// 用户相关API
		users := api.Group("/users")
		{
			// 获取用户信息
			users.GET("/profile", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "用户信息不存在", "")
					return
				}
				userID := userIDInterface.(uint)

				standardSuccessResponse(c, gin.H{
					"user_id":  userID,
					"username": c.GetString("username"),
					"role":     c.GetString("role"),
				}, "获取用户信息成功")
			})

			// 更新用户信息
			users.PUT("/profile", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "用户信息不存在", "")
					return
				}
				userID := userIDInterface.(uint)

				var req struct {
					Email string `json:"email"`
					Phone string `json:"phone"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 这里可以添加更新用户信息的逻辑
				standardSuccessResponse(c, gin.H{
					"user_id": userID,
					"email":   req.Email,
					"phone":   req.Phone,
				}, "更新用户信息成功")
			})
		}
	}
}

// registerToConsul 注册服务到Consul
func registerToConsul(serviceName, serviceHost string, servicePort int) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Tags:    []string{"user", "auth", "rbac", "jobfirst", "microservice", "version:3.1.0"},
		Port:    servicePort,
		Address: serviceHost,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceHost, servicePort),
			Timeout:                        "3s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("%s registered with Consul successfully", serviceName)
	}
}

// standardSuccessResponse 标准成功响应
func standardSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	response := gin.H{
		"success": true,
		"data":    data,
		"service": "user-service",
		"time":    time.Now().Format(time.RFC3339),
	}
	if len(message) > 0 {
		response["message"] = message[0]
	}
	c.JSON(http.StatusOK, response)
}

// standardErrorResponse 标准错误响应
func standardErrorResponse(c *gin.Context, statusCode int, message, details string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
		"details": details,
		"service": "user-service",
		"time":    time.Now().Format(time.RFC3339),
	})
}
