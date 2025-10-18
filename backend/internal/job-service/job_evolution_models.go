package main

import (
	"time"
)

// 远程工作支持平台数据模型

// RemoteWorkJob 远程工作职位
type RemoteWorkJob struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	JobID               uint      `json:"job_id" gorm:"not null"`
	RemoteType          string    `json:"remote_type" gorm:"size:50"` // fully_remote, hybrid, flexible
	TimeZone            string    `json:"time_zone" gorm:"size:50"`
	WorkHours           string    `json:"work_hours" gorm:"size:100"`            // 工作时间要求
	CommunicationTools  string    `json:"communication_tools" gorm:"type:text"`  // 沟通工具
	EquipmentProvided   string    `json:"equipment_provided" gorm:"type:text"`   // 设备提供
	FlexibilityLevel    string    `json:"flexibility_level" gorm:"size:50"`      // 灵活度级别
	LocationRequirement string    `json:"location_requirement" gorm:"type:text"` // 地点要求
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// 关联
	Job Job `json:"job" gorm:"foreignKey:JobID"`
}

// FlexibleEmployment 灵活用工管理
type FlexibleEmployment struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	JobID             uint      `json:"job_id" gorm:"not null"`
	EmploymentType    string    `json:"employment_type" gorm:"size:50"`      // contract, freelance, part_time, project_based
	Duration          string    `json:"duration" gorm:"size:100"`            // 工作期限
	PaymentType       string    `json:"payment_type" gorm:"size:50"`         // hourly, project, milestone
	PaymentRate       float64   `json:"payment_rate"`                        // 支付费率
	FlexibilityLevel  string    `json:"flexibility_level" gorm:"size:50"`    // 灵活度
	SkillRequirements string    `json:"skill_requirements" gorm:"type:text"` // 技能要求
	ProjectScope      string    `json:"project_scope" gorm:"type:text"`      // 项目范围
	Timeline          string    `json:"timeline" gorm:"size:200"`            // 时间线
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// 关联
	Job Job `json:"job" gorm:"foreignKey:JobID"`
}

// SmartMatching 智能匹配引擎
type SmartMatching struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	JobID            uint      `json:"job_id" gorm:"not null"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	MatchScore       float64   `json:"match_score"`                           // 匹配分数
	SkillMatch       float64   `json:"skill_match"`                           // 技能匹配度
	ExperienceMatch  float64   `json:"experience_match"`                      // 经验匹配度
	LocationMatch    float64   `json:"location_match"`                        // 地点匹配度
	SalaryMatch      float64   `json:"salary_match"`                          // 薪资匹配度
	CultureMatch     float64   `json:"culture_match"`                         // 文化匹配度
	AIRecommendation string    `json:"ai_recommendation" gorm:"type:text"`    // AI推荐理由
	MatchFactors     string    `json:"match_factors" gorm:"type:json"`        // 匹配因子
	Status           string    `json:"status" gorm:"size:20;default:pending"` // pending, accepted, rejected
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联
	Job  Job  `json:"job" gorm:"foreignKey:JobID"`
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// CareerDevelopment 个性化职业发展
type CareerDevelopment struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	CurrentLevel     string    `json:"current_level" gorm:"size:50"`       // 当前职业级别
	TargetLevel      string    `json:"target_level" gorm:"size:50"`        // 目标职业级别
	SkillsGap        string    `json:"skills_gap" gorm:"type:text"`        // 技能差距
	DevelopmentPlan  string    `json:"development_plan" gorm:"type:text"`  // 发展计划
	RecommendedJobs  string    `json:"recommended_jobs" gorm:"type:json"`  // 推荐职位
	SkillAssessment  string    `json:"skill_assessment" gorm:"type:json"`  // 技能评估
	CareerGoals      string    `json:"career_goals" gorm:"type:text"`      // 职业目标
	ProgressTracking string    `json:"progress_tracking" gorm:"type:json"` // 进度跟踪
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// WorkLifeBalance 工作生活平衡
type WorkLifeBalance struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	JobID            uint      `json:"job_id" gorm:"not null"`
	FlexibleHours    bool      `json:"flexible_hours"`                     // 灵活工作时间
	RemoteWorkDays   int       `json:"remote_work_days"`                   // 远程工作天数
	VacationDays     int       `json:"vacation_days"`                      // 假期天数
	HealthBenefits   string    `json:"health_benefits" gorm:"type:text"`   // 健康福利
	WellnessPrograms string    `json:"wellness_programs" gorm:"type:text"` // 健康项目
	FamilySupport    string    `json:"family_support" gorm:"type:text"`    // 家庭支持
	WorkloadBalance  string    `json:"workload_balance" gorm:"size:100"`   // 工作负载平衡
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联
	Job Job `json:"job" gorm:"foreignKey:JobID"`
}

// SkillAssessment 技能评估
type SkillAssessment struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	SkillName        string    `json:"skill_name" gorm:"size:100"`
	SkillLevel       int       `json:"skill_level"`                      // 1-5级别
	AssessmentMethod string    `json:"assessment_method" gorm:"size:50"` // 评估方法
	AssessmentScore  float64   `json:"assessment_score"`                 // 评估分数
	Certification    string    `json:"certification" gorm:"size:200"`    // 认证
	ExperienceYears  int       `json:"experience_years"`                 // 经验年数
	LastUpdated      time.Time `json:"last_updated"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// JobRecommendation AI职位推荐
type JobRecommendation struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	UserID               uint      `json:"user_id" gorm:"not null"`
	JobID                uint      `json:"job_id" gorm:"not null"`
	RecommendationScore  float64   `json:"recommendation_score"`                   // 推荐分数
	RecommendationReason string    `json:"recommendation_reason" gorm:"type:text"` // 推荐理由
	AIInsights           string    `json:"ai_insights" gorm:"type:text"`           // AI洞察
	MatchFactors         string    `json:"match_factors" gorm:"type:json"`         // 匹配因子
	Status               string    `json:"status" gorm:"size:20;default:active"`   // active, viewed, applied, dismissed
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
	Job  Job  `json:"job" gorm:"foreignKey:JobID"`
}

// 枚举常量定义

// 远程工作类型
const (
	RemoteTypeFullyRemote = "fully_remote"
	RemoteTypeHybrid      = "hybrid"
	RemoteTypeFlexible    = "flexible"
)

// 灵活用工类型
const (
	EmploymentTypeContract     = "contract"
	EmploymentTypeFreelance    = "freelance"
	EmploymentTypePartTime     = "part_time"
	EmploymentTypeProjectBased = "project_based"
)

// 支付类型
const (
	PaymentTypeHourly    = "hourly"
	PaymentTypeProject   = "project"
	PaymentTypeMilestone = "milestone"
)

// 灵活度级别
const (
	FlexibilityLevelHigh   = "high"
	FlexibilityLevelMedium = "medium"
	FlexibilityLevelLow    = "low"
)

// 匹配状态
const (
	MatchStatusPending  = "pending"
	MatchStatusAccepted = "accepted"
	MatchStatusRejected = "rejected"
)

// 推荐状态
const (
	RecommendationStatusActive    = "active"
	RecommendationStatusViewed    = "viewed"
	RecommendationStatusApplied   = "applied"
	RecommendationStatusDismissed = "dismissed"
)

// 评估方法
const (
	AssessmentMethodSelf      = "self"
	AssessmentMethodTest      = "test"
	AssessmentMethodProject   = "project"
	AssessmentMethodInterview = "interview"
)
