package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/jobfirst/jobfirst-core"
)

func main() {
	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("../../configs/jobfirst-core-config.yaml")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 初始化模板增强服务
	enhancedService, err := NewTemplateEnhancedService(core)
	if err != nil {
		log.Printf("初始化模板增强服务失败: %v", err)
		log.Println("继续以基础模式运行...")
		enhancedService = nil
	}
	if enhancedService != nil {
		defer enhancedService.Close()
		log.Println("模板增强服务初始化成功")
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 设置标准路由 (使用jobfirst-core统一模板)
	setupStandardRoutes(r, core)

	// 设置业务路由 (保持现有API)
	setupBusinessRoutes(r, core)

	// 设置增强路由
	if enhancedService != nil {
		setupEnhancedRoutes(r, core, enhancedService)
		log.Println("模板增强API路由已设置")
	}

	// 注册到Consul
	registerToConsul("template-service", "127.0.0.1", 7532)

	// 启动服务器
	log.Println("Starting Template Service with jobfirst-core on 0.0.0.0:7532")
	if enhancedService != nil {
		log.Println("多数据库架构已启用")
	}
	if err := r.Run(":7532"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由 (使用jobfirst-core统一模板)
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查 (统一格式)
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "template-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "template-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "template-service",
			"version":    "3.1.0",
			"port":       7532,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由 (保持现有API)
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 公开API路由（不需要认证）
	public := r.Group("/api/v1/template/public")
	{
		// 获取模板列表（支持搜索、排序）
		public.GET("/templates", func(c *gin.Context) {
			category := c.Query("category")
			search := c.Query("search")
			sortBy := c.DefaultQuery("sort", "created_at")
			order := c.DefaultQuery("order", "desc")
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

			// 使用核心包的数据库管理器
			db := core.GetDB()
			var templates []Template
			offset := (page - 1) * pageSize

			query := db.Where("is_active = true")

			// 分类筛选
			if category != "" {
				query = query.Where("category = ?", category)
			}

			// 搜索功能
			if search != "" {
				searchPattern := "%" + search + "%"
				query = query.Where("name LIKE ? OR description LIKE ? OR content LIKE ?",
					searchPattern, searchPattern, searchPattern)
			}

			// 排序
			validSortFields := map[string]bool{
				"created_at": true,
				"updated_at": true,
				"usage":      true,
				"rating":     true,
				"name":       true,
			}
			if validSortFields[sortBy] {
				if order == "desc" {
					query = query.Order(fmt.Sprintf("%s DESC", sortBy))
				} else {
					query = query.Order(fmt.Sprintf("%s ASC", sortBy))
				}
			} else {
				query = query.Order("created_at DESC")
			}

			if err := query.Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get template list", err.Error())
				return
			}

			var total int64
			query.Count(&total)

			standardSuccessResponse(c, gin.H{
				"templates": templates,
				"total":     total,
				"page":      page,
				"size":      pageSize,
			}, "Template list retrieved successfully")
		})

		// 获取单个模板
		public.GET("/templates/:id", func(c *gin.Context) {
			templateID, _ := strconv.Atoi(c.Param("id"))

			// 使用核心包的数据库管理器
			db := core.GetDB()
			var template Template
			if err := db.First(&template, templateID).Error; err != nil {
				standardErrorResponse(c, http.StatusNotFound, "Template not found", err.Error())
				return
			}

			// 增加使用次数
			db.Model(&template).Update("usage", template.Usage+1)

			standardSuccessResponse(c, template, "Template retrieved successfully")
		})

		// 获取热门模板
		public.GET("/templates/popular", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if limit > 50 {
				limit = 50
			}

			db := core.GetDB()
			var templates []Template
			if err := db.Where("is_active = true").
				Order("usage DESC, rating DESC").
				Limit(limit).Find(&templates).Error; err != nil {
				standardErrorResponse(c, http.StatusInternalServerError, "Failed to get popular templates", err.Error())
				return
			}

			standardSuccessResponse(c, templates, "Popular templates retrieved successfully")
		})

		// 获取模板分类
		public.GET("/categories", func(c *gin.Context) {
			categories := []string{
				"简历模板",
				"求职信模板",
				"项目介绍模板",
				"技能展示模板",
				"其他",
			}
			standardSuccessResponse(c, categories, "Categories retrieved successfully")
		})
	}

	// 需要认证的API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	api := r.Group("/api/v1/template")
	api.Use(authMiddleware)
	{
		// 模板管理API
		templates := api.Group("/templates")
		{
			// 创建模板
			templates.POST("/", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				userID := userIDInterface.(uint)

				var template Template
				if err := c.ShouldBindJSON(&template); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				template.CreatedBy = userID
				template.CreatedAt = time.Now()
				template.UpdatedAt = time.Now()
				template.IsActive = true
				template.Usage = 0
				template.Rating = 0.0

				if err := db.Create(&template).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to create template", err.Error())
					return
				}

				standardSuccessResponse(c, template, "Template created successfully")
			})

			// 更新模板
			templates.PUT("/:id", func(c *gin.Context) {
				templateID, _ := strconv.Atoi(c.Param("id"))

				var updateData Template
				if err := c.ShouldBindJSON(&updateData); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var template Template
				if err := db.First(&template, templateID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Template not found", err.Error())
					return
				}

				// 检查权限：只有模板创建者或管理员可以更新
				userIDInterface, _ := c.Get("user_id")
				userID := userIDInterface.(uint)
				role := c.GetString("role")

				if template.CreatedBy != userID && role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				updateData.ID = uint(templateID)
				updateData.UpdatedAt = time.Now()
				// 不允许更新使用次数和评分
				updateData.Usage = template.Usage
				updateData.Rating = template.Rating

				if err := db.Save(&updateData).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to update template", err.Error())
					return
				}

				standardSuccessResponse(c, updateData, "Template updated successfully")
			})

			// 删除模板
			templates.DELETE("/:id", func(c *gin.Context) {
				templateID, _ := strconv.Atoi(c.Param("id"))

				// 使用核心包的数据库管理器
				db := core.GetDB()
				var template Template
				if err := db.First(&template, templateID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Template not found", err.Error())
					return
				}

				// 检查权限：只有模板创建者或管理员可以删除
				userIDInterface, _ := c.Get("user_id")
				userID := userIDInterface.(uint)
				role := c.GetString("role")

				if template.CreatedBy != userID && role != "admin" && role != "super_admin" {
					standardErrorResponse(c, http.StatusForbidden, "Insufficient permissions", "")
					return
				}

				if err := db.Delete(&template).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to delete template", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{"deleted": true}, "Template deleted successfully")
			})

			// 评分模板
			templates.POST("/:id/rate", func(c *gin.Context) {
				templateID, _ := strconv.Atoi(c.Param("id"))

				var ratingRequest struct {
					Rating float64 `json:"rating" binding:"required,min=0,max=5"`
				}

				if err := c.ShouldBindJSON(&ratingRequest); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Rating must be between 0-5", err.Error())
					return
				}

				db := core.GetDB()
				var template Template
				if err := db.First(&template, templateID).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Template not found", err.Error())
					return
				}

				// 更新评分（简单平均算法）
				userIDInterface, _ := c.Get("user_id")
				userID := userIDInterface.(uint)

				// 检查用户是否已经评分过
				var existingRating Rating
				err := db.Where("template_id = ? AND user_id = ?", templateID, userID).First(&existingRating).Error

				if err == nil {
					// 用户已经评分过，更新评分
					existingRating.Rating = ratingRequest.Rating
					existingRating.UpdatedAt = time.Now()
					db.Save(&existingRating)
				} else {
					// 新评分
					newRating := Rating{
						TemplateID: uint(templateID),
						UserID:     userID,
						Rating:     ratingRequest.Rating,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}
					db.Create(&newRating)
				}

				// 重新计算平均评分
				var avgRating float64
				db.Model(&Rating{}).Where("template_id = ?", templateID).Select("AVG(rating)").Scan(&avgRating)

				// 更新模板评分
				db.Model(&template).Update("rating", avgRating)

				standardSuccessResponse(c, gin.H{
					"rating": avgRating,
				}, "Template rated successfully")
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
			"template",
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
		"service": "template-service",
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
		"service": "template-service",
		"time":    time.Now().Format(time.RFC3339),
	}
	if len(details) > 0 {
		response["details"] = details[0]
	}
	c.JSON(statusCode, response)
}

// 数据模型定义
type Template struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:200;not null"`
	Category    string    `json:"category" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Content     string    `json:"content" gorm:"type:text"`
	Variables   string    `json:"variables" gorm:"type:json"`
	Preview     string    `json:"preview" gorm:"type:text"`                  // 新增：预览内容
	Usage       int       `json:"usage" gorm:"column:usage_count;default:0"` // 新增：使用次数
	Rating      float64   `json:"rating" gorm:"default:0"`                   // 新增：评分
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedBy   uint      `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 评分模型
type Rating struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	TemplateID uint      `json:"template_id" gorm:"not null"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	Rating     float64   `json:"rating" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
