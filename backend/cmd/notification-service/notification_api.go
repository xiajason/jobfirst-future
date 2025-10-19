package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// setupNotificationBusinessRoutes 设置完整的通知业务API路由
func setupNotificationBusinessRoutes(r *gin.Engine, nb *NotificationBusiness) {
	// 通知管理API组
	notificationAPI := r.Group("/api/v1/notification")
	{
		// 获取用户通知列表
		notificationAPI.GET("/user/:user_id/notifications", func(c *gin.Context) {
			userIDStr := c.Param("user_id")
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			// 获取查询参数
			limitStr := c.DefaultQuery("limit", "10")
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				limit = 10
			}

			notifications, err := nb.GetUserNotifications(uint(userID), limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取通知列表失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   notifications,
			})
		})

		// 标记通知为已读
		notificationAPI.PUT("/user/:user_id/notifications/:id/read", func(c *gin.Context) {
			notificationIDStr := c.Param("id")
			notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
				return
			}

			userIDStr := c.Query("user_id")
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			err = nb.MarkAsRead(uint(notificationID), uint(userID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "标记通知为已读失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "通知已标记为已读",
			})
		})

		// 获取用户通知统计
		notificationAPI.GET("/user/:user_id/stats", func(c *gin.Context) {
			userIDStr := c.Param("user_id")
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			stats, err := nb.GetNotificationStats(uint(userID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取通知统计失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   stats,
			})
		})
	}

	// 订阅管理通知API组
	subscriptionAPI := r.Group("/api/v1/notification/subscription")
	{
		// 发送订阅即将到期通知
		subscriptionAPI.POST("/expiring", func(c *gin.Context) {
			var req struct {
				UserID   uint   `json:"user_id" binding:"required"`
				DaysLeft int    `json:"days_left"`
				PlanName string `json:"plan_name"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			title := "订阅即将到期提醒"
			content := "您的" + req.PlanName + "订阅将在" + strconv.Itoa(req.DaysLeft) + "天后到期，请及时续费以继续享受服务。"

			err := nb.SendSubscriptionNotification(req.UserID, "subscription_expiring", title, content)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "订阅到期提醒通知已发送",
			})
		})

		// 发送订阅升级成功通知
		subscriptionAPI.POST("/upgraded", func(c *gin.Context) {
			var req struct {
				UserID  uint   `json:"user_id" binding:"required"`
				OldPlan string `json:"old_plan"`
				NewPlan string `json:"new_plan"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			title := "订阅升级成功"
			content := "恭喜！您已成功从" + req.OldPlan + "升级到" + req.NewPlan + "，现在可以享受更多功能和服务。"

			err := nb.SendSubscriptionNotification(req.UserID, "subscription_upgraded", title, content)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "订阅升级成功通知已发送",
			})
		})

		// 发送订阅已过期通知
		subscriptionAPI.POST("/expired", func(c *gin.Context) {
			var req struct {
				UserID   uint   `json:"user_id" binding:"required"`
				PlanName string `json:"plan_name"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			title := "订阅已过期"
			content := "您的" + req.PlanName + "订阅已过期，部分功能将受到限制。请及时续费以恢复完整功能。"

			err := nb.SendSubscriptionNotification(req.UserID, "subscription_expired", title, content)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "订阅过期通知已发送",
			})
		})

		// 发送订阅续费成功通知
		subscriptionAPI.POST("/renewed", func(c *gin.Context) {
			var req struct {
				UserID   uint   `json:"user_id" binding:"required"`
				PlanName string `json:"plan_name"`
				Duration string `json:"duration"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			title := "订阅续费成功"
			content := "您的" + req.PlanName + "订阅已成功续费" + req.Duration + "，感谢您的支持！"

			err := nb.SendSubscriptionNotification(req.UserID, "subscription_renewed", title, content)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "订阅续费成功通知已发送",
			})
		})
	}

	// AI服务通知API组
	aiServiceAPI := r.Group("/api/v1/notification/ai-service")
	{
		// 发送AI服务使用限制警告通知
		aiServiceAPI.POST("/limit-warning", func(c *gin.Context) {
			var req struct {
				UserID      uint    `json:"user_id" binding:"required"`
				ServiceType string  `json:"service_type"`
				Usage       float64 `json:"usage"`
				Limit       float64 `json:"limit"`
				Percentage  float64 `json:"percentage"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			title := "AI服务使用量警告"
			content := "您的" + req.ServiceType + "服务使用量已达到" + strconv.FormatFloat(req.Percentage, 'f', 1, 64) + "%，当前使用量：" + strconv.FormatFloat(req.Usage, 'f', 0, 64) + "/" + strconv.FormatFloat(req.Limit, 'f', 0, 64)

			usageData := map[string]interface{}{
				"service_type": req.ServiceType,
				"usage":        req.Usage,
				"limit":        req.Limit,
				"percentage":   req.Percentage,
			}

			err := nb.SendAIServiceNotification(req.UserID, "ai_service_limit_warning", title, content, usageData)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "AI服务使用限制警告通知已发送",
			})
		})

		// 发送AI服务使用限制超出通知
		aiServiceAPI.POST("/limit-exceeded", func(c *gin.Context) {
			var req struct {
				UserID      uint    `json:"user_id" binding:"required"`
				ServiceType string  `json:"service_type"`
				Usage       float64 `json:"usage"`
				Limit       float64 `json:"limit"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			title := "AI服务使用限制超出"
			content := "您的" + req.ServiceType + "服务使用量已超出限制，当前使用量：" + strconv.FormatFloat(req.Usage, 'f', 0, 64) + "，限制：" + strconv.FormatFloat(req.Limit, 'f', 0, 64) + "。请升级订阅或等待配额重置。"

			usageData := map[string]interface{}{
				"service_type": req.ServiceType,
				"usage":        req.Usage,
				"limit":        req.Limit,
			}

			err := nb.SendAIServiceNotification(req.UserID, "ai_service_limit_exceeded", title, content, usageData)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "AI服务使用限制超出通知已发送",
			})
		})

		// 发送AI服务配额重置通知
		aiServiceAPI.POST("/quota-reset", func(c *gin.Context) {
			var req struct {
				UserID      uint   `json:"user_id" binding:"required"`
				ServiceType string `json:"service_type"`
				ResetType   string `json:"reset_type"` // daily, monthly
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			title := "AI服务配额已重置"
			content := "您的" + req.ServiceType + "服务" + req.ResetType + "配额已重置，现在可以继续使用服务。"

			usageData := map[string]interface{}{
				"service_type": req.ServiceType,
				"reset_type":   req.ResetType,
			}

			err := nb.SendAIServiceNotification(req.UserID, "ai_service_quota_reset", title, content, usageData)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "AI服务配额重置通知已发送",
			})
		})
	}
}
