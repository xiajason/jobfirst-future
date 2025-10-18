package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// setupServiceIntegrationRoutes 设置服务间集成API路由
func setupServiceIntegrationRoutes(r *gin.Engine, si *ServiceIntegration) {
	// 服务间集成API组
	integrationAPI := r.Group("/api/v1/integration")
	{
		// 检查用户配额并发送通知
		integrationAPI.POST("/check-quota/:user_id", func(c *gin.Context) {
			userIDStr := c.Param("user_id")
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			err = si.CheckUserQuotaAndSendNotification(uint(userID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "检查用户配额失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "用户配额检查完成，相关通知已发送",
			})
		})

		// 处理订阅状态变更
		integrationAPI.POST("/subscription-status-change", func(c *gin.Context) {
			var req struct {
				UserID    uint   `json:"user_id" binding:"required"`
				OldStatus string `json:"old_status" binding:"required"`
				NewStatus string `json:"new_status" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := si.HandleSubscriptionStatusChange(req.UserID, req.OldStatus, req.NewStatus)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "处理订阅状态变更失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "订阅状态变更处理完成，相关通知已发送",
			})
		})

		// 发送欢迎通知
		integrationAPI.POST("/welcome/:user_id", func(c *gin.Context) {
			userIDStr := c.Param("user_id")
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			err = si.SendWelcomeNotification(uint(userID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "发送欢迎通知失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "欢迎通知已发送",
			})
		})

		// 批量检查用户配额
		integrationAPI.POST("/batch-check-quota", func(c *gin.Context) {
			var req struct {
				UserIDs []uint `json:"user_ids" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			var results []map[string]interface{}
			var errors []string

			for _, userID := range req.UserIDs {
				err := si.CheckUserQuotaAndSendNotification(userID)
				if err != nil {
					errors = append(errors, fmt.Sprintf("用户%d: %v", userID, err))
					results = append(results, map[string]interface{}{
						"user_id": userID,
						"status":  "failed",
						"error":   err.Error(),
					})
				} else {
					results = append(results, map[string]interface{}{
						"user_id": userID,
						"status":  "success",
					})
				}
			}

			response := gin.H{
				"status":  "completed",
				"results": results,
			}

			if len(errors) > 0 {
				response["errors"] = errors
			}

			c.JSON(http.StatusOK, response)
		})
	}

	// 系统事件API组 - 供其他服务调用
	eventAPI := r.Group("/api/v1/events")
	{
		// 用户注册事件
		eventAPI.POST("/user-registered", func(c *gin.Context) {
			var req struct {
				UserID uint `json:"user_id" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := si.SendWelcomeNotification(req.UserID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "处理用户注册事件失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "用户注册事件处理完成",
			})
		})

		// AI服务调用事件
		eventAPI.POST("/ai-service-called", func(c *gin.Context) {
			var req struct {
				UserID      uint    `json:"user_id" binding:"required"`
				ServiceType string  `json:"service_type" binding:"required"`
				Cost        float64 `json:"cost"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 检查用户配额并发送相应通知
			err := si.CheckUserQuotaAndSendNotification(req.UserID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "处理AI服务调用事件失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "AI服务调用事件处理完成",
			})
		})

		// 订阅变更事件
		eventAPI.POST("/subscription-changed", func(c *gin.Context) {
			var req struct {
				UserID    uint   `json:"user_id" binding:"required"`
				OldStatus string `json:"old_status" binding:"required"`
				NewStatus string `json:"new_status" binding:"required"`
				PlanName  string `json:"plan_name"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := si.HandleSubscriptionStatusChange(req.UserID, req.OldStatus, req.NewStatus)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "处理订阅变更事件失败",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "订阅变更事件处理完成",
			})
		})
	}
}
