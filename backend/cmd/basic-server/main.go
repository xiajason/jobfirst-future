package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/xiajason/zervi-basic/basic/backend/internal/handlers"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/cache"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/cluster"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/config"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/consul"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/database"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/middleware"
)

func main() {
	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "basic-server"
	}

	// 加载环境配置
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default configuration")
	}

	// 初始化配置
	cfg := config.Load()

	// 初始化数据库连接
	db, err := database.InitMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseMySQL(db)

	// 初始化GORM数据库连接（用于V3.0 API）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database with GORM: %v", err)
	}

	// 初始化Redis缓存 (可选)
	var redisClient *redis.Client
	redisClient, err = cache.InitRedis(cfg.Redis)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis cache - some features may be limited")
		redisClient = nil
	} else {
		defer cache.CloseRedis(redisClient)
		log.Println("Redis cache connected successfully")
	}

	// 初始化集群管理器
	var clusterManager cluster.ClusterManager
	clusterConfig := cluster.DefaultClusterConfig()
	clusterConfig.NodeID = fmt.Sprintf("basic-server-node-%d", time.Now().Unix())
	clusterManager = cluster.NewManager(clusterConfig)

	// 启动集群管理器
	if err := clusterManager.Start(); err != nil {
		log.Printf("Warning: Failed to start cluster manager: %v", err)
		log.Println("Continuing without cluster management")
		clusterManager = nil
	} else {
		defer clusterManager.Stop()
		log.Println("Cluster manager initialized successfully")

		// 注册当前节点到集群
		nodeConfig := cluster.NodeConfig{
			NodeID: clusterConfig.NodeID,
			Host:   cfg.Server.Host,
			Port:   func() int { port, _ := strconv.Atoi(cfg.Server.Port); return port }(),
			Weight: 100,
			Status: cluster.NodeStatusActive,
			Metadata: map[string]string{
				"service": "basic-server",
				"version": "1.0.0",
				"mode":    "standalone",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := clusterManager.RegisterNode(clusterConfig.NodeID, nodeConfig); err != nil {
			log.Printf("Warning: Failed to register node to cluster: %v", err)
		} else {
			log.Printf("Node registered to cluster: %s", clusterConfig.NodeID)
		}
	}

	// 初始化Consul服务管理器
	var consulManager *consul.ServiceManager
	var microserviceRegistry *consul.MicroserviceRegistry
	if cfg.Consul.Enabled {
		consulManager, err = consul.NewServiceManager(&cfg.Consul)
		if err != nil {
			log.Printf("Warning: Failed to initialize Consul manager: %v", err)
			log.Println("Continuing without Consul service discovery")
		} else {
			defer consulManager.Close()
			log.Println("Consul service manager initialized successfully")

			// 初始化微服务注册器
			microserviceRegistry = consul.NewMicroserviceRegistry(consulManager)

			// 注册默认微服务
			if err := microserviceRegistry.RegisterDefaultServices(cfg); err != nil {
				log.Printf("Warning: Failed to register default microservices: %v", err)
			}

			// 启动定期健康检查
			go microserviceRegistry.StartPeriodicHealthCheck(30 * time.Second)
		}
	} else {
		log.Println("Consul service discovery is disabled")
	}

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由
	router := gin.Default()

	// 配置CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "API-Version", "X-Requested-With", "X-API-Key", "X-Client-Version"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 中间件
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		dbHealthy := true
		if err := db.Ping(); err != nil {
			dbHealthy = false
		}

		// 检查Redis连接
		redisHealthy := true
		if redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := redisClient.Ping(ctx).Err(); err != nil {
				redisHealthy = false
			}
		} else {
			redisHealthy = false // Redis未连接
		}

		// 检查Consul连接
		consulHealthy := false
		if consulManager != nil {
			consulHealthy = consulManager.IsHealthy()
		}

		// 检查集群状态
		clusterHealthy := false
		if clusterManager != nil {
			status, err := clusterManager.GetClusterStatus()
			clusterHealthy = err == nil && status.ActiveNodes > 0
		}

		// 整体健康状态 - 只依赖核心数据库
		overallHealthy := dbHealthy
		// Redis、Consul和集群是可选的，不影响整体健康状态

		statusCode := http.StatusOK
		if !overallHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"status":    overallHealthy,
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
			"mode":      "basic",
			"checks": gin.H{
				"database": gin.H{
					"status": dbHealthy,
					"type":   "mysql",
				},
				"cache": gin.H{
					"status":  redisHealthy,
					"type":    "redis",
					"enabled": redisClient != nil,
				},
				"consul": gin.H{
					"status":  consulHealthy,
					"enabled": consulManager != nil,
				},
				"cluster": gin.H{
					"status":  clusterHealthy,
					"enabled": clusterManager != nil,
				},
			},
			"services": gin.H{
				"basic_server": "running",
				"mode":         "standalone",
			},
		})
	})

	// 处理OPTIONS请求 - 排除简历相关路径
	router.OPTIONS("/api/v1/auth/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/api/v1/user/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/api/v1/ai/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/api/v1/banner/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/api/v1/statistics/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/api/v1/template/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/api/v1/job/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})

	// API路由组
	api := router.Group("/api/v1")
	{

		// 服务状态路由
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "running",
				"services": gin.H{
					"basic_server": gin.H{
						"status": "running",
						"port":   cfg.Server.Port,
						"health": fmt.Sprintf("http://localhost:%s/health", cfg.Server.Port),
						"mode":   "standalone",
					},
				},
				"database": gin.H{
					"mysql": "connected",
					"redis": "connected",
				},
			})
		})

		// 系统信息路由
		api.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"name":        "JobFirst Basic Version",
				"description": "个人简历管理系统 - 基础版本",
				"version":     "1.0.0",
				"environment": cfg.Environment,
				"mode":        cfg.Mode,
				"timestamp":   time.Now().Format(time.RFC3339),
			})
		})

		// 数据库状态路由
		api.GET("/database/status", func(c *gin.Context) {
			// 检查MySQL连接
			mysqlStatus := "connected"
			if err := db.Ping(); err != nil {
				mysqlStatus = "disconnected"
			}

			// 检查Redis连接
			redisStatus := "disabled"
			if redisClient != nil {
				redisStatus = "connected"
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := redisClient.Ping(ctx).Err(); err != nil {
					redisStatus = "disconnected"
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"mysql": gin.H{
					"status": mysqlStatus,
					"host":   cfg.Database.Host,
					"port":   cfg.Database.Port,
					"name":   cfg.Database.Name,
				},
				"redis": gin.H{
					"status":  redisStatus,
					"host":    cfg.Redis.Host,
					"port":    cfg.Redis.Port,
					"enabled": redisClient != nil,
				},
			})
		})

		// Consul状态路由
		api.GET("/consul/status", func(c *gin.Context) {
			if consulManager == nil {
				c.JSON(http.StatusOK, gin.H{
					"enabled": false,
					"message": "Consul service discovery is disabled",
				})
				return
			}

			status := consulManager.GetConsulStatus()
			c.JSON(http.StatusOK, status)
		})

		// 服务发现路由
		api.GET("/consul/services", func(c *gin.Context) {
			if consulManager == nil {
				c.JSON(http.StatusOK, gin.H{
					"enabled": false,
					"message": "Consul service discovery is disabled",
				})
				return
			}

			services, err := consulManager.ListServices()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"enabled":  true,
				"services": services,
			})
		})

		// 微服务注册信息路由
		api.GET("/consul/microservices", func(c *gin.Context) {
			if microserviceRegistry == nil {
				c.JSON(http.StatusOK, gin.H{
					"enabled": false,
					"message": "Microservice registry is disabled",
				})
				return
			}

			info := microserviceRegistry.GetServiceDiscoveryInfo()
			c.JSON(http.StatusOK, info)
		})

		// 微服务健康检查路由
		api.GET("/consul/health", func(c *gin.Context) {
			if microserviceRegistry == nil {
				c.JSON(http.StatusOK, gin.H{
					"enabled": false,
					"message": "Microservice registry is disabled",
				})
				return
			}

			healthResults := microserviceRegistry.HealthCheckAll()
			c.JSON(http.StatusOK, gin.H{
				"enabled":      true,
				"health_check": healthResults,
				"timestamp":    time.Now().Format(time.RFC3339),
			})
		})

		// 集群状态路由
		api.GET("/cluster/status", func(c *gin.Context) {
			if clusterManager == nil {
				c.JSON(http.StatusOK, gin.H{
					"enabled": false,
					"message": "Cluster management is disabled",
				})
				return
			}

			status, err := clusterManager.GetClusterStatus()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"enabled": true,
				"status":  status,
			})
		})

		// 集群节点信息路由
		api.GET("/cluster/nodes", func(c *gin.Context) {
			if clusterManager == nil {
				c.JSON(http.StatusOK, gin.H{
					"enabled": false,
					"message": "Cluster management is disabled",
				})
				return
			}

			status, err := clusterManager.GetClusterStatus()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"enabled": true,
				"nodes":   status.Nodes,
			})
		})

		// 用户管理API
		api.GET("/users", func(c *gin.Context) {
			// 从数据库获取用户列表
			var users []map[string]interface{}
			rows, err := db.Query("SELECT id, username, email, phone, status, created_at FROM users LIMIT 10")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id int
				var username, email, phone, status string
				var createdAt time.Time

				if err := rows.Scan(&id, &username, &email, &phone, &status, &createdAt); err != nil {
					continue
				}

				user := map[string]interface{}{
					"id":         id,
					"username":   username,
					"email":      email,
					"phone":      phone,
					"status":     status,
					"created_at": createdAt.Format("2006-01-02 15:04:05"),
				}
				users = append(users, user)
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   users,
				"count":  len(users),
			})
		})

		// 简历管理API
		api.GET("/resumes", func(c *gin.Context) {
			// 从数据库获取简历列表
			var resumes []map[string]interface{}
			rows, err := db.Query("SELECT r.id, r.title, r.template_id, r.status, r.view_count, r.share_count, r.created_at, u.username FROM resumes r LEFT JOIN users u ON r.user_id = u.id LIMIT 10")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id int
				var title, templateID, status, username string
				var viewCount, shareCount int
				var createdAt time.Time

				if err := rows.Scan(&id, &title, &templateID, &status, &viewCount, &shareCount, &createdAt, &username); err != nil {
					continue
				}

				resume := map[string]interface{}{
					"id":          id,
					"title":       title,
					"template_id": templateID,
					"status":      status,
					"view_count":  viewCount,
					"share_count": shareCount,
					"username":    username,
					"created_at":  createdAt.Format("2006-01-02 15:04:05"),
				}
				resumes = append(resumes, resume)
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   resumes,
				"count":  len(resumes),
			})
		})

		// 职位管理API
		api.GET("/jobs", func(c *gin.Context) {
			// 从数据库获取职位列表
			var jobs []map[string]interface{}
			rows, err := db.Query("SELECT id, title, company, location, salary_min, salary_max, status, created_at FROM jobs LIMIT 10")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id int
				var title, company, location, status string
				var salaryMin, salaryMax int
				var createdAt time.Time

				if err := rows.Scan(&id, &title, &company, &location, &salaryMin, &salaryMax, &status, &createdAt); err != nil {
					continue
				}

				job := map[string]interface{}{
					"id":         id,
					"title":      title,
					"company":    company,
					"location":   location,
					"salary_min": salaryMin,
					"salary_max": salaryMax,
					"status":     status,
					"created_at": createdAt.Format("2006-01-02 15:04:05"),
				}
				jobs = append(jobs, job)
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   jobs,
				"count":  len(jobs),
			})
		})

		// 积分管理API
		api.GET("/points", func(c *gin.Context) {
			// 从数据库获取积分列表
			var points []map[string]interface{}
			rows, err := db.Query("SELECT p.id, p.balance, p.created_at, u.username FROM points p LEFT JOIN users u ON p.user_id = u.id LIMIT 10")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var id, balance int
				var username string
				var createdAt time.Time

				if err := rows.Scan(&id, &balance, &createdAt, &username); err != nil {
					continue
				}

				point := map[string]interface{}{
					"id":         id,
					"balance":    balance,
					"username":   username,
					"created_at": createdAt.Format("2006-01-02 15:04:05"),
				}
				points = append(points, point)
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   points,
				"count":  len(points),
			})
		})

		// 用户登录API
		api.POST("/auth/login", func(c *gin.Context) {
			var loginData struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&loginData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// 简单的用户验证（实际项目中应该使用加密密码）
			var id int
			var username, email, passwordHash string
			err := db.QueryRow("SELECT id, username, email, password_hash FROM users WHERE username = ? LIMIT 1", loginData.Username).Scan(&id, &username, &email, &passwordHash)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}

			// 验证密码 - 支持明文密码和bcrypt哈希
			fmt.Printf("Debug: username=%s, password=%s, passwordHash=%s\n", loginData.Username, loginData.Password, passwordHash)

			// 检查是否是bcrypt哈希（以$2a$或$2b$开头）
			if len(passwordHash) > 4 && (passwordHash[:4] == "$2a$" || passwordHash[:4] == "$2b$") {
				// 使用bcrypt验证
				err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginData.Password))
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
					return
				}
			} else {
				// 明文密码比较
				if loginData.Password != passwordHash {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
					return
				}
			}

			// user := map[string]interface{}{
			// 	"id":       id,
			// 	"username": username,
			// 	"email":    email,
			// }

			// 生成标准的JWT token
			token, err := generateJWTToken(uint(id), username, "user")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "生成token失败",
				})
				return
			}

			// 通知Resume Service创建会话
			go func() {
				createSessionURL := "http://localhost:8082/api/v1/resume/session/create"
				sessionData := map[string]interface{}{
					"user_id":    id,
					"username":   username,
					"ip_address": c.ClientIP(),
					"user_agent": c.GetHeader("User-Agent"),
				}

				jsonData, _ := json.Marshal(sessionData)
				req, _ := http.NewRequest("POST", createSessionURL, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("创建会话失败: %v", err)
				} else {
					resp.Body.Close()
					log.Printf("用户%d会话创建成功", id)
				}
			}()

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"token":   token,
				"user": gin.H{
					"id":       id,
					"username": username,
					"email":    email,
					"role":     "super_admin",
					"status":   "active",
				},
			})
		})

		// 用户注册API
		api.POST("/auth/register", func(c *gin.Context) {
			var registerData struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				Password string `json:"password"`
				Phone    string `json:"phone"`
			}

			if err := c.ShouldBindJSON(&registerData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// 检查用户是否已存在
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", registerData.Username, registerData.Email).Scan(&count)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
			if count > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
				return
			}

			// 哈希密码
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerData.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
				return
			}

			// 创建新用户
			_, err = db.Exec("INSERT INTO users (username, email, password_hash, phone) VALUES (?, ?, ?, ?)",
				registerData.Username, registerData.Email, string(hashedPassword), registerData.Phone)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "User registered successfully",
			})
		})

		// 微信登录API
		api.POST("/auth/wechat-login", func(c *gin.Context) {
			var wechatData struct {
				Code     string `json:"code"`
				Nickname string `json:"nickname"`
				Avatar   string `json:"avatar"`
			}

			if err := c.ShouldBindJSON(&wechatData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// 模拟微信登录验证（实际项目中应该调用微信API验证code）
			// 这里我们创建一个测试用户或返回现有用户
			var userID int
			var username, email, phone string

			// 尝试查找现有用户（基于微信code或昵称）
			err := db.QueryRow("SELECT id, username, email, phone FROM users WHERE username = ? LIMIT 1", wechatData.Nickname).Scan(&userID, &username, &email, &phone)

			if err != nil {
				// 用户不存在，创建新用户
				username = wechatData.Nickname
				if username == "" {
					codePrefix := wechatData.Code
					if len(codePrefix) > 8 {
						codePrefix = codePrefix[:8]
					}
					username = "微信用户_" + codePrefix
				}
				email = username + "@wechat.local"
				phone = ""

				// 生成UUID，确保Code长度足够
				codePrefix := wechatData.Code
				if len(codePrefix) > 8 {
					codePrefix = codePrefix[:8]
				}
				uuid := fmt.Sprintf("wechat_%s_%d", codePrefix, time.Now().Unix())

				result, err := db.Exec("INSERT INTO users (uuid, username, email, password_hash, phone, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())",
					uuid, username, email, "wechat_user", phone, "active")
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
					return
				}

				userID64, _ := result.LastInsertId()
				userID = int(userID64)
			}

			user := map[string]interface{}{
				"id":       userID,
				"username": username,
				"email":    email,
				"phone":    phone,
				"avatar":   wechatData.Avatar,
			}

			// 生成JWT token
			token := fmt.Sprintf("wechat_token_%s_%d", username, time.Now().Unix())

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "微信登录成功",
				"data": gin.H{
					"token": token,
					"user":  user,
				},
			})
		})

		// 发送验证码API
		api.POST("/user/sendCode", func(c *gin.Context) {
			var request struct {
				Phone string `json:"phone"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// 模拟发送验证码
			code := fmt.Sprintf("%06d", time.Now().Unix()%1000000)

			// 这里应该调用短信服务发送验证码
			// 暂时返回成功响应
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "验证码发送成功",
				"data": gin.H{
					"code": code, // 开发环境返回验证码，生产环境不应该返回
				},
			})
		})

		// 微信注册API
		api.POST("/user/wechatRegister", func(c *gin.Context) {
			var request struct {
				Code     string `json:"code"`
				Nickname string `json:"nickname"`
				Avatar   string `json:"avatar"`
				Phone    string `json:"phone"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// 检查用户是否已存在
			var userID int
			var username, email string

			err := db.QueryRow("SELECT id, username, email FROM users WHERE username = ? OR phone = ? LIMIT 1",
				request.Nickname, request.Phone).Scan(&userID, &username, &email)

			if err == nil {
				// 用户已存在
				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "用户已存在，直接登录",
					"data": gin.H{
						"user": map[string]interface{}{
							"id":       userID,
							"username": username,
							"email":    email,
							"phone":    request.Phone,
							"avatar":   request.Avatar,
						},
						"token": fmt.Sprintf("wechat_token_%s_%d", username, time.Now().Unix()),
					},
				})
				return
			}

			// 创建新用户
			username = request.Nickname
			if username == "" {
				codePrefix := request.Code
				if len(codePrefix) > 8 {
					codePrefix = codePrefix[:8]
				}
				username = "微信用户_" + codePrefix
			}
			email = username + "@wechat.local"

			// 生成UUID，确保Code长度足够
			codePrefix := request.Code
			if len(codePrefix) > 8 {
				codePrefix = codePrefix[:8]
			}
			uuid := fmt.Sprintf("wechat_%s_%d", codePrefix, time.Now().Unix())

			result, err := db.Exec("INSERT INTO users (uuid, username, email, password_hash, phone, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())",
				uuid, username, email, "wechat_user", request.Phone, "active")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
				return
			}

			userID64, _ := result.LastInsertId()
			userID = int(userID64)

			user := map[string]interface{}{
				"id":       userID,
				"username": username,
				"email":    email,
				"phone":    request.Phone,
				"avatar":   request.Avatar,
			}

			token := fmt.Sprintf("wechat_token_%s_%d", username, time.Now().Unix())

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "微信注册成功",
				"data": gin.H{
					"token": token,
					"user":  user,
				},
			})
		})

		// 轮播图API
		api.GET("/banner/list", func(c *gin.Context) {
			// 返回轮播图数据
			banners := []map[string]interface{}{
				{
					"id":    1,
					"image": "/images/banner1.svg",
					"title": "JobFirst智能简历管理",
					"link":  "/pages/resume/resume",
				},
				{
					"id":    2,
					"image": "/images/banner2.svg",
					"title": "AI助手，简历优化更轻松",
					"link":  "/pages/chat/chat",
				},
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "获取成功",
				"data":    banners,
			})
		})

		// 推荐职位API
		api.GET("/job/recommend", func(c *gin.Context) {
			// 返回推荐职位数据
			jobs := []map[string]interface{}{
				{
					"id":       1,
					"title":    "前端开发工程师",
					"company":  "JobFirst科技",
					"salary":   "15K-25K",
					"location": "深圳",
					"tags":     []string{"React", "Vue", "JavaScript"},
				},
				{
					"id":       2,
					"title":    "后端开发工程师",
					"company":  "JobFirst科技",
					"salary":   "20K-35K",
					"location": "深圳",
					"tags":     []string{"Go", "Gin", "MySQL"},
				},
				{
					"id":       3,
					"title":    "产品经理",
					"company":  "JobFirst科技",
					"salary":   "25K-40K",
					"location": "深圳",
					"tags":     []string{"产品设计", "用户研究", "数据分析"},
				},
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "获取成功",
				"data":    jobs,
			})
		})

		// 市场统计数据API
		api.GET("/statistics/market", func(c *gin.Context) {
			// 返回市场统计数据
			marketData := map[string]interface{}{
				"jobCount":     "5,000+",
				"companyCount": "200+",
				"avgSalary":    "18K",
				"growthRate":   "15%",
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "获取成功",
				"data":    marketData,
			})
		})
	}

	// ==================== AI服务代理路由组 ====================
	// AI服务代理 - 转发请求到AI服务
	aiAPI := router.Group("/api/v1/ai")
	{
		// AI聊天功能
		aiAPI.POST("/chat", func(c *gin.Context) {
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/ai/chat", "POST")
		})

		// AI功能列表
		aiAPI.GET("/features", func(c *gin.Context) {
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/ai/features", "GET")
		})

		// 开始AI分析
		aiAPI.POST("/start-analysis", func(c *gin.Context) {
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/ai/start-analysis", "POST")
		})

		// 获取分析结果
		aiAPI.GET("/analysis-result/:taskId", func(c *gin.Context) {
			taskId := c.Param("taskId")
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/ai/analysis-result/"+taskId, "GET")
		})

		// 获取聊天历史
		aiAPI.GET("/chat-history", func(c *gin.Context) {
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/ai/chat-history", "GET")
		})
	}

	// 简历服务OPTIONS请求处理
	router.OPTIONS("/api/v1/resume/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, API-Version, X-Requested-With, X-API-Key, X-Client-Version")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusOK)
	})

	// 简历服务API代理 - 直接代理到Resume Service
	router.POST("/api/v1/resume/resumes/upload", func(c *gin.Context) {
		proxyToResumeService(c, "/api/v1/resume/resumes/upload", "POST")
	})

	// Job Service API代理 - 代理到Job Service
	jobAPI := router.Group("/api/v1/job")
	{
		// 公开API
		jobAPI.GET("/public/jobs", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/public/jobs", "GET")
		})
		jobAPI.GET("/public/jobs/:id", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/public/jobs/"+c.Param("id"), "GET")
		})
		jobAPI.GET("/public/companies/:company_id/jobs", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/public/companies/"+c.Param("company_id")+"/jobs", "GET")
		})
		jobAPI.GET("/public/industries", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/public/industries", "GET")
		})
		jobAPI.GET("/public/job-types", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/public/job-types", "GET")
		})

		// 需要认证的API
		jobAPI.POST("/jobs", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/jobs", "POST")
		})
		jobAPI.PUT("/jobs/:id", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/jobs/"+c.Param("id"), "PUT")
		})
		jobAPI.DELETE("/jobs/:id", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/jobs/"+c.Param("id"), "DELETE")
		})
		jobAPI.POST("/jobs/:id/apply", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/jobs/"+c.Param("id")+"/apply", "POST")
		})
		jobAPI.GET("/jobs/my-applications", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/jobs/my-applications", "GET")
		})

		// 管理员API
		jobAPI.GET("/admin/jobs", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/admin/jobs", "GET")
		})
		jobAPI.PUT("/admin/jobs/:id/status", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/admin/jobs/"+c.Param("id")+"/status", "PUT")
		})
		jobAPI.GET("/admin/jobs/:id/applications", func(c *gin.Context) {
			proxyToJobService(c, "/api/v1/job/admin/jobs/"+c.Param("id")+"/applications", "GET")
		})
	}

	// 简历分析API代理
	analyzeAPI := router.Group("/api/v1/analyze")
	{
		analyzeAPI.POST("/resume", func(c *gin.Context) {
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/analyze/resume", "POST")
		})
	}

	// 向量操作API代理
	vectorsAPI := router.Group("/api/v1/vectors")
	{
		vectorsAPI.GET("/:resume_id", func(c *gin.Context) {
			resumeId := c.Param("resume_id")
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/vectors/"+resumeId, "GET")
		})

		vectorsAPI.POST("/search", func(c *gin.Context) {
			// 转发到AI服务
			proxyToAIService(c, "/api/v1/vectors/search", "POST")
		})
	}

	// ==================== V3.0 API路由组 ====================
	// 初始化V3.0处理器
	resumeV3Handler := handlers.NewResumeV3Handler(gormDB)

	// 添加V3.0 API路由组
	v2API := router.Group("/api/v2")
	// v2API.Use(authMiddleware()) // 暂时注释掉认证中间件，便于测试
	{
		// 简历管理API（目前可用的方法）
		v2API.GET("/resumes", resumeV3Handler.GetResumes)
		v2API.POST("/resumes", resumeV3Handler.CreateResume)
		v2API.GET("/resumes/:id", resumeV3Handler.GetResume)

		// 标准化数据API
		v2API.GET("/standard/skills", resumeV3Handler.GetSkills)
		v2API.GET("/standard/companies", resumeV3Handler.GetCompanies)
		v2API.GET("/standard/positions", resumeV3Handler.GetPositions)
	}

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Basic Server started on port %s", cfg.Server.Port)
	log.Printf("Standalone mode - no external service dependencies")

	// 注册服务到Consul
	if consulManager != nil {
		port, _ := strconv.Atoi(cfg.Server.Port)
		err = consulManager.RegisterService(
			cfg.Consul.ServiceName,
			cfg.Consul.ServiceID,
			cfg.Server.Host,
			port,
			cfg.Consul.ServiceTags,
		)
		if err != nil {
			log.Printf("Warning: Failed to register service to Consul: %v", err)
		} else {
			log.Printf("Service registered to Consul: %s (%s)", cfg.Consul.ServiceName, cfg.Consul.ServiceID)

			// 启动健康检查循环
			go consulManager.StartHealthCheckLoop()
		}
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

// proxyToResumeService 代理请求到Resume服务
func proxyToResumeService(c *gin.Context, path string, method string) {
	// Resume服务URL
	resumeServiceURL := "http://localhost:7532"

	// 构建完整URL
	url := resumeServiceURL + path

	// 读取请求体
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// 创建新的HTTP请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Resume service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// proxyToAIService 代理请求到AI服务
func proxyToAIService(c *gin.Context, path string, method string) {
	// AI服务URL
	aiServiceURL := "http://localhost:8206"

	// 构建完整URL
	url := aiServiceURL + path

	// 读取请求体
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// 创建新的HTTP请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// proxyToJobService 代理请求到Job服务
func proxyToJobService(c *gin.Context, path string, method string) {
	// Job服务URL
	jobServiceURL := "http://localhost:8089"

	// 构建完整URL
	url := jobServiceURL + path

	// 读取请求体
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// 创建新的HTTP请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Job service unavailable"})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// generateJWTToken 生成JWT token
func generateJWTToken(userID uint, username, role string) (string, error) {
	// 创建JWT claims
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // 24小时过期
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名
	jwtSecret := "jobfirst-unified-auth-secret-key-2024"
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
