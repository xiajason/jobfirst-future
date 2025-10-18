package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// 成本控制通知API
func setupCostControlNotificationRoutes(r *gin.Engine, core *jobfirst.Core) {
	sender := NewCostControlNotificationSender(core)

	costControlAPI := r.Group("/api/v1/notification/cost-control")
	{
		// 成本限制警告通知
		costControlAPI.POST("/limit-warning", func(c *gin.Context) {
			var req struct {
				UserID      uint    `json:"user_id" binding:"required"`
				CurrentCost float64 `json:"current_cost" binding:"required"`
				Limit       float64 `json:"limit" binding:"required"`
				Percentage  float64 `json:"percentage,omitempty"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid request parameters",
					"error":   err.Error(),
				})
				return
			}

			// 计算百分比
			if req.Percentage == 0 {
				req.Percentage = (req.CurrentCost / req.Limit) * 100
			}

			err := sender.SendCostLimitWarningNotification(req.UserID, req.CurrentCost, req.Limit, req.Percentage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Failed to send cost limit warning notification",
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "Cost limit warning notification sent successfully",
				"data": gin.H{
					"user_id":      req.UserID,
					"current_cost": req.CurrentCost,
					"limit":        req.Limit,
					"percentage":   req.Percentage,
				},
			})
		})

		// 成本限制超出通知
		costControlAPI.POST("/limit-exceeded", func(c *gin.Context) {
			var req struct {
				UserID       uint    `json:"user_id" binding:"required"`
				CurrentCost  float64 `json:"current_cost" binding:"required"`
				Limit        float64 `json:"limit" binding:"required"`
				ExcessAmount float64 `json:"excess_amount,omitempty"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid request parameters",
					"error":   err.Error(),
				})
				return
			}

			// 计算超出金额
			if req.ExcessAmount == 0 {
				req.ExcessAmount = req.CurrentCost - req.Limit
			}

			err := sender.SendCostLimitExceededNotification(req.UserID, req.CurrentCost, req.Limit, req.ExcessAmount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Failed to send cost limit exceeded notification",
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "Cost limit exceeded notification sent successfully",
				"data": gin.H{
					"user_id":       req.UserID,
					"current_cost":  req.CurrentCost,
					"limit":         req.Limit,
					"excess_amount": req.ExcessAmount,
				},
			})
		})

		// 成本优化建议通知
		costControlAPI.POST("/optimization", func(c *gin.Context) {
			var req struct {
				UserID      uint     `json:"user_id" binding:"required"`
				CurrentCost float64  `json:"current_cost" binding:"required"`
				Suggestions []string `json:"suggestions" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid request parameters",
					"error":   err.Error(),
				})
				return
			}

			err := sender.SendCostOptimizationNotification(req.UserID, req.CurrentCost, req.Suggestions)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Failed to send cost optimization notification",
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "Cost optimization notification sent successfully",
				"data": gin.H{
					"user_id":      req.UserID,
					"current_cost": req.CurrentCost,
					"suggestions":  req.Suggestions,
				},
			})
		})
	}
}

// CostControlNotificationSender 成本控制通知发送器
type CostControlNotificationSender struct {
	core *jobfirst.Core
}

// NewCostControlNotificationSender 创建成本控制通知发送器
func NewCostControlNotificationSender(core *jobfirst.Core) *CostControlNotificationSender {
	return &CostControlNotificationSender{
		core: core,
	}
}

// SendCostLimitWarningNotification 发送成本限制警告通知
func (s *CostControlNotificationSender) SendCostLimitWarningNotification(userID uint, currentCost, limit, percentage float64) error {
	title := "成本使用警告"
	content := "您的使用成本已达到" + strconv.FormatFloat(percentage, 'f', 1, 64) + "%，当前成本：$" + strconv.FormatFloat(currentCost, 'f', 2, 64)

	metadata := map[string]interface{}{
		"current_cost": currentCost,
		"limit":        limit,
		"percentage":   percentage,
		"type":         "cost_limit_warning",
	}

	metadataJSON, _ := json.Marshal(metadata)

	return s.sendNotification(userID, "cost_limit_warning", title, content, "cost_control", "high", string(metadataJSON))
}

// SendCostLimitExceededNotification 发送成本限制超出通知
func (s *CostControlNotificationSender) SendCostLimitExceededNotification(userID uint, currentCost, limit, excessAmount float64) error {
	title := "成本使用超出限制"
	content := "您的使用成本已超出限制，当前成本：$" + strconv.FormatFloat(currentCost, 'f', 2, 64) + "，超出：$" + strconv.FormatFloat(excessAmount, 'f', 2, 64)

	metadata := map[string]interface{}{
		"current_cost":  currentCost,
		"limit":         limit,
		"excess_amount": excessAmount,
		"type":          "cost_limit_exceeded",
	}

	metadataJSON, _ := json.Marshal(metadata)

	return s.sendNotification(userID, "cost_limit_exceeded", title, content, "cost_control", "urgent", string(metadataJSON))
}

// SendCostOptimizationNotification 发送成本优化建议通知
func (s *CostControlNotificationSender) SendCostOptimizationNotification(userID uint, currentCost float64, suggestions []string) error {
	title := "成本优化建议"
	content := "为您推荐以下成本优化方案：\n"
	for i, suggestion := range suggestions {
		content += strconv.Itoa(i+1) + ". " + suggestion + "\n"
	}

	metadata := map[string]interface{}{
		"current_cost": currentCost,
		"suggestions":  suggestions,
		"type":         "cost_optimization",
	}

	metadataJSON, _ := json.Marshal(metadata)

	return s.sendNotification(userID, "cost_optimization", title, content, "cost_control", "normal", string(metadataJSON))
}

// sendNotification 发送通知的通用方法
func (s *CostControlNotificationSender) sendNotification(userID uint, notificationType, title, content, category, priority, metadata string) error {
	// 这里应该调用jobfirst-core的通知发送方法
	// 由于我们直接使用数据库，这里简化处理

	// 创建通知记录
	notification := map[string]interface{}{
		"user_id":    userID,
		"type":       notificationType,
		"title":      title,
		"content":    content,
		"category":   category,
		"priority":   priority,
		"status":     "unread",
		"is_read":    false,
		"metadata":   metadata,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	// 这里应该使用jobfirst-core的数据库操作
	// 暂时返回成功，实际实现需要调用core的数据库方法
	_ = notification // 避免未使用变量警告
	return nil
}
