package aiquota

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

// UserAIQuota 用户AI配额
type UserAIQuota struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	UserID           uint       `json:"user_id" gorm:"not null"`
	SubscriptionType string     `json:"subscription_type" gorm:"default:'trial'"`
	ServiceType      string     `json:"service_type" gorm:"not null"`
	DailyLimit       int        `json:"daily_limit" gorm:"default:0"`
	MonthlyLimit     int        `json:"monthly_limit" gorm:"default:0"`
	DailyUsed        int        `json:"daily_used" gorm:"default:0"`
	MonthlyUsed      int        `json:"monthly_used" gorm:"default:0"`
	DailyCostLimit   float64    `json:"daily_cost_limit" gorm:"type:decimal(10,6);default:0.000000"`
	MonthlyCostLimit float64    `json:"monthly_cost_limit" gorm:"type:decimal(10,6);default:0.000000"`
	DailyCostUsed    float64    `json:"daily_cost_used" gorm:"type:decimal(10,6);default:0.000000"`
	MonthlyCostUsed  float64    `json:"monthly_cost_used" gorm:"type:decimal(10,6);default:0.000000"`
	QuotaResetDate   *time.Time `json:"quota_reset_date"`
	IsActive         bool       `json:"is_active" gorm:"default:true"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (UserAIQuota) TableName() string {
	return "user_ai_quotas"
}

// CheckQuota 检查用户配额
func (m *QuotaMiddleware) CheckQuota(userID uint, serviceType string, estimatedCost float64) (*QuotaCheckResult, error) {
	// 获取用户配额
	var quota UserAIQuota
	err := m.db.Where("user_id = ? AND service_type = ? AND is_active = ?", userID, serviceType, true).First(&quota).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有配额记录，创建默认配额
			quota, err = m.createDefaultQuota(userID, serviceType)
			if err != nil {
				return nil, fmt.Errorf("failed to create default quota: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get quota: %v", err)
		}
	}

	// 检查是否需要重置配额
	if err := m.resetQuotaIfNeeded(&quota); err != nil {
		return nil, fmt.Errorf("failed to reset quota: %v", err)
	}

	// 检查每日限制
	if quota.DailyUsed >= quota.DailyLimit {
		return &QuotaCheckResult{
			Allowed:      false,
			Reason:       "daily_limit_exceeded",
			DailyUsed:    quota.DailyUsed,
			DailyLimit:   quota.DailyLimit,
			MonthlyUsed:  quota.MonthlyUsed,
			MonthlyLimit: quota.MonthlyLimit,
			CostUsed:     quota.DailyCostUsed,
			CostLimit:    quota.DailyCostLimit,
			ResetTime:    quota.QuotaResetDate.Add(24 * time.Hour).Format(time.RFC3339),
		}, nil
	}

	// 检查每月限制
	if quota.MonthlyUsed >= quota.MonthlyLimit {
		return &QuotaCheckResult{
			Allowed:      false,
			Reason:       "monthly_limit_exceeded",
			DailyUsed:    quota.DailyUsed,
			DailyLimit:   quota.DailyLimit,
			MonthlyUsed:  quota.MonthlyUsed,
			MonthlyLimit: quota.MonthlyLimit,
			CostUsed:     quota.MonthlyCostUsed,
			CostLimit:    quota.MonthlyCostLimit,
		}, nil
	}

	// 检查每日成本限制
	if quota.DailyCostUsed+estimatedCost > quota.DailyCostLimit {
		return &QuotaCheckResult{
			Allowed:      false,
			Reason:       "daily_cost_limit_exceeded",
			DailyUsed:    quota.DailyUsed,
			DailyLimit:   quota.DailyLimit,
			MonthlyUsed:  quota.MonthlyUsed,
			MonthlyLimit: quota.MonthlyLimit,
			CostUsed:     quota.DailyCostUsed,
			CostLimit:    quota.DailyCostLimit,
			ResetTime:    quota.QuotaResetDate.Add(24 * time.Hour).Format(time.RFC3339),
		}, nil
	}

	// 检查每月成本限制
	if quota.MonthlyCostUsed+estimatedCost > quota.MonthlyCostLimit {
		return &QuotaCheckResult{
			Allowed:      false,
			Reason:       "monthly_cost_limit_exceeded",
			DailyUsed:    quota.DailyUsed,
			DailyLimit:   quota.DailyLimit,
			MonthlyUsed:  quota.MonthlyUsed,
			MonthlyLimit: quota.MonthlyLimit,
			CostUsed:     quota.MonthlyCostUsed,
			CostLimit:    quota.MonthlyCostLimit,
		}, nil
	}

	// 配额检查通过
	return &QuotaCheckResult{
		Allowed:      true,
		DailyUsed:    quota.DailyUsed,
		DailyLimit:   quota.DailyLimit,
		MonthlyUsed:  quota.MonthlyUsed,
		MonthlyLimit: quota.MonthlyLimit,
		CostUsed:     quota.DailyCostUsed,
		CostLimit:    quota.DailyCostLimit,
	}, nil
}

// RecordUsage 记录AI服务使用
func (m *QuotaMiddleware) RecordUsage(userID uint, serviceType, serviceName, requestID string, inputTokens, outputTokens int, costUSD float64, processingTime int, status string, errorMessage string) error {
	// 记录使用统计
	usage := AIUsageRecord{
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

	if err := m.db.Create(&usage).Error; err != nil {
		return fmt.Errorf("failed to record usage: %v", err)
	}

	// 更新配额使用量
	if status == "success" {
		return m.updateQuotaUsage(userID, serviceType, costUSD)
	}

	return nil
}

// createDefaultQuota 创建默认配额
func (m *QuotaMiddleware) createDefaultQuota(userID uint, serviceType string) (UserAIQuota, error) {
	// 获取用户订阅类型
	var subscriptionType string
	err := m.db.Raw("SELECT COALESCE(subscription_status, 'trial') FROM users WHERE id = ?", userID).Scan(&subscriptionType).Error
	if err != nil {
		return UserAIQuota{}, fmt.Errorf("failed to get user subscription type: %v", err)
	}

	// 获取默认限制配置
	var limits struct {
		DailyLimit       int     `json:"daily_limit"`
		MonthlyLimit     int     `json:"monthly_limit"`
		DailyCostLimit   float64 `json:"daily_cost_limit"`
		MonthlyCostLimit float64 `json:"monthly_cost_limit"`
	}

	err = m.db.Raw(`
		SELECT daily_limit, monthly_limit, daily_cost_limit, monthly_cost_limit 
		FROM subscription_limits 
		WHERE subscription_type = ? AND service_type = ? AND is_active = ?
	`, subscriptionType, serviceType, true).Scan(&limits).Error

	if err != nil {
		return UserAIQuota{}, fmt.Errorf("failed to get default limits: %v", err)
	}

	// 创建配额记录
	quota := UserAIQuota{
		UserID:           userID,
		SubscriptionType: subscriptionType,
		ServiceType:      serviceType,
		DailyLimit:       limits.DailyLimit,
		MonthlyLimit:     limits.MonthlyLimit,
		DailyCostLimit:   limits.DailyCostLimit,
		MonthlyCostLimit: limits.MonthlyCostLimit,
		QuotaResetDate:   &[]time.Time{time.Now()}[0],
		IsActive:         true,
	}

	if err := m.db.Create(&quota).Error; err != nil {
		return UserAIQuota{}, fmt.Errorf("failed to create quota: %v", err)
	}

	return quota, nil
}

// resetQuotaIfNeeded 检查并重置配额
func (m *QuotaMiddleware) resetQuotaIfNeeded(quota *UserAIQuota) error {
	now := time.Now()

	// 检查是否需要重置每日配额
	if quota.QuotaResetDate == nil || now.After(quota.QuotaResetDate.Add(24*time.Hour)) {
		quota.DailyUsed = 0
		quota.DailyCostUsed = 0
		quota.QuotaResetDate = &now
	}

	// 检查是否需要重置每月配额（每月1号）
	if now.Day() == 1 {
		quota.MonthlyUsed = 0
		quota.MonthlyCostUsed = 0
	}

	return m.db.Save(quota).Error
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

// GetUserQuota 获取用户配额信息
func (m *QuotaMiddleware) GetUserQuota(userID uint, serviceType string) (*QuotaCheckResult, error) {
	var quota UserAIQuota
	err := m.db.Where("user_id = ? AND service_type = ? AND is_active = ?", userID, serviceType, true).First(&quota).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有配额记录，创建默认配额
			quota, err = m.createDefaultQuota(userID, serviceType)
			if err != nil {
				return nil, fmt.Errorf("failed to create default quota: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get quota: %v", err)
		}
	}

	// 检查是否需要重置配额
	if err := m.resetQuotaIfNeeded(&quota); err != nil {
		return nil, fmt.Errorf("failed to reset quota: %v", err)
	}

	return &QuotaCheckResult{
		Allowed:      true,
		DailyUsed:    quota.DailyUsed,
		DailyLimit:   quota.DailyLimit,
		MonthlyUsed:  quota.MonthlyUsed,
		MonthlyLimit: quota.MonthlyLimit,
		CostUsed:     quota.DailyCostUsed,
		CostLimit:    quota.DailyCostLimit,
	}, nil
}

// QuotaCheckMiddleware Gin中间件
func (m *QuotaMiddleware) QuotaCheckMiddleware(serviceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从JWT token中获取用户ID
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user not authenticated",
			})
			c.Abort()
			return
		}

		userID, ok := userIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user ID",
			})
			c.Abort()
			return
		}

		// 估算成本（这里可以根据实际情况调整）
		estimatedCost := 0.01 // 默认估算成本

		// 检查配额
		result, err := m.CheckQuota(userID, serviceType, estimatedCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check quota",
			})
			c.Abort()
			return
		}

		if !result.Allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "quota exceeded",
				"quota": result,
			})
			c.Abort()
			return
		}

		// 将配额信息存储到上下文中
		c.Set("quota_info", result)
		c.Next()
	}
}

// RecordUsageMiddleware 记录使用量的中间件
func (m *QuotaMiddleware) RecordUsageMiddleware(serviceType, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 获取用户ID
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			return
		}

		userID, ok := userIDInterface.(uint)
		if !ok {
			return
		}

		// 计算处理时间
		processingTime := int(time.Since(startTime).Milliseconds())

		// 获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
		}

		// 估算token和成本（这里可以根据实际情况调整）
		inputTokens := 100 // 默认输入token
		outputTokens := 50 // 默认输出token
		costUSD := 0.01    // 默认成本

		// 确定状态
		status := "success"
		errorMessage := ""
		if c.Writer.Status() >= 400 {
			status = "failed"
			errorMessage = "HTTP error"
		}

		// 记录使用量
		m.RecordUsage(userID, serviceType, serviceName, requestID, inputTokens, outputTokens, costUSD, processingTime, status, errorMessage)
	}
}
