package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/consul/api"
	"github.com/jobfirst/jobfirst-core"
)

func main() {
	// 从环境变量获取端口，默认为8083
	port := os.Getenv("COMPANY_SERVICE_PORT")
	if port == "" {
		port = "8083"
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

	// 设置文档API路由
	documentAPI := NewDocumentAPI(core)
	documentAPI.SetupDocumentRoutes(r)

	// 设置企业画像API路由
	profileAPI := NewCompanyProfileAPI(core)
	profileAPI.SetupCompanyProfileRoutes(r)

	// 设置AI配额API路由
	quotaAPI := NewQuotaAPI(core.GetDB())
	quotaAPI.RegisterRoutes(r.Group("/api/v1"))

	// 设置管理员配额API路由
	adminAPI := NewAdminAPI(core.GetDB())
	adminAPI.RegisterAdminRoutes(r.Group("/api/v1"))

	// 初始化企业权限管理器
	var redisClient *redis.Client
	if redisManager := core.Database.GetRedis(); redisManager != nil {
		redisClient = redisManager.GetClient()
	}
	permissionManager := NewCompanyPermissionManager(core.GetDB(), redisClient)

	// 初始化企业数据同步服务
	dataSyncService := NewCompanyDataSyncService(core.GetDB(), nil, nil, redisClient)

	// 设置企业认证增强API路由
	authAPI := NewCompanyAuthAPI(core, permissionManager, dataSyncService)
	authAPI.SetupCompanyAuthRoutes(r)

	// 设置企业增强API路由
	setupCompanyEnhancedRoutes(r, core, dataSyncService)

	// 设置DAO功能路由
	setupDAORoutes(r, core)

	// 设置AI优化功能路由
	setupAIOptimizationRoutes(r, core)

	// 设置企业信用信息API路由
	creditInfoAPI := NewCreditInfoAPI(core)
	creditInfoAPI.SetupCreditInfoRoutes(r)

	// 注册到Consul
	registerToConsul("company-service", "127.0.0.1", portInt)

	// 启动服务器
	log.Printf("Starting Company Service with jobfirst-core on 0.0.0.0:%s", port)
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
			"service":     "company-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "company-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "company-service",
			"version":    "3.1.0",
			"port":       8083,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由 (保持现有API)
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1/company/public")
	{
		// 获取企业列表
		public.GET("/companies", func(c *gin.Context) {
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
			industry := c.Query("industry")
			location := c.Query("location")

			// 使用核心包的数据库管理器
			db := core.GetDB()
			var companies []Company
			offset := (page - 1) * pageSize

			query := db.Model(&Company{}).Where("status = 'active'")
			if industry != "" {
				query = query.Where("industry = ?", industry)
			}
			if location != "" {
				query = query.Where("location LIKE ?", "%"+location+"%")
			}

			if err := query.Offset(offset).Limit(pageSize).Find(&companies).Error; err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get company list", err.Error())
				return
			}

			var total int64
			query.Count(&total)

			standardSuccessResponse(c, gin.H{
				"companies": companies,
				"total":     total,
				"page":      page,
				"size":      pageSize,
			}, "Company list retrieved successfully")
		})

		// 获取单个企业信息
		public.GET("/companies/:id", func(c *gin.Context) {
			companyID, _ := strconv.Atoi(c.Param("id"))

			// 使用核心包的数据库管理器
			db := core.GetDB()
			var company Company
			if err := db.First(&company, companyID).Error; err != nil {
				standardErrorResponse(c, http.StatusNotFound, "Company not found", err.Error())
				return
			}

			// 增加浏览次数
			db.Model(&company).Update("view_count", company.ViewCount+1)

			standardSuccessResponse(c, company, "Company information retrieved successfully")
		})

		// 获取行业列表
		public.GET("/industries", func(c *gin.Context) {
			industries := []string{
				"互联网/电子商务",
				"计算机软件",
				"金融/投资/证券",
				"教育培训",
				"医疗/健康",
				"房地产/建筑",
				"制造业",
				"零售/批发",
				"广告/媒体",
				"其他",
			}
			standardSuccessResponse(c, industries, "Industries retrieved successfully")
		})

		// 获取公司规模列表
		public.GET("/company-sizes", func(c *gin.Context) {
			sizes := []string{
				"1-20人",
				"21-50人",
				"51-100人",
				"101-500人",
				"501-1000人",
				"1000人以上",
			}
			standardSuccessResponse(c, sizes, "Company sizes retrieved successfully")
		})
	}

	// 需要认证的API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	api := r.Group("/api/v1/company")
	api.Use(authMiddleware)
	{
		// 企业管理API
		companies := api.Group("/companies")
		{
			// 创建企业
			companies.POST("/", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				userID := userIDInterface.(uint)

				var company Company
				if err := c.ShouldBindJSON(&company); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				company.CreatedBy = userID
				company.CreatedAt = time.Now()
				company.UpdatedAt = time.Now()
				company.Status = "pending"
				company.VerificationLevel = "unverified"

				if err := db.Create(&company).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create company", err.Error())
					return
				}

				standardSuccessResponse(c, company, "Company created successfully")
			})

			// 更新企业信息
			companies.PUT("/:id", func(c *gin.Context) {
				companyID, _ := strconv.Atoi(c.Param("id"))

				var updateData Company
				if err := c.ShouldBindJSON(&updateData); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var company Company
				if err := db.First(&company, companyID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Company not found", err.Error())
					return
				}

				// 检查权限：只有企业创建者或管理员可以更新
				userIDInterface, _ := c.Get("user_id")
				userID := userIDInterface.(uint)
				role := c.GetString("role")

				if company.CreatedBy != userID && role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				updateData.ID = uint(companyID)
				updateData.UpdatedAt = time.Now()

				if err := db.Save(&updateData).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to update company", err.Error())
					return
				}

				standardSuccessResponse(c, updateData, "Company updated successfully")
			})

			// 删除企业
			companies.DELETE("/:id", func(c *gin.Context) {
				companyID, _ := strconv.Atoi(c.Param("id"))

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var company Company
				if err := db.First(&company, companyID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Company not found", err.Error())
					return
				}

				// 检查权限：只有企业创建者或管理员可以删除
				userIDInterface, _ := c.Get("user_id")
				userID := userIDInterface.(uint)
				role := c.GetString("role")

				if company.CreatedBy != userID && role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				if err := db.Delete(&company).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to delete company", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{"deleted": true}, "Company deleted successfully")
			})

			// 获取用户创建的企业列表
			companies.GET("/my-companies", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				userID := userIDInterface.(uint)

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var companies []Company
				if err := db.Where("created_by = ?", userID).Find(&companies).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get user companies", err.Error())
					return
				}

				standardSuccessResponse(c, companies, "User companies retrieved successfully")
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
			"company",
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

