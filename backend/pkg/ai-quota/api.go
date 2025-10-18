package aiquota

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QuotaAPI AI配额API
type QuotaAPI struct {
	db         *gorm.DB
	middleware *QuotaMiddleware
}

// NewQuotaAPI 创建配额API
func NewQuotaAPI(db *gorm.DB) *QuotaAPI {
	return &QuotaAPI{
		db:         db,
		middleware: NewQuotaMiddleware(db),
	}
}

// GetUserQuota 获取用户配额信息
// @Summary 获取用户配额信息
// @Description 获取指定用户的AI服务配额信息
// @Tags AI Quota
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param service_type query string false "服务类型"
// @Success 200 {object} QuotaCheckResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/quota/user/{user_id} [get]
func (api *QuotaAPI) GetUserQuota(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	serviceType := c.Query("service_type")
	if serviceType == "" {
		serviceType = "document_parsing" // 默认服务类型
	}

	result, err := api.middleware.GetUserQuota(uint(userID), serviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get quota",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetUserAllQuotas 获取用户所有服务配额信息
// @Summary 获取用户所有服务配额信息
// @Description 获取指定用户的所有AI服务配额信息
// @Tags AI Quota
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200 {object} map[string]QuotaCheckResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/quota/user/{user_id}/all [get]
func (api *QuotaAPI) GetUserAllQuotas(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// 获取所有服务类型
	serviceTypes := []string{"document_parsing", "text_analysis", "ai_chat"}
	quotas := make(map[string]*QuotaCheckResult)

	for _, serviceType := range serviceTypes {
		result, err := api.middleware.GetUserQuota(uint(userID), serviceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to get quota for service: " + serviceType,
				"details": err.Error(),
			})
			return
		}
		quotas[serviceType] = result
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    quotas,
	})
}

// CheckQuota 检查用户配额
// @Summary 检查用户配额
// @Description 检查用户是否可以调用指定的AI服务
// @Tags AI Quota
// @Accept json
// @Produce json
// @Param request body QuotaCheckRequest true "配额检查请求"
// @Success 200 {object} QuotaCheckResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/quota/check [post]
func (api *QuotaAPI) CheckQuota(c *gin.Context) {
	var request QuotaCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	result, err := api.middleware.CheckQuota(request.UserID, request.ServiceType, request.EstimatedCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to check quota",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetUsageStats 获取使用统计
// @Summary 获取使用统计
// @Description 获取用户的AI服务使用统计信息
// @Tags AI Quota
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param service_type query string false "服务类型"
// @Param days query int false "统计天数，默认30天"
// @Success 200 {object} UsageStats
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/quota/user/{user_id}/usage [get]
func (api *QuotaAPI) GetUsageStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	serviceType := c.Query("service_type")
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 30
	}

	stats, err := api.getUsageStats(uint(userID), serviceType, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get usage stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ResetQuota 重置用户配额
// @Summary 重置用户配额
// @Description 重置指定用户的配额使用量
// @Tags AI Quota
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Param service_type query string false "服务类型，不指定则重置所有服务"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/quota/user/{user_id}/reset [post]
func (api *QuotaAPI) ResetQuota(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	serviceType := c.Query("service_type")

	err = api.resetUserQuota(uint(userID), serviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to reset quota",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "quota reset successfully",
	})
}

// QuotaCheckRequest 配额检查请求
type QuotaCheckRequest struct {
	UserID        uint    `json:"user_id" binding:"required"`
	ServiceType   string  `json:"service_type" binding:"required"`
	EstimatedCost float64 `json:"estimated_cost"`
}

// UsageStats 使用统计
type UsageStats struct {
	UserID          uint         `json:"user_id"`
	ServiceType     string       `json:"service_type"`
	TotalCalls      int          `json:"total_calls"`
	SuccessCalls    int          `json:"success_calls"`
	FailedCalls     int          `json:"failed_calls"`
	TotalCost       float64      `json:"total_cost"`
	TotalTokens     int          `json:"total_tokens"`
	AvgResponseTime float64      `json:"avg_response_time"`
	DailyStats      []DailyUsage `json:"daily_stats"`
}

// DailyUsage 每日使用统计
type DailyUsage struct {
	Date            string  `json:"date"`
	Calls           int     `json:"calls"`
	SuccessCalls    int     `json:"success_calls"`
	FailedCalls     int     `json:"failed_calls"`
	Cost            float64 `json:"cost"`
	Tokens          int     `json:"tokens"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// getUsageStats 获取使用统计
func (api *QuotaAPI) getUsageStats(userID uint, serviceType string, days int) (*UsageStats, error) {
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

	// 获取每日统计
	var dailyStats []DailyUsage
	err = query.Select(`
		DATE(created_at) as date,
		COUNT(*) as calls,
		SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_calls,
		SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_calls,
		SUM(cost_usd) as cost,
		SUM(total_tokens) as tokens,
		AVG(processing_time_ms) as avg_response_time
	`).Where("created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)", days).
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&dailyStats).Error

	if err != nil {
		return nil, err
	}

	stats.DailyStats = dailyStats

	return &stats, nil
}

// resetUserQuota 重置用户配额
func (api *QuotaAPI) resetUserQuota(userID uint, serviceType string) error {
	query := api.db.Model(&UserAIQuota{}).Where("user_id = ?", userID)
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	return query.Updates(map[string]interface{}{
		"daily_used":        0,
		"monthly_used":      0,
		"daily_cost_used":   0,
		"monthly_cost_used": 0,
		"quota_reset_date":  "NOW()",
	}).Error
}

// RegisterRoutes 注册路由
func (api *QuotaAPI) RegisterRoutes(r *gin.RouterGroup) {
	quota := r.Group("/quota")
	{
		quota.GET("/user/:user_id", api.GetUserQuota)
		quota.GET("/user/:user_id/all", api.GetUserAllQuotas)
		quota.POST("/check", api.CheckQuota)
		quota.GET("/user/:user_id/usage", api.GetUsageStats)
		quota.POST("/user/:user_id/reset", api.ResetQuota)
	}
}
