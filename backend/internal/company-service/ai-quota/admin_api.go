package aiquota

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminAPI 管理员配额管理API
type AdminAPI struct {
	db         *gorm.DB
	middleware *QuotaMiddleware
}

// NewAdminAPI 创建管理员API
func NewAdminAPI(db *gorm.DB) *AdminAPI {
	return &AdminAPI{
		db:         db,
		middleware: NewQuotaMiddleware(db),
	}
}

// UpdateUserSubscription 更新用户订阅状态
// @Summary 更新用户订阅状态
// @Description 管理员更新用户订阅状态并自动调整配额
// @Tags Admin Quota
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param request body UpdateSubscriptionRequest true "订阅更新请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/quota/user/{user_id}/subscription [put]
func (api *AdminAPI) UpdateUserSubscription(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	var request UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// 验证订阅类型
	validTypes := []string{"trial", "basic", "premium", "enterprise"}
	if !contains(validTypes, request.SubscriptionStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid subscription status",
			"valid_types": validTypes,
		})
		return
	}

	// 更新用户订阅状态
	err = api.db.Exec("UPDATE users SET subscription_status = ? WHERE id = ?",
		request.SubscriptionStatus, userID).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update user subscription",
			"details": err.Error(),
		})
		return
	}

	// 更新所有服务的配额限制
	serviceTypes := []string{"document_parsing", "text_analysis", "ai_chat"}
	updatedServices := []string{}

	for _, serviceType := range serviceTypes {
		// 获取新的限制配置
		var limits struct {
			DailyLimit       int     `json:"daily_limit"`
			MonthlyLimit     int     `json:"monthly_limit"`
			DailyCostLimit   float64 `json:"daily_cost_limit"`
			MonthlyCostLimit float64 `json:"monthly_cost_limit"`
		}

		err = api.db.Raw(`
			SELECT daily_limit, monthly_limit, daily_cost_limit, monthly_cost_limit 
			FROM subscription_limits 
			WHERE subscription_type = ? AND service_type = ? AND is_active = true
		`, request.SubscriptionStatus, serviceType).Scan(&limits).Error

		if err != nil {
			continue // 跳过没有配置的服务
		}

		// 更新配额限制
		err = api.db.Model(&UserAIQuota{}).Where("user_id = ? AND service_type = ?", userID, serviceType).Updates(map[string]interface{}{
			"subscription_type":  request.SubscriptionStatus,
			"daily_limit":        limits.DailyLimit,
			"monthly_limit":      limits.MonthlyLimit,
			"daily_cost_limit":   limits.DailyCostLimit,
			"monthly_cost_limit": limits.MonthlyCostLimit,
		}).Error

		if err == nil {
			updatedServices = append(updatedServices, serviceType)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user subscription updated successfully",
		"data": gin.H{
			"user_id":                 userID,
			"new_subscription_status": request.SubscriptionStatus,
			"updated_services":        updatedServices,
		},
	})
}

// ResetUserQuota 重置用户配额
// @Summary 重置用户配额
// @Description 管理员重置用户配额使用量
// @Tags Admin Quota
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param service_type query string false "服务类型，不指定则重置所有服务"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/quota/user/{user_id}/reset [post]
func (api *AdminAPI) ResetUserQuota(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	serviceType := c.Query("service_type")

	// 重置配额使用量
	query := api.db.Model(&UserAIQuota{}).Where("user_id = ?", userID)
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	err = query.Updates(map[string]interface{}{
		"daily_used":        0,
		"monthly_used":      0,
		"daily_cost_used":   0,
		"monthly_cost_used": 0,
	}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to reset quota",
			"details": err.Error(),
		})
		return
	}

	message := "quota reset successfully"
	if serviceType != "" {
		message = "quota reset successfully for service: " + serviceType
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data": gin.H{
			"user_id":      userID,
			"service_type": serviceType,
		},
	})
}