// standardSuccessResponse 标准成功响应
func standardSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	response := gin.H{
		"success": true,
		"data":    data,
		"service": "company-service",
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
		"service": "company-service",
		"time":    time.Now().Format(time.RFC3339),
	}
	if len(details) > 0 {
		response["details"] = details[0]
	}
	c.JSON(statusCode, response)
}

// Company 公司信息结构体
type Company struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name" gorm:"size:200;not null"`
	ShortName         string    `json:"short_name" gorm:"size:100"`
	LogoURL           string    `json:"logo_url" gorm:"size:500"`
	Industry          string    `json:"industry" gorm:"size:100"`
	CompanySize       string    `json:"company_size" gorm:"size:50"`
	Location          string    `json:"location" gorm:"size:200"`
	Website           string    `json:"website" gorm:"size:200"`
	Description       string    `json:"description" gorm:"type:text"`
	FoundedYear       int       `json:"founded_year"`
	Status            string    `json:"status" gorm:"size:20;default:pending"`                // pending, active, inactive, rejected
	VerificationLevel string    `json:"verification_level" gorm:"size:20;default:unverified"` // verified, unverified
	JobCount          int       `json:"job_count" gorm:"default:0"`
	ViewCount         int       `json:"view_count" gorm:"default:0"`
	CreatedBy         uint      `json:"created_by" gorm:"not null"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
