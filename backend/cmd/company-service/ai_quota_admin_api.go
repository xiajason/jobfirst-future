package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminAPI 管理员配额API
type AdminAPI struct {
	db *gorm.DB
}

// NewAdminAPI 创建管理员API
func NewAdminAPI(db *gorm.DB) *AdminAPI {
	return &AdminAPI{db: db}
}

// GetUserQuotaDetails 获取用户配额详情
func (api *AdminAPI) GetUserQuotaDetails(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	serviceType := c.Query("service_type")
	if serviceType == "" {
		serviceType = "document_parsing"
	}

	// 获取用户配额
	var quota UserAIQuota
	if err := api.db.Where("user_id = ? AND service_type = ?", userID, serviceType).First(&quota).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "quota not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取用户信息
	var user struct {
		ID                 uint   `json:"id"`
		Username           string `json:"username"`
		SubscriptionStatus string `json:"subscription_status"`
	}
	if err := api.db.Raw("SELECT id, username, COALESCE(subscription_status, 'trial') as subscription_status FROM users WHERE id = ?", userID).Scan(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"quota":        quota,
		"service_type": serviceType,
	})
}

// UpdateUserSubscription 更新用户订阅状态
func (api *AdminAPI) UpdateUserSubscription(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var request struct {
		SubscriptionStatus string `json:"subscription_status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证订阅状态
	validStatuses := []string{"trial", "free", "premium", "enterprise"}
	isValid := false
	for _, status := range validStatuses {
		if request.SubscriptionStatus == status {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription status"})
		return
	}

	// 更新用户订阅状态
	if err := api.db.Exec("UPDATE users SET subscription_status = ? WHERE id = ?", request.SubscriptionStatus, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新所有相关配额
	if err := api.updateQuotasForSubscription(uint(userID), request.SubscriptionStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "subscription updated successfully",
		"user_id":             userID,
		"subscription_status": request.SubscriptionStatus,
	})
}

// ResetUserQuota 重置用户配额
func (api *AdminAPI) ResetUserQuota(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	serviceType := c.Query("service_type")
	if serviceType == "" {
		// 重置所有服务的配额
		if err := api.db.Model(&UserAIQuota{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"daily_used":   0,
				"monthly_used": 0,
				"cost_used":    0,
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// 重置特定服务的配额
		if err := api.db.Model(&UserAIQuota{}).
			Where("user_id = ? AND service_type = ?", userID, serviceType).
			Updates(map[string]interface{}{
				"daily_used":        0,
				"monthly_used":      0,
				"daily_cost_used":   0,
				"monthly_cost_used": 0,
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "quota reset successfully",
		"user_id":      userID,
		"service_type": serviceType,
	})
}

// GetSystemStats 获取系统统计
func (api *AdminAPI) GetSystemStats(c *gin.Context) {
	// 获取总用户数
	var totalUsers int64
	api.db.Model(&struct{}{}).Table("users").Count(&totalUsers)

	// 获取配额统计
	var quotaStats struct {
		TotalQuotas      int64   `json:"total_quotas"`
		TotalDailyUsed   int64   `json:"total_daily_used"`
		TotalMonthlyUsed int64   `json:"total_monthly_used"`
		TotalCostUsed    float64 `json:"total_cost_used"`
	}

	api.db.Model(&UserAIQuota{}).Count(&quotaStats.TotalQuotas)
	api.db.Model(&UserAIQuota{}).Select("COALESCE(SUM(daily_used), 0)").Scan(&quotaStats.TotalDailyUsed)
	api.db.Model(&UserAIQuota{}).Select("COALESCE(SUM(monthly_used), 0)").Scan(&quotaStats.TotalMonthlyUsed)
	api.db.Model(&UserAIQuota{}).Select("COALESCE(SUM(cost_used), 0)").Scan(&quotaStats.TotalCostUsed)

	// 获取使用记录统计
	var usageStats struct {
		TotalRecords int64 `json:"total_records"`
		SuccessCount int64 `json:"success_count"`
		ErrorCount   int64 `json:"error_count"`
	}

	api.db.Model(&AIUsageRecord{}).Count(&usageStats.TotalRecords)
	api.db.Model(&AIUsageRecord{}).Where("status = ?", "success").Count(&usageStats.SuccessCount)
	api.db.Model(&AIUsageRecord{}).Where("status = ?", "error").Count(&usageStats.ErrorCount)

	c.JSON(http.StatusOK, gin.H{
		"total_users": totalUsers,
		"quota_stats": quotaStats,
		"usage_stats": usageStats,
	})
}

// updateQuotasForSubscription 根据订阅状态更新配额
func (api *AdminAPI) updateQuotasForSubscription(userID uint, subscriptionStatus string) error {
	// 定义不同订阅类型的配额
	var dailyLimit, monthlyLimit int
	var dailyCostLimit, monthlyCostLimit float64

	switch subscriptionStatus {
	case "premium":
		dailyLimit = 50
		monthlyLimit = 1000
		dailyCostLimit = 50.00
		monthlyCostLimit = 500.00
	case "enterprise":
		dailyLimit = 200
		monthlyLimit = 5000
		dailyCostLimit = 200.00
		monthlyCostLimit = 2000.00
	default: // trial, free
		dailyLimit = 5
		monthlyLimit = 100
		dailyCostLimit = 10.00
		monthlyCostLimit = 50.00
	}

	// 更新所有现有配额
	return api.db.Model(&UserAIQuota{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"subscription_type":  subscriptionStatus,
			"daily_limit":        dailyLimit,
			"monthly_limit":      monthlyLimit,
			"daily_cost_limit":   dailyCostLimit,
			"monthly_cost_limit": monthlyCostLimit,
		}).Error
}

// RegisterAdminRoutes 注册管理员路由
func (api *AdminAPI) RegisterAdminRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin/quota")
	{
		// 支持两种路径格式
		admin.GET("/user/:user_id", api.GetUserQuotaDetails)
		admin.GET("/user/:user_id/details", api.GetUserQuotaDetails)
		admin.PUT("/user/:user_id/subscription", api.UpdateUserSubscription)
		admin.POST("/user/:user_id/reset", api.ResetUserQuota)
		admin.GET("/system/stats", api.GetSystemStats)
	}
}