// GetUserQuotaDetails 获取用户配额详情
// @Summary 获取用户配额详情
// @Description 管理员查看用户配额详情
// @Tags Admin Quota
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/quota/user/{user_id}/details [get]
func (api *AdminAPI) GetUserQuotaDetails(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// 获取用户信息
	var user struct {
		ID                 uint   `json:"id"`
		Username           string `json:"username"`
		SubscriptionStatus string `json:"subscription_status"`
	}
	err = api.db.Raw("SELECT id, username, subscription_status FROM users WHERE id = ?", userID).Scan(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	// 获取所有服务的配额信息
	serviceTypes := []string{"document_parsing", "text_analysis", "ai_chat"}
	quotas := make(map[string]*QuotaCheckResult)

	for _, serviceType := range serviceTypes {
		result, err := api.middleware.GetUserQuota(uint(userID), serviceType)
		if err != nil {
			quotas[serviceType] = &QuotaCheckResult{
				Allowed: false,
				Reason:  "error",
			}
		} else {
			quotas[serviceType] = result
		}
	}

	// 获取使用统计
	usageStats := make(map[string]interface{})
	for _, serviceType := range serviceTypes {
		stats, err := api.getUsageStats(uint(userID), serviceType, 30)
		if err == nil {
			usageStats[serviceType] = stats
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":        user,
			"quotas":      quotas,
			"usage_stats": usageStats,
		},
	})
}

// GetSystemQuotaStats 获取系统配额统计
// @Summary 获取系统配额统计
// @Description 管理员查看系统配额使用统计
// @Tags Admin Quota
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/admin/quota/system/stats [get]
func (api *AdminAPI) GetSystemQuotaStats(c *gin.Context) {
	// 获取各订阅类型的用户数量
	var subscriptionStats []struct {
		SubscriptionStatus string `json:"subscription_status"`
		UserCount          int    `json:"user_count"`
	}

	err := api.db.Raw(`
		SELECT subscription_status, COUNT(*) as user_count 
		FROM users 
		GROUP BY subscription_status
	`).Scan(&subscriptionStats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get subscription stats",
		})
		return
	}

	// 获取今日AI服务使用统计
	var todayUsage struct {
		TotalCalls   int     `json:"total_calls"`
		TotalCost    float64 `json:"total_cost"`
		SuccessCalls int     `json:"success_calls"`
		FailedCalls  int     `json:"failed_calls"`
	}

	err = api.db.Raw(`
		SELECT 
			COUNT(*) as total_calls,
			SUM(cost_usd) as total_cost,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_calls,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_calls
		FROM ai_service_usage 
		WHERE DATE(created_at) = CURDATE()
	`).Scan(&todayUsage).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get usage stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"subscription_stats": subscriptionStats,
			"today_usage":        todayUsage,
		},
	})
}

// UpdateSubscriptionRequest 更新订阅请求
type UpdateSubscriptionRequest struct {
	SubscriptionStatus string `json:"subscription_status" binding:"required"`
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getUsageStats 获取使用统计（复用QuotaAPI的方法）
func (api *AdminAPI) getUsageStats(userID uint, serviceType string, days int) (*UsageStats, error) {
	var stats UsageStats
	stats.UserID = userID
	stats.ServiceType = serviceType

	// 构建查询条件
	query := api.db.Model(&AIUsageRecord{}).Where("user_id = ?", userID)
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	// 获取总体统计
	var totalStats struct {
		TotalCalls      int     `json:"total_calls"`
		SuccessCalls    int     `json:"success_calls"`
		FailedCalls     int     `json:"failed_calls"`
		TotalCost       float64 `json:"total_cost"`
		TotalTokens     int     `json:"total_tokens"`
		AvgResponseTime float64 `json:"avg_response_time"`
	}

	err := query.Select(`
		COUNT(*) as total_calls,
		SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_calls,
		SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_calls,
		SUM(cost_usd) as total_cost,
		SUM(total_tokens) as total_tokens,
		AVG(processing_time_ms) as avg_response_time
	`).Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", days).Scan(&totalStats).Error

	if err != nil {
		return nil, err
	}

	stats.TotalCalls = totalStats.TotalCalls
	stats.SuccessCalls = totalStats.SuccessCalls
	stats.FailedCalls = totalStats.FailedCalls
	stats.TotalCost = totalStats.TotalCost
	stats.TotalTokens = totalStats.TotalTokens
	stats.AvgResponseTime = totalStats.AvgResponseTime

	return &stats, nil
}

// RegisterAdminRoutes 注册管理员路由
func (api *AdminAPI) RegisterAdminRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin/quota")
	{
		admin.PUT("/user/:user_id/subscription", api.UpdateUserSubscription)
		admin.POST("/user/:user_id/reset", api.ResetUserQuota)
		admin.GET("/user/:user_id/details", api.GetUserQuotaDetails)
		admin.GET("/system/stats", api.GetSystemQuotaStats)
	}
}
