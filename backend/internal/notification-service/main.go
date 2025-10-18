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

	// 初始化通知业务逻辑
	notificationBusiness := NewNotificationBusiness(core)

	// 自动迁移数据库表
	if err := notificationBusiness.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化服务间集成
	serviceIntegration := NewServiceIntegration(notificationBusiness)

	// 启动配额监控
	serviceIntegration.StartQuotaMonitoring()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 设置标准路由 (使用jobfirst-core统一模板)
	setupStandardRoutes(r, core)

	// 设置业务路由 (保持现有API)
	setupBusinessRoutes(r, core)

	// 设置完整的通知业务API路由
	setupNotificationBusinessRoutes(r, notificationBusiness)

	// 设置服务间集成API路由
	setupServiceIntegrationRoutes(r, serviceIntegration)

	// 设置成本控制通知API路由
	setupCostControlNotificationRoutes(r, core)

	// 注册到Consul
	registerToConsul("notification-service", "127.0.0.1", 7534)

	// 启动服务器
	log.Println("Starting Notification Service with jobfirst-core on 0.0.0.0:7534")
	if err := r.Run(":7534"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// setupStandardRoutes 设置标准路由 (使用jobfirst-core统一模板)
func setupStandardRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 健康检查 (统一格式)
	r.GET("/health", func(c *gin.Context) {
		health := core.Health()
		c.JSON(http.StatusOK, gin.H{
			"service":     "notification-service",
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "3.1.0",
			"description": "Notification Management Service",
			"core_health": health,
		})
	})

	// 版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "notification-service",
			"version": "3.1.0",
			"build":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 服务信息
	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "notification-service",
			"version":    "3.1.0",
			"port":       7534,
			"status":     "running",
			"started_at": time.Now().Format(time.RFC3339),
		})
	})
}

// setupBusinessRoutes 设置业务路由 (保持现有API)
func setupBusinessRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 需要认证的API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	api := r.Group("/api/v1/notification")
	api.Use(authMiddleware)
	{
		// 通知管理API
		notifications := api.Group("/notifications")
		{
			// 获取用户通知列表
			notifications.GET("/", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				userID := userIDInterface.(uint)

				page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
				pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
				readStatus := c.Query("read_status")

				if page <= 0 {
					page = 1
				}
				if pageSize <= 0 {
					pageSize = 10
				}

				db := core.GetDB()
				var notifications []Notification
				offset := (page - 1) * pageSize

				query := db.Where("user_id = ?", userID)
				if readStatus != "" {
					query = query.Where("status = ?", readStatus)
				}

				if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to get notifications", err.Error())
					return
				}

				var total int64
				query.Count(&total)

				standardSuccessResponse(c, gin.H{
					"notifications": notifications,
					"total":         total,
					"page":          page,
					"size":          pageSize,
				}, "Notifications retrieved successfully")
			})

			// 标记通知为已读
			notifications.PUT("/:id/read", func(c *gin.Context) {
				notificationID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
					return
				}

				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				userID := userIDInterface.(uint)

				db := core.GetDB()
				var notification Notification
				if err := db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Notification not found", err.Error())
					return
				}

				now := time.Now()
				notification.IsRead = true
				notification.Status = "read"
				notification.ReadAt = &now
				notification.UpdatedAt = now

				if err := db.Save(&notification).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to mark notification as read", err.Error())
					return
				}

				standardSuccessResponse(c, notification, "Notification marked as read successfully")
			})

			// 删除通知
			notifications.DELETE("/:id", func(c *gin.Context) {
				notificationID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", err.Error())
					return
				}

				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				userID := userIDInterface.(uint)

				db := core.GetDB()
				var notification Notification
				if err := db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
					standardErrorResponse(c, http.StatusNotFound, "Notification not found", err.Error())
					return
				}

				if err := db.Delete(&notification).Error; err != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification", err.Error())
					return
				}

				standardSuccessResponse(c, gin.H{"deleted": true}, "Notification deleted successfully")
			})

			// 批量标记为已读
			notifications.PUT("/batch/read", func(c *gin.Context) {
				var req struct {
					IDs []uint `json:"ids" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				userID := userIDInterface.(uint)

				db := core.GetDB()
				now := time.Now()
				result := db.Model(&Notification{}).Where("id IN ? AND user_id = ?", req.IDs, userID).Updates(map[string]interface{}{
					"is_read":    true,
					"status":     "read",
					"read_at":    &now,
					"updated_at": now,
				})

				if result.Error != nil {
					standardErrorResponse(c, http.StatusInternalServerError, "Failed to mark notifications as read", result.Error.Error())
					return
				}

				standardSuccessResponse(c, gin.H{
					"updated_count": result.RowsAffected,
				}, "Notifications marked as read successfully")
			})
		}

		// 通知设置API
		settings := api.Group("/settings")
		{
			// 获取用户通知设置
			settings.GET("/", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				_ = userIDInterface.(uint)

				// 这里应该从数据库获取用户通知设置
				// 暂时返回默认设置
				settings := map[string]interface{}{
					"email_notifications":    true,
					"push_notifications":     true,
					"sms_notifications":      false,
					"notification_frequency": "immediate",
					"quiet_hours": map[string]interface{}{
						"enabled": false,
						"start":   "22:00",
						"end":     "08:00",
					},
				}

				standardSuccessResponse(c, settings, "Notification settings retrieved successfully")
			})

			// 更新用户通知设置
			settings.PUT("/", func(c *gin.Context) {
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
					return
				}
				_ = userIDInterface.(uint)

				var settings map[string]interface{}
				if err := c.ShouldBindJSON(&settings); err != nil {
					standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
					return
				}

				// 这里应该保存用户通知设置到数据库
				// 暂时返回成功响应
				standardSuccessResponse(c, settings, "Notification settings updated successfully")
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
			"notification",
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
		"service": "notification-service",
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
		"service": "notification-service",
		"time":    time.Now().Format(time.RFC3339),
	}
	if len(details) > 0 {
		response["details"] = details[0]
	}
	c.JSON(statusCode, response)
}
