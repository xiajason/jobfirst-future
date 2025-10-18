package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QuotaAPI AI配额API
type QuotaAPI struct {
	db *gorm.DB
}

// NewQuotaAPI 创建配额API
func NewQuotaAPI(db *gorm.DB) *QuotaAPI {
	return &QuotaAPI{db: db}
}

// GetUserQuota 获取用户配额
func (api *QuotaAPI) GetUserQuota(c *gin.Context) {
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

	middleware := NewQuotaMiddleware(api.db)
	result, err := middleware.CheckQuota(uint(userID), serviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":      userID,
		"service_type": serviceType,
		"quota_info":   result,
	})
}

// GetUserAllQuotas 获取用户所有配额
func (api *QuotaAPI) GetUserAllQuotas(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var quotas []UserAIQuota
	if err := api.db.Where("user_id = ?", userID).Find(&quotas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"quotas":  quotas,
	})
}

// CheckQuota 检查配额
func (api *QuotaAPI) CheckQuota(c *gin.Context) {
	var request struct {
		UserID      uint   `json:"user_id" binding:"required"`
		ServiceType string `json:"service_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware := NewQuotaMiddleware(api.db)
	result, err := middleware.CheckQuota(request.UserID, request.ServiceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quota_info": result,
	})
}

// GetUsageStats 获取使用统计
func (api *QuotaAPI) GetUsageStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	serviceType := c.Query("service_type")

	var records []AIUsageRecord
	query := api.db.Where("user_id = ?", userID)
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	if err := query.Order("created_at DESC").Limit(100).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"service_type":  serviceType,
		"usage_records": records,
	})
}

// ResetQuota 重置配额
func (api *QuotaAPI) ResetQuota(c *gin.Context) {
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

	// 重置配额
	if err := api.db.Model(&UserAIQuota{}).
		Where("user_id = ? AND service_type = ?", userID, serviceType).
		Updates(map[string]interface{}{
			"daily_used":   0,
			"monthly_used": 0,
			"cost_used":    0,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "quota reset successfully",
		"user_id":      userID,
		"service_type": serviceType,
	})
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
