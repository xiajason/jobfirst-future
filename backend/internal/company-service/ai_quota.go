package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QuotaMiddleware AI服务配额检查中间件
type QuotaMiddleware struct {
	db *gorm.DB
}

// NewQuotaMiddleware 创建配额中间件
func NewQuotaMiddleware(db *gorm.DB) *QuotaMiddleware {
	return &QuotaMiddleware{db: db}
}

// QuotaCheckResult 配额检查结果
type QuotaCheckResult struct {
	Allowed      bool    `json:"allowed"`
	Reason       string  `json:"reason,omitempty"`
	DailyUsed    int     `json:"daily_used"`
	DailyLimit   int     `json:"daily_limit"`
	MonthlyUsed  int     `json:"monthly_used"`
	MonthlyLimit int     `json:"monthly_limit"`
	CostUsed     float64 `json:"cost_used"`
	CostLimit    float64 `json:"cost_limit"`
	ResetTime    string  `json:"reset_time,omitempty"`
}

// AIUsageRecord AI服务使用记录
type AIUsageRecord struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"not null"`
	ServiceType    string    `json:"service_type" gorm:"not null"`
	ServiceName    string    `json:"service_name" gorm:"not null"`
	RequestID      string    `json:"request_id"`
	InputTokens    int       `json:"input_tokens" gorm:"default:0"`
	OutputTokens   int       `json:"output_tokens" gorm:"default:0"`
	TotalTokens    int       `json:"total_tokens" gorm:"default:0"`
	CostUSD        float64   `json:"cost_usd" gorm:"type:decimal(10,6);default:0.000000"`
	ProcessingTime int       `json:"processing_time_ms" gorm:"column:processing_time_ms;default:0"`
	Status         string    `json:"status" gorm:"default:'success'"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AIUsageRecord) TableName() string {
	return "ai_service_usage"
}

// UserAIQuota 用户AI服务配额
type UserAIQuota struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	SubscriptionType string    `json:"subscription_type" gorm:"type:enum('trial','basic','premium','enterprise');default:'trial'"`
	ServiceType      string    `json:"service_type" gorm:"not null"`
	DailyLimit       int       `json:"daily_limit" gorm:"default:0"`
	MonthlyLimit     int       `json:"monthly_limit" gorm:"default:0"`
	DailyUsed        int       `json:"daily_used" gorm:"default:0"`
	MonthlyUsed      int       `json:"monthly_used" gorm:"default:0"`
	DailyCostLimit   float64   `json:"daily_cost_limit" gorm:"type:decimal(10,6);default:0.000000"`
	MonthlyCostLimit float64   `json:"monthly_cost_limit" gorm:"type:decimal(10,6);default:0.000000"`
	DailyCostUsed    float64   `json:"daily_cost_used" gorm:"type:decimal(10,6);default:0.000000"`
	MonthlyCostUsed  float64   `json:"monthly_cost_used" gorm:"type:decimal(10,6);default:0.000000"`
	QuotaResetDate   time.Time `json:"quota_reset_date" gorm:"type:date"`
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UserAIQuota) TableName() string {
	return "user_ai_quotas"
}

// CheckQuota 检查用户配额
func (m *QuotaMiddleware) CheckQuota(userID uint, serviceType string) (*QuotaCheckResult, error) {
	// 获取或创建用户配额
	quota, err := m.getOrCreateQuota(userID, serviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get user quota: %v", err)
	}

	// 检查是否需要重置配额
	if err := m.resetQuotaIfNeeded(quota); err != nil {
		return nil, fmt.Errorf("failed to reset quota: %v", err)
	}

	// 检查配额限制
	result := &QuotaCheckResult{
		DailyUsed:    quota.DailyUsed,
		DailyLimit:   quota.DailyLimit,
		MonthlyUsed:  quota.MonthlyUsed,
		MonthlyLimit: quota.MonthlyLimit,
		CostUsed:     quota.DailyCostUsed,
		CostLimit:    quota.DailyCostLimit,
		ResetTime:    quota.QuotaResetDate.Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	// 检查每日限制
	if quota.DailyUsed >= quota.DailyLimit {
		result.Allowed = false
		result.Reason = "daily limit exceeded"
		return result, nil
	}

	// 检查每月限制
	if quota.MonthlyUsed >= quota.MonthlyLimit {
		result.Allowed = false
		result.Reason = "monthly limit exceeded"
		return result, nil
	}

	// 检查每日成本限制
	if quota.DailyCostUsed >= quota.DailyCostLimit {
		result.Allowed = false
		result.Reason = "daily cost limit exceeded"
		return result, nil
	}

	result.Allowed = true
	return result, nil
}

// RecordUsage 记录使用情况
func (m *QuotaMiddleware) RecordUsage(userID uint, serviceType, serviceName, requestID string, inputTokens, outputTokens int, costUSD float64, processingTime int, status, errorMessage string) error {
	// 记录使用情况
	record := AIUsageRecord{
		UserID:         userID,
		ServiceType:    serviceType,
		ServiceName:    serviceName,
		RequestID:      requestID,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		TotalTokens:    inputTokens + outputTokens,
		CostUSD:        costUSD,
		ProcessingTime: processingTime,
		Status:         status,
		ErrorMessage:   errorMessage,
	}

	if err := m.db.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to record usage: %v", err)
	}

	// 更新配额使用量
	if status == "success" {
		if err := m.updateQuotaUsage(userID, serviceType, costUSD); err != nil {
			return fmt.Errorf("failed to update quota usage: %v", err)
		}
	}

	return nil
}

