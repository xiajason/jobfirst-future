package main

import (
	"fmt"
	"time"
)

// AI服务分层管理
type AIServiceLayering struct {
	UserID             uint               `json:"user_id"`
	ServiceLevel       string             `json:"service_level"`       // basic, premium, enterprise
	AuthorizationLevel string             `json:"authorization_level"` // no_consent, partial_consent, full_consent
	AvailableServices  []AIService        `json:"available_services"`
	UsageLimits        ServiceUsageLimits `json:"usage_limits"`
	BillingInfo        ServiceBillingInfo `json:"billing_info"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// AI服务定义
type AIService struct {
	ServiceID     string           `json:"service_id"`
	ServiceName   string           `json:"service_name"`
	ServiceType   string           `json:"service_type"`   // resume_analysis, job_matching, career_guidance
	RequiredLevel string           `json:"required_level"` // basic, premium, enterprise
	RequiredAuth  string           `json:"required_auth"`  // no_consent, partial_consent, full_consent
	Features      []ServiceFeature `json:"features"`
	Pricing       ServicePricing   `json:"pricing"`
	Availability  bool             `json:"availability"`
	Description   string           `json:"description"`
}

// 服务功能特性
type ServiceFeature struct {
	FeatureID     string `json:"feature_id"`
	FeatureName   string `json:"feature_name"`
	FeatureType   string `json:"feature_type"` // analysis, matching, guidance, optimization
	RequiredLevel string `json:"required_level"`
	RequiredAuth  string `json:"required_auth"`
	Description   string `json:"description"`
	IsAvailable   bool   `json:"is_available"`
}

// 服务使用限制
type ServiceUsageLimits struct {
	DailyRequests    int    `json:"daily_requests"`
	MonthlyRequests  int    `json:"monthly_requests"`
	ConcurrentJobs   int    `json:"concurrent_jobs"`
	DataStorageLimit int64  `json:"data_storage_limit"` // bytes
	APICallLimit     int    `json:"api_call_limit"`
	ResetPeriod      string `json:"reset_period"` // daily, monthly
}

// 服务计费信息
type ServiceBillingInfo struct {
	BillingModel    string             `json:"billing_model"` // free, pay_per_use, subscription
	BasePrice       float64            `json:"base_price"`
	PricePerRequest float64            `json:"price_per_request"`
	PricePerFeature map[string]float64 `json:"price_per_feature"`
	DiscountRate    float64            `json:"discount_rate"`
	TotalCost       float64            `json:"total_cost"`
	LastBilling     time.Time          `json:"last_billing"`
	NextBilling     time.Time          `json:"next_billing"`
}

// 服务定价
type ServicePricing struct {
	BasePrice       float64            `json:"base_price"`
	PricePerRequest float64            `json:"price_per_request"`
	PricePerFeature map[string]float64 `json:"price_per_feature"`
	FreeRequests    int                `json:"free_requests"`
	DiscountTiers   []DiscountTier     `json:"discount_tiers"`
}

// 折扣层级
type DiscountTier struct {
	MinRequests  int     `json:"min_requests"`
	MaxRequests  int     `json:"max_requests"`
	DiscountRate float64 `json:"discount_rate"`
	Description  string  `json:"description"`
}

// 用户服务使用记录
type UserServiceUsage struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	UserID             uint      `json:"user_id" gorm:"not null"`
	ServiceID          string    `json:"service_id" gorm:"not null"`
	FeatureID          string    `json:"feature_id"`
	RequestType        string    `json:"request_type"`    // resume_analysis, job_matching, career_guidance
	DataSize           int64     `json:"data_size"`       // bytes
	ProcessingTime     int64     `json:"processing_time"` // milliseconds
	Cost               float64   `json:"cost"`
	AuthorizationLevel string    `json:"authorization_level"`
	Anonymized         bool      `json:"anonymized"`
	CreatedAt          time.Time `json:"created_at"`
}

// 服务分层管理器
type AIServiceLayeringManager struct {
	services     map[string]AIService
	userLayering map[uint]AIServiceLayering
	usageTracker map[uint][]UserServiceUsage
}

// 创建AI服务分层管理器
func NewAIServiceLayeringManager() *AIServiceLayeringManager {
	manager := &AIServiceLayeringManager{
		services:     make(map[string]AIService),
		userLayering: make(map[uint]AIServiceLayering),
		usageTracker: make(map[uint][]UserServiceUsage),
	}

	// 初始化默认服务
	manager.initializeDefaultServices()

	return manager
}

// 初始化默认服务
func (m *AIServiceLayeringManager) initializeDefaultServices() {
	// 基础服务 - 简历分析
	m.services["resume_analysis_basic"] = AIService{
		ServiceID:     "resume_analysis_basic",
		ServiceName:   "基础简历分析",
		ServiceType:   "resume_analysis",
		RequiredLevel: "basic",
		RequiredAuth:  "partial_consent",
		Features: []ServiceFeature{
			{
				FeatureID:     "basic_parsing",
				FeatureName:   "基础解析",
				FeatureType:   "analysis",
				RequiredLevel: "basic",
				RequiredAuth:  "partial_consent",
				Description:   "基础简历信息提取和解析",
				IsAvailable:   true,
			},
			{
				FeatureID:     "skill_extraction",
				FeatureName:   "技能提取",
				FeatureType:   "analysis",
				RequiredLevel: "basic",
				RequiredAuth:  "partial_consent",
				Description:   "从简历中提取技能信息",
				IsAvailable:   true,
			},
		},
		Pricing: ServicePricing{
			BasePrice:       0.0,
			PricePerRequest: 0.0,
			PricePerFeature: map[string]float64{
				"basic_parsing":    0.0,
				"skill_extraction": 0.0,
			},
			FreeRequests: 10,
		},
		Availability: true,
		Description:  "基础简历分析服务，免费提供基础功能",
	}

	// 高级服务 - 智能简历优化
	m.services["resume_optimization_premium"] = AIService{
		ServiceID:     "resume_optimization_premium",
		ServiceName:   "智能简历优化",
		ServiceType:   "resume_analysis",
		RequiredLevel: "premium",
		RequiredAuth:  "full_consent",
		Features: []ServiceFeature{
			{
				FeatureID:     "ai_optimization",
				FeatureName:   "AI优化建议",
				FeatureType:   "optimization",
				RequiredLevel: "premium",
				RequiredAuth:  "full_consent",
				Description:   "基于AI的简历优化建议",
				IsAvailable:   true,
			},
			{
				FeatureID:     "keyword_optimization",
				FeatureName:   "关键词优化",
				FeatureType:   "optimization",
				RequiredLevel: "premium",
				RequiredAuth:  "full_consent",
				Description:   "针对特定职位的关键词优化",
				IsAvailable:   true,
			},
			{
				FeatureID:     "format_optimization",
				FeatureName:   "格式优化",
				FeatureType:   "optimization",
				RequiredLevel: "premium",
				RequiredAuth:  "full_consent",
				Description:   "简历格式和布局优化",
				IsAvailable:   true,
			},
		},
		Pricing: ServicePricing{
			BasePrice:       9.99,
			PricePerRequest: 0.50,
			PricePerFeature: map[string]float64{
				"ai_optimization":      2.00,
				"keyword_optimization": 1.00,
				"format_optimization":  0.50,
			},
			FreeRequests: 5,
		},
		Availability: true,
		Description:  "高级简历优化服务，提供AI驱动的优化建议",
	}

	// 企业服务 - 职业发展指导
	m.services["career_guidance_enterprise"] = AIService{
		ServiceID:     "career_guidance_enterprise",
		ServiceName:   "职业发展指导",
		ServiceType:   "career_guidance",
		RequiredLevel: "enterprise",
		RequiredAuth:  "full_consent",
		Features: []ServiceFeature{
			{
				FeatureID:     "career_path_analysis",
				FeatureName:   "职业路径分析",
				FeatureType:   "guidance",
				RequiredLevel: "enterprise",
				RequiredAuth:  "full_consent",
				Description:   "基于用户数据的职业路径分析",
				IsAvailable:   true,
			},
			{
				FeatureID:     "skill_gap_analysis",
				FeatureName:   "技能差距分析",
				FeatureType:   "guidance",
				RequiredLevel: "enterprise",
				RequiredAuth:  "full_consent",
				Description:   "分析用户技能与目标职位的差距",
				IsAvailable:   true,
			},
			{
				FeatureID:     "learning_recommendations",
				FeatureName:   "学习推荐",
				FeatureType:   "guidance",
				RequiredLevel: "enterprise",
				RequiredAuth:  "full_consent",
				Description:   "个性化的学习路径推荐",
				IsAvailable:   true,
			},
		},
		Pricing: ServicePricing{
			BasePrice:       29.99,
			PricePerRequest: 1.00,
			PricePerFeature: map[string]float64{
				"career_path_analysis":     5.00,
				"skill_gap_analysis":       3.00,
				"learning_recommendations": 2.00,
			},
			FreeRequests: 3,
		},
		Availability: true,
		Description:  "企业级职业发展指导服务，提供全面的职业规划支持",
	}
}

// 获取用户可用的服务
func (m *AIServiceLayeringManager) GetUserAvailableServices(userID uint, authorizationLevel string) ([]AIService, error) {
	var availableServices []AIService

	for _, service := range m.services {
		// 检查授权级别
		if m.checkAuthorizationLevel(authorizationLevel, service.RequiredAuth) {
			// 检查服务可用性
			if service.Availability {
				availableServices = append(availableServices, service)
			}
		}
	}

	return availableServices, nil
}

// 检查授权级别
func (m *AIServiceLayeringManager) checkAuthorizationLevel(userAuth, requiredAuth string) bool {
	authLevels := map[string]int{
		"no_consent":      0,
		"partial_consent": 1,
		"full_consent":    2,
	}

	userLevel := authLevels[userAuth]
	requiredLevel := authLevels[requiredAuth]

	return userLevel >= requiredLevel
}

// 记录服务使用
func (m *AIServiceLayeringManager) RecordServiceUsage(userID uint, serviceID string, featureID string, requestType string, dataSize int64, processingTime int64, authorizationLevel string, anonymized bool) error {
	// 获取服务定价
	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("服务不存在: %s", serviceID)
	}

	// 计算费用
	cost := m.calculateServiceCost(service, featureID, dataSize)

	// 创建使用记录
	usage := UserServiceUsage{
		UserID:             userID,
		ServiceID:          serviceID,
		FeatureID:          featureID,
		RequestType:        requestType,
		DataSize:           dataSize,
		ProcessingTime:     processingTime,
		Cost:               cost,
		AuthorizationLevel: authorizationLevel,
		Anonymized:         anonymized,
		CreatedAt:          time.Now(),
	}

	// 添加到使用跟踪器
	m.usageTracker[userID] = append(m.usageTracker[userID], usage)

	return nil
}

// 计算服务费用
func (m *AIServiceLayeringManager) calculateServiceCost(service AIService, featureID string, dataSize int64) float64 {
	// 基础费用
	cost := service.Pricing.BasePrice

	// 按请求收费
	cost += service.Pricing.PricePerRequest

	// 按功能收费
	if featureCost, exists := service.Pricing.PricePerFeature[featureID]; exists {
		cost += featureCost
	}

	// 按数据大小收费（每MB 0.01元）
	dataSizeMB := float64(dataSize) / (1024 * 1024)
	cost += dataSizeMB * 0.01

	return cost
}

// 获取用户服务使用统计
func (m *AIServiceLayeringManager) GetUserUsageStats(userID uint) (map[string]interface{}, error) {
	usageRecords, exists := m.usageTracker[userID]
	if !exists {
		return map[string]interface{}{
			"total_requests": 0,
			"total_cost":     0.0,
			"services_used":  []string{},
		}, nil
	}

	stats := map[string]interface{}{
		"total_requests":  len(usageRecords),
		"total_cost":      0.0,
		"services_used":   []string{},
		"features_used":   []string{},
		"data_processed":  int64(0),
		"processing_time": int64(0),
	}

	servicesUsed := make(map[string]bool)
	featuresUsed := make(map[string]bool)

	for _, record := range usageRecords {
		stats["total_cost"] = stats["total_cost"].(float64) + record.Cost
		stats["data_processed"] = stats["data_processed"].(int64) + record.DataSize
		stats["processing_time"] = stats["processing_time"].(int64) + record.ProcessingTime

		servicesUsed[record.ServiceID] = true
		featuresUsed[record.FeatureID] = true
	}

	// 转换map为slice
	for service := range servicesUsed {
		stats["services_used"] = append(stats["services_used"].([]string), service)
	}

	for feature := range featuresUsed {
		stats["features_used"] = append(stats["features_used"].([]string), feature)
	}

	return stats, nil
}

// 检查用户服务限制
func (m *AIServiceLayeringManager) CheckServiceLimits(userID uint, serviceID string) (bool, error) {
	// 获取用户分层信息
	layering, exists := m.userLayering[userID]
	if !exists {
		// 默认基础用户
		layering = AIServiceLayering{
			UserID:             userID,
			ServiceLevel:       "basic",
			AuthorizationLevel: "partial_consent",
			UsageLimits: ServiceUsageLimits{
				DailyRequests:    10,
				MonthlyRequests:  100,
				ConcurrentJobs:   2,
				DataStorageLimit: 100 * 1024 * 1024, // 100MB
				APICallLimit:     50,
				ResetPeriod:      "daily",
			},
		}
		m.userLayering[userID] = layering
	}

	// 检查使用限制
	usageRecords := m.usageTracker[userID]

	// 检查每日请求限制
	today := time.Now().Truncate(24 * time.Hour)
	dailyCount := 0
	for _, record := range usageRecords {
		if record.CreatedAt.After(today) {
			dailyCount++
		}
	}

	if dailyCount >= layering.UsageLimits.DailyRequests {
		return false, fmt.Errorf("已达到每日请求限制: %d", layering.UsageLimits.DailyRequests)
	}

	// 检查月度请求限制
	monthStart := time.Now().Truncate(24 * time.Hour * 30)
	monthlyCount := 0
	for _, record := range usageRecords {
		if record.CreatedAt.After(monthStart) {
			monthlyCount++
		}
	}

	if monthlyCount >= layering.UsageLimits.MonthlyRequests {
		return false, fmt.Errorf("已达到月度请求限制: %d", layering.UsageLimits.MonthlyRequests)
	}

	return true, nil
}

// 升级用户服务层级
func (m *AIServiceLayeringManager) UpgradeUserServiceLevel(userID uint, newLevel string, newAuthLevel string) error {
	layering, exists := m.userLayering[userID]
	if !exists {
		layering = AIServiceLayering{
			UserID:    userID,
			CreatedAt: time.Now(),
		}
	}

	layering.ServiceLevel = newLevel
	layering.AuthorizationLevel = newAuthLevel
	layering.UpdatedAt = time.Now()

	// 根据新层级更新使用限制
	switch newLevel {
	case "premium":
		layering.UsageLimits = ServiceUsageLimits{
			DailyRequests:    50,
			MonthlyRequests:  500,
			ConcurrentJobs:   5,
			DataStorageLimit: 500 * 1024 * 1024, // 500MB
			APICallLimit:     200,
			ResetPeriod:      "daily",
		}
	case "enterprise":
		layering.UsageLimits = ServiceUsageLimits{
			DailyRequests:    200,
			MonthlyRequests:  2000,
			ConcurrentJobs:   10,
			DataStorageLimit: 2 * 1024 * 1024 * 1024, // 2GB
			APICallLimit:     1000,
			ResetPeriod:      "daily",
		}
	}

	m.userLayering[userID] = layering

	return nil
}

// 获取服务分层信息
func (m *AIServiceLayeringManager) GetServiceLayeringInfo() map[string]interface{} {
	return map[string]interface{}{
		"total_services":       len(m.services),
		"service_levels":       []string{"basic", "premium", "enterprise"},
		"authorization_levels": []string{"no_consent", "partial_consent", "full_consent"},
		"services":             m.services,
	}
}