// getOrCreateQuota 获取或创建用户配额
func (m *QuotaMiddleware) getOrCreateQuota(userID uint, serviceType string) (*UserAIQuota, error) {
	var quota UserAIQuota
	err := m.db.Where("user_id = ? AND service_type = ?", userID, serviceType).First(&quota).Error

	if err == gorm.ErrRecordNotFound {
		// 创建默认配额
		quota, err = m.createDefaultQuota(userID, serviceType)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &quota, nil
}

// createDefaultQuota 创建默认配额
func (m *QuotaMiddleware) createDefaultQuota(userID uint, serviceType string) (UserAIQuota, error) {
	// 获取用户订阅类型
	var subscriptionType string
	err := m.db.Raw("SELECT COALESCE(subscription_status, 'trial') FROM users WHERE id = ?", userID).Scan(&subscriptionType).Error
	if err != nil {
		return UserAIQuota{}, fmt.Errorf("failed to get user subscription type: %v", err)
	}

	// 根据订阅类型设置配额
	quota := UserAIQuota{
		UserID:           userID,
		SubscriptionType: subscriptionType,
		ServiceType:      serviceType,
		QuotaResetDate:   time.Now(),
	}

	switch subscriptionType {
	case "premium":
		quota.DailyLimit = 50
		quota.MonthlyLimit = 1000
		quota.DailyCostLimit = 50.00
		quota.MonthlyCostLimit = 500.00
	case "enterprise":
		quota.DailyLimit = 200
		quota.MonthlyLimit = 5000
		quota.DailyCostLimit = 200.00
		quota.MonthlyCostLimit = 2000.00
	default: // trial, free
		quota.DailyLimit = 5
		quota.MonthlyLimit = 100
		quota.DailyCostLimit = 10.00
		quota.MonthlyCostLimit = 50.00
	}

	if err := m.db.Create(&quota).Error; err != nil {
		return UserAIQuota{}, fmt.Errorf("failed to create default quota: %v", err)
	}

	return quota, nil
}

// resetQuotaIfNeeded 如果需要则重置配额
func (m *QuotaMiddleware) resetQuotaIfNeeded(quota *UserAIQuota) error {
	now := time.Now()

	// 检查是否需要重置每日配额
	if now.Sub(quota.QuotaResetDate) >= 24*time.Hour {
		quota.DailyUsed = 0
		quota.DailyCostUsed = 0
		quota.QuotaResetDate = now

		// 检查是否需要重置每月配额
		if now.Day() == 1 {
			quota.MonthlyUsed = 0
			quota.MonthlyCostUsed = 0
		}

		if err := m.db.Save(quota).Error; err != nil {
			return err
		}
	}

	return nil
}

// updateQuotaUsage 更新配额使用量
func (m *QuotaMiddleware) updateQuotaUsage(userID uint, serviceType string, costUSD float64) error {
	return m.db.Model(&UserAIQuota{}).
		Where("user_id = ? AND service_type = ?", userID, serviceType).
		Updates(map[string]interface{}{
			"daily_used":        gorm.Expr("daily_used + 1"),
			"monthly_used":      gorm.Expr("monthly_used + 1"),
			"daily_cost_used":   gorm.Expr("daily_cost_used + ?", costUSD),
			"monthly_cost_used": gorm.Expr("monthly_cost_used + ?", costUSD),
		}).Error
}

// QuotaCheckMiddleware 配额检查中间件
func (m *QuotaMiddleware) QuotaCheckMiddleware(serviceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		userID, ok := userIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			c.Abort()
			return
		}

		// 检查配额
		result, err := m.CheckQuota(userID, serviceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "quota check failed"})
			c.Abort()
			return
		}

		if !result.Allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "quota exceeded",
				"reason":     result.Reason,
				"quota_info": result,
			})
			c.Abort()
			return
		}

		// 将配额信息存储到上下文
		c.Set("quota_info", result)
		c.Next()
	}
}

// RecordUsageMiddleware 使用记录中间件
func (m *QuotaMiddleware) RecordUsageMiddleware(serviceType, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 从上下文获取用户ID
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			return
		}

		userID, ok := userIDInterface.(uint)
		if !ok {
			return
		}

		// 生成请求ID
		requestID := fmt.Sprintf("%d_%d", userID, time.Now().UnixNano())

		// 记录使用情况（简化版本，实际应该从响应中获取token信息）
		_ = m.RecordUsage(userID, serviceType, serviceName, requestID, 0, 0, 0.01, 0, "success", "")
	}
}
