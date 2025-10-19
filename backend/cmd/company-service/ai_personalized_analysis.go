package main

import (
	"fmt"
	"time"
)

// 个性化分析管理器
type PersonalizedAnalysisManager struct {
	userProfiles    map[uint]UserProfile
	analysisHistory map[uint][]AnalysisResult
	recommendations map[uint][]Recommendation
}

// 用户画像
type UserProfile struct {
	UserID              uint             `json:"user_id"`
	BasicInfo           UserBasicInfo    `json:"basic_info"`
	SkillsProfile       SkillsProfile    `json:"skills_profile"`
	CareerProfile       CareerProfile    `json:"career_profile"`
	LearningProfile     LearningProfile  `json:"learning_profile"`
	Preferences         UserPreferences  `json:"preferences"`
	BehaviorPatterns    BehaviorPatterns `json:"behavior_patterns"`
	LastUpdated         time.Time        `json:"last_updated"`
	ProfileCompleteness float64          `json:"profile_completeness"`
}

// 用户基本信息
type UserBasicInfo struct {
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Location    string `json:"location"`
	Education   string `json:"education"`
	Experience  int    `json:"experience"` // years
	Industry    string `json:"industry"`
	CurrentRole string `json:"current_role"`
	CareerLevel string `json:"career_level"` // entry, mid, senior, lead, manager
	SalaryRange string `json:"salary_range"`
	WorkStyle   string `json:"work_style"` // remote, hybrid, office
}

// 技能画像
type SkillsProfile struct {
	TechnicalSkills   []Skill         `json:"technical_skills"`
	SoftSkills        []Skill         `json:"soft_skills"`
	LanguageSkills    []Skill         `json:"language_skills"`
	Certifications    []Certification `json:"certifications"`
	SkillGaps         []SkillGap      `json:"skill_gaps"`
	SkillTrends       []SkillTrend    `json:"skill_trends"`
	OverallSkillLevel float64         `json:"overall_skill_level"`
}

// 技能定义
type Skill struct {
	SkillID       string    `json:"skill_id"`
	SkillName     string    `json:"skill_name"`
	SkillCategory string    `json:"skill_category"`
	Proficiency   int       `json:"proficiency"` // 1-5
	Experience    int       `json:"experience"`  // months
	LastUsed      time.Time `json:"last_used"`
	Relevance     float64   `json:"relevance"` // 0-1
	Demand        float64   `json:"demand"`    // 0-1
}

// 认证信息
type Certification struct {
	CertID          string     `json:"cert_id"`
	CertName        string     `json:"cert_name"`
	Issuer          string     `json:"issuer"`
	IssueDate       time.Time  `json:"issue_date"`
	ExpiryDate      *time.Time `json:"expiry_date"`
	CredentialID    string     `json:"credential_id"`
	VerificationURL string     `json:"verification_url"`
}

// 技能差距
type SkillGap struct {
	SkillID       string             `json:"skill_id"`
	SkillName     string             `json:"skill_name"`
	CurrentLevel  int                `json:"current_level"`
	RequiredLevel int                `json:"required_level"`
	GapSize       int                `json:"gap_size"`
	Priority      string             `json:"priority"` // high, medium, low
	LearningPath  []LearningResource `json:"learning_path"`
}

// 技能趋势
type SkillTrend struct {
	SkillID      string  `json:"skill_id"`
	SkillName    string  `json:"skill_name"`
	Trend        string  `json:"trend"` // rising, stable, declining
	GrowthRate   float64 `json:"growth_rate"`
	MarketDemand float64 `json:"market_demand"`
	FutureValue  float64 `json:"future_value"`
}

// 职业画像
type CareerProfile struct {
	CareerGoals        []CareerGoal       `json:"career_goals"`
	CareerPath         []CareerStep       `json:"career_path"`
	IndustryFocus      []IndustryFocus    `json:"industry_focus"`
	RolePreferences    []RolePreference   `json:"role_preferences"`
	SalaryExpectations SalaryExpectations `json:"salary_expectations"`
	WorkLifeBalance    WorkLifeBalance    `json:"work_life_balance"`
	CareerStage        string             `json:"career_stage"` // early, mid, senior, executive
}

// 职业目标
type CareerGoal struct {
	GoalID         string  `json:"goal_id"`
	GoalName       string  `json:"goal_name"`
	GoalType       string  `json:"goal_type"` // promotion, skill_development, career_change
	TargetRole     string  `json:"target_role"`
	TargetIndustry string  `json:"target_industry"`
	TargetSalary   int     `json:"target_salary"`
	Timeline       int     `json:"timeline"` // months
	Priority       string  `json:"priority"` // high, medium, low
	Status         string  `json:"status"`   // active, completed, paused
	Progress       float64 `json:"progress"` // 0-1
}

// 职业步骤
type CareerStep struct {
	StepID         string   `json:"step_id"`
	StepName       string   `json:"step_name"`
	StepType       string   `json:"step_type"` // skill_development, networking, certification, experience
	Description    string   `json:"description"`
	RequiredSkills []string `json:"required_skills"`
	EstimatedTime  int      `json:"estimated_time"` // months
	Difficulty     string   `json:"difficulty"`     // easy, medium, hard
	Status         string   `json:"status"`         // pending, in_progress, completed
	Progress       float64  `json:"progress"`       // 0-1
}

// 行业关注
type IndustryFocus struct {
	IndustryID      string      `json:"industry_id"`
	IndustryName    string      `json:"industry_name"`
	InterestLevel   int         `json:"interest_level"` // 1-5
	Experience      int         `json:"experience"`     // months
	GrowthPotential float64     `json:"growth_potential"`
	SalaryRange     SalaryRange `json:"salary_range"`
	KeySkills       []string    `json:"key_skills"`
}

// 角色偏好
type RolePreference struct {
	RoleID         string      `json:"role_id"`
	RoleName       string      `json:"role_name"`
	RoleType       string      `json:"role_type"`      // individual_contributor, manager, director, executive
	InterestLevel  int         `json:"interest_level"` // 1-5
	RequiredSkills []string    `json:"required_skills"`
	SalaryRange    SalaryRange `json:"salary_range"`
	WorkStyle      string      `json:"work_style"` // remote, hybrid, office
}

// 薪资期望
type SalaryExpectations struct {
	CurrentSalary      int    `json:"current_salary"`
	TargetSalary       int    `json:"target_salary"`
	MinAcceptable      int    `json:"min_acceptable"`
	MaxExpected        int    `json:"max_expected"`
	Currency           string `json:"currency"`
	Negotiable         bool   `json:"negotiable"`
	BonusExpectations  int    `json:"bonus_expectations"`
	EquityExpectations int    `json:"equity_expectations"`
}

// 薪资范围
type SalaryRange struct {
	Min       int    `json:"min"`
	Max       int    `json:"max"`
	Currency  string `json:"currency"`
	Frequency string `json:"frequency"` // annual, monthly, hourly
}

// 工作生活平衡
type WorkLifeBalance struct {
	PreferredHours    int      `json:"preferred_hours"`
	Flexibility       string   `json:"flexibility"`        // high, medium, low
	RemoteWork        string   `json:"remote_work"`        // required, preferred, optional, not_interested
	TravelWillingness string   `json:"travel_willingness"` // none, minimal, moderate, extensive
	WorkEnvironment   []string `json:"work_environment"`   // startup, corporate, non_profit, government
	Values            []string `json:"values"`             // innovation, stability, growth, impact
}

// 学习画像
type LearningProfile struct {
	LearningStyle     string             `json:"learning_style"`    // visual, auditory, kinesthetic, reading
	PreferredMethods  []string           `json:"preferred_methods"` // online, classroom, hands_on, mentoring
	LearningGoals     []LearningGoal     `json:"learning_goals"`
	LearningHistory   []LearningActivity `json:"learning_history"`
	LearningResources []LearningResource `json:"learning_resources"`
	LearningBudget    int                `json:"learning_budget"`
	TimeAvailability  int                `json:"time_availability"` // hours per week
}

// 学习目标
type LearningGoal struct {
	GoalID       string  `json:"goal_id"`
	GoalName     string  `json:"goal_name"`
	SkillID      string  `json:"skill_id"`
	TargetLevel  int     `json:"target_level"`  // 1-5
	CurrentLevel int     `json:"current_level"` // 1-5
	Timeline     int     `json:"timeline"`      // months
	Priority     string  `json:"priority"`      // high, medium, low
	Status       string  `json:"status"`        // active, completed, paused
	Progress     float64 `json:"progress"`      // 0-1
}

// 学习活动
type LearningActivity struct {
	ActivityID   string     `json:"activity_id"`
	ActivityName string     `json:"activity_name"`
	ActivityType string     `json:"activity_type"` // course, workshop, certification, project
	Provider     string     `json:"provider"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Status       string     `json:"status"` // completed, in_progress, planned
	Score        float64    `json:"score"`
	SkillsGained []string   `json:"skills_gained"`
	Cost         float64    `json:"cost"`
}

// 学习资源
type LearningResource struct {
	ResourceID          string   `json:"resource_id"`
	ResourceName        string   `json:"resource_name"`
	ResourceType        string   `json:"resource_type"` // course, book, video, tutorial, certification
	Provider            string   `json:"provider"`
	URL                 string   `json:"url"`
	Cost                float64  `json:"cost"`
	Duration            int      `json:"duration"`   // hours
	Difficulty          string   `json:"difficulty"` // beginner, intermediate, advanced
	Skills              []string `json:"skills"`
	Rating              float64  `json:"rating"`
	RecommendationScore float64  `json:"recommendation_score"`
}

// 用户偏好
type UserPreferences struct {
	JobSearchPreferences    JobSearchPreferences    `json:"job_search_preferences"`
	NotificationPreferences NotificationPreferences `json:"notification_preferences"`
	PrivacyPreferences      PrivacyPreferences      `json:"privacy_preferences"`
	DisplayPreferences      DisplayPreferences      `json:"display_preferences"`
}

// 求职偏好
type JobSearchPreferences struct {
	JobTypes     []string    `json:"job_types"` // full_time, part_time, contract, freelance
	Industries   []string    `json:"industries"`
	Locations    []string    `json:"locations"`
	RemoteWork   string      `json:"remote_work"`   // required, preferred, optional, not_interested
	CompanySize  []string    `json:"company_size"`  // startup, small, medium, large, enterprise
	CompanyStage []string    `json:"company_stage"` // startup, growth, mature, public
	SalaryRange  SalaryRange `json:"salary_range"`
	Benefits     []string    `json:"benefits"`
	WorkCulture  []string    `json:"work_culture"`
}

// 通知偏好
type NotificationPreferences struct {
	EmailNotifications bool   `json:"email_notifications"`
	SMSNotifications   bool   `json:"sms_notifications"`
	PushNotifications  bool   `json:"push_notifications"`
	JobAlerts          bool   `json:"job_alerts"`
	SkillUpdates       bool   `json:"skill_updates"`
	CareerAdvice       bool   `json:"career_advice"`
	Frequency          string `json:"frequency"` // immediate, daily, weekly, monthly
}

// 隐私偏好
type PrivacyPreferences struct {
	ProfileVisibility string `json:"profile_visibility"` // public, connections, private
	DataSharing       string `json:"data_sharing"`       // full, partial, minimal
	AnalyticsOptIn    bool   `json:"analytics_opt_in"`
	MarketingOptIn    bool   `json:"marketing_opt_in"`
	ThirdPartySharing bool   `json:"third_party_sharing"`
}

// 显示偏好
type DisplayPreferences struct {
	Language     string `json:"language"`
	Currency     string `json:"currency"`
	TimeZone     string `json:"time_zone"`
	DateFormat   string `json:"date_format"`
	NumberFormat string `json:"number_format"`
	Theme        string `json:"theme"` // light, dark, auto
}

// 行为模式
type BehaviorPatterns struct {
	JobSearchBehavior   JobSearchBehavior   `json:"job_search_behavior"`
	LearningBehavior    LearningBehavior    `json:"learning_behavior"`
	NetworkingBehavior  NetworkingBehavior  `json:"networking_behavior"`
	ApplicationBehavior ApplicationBehavior `json:"application_behavior"`
}

// 求职行为
type JobSearchBehavior struct {
	SearchFrequency     string   `json:"search_frequency"` // daily, weekly, monthly, occasionally
	SearchKeywords      []string `json:"search_keywords"`
	AppliedJobs         int      `json:"applied_jobs"`
	InterviewRate       float64  `json:"interview_rate"`
	OfferRate           float64  `json:"offer_rate"`
	AverageResponseTime int      `json:"average_response_time"` // days
}

// 学习行为
type LearningBehavior struct {
	LearningFrequency string   `json:"learning_frequency"` // daily, weekly, monthly
	PreferredTime     string   `json:"preferred_time"`     // morning, afternoon, evening, night
	LearningDuration  int      `json:"learning_duration"`  // minutes per session
	CompletionRate    float64  `json:"completion_rate"`
	PreferredFormat   []string `json:"preferred_format"` // video, text, audio, interactive
}

// 网络行为
type NetworkingBehavior struct {
	NetworkingFrequency string   `json:"networking_frequency"` // daily, weekly, monthly, rarely
	PlatformsUsed       []string `json:"platforms_used"`
	ConnectionsMade     int      `json:"connections_made"`
	EventsAttended      int      `json:"events_attended"`
	MentorshipSeeking   bool     `json:"mentorship_seeking"`
}

// 申请行为
type ApplicationBehavior struct {
	ApplicationStrategy string `json:"application_strategy"` // targeted, broad, selective
	CustomizationLevel  string `json:"customization_level"`  // high, medium, low
	FollowUpFrequency   string `json:"follow_up_frequency"`  // always, sometimes, rarely, never
	ResponseTime        int    `json:"response_time"`        // hours
}

// 分析结果
type AnalysisResult struct {
	AnalysisID      string                 `json:"analysis_id"`
	UserID          uint                   `json:"user_id"`
	AnalysisType    string                 `json:"analysis_type"` // resume, skills, career, learning
	AnalysisData    map[string]interface{} `json:"analysis_data"`
	Insights        []Insight              `json:"insights"`
	Recommendations []Recommendation       `json:"recommendations"`
	Confidence      float64                `json:"confidence"` // 0-1
	CreatedAt       time.Time              `json:"created_at"`
}

// 洞察
type Insight struct {
	InsightID   string  `json:"insight_id"`
	InsightType string  `json:"insight_type"` // strength, weakness, opportunity, threat
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`     // high, medium, low
	Confidence  float64 `json:"confidence"` // 0-1
	Actionable  bool    `json:"actionable"`
	Priority    string  `json:"priority"` // high, medium, low
}

// 推荐
type Recommendation struct {
	RecommendationID   string             `json:"recommendation_id"`
	RecommendationType string             `json:"recommendation_type"` // skill_development, career_move, learning, networking
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	Priority           string             `json:"priority"` // high, medium, low
	Effort             string             `json:"effort"`   // low, medium, high
	Timeline           int                `json:"timeline"` // days
	ExpectedOutcome    string             `json:"expected_outcome"`
	Resources          []LearningResource `json:"resources"`
	SuccessRate        float64            `json:"success_rate"`       // 0-1
	PersonalizedScore  float64            `json:"personalized_score"` // 0-1
}

// 创建个性化分析管理器
func NewPersonalizedAnalysisManager() *PersonalizedAnalysisManager {
	return &PersonalizedAnalysisManager{
		userProfiles:    make(map[uint]UserProfile),
		analysisHistory: make(map[uint][]AnalysisResult),
		recommendations: make(map[uint][]Recommendation),
	}
}

// 更新用户画像
func (m *PersonalizedAnalysisManager) UpdateUserProfile(userID uint, profile UserProfile) error {
	profile.UserID = userID
	profile.LastUpdated = time.Now()

	// 计算画像完整性
	profile.ProfileCompleteness = m.calculateProfileCompleteness(profile)

	m.userProfiles[userID] = profile

	return nil
}

// 计算画像完整性
func (m *PersonalizedAnalysisManager) calculateProfileCompleteness(profile UserProfile) float64 {
	completeness := 0.0
	totalFields := 0.0

	// 基本信息完整性
	if profile.BasicInfo.Age > 0 {
		completeness += 1
	}
	if profile.BasicInfo.Experience > 0 {
		completeness += 1
	}
	if profile.BasicInfo.Industry != "" {
		completeness += 1
	}
	if profile.BasicInfo.CurrentRole != "" {
		completeness += 1
	}
	totalFields += 4

	// 技能画像完整性
	if len(profile.SkillsProfile.TechnicalSkills) > 0 {
		completeness += 1
	}
	if len(profile.SkillsProfile.SoftSkills) > 0 {
		completeness += 1
	}
	totalFields += 2

	// 职业画像完整性
	if len(profile.CareerProfile.CareerGoals) > 0 {
		completeness += 1
	}
	if profile.CareerProfile.CareerStage != "" {
		completeness += 1
	}
	totalFields += 2

	// 学习画像完整性
	if len(profile.LearningProfile.LearningGoals) > 0 {
		completeness += 1
	}
	if profile.LearningProfile.LearningStyle != "" {
		completeness += 1
	}
	totalFields += 2

	return completeness / totalFields
}

// 执行个性化分析
func (m *PersonalizedAnalysisManager) PerformPersonalizedAnalysis(userID uint, analysisType string) (*AnalysisResult, error) {
	profile, exists := m.userProfiles[userID]
	if !exists {
		return nil, fmt.Errorf("用户画像不存在: %d", userID)
	}

	analysisResult := &AnalysisResult{
		AnalysisID:   fmt.Sprintf("analysis_%d_%s_%d", userID, analysisType, time.Now().Unix()),
		UserID:       userID,
		AnalysisType: analysisType,
		AnalysisData: make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}

	// 根据分析类型执行不同的分析
	switch analysisType {
	case "resume":
		analysisResult.Insights, analysisResult.Recommendations = m.analyzeResume(profile)
	case "skills":
		analysisResult.Insights, analysisResult.Recommendations = m.analyzeSkills(profile)
	case "career":
		analysisResult.Insights, analysisResult.Recommendations = m.analyzeCareer(profile)
	case "learning":
		analysisResult.Insights, analysisResult.Recommendations = m.analyzeLearning(profile)
	default:
		return nil, fmt.Errorf("不支持的分析类型: %s", analysisType)
	}

	// 计算整体置信度
	analysisResult.Confidence = m.calculateAnalysisConfidence(analysisResult)

	// 保存分析历史
	m.analysisHistory[userID] = append(m.analysisHistory[userID], *analysisResult)

	return analysisResult, nil
}

// 分析简历
func (m *PersonalizedAnalysisManager) analyzeResume(profile UserProfile) ([]Insight, []Recommendation) {
	var insights []Insight
	var recommendations []Recommendation

	// 简历完整性分析
	if profile.ProfileCompleteness < 0.5 {
		insights = append(insights, Insight{
			InsightID:   "resume_incomplete",
			InsightType: "weakness",
			Title:       "简历信息不完整",
			Description: "您的简历信息完整度较低，建议补充更多详细信息以提高匹配度",
			Impact:      "high",
			Confidence:  0.9,
			Actionable:  true,
			Priority:    "high",
		})

		recommendations = append(recommendations, Recommendation{
			RecommendationID:   "complete_profile",
			RecommendationType: "skill_development",
			Title:              "完善个人资料",
			Description:        "建议补充完整的个人信息、技能和经验描述",
			Priority:           "high",
			Effort:             "low",
			Timeline:           7,
			ExpectedOutcome:    "提高简历匹配度和求职成功率",
			SuccessRate:        0.8,
			PersonalizedScore:  0.9,
		})
	}

	// 技能匹配分析
	if len(profile.SkillsProfile.TechnicalSkills) > 0 {
		insights = append(insights, Insight{
			InsightID:   "skills_strength",
			InsightType: "strength",
			Title:       "技术技能优势",
			Description: fmt.Sprintf("您拥有%d项技术技能，这是您的核心优势", len(profile.SkillsProfile.TechnicalSkills)),
			Impact:      "high",
			Confidence:  0.8,
			Actionable:  false,
			Priority:    "medium",
		})
	}

	return insights, recommendations
}

// 分析技能
func (m *PersonalizedAnalysisManager) analyzeSkills(profile UserProfile) ([]Insight, []Recommendation) {
	var insights []Insight
	var recommendations []Recommendation

	// 技能差距分析
	for _, gap := range profile.SkillsProfile.SkillGaps {
		insights = append(insights, Insight{
			InsightID:   fmt.Sprintf("skill_gap_%s", gap.SkillID),
			InsightType: "weakness",
			Title:       fmt.Sprintf("%s技能差距", gap.SkillName),
			Description: fmt.Sprintf("您的%s技能当前水平为%d，但目标职位要求%d", gap.SkillName, gap.CurrentLevel, gap.RequiredLevel),
			Impact:      gap.Priority,
			Confidence:  0.9,
			Actionable:  true,
			Priority:    gap.Priority,
		})

		recommendations = append(recommendations, Recommendation{
			RecommendationID:   fmt.Sprintf("develop_skill_%s", gap.SkillID),
			RecommendationType: "skill_development",
			Title:              fmt.Sprintf("提升%s技能", gap.SkillName),
			Description:        fmt.Sprintf("建议通过学习和实践提升%s技能到目标水平", gap.SkillName),
			Priority:           gap.Priority,
			Effort:             "medium",
			Timeline:           30,
			ExpectedOutcome:    fmt.Sprintf("将%s技能提升到%d级", gap.SkillName, gap.RequiredLevel),
			Resources:          gap.LearningPath,
			SuccessRate:        0.7,
			PersonalizedScore:  0.8,
		})
	}

	return insights, recommendations
}

// 分析职业发展
func (m *PersonalizedAnalysisManager) analyzeCareer(profile UserProfile) ([]Insight, []Recommendation) {
	var insights []Insight
	var recommendations []Recommendation

	// 职业目标分析
	for _, goal := range profile.CareerProfile.CareerGoals {
		if goal.Status == "active" {
			insights = append(insights, Insight{
				InsightID:   fmt.Sprintf("career_goal_%s", goal.GoalID),
				InsightType: "opportunity",
				Title:       fmt.Sprintf("职业目标: %s", goal.GoalName),
				Description: fmt.Sprintf("您正在追求%s目标，当前进度为%.1f%%", goal.GoalName, goal.Progress*100),
				Impact:      "high",
				Confidence:  0.8,
				Actionable:  true,
				Priority:    goal.Priority,
			})
		}
	}

	// 职业发展建议
	recommendations = append(recommendations, Recommendation{
		RecommendationID:   "career_development",
		RecommendationType: "career_move",
		Title:              "职业发展规划",
		Description:        "基于您的技能和职业目标，建议制定详细的职业发展计划",
		Priority:           "high",
		Effort:             "high",
		Timeline:           90,
		ExpectedOutcome:    "实现职业目标，提升职业发展水平",
		SuccessRate:        0.6,
		PersonalizedScore:  0.9,
	})

	return insights, recommendations
}

// 分析学习需求
func (m *PersonalizedAnalysisManager) analyzeLearning(profile UserProfile) ([]Insight, []Recommendation) {
	var insights []Insight
	var recommendations []Recommendation

	// 学习目标分析
	for _, goal := range profile.LearningProfile.LearningGoals {
		insights = append(insights, Insight{
			InsightID:   fmt.Sprintf("learning_goal_%s", goal.GoalID),
			InsightType: "opportunity",
			Title:       fmt.Sprintf("学习目标: %s", goal.GoalName),
			Description: fmt.Sprintf("您正在学习%s，当前进度为%.1f%%", goal.GoalName, goal.Progress*100),
			Impact:      "medium",
			Confidence:  0.8,
			Actionable:  true,
			Priority:    goal.Priority,
		})
	}

	// 学习建议
	recommendations = append(recommendations, Recommendation{
		RecommendationID:   "learning_plan",
		RecommendationType: "learning",
		Title:              "个性化学习计划",
		Description:        "基于您的学习偏好和目标，制定个性化的学习计划",
		Priority:           "medium",
		Effort:             "medium",
		Timeline:           60,
		ExpectedOutcome:    "提升技能水平，实现学习目标",
		SuccessRate:        0.7,
		PersonalizedScore:  0.8,
	})

	return insights, recommendations
}

// 计算分析置信度
func (m *PersonalizedAnalysisManager) calculateAnalysisConfidence(result *AnalysisResult) float64 {
	// 基于画像完整性和分析质量计算置信度
	profile, exists := m.userProfiles[result.UserID]
	if !exists {
		return 0.5
	}

	baseConfidence := profile.ProfileCompleteness

	// 根据洞察数量调整置信度
	insightCount := len(result.Insights)
	if insightCount > 5 {
		baseConfidence += 0.1
	} else if insightCount < 2 {
		baseConfidence -= 0.1
	}

	// 确保置信度在0-1范围内
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	} else if baseConfidence < 0.0 {
		baseConfidence = 0.0
	}

	return baseConfidence
}

// 获取个性化推荐
func (m *PersonalizedAnalysisManager) GetPersonalizedRecommendations(userID uint, limit int) ([]Recommendation, error) {
	profile, exists := m.userProfiles[userID]
	if !exists {
		return nil, fmt.Errorf("用户画像不存在: %d", userID)
	}

	var recommendations []Recommendation

	// 基于用户画像生成推荐
	recommendations = append(recommendations, m.generateSkillRecommendations(profile)...)
	recommendations = append(recommendations, m.generateCareerRecommendations(profile)...)
	recommendations = append(recommendations, m.generateLearningRecommendations(profile)...)

	// 按个性化分数排序
	for i := 0; i < len(recommendations); i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[i].PersonalizedScore < recommendations[j].PersonalizedScore {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	// 限制返回数量
	if limit > 0 && len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

// 生成技能推荐
func (m *PersonalizedAnalysisManager) generateSkillRecommendations(profile UserProfile) []Recommendation {
	var recommendations []Recommendation

	// 基于技能差距生成推荐
	for _, gap := range profile.SkillsProfile.SkillGaps {
		if gap.Priority == "high" {
			recommendations = append(recommendations, Recommendation{
				RecommendationID:   fmt.Sprintf("skill_%s", gap.SkillID),
				RecommendationType: "skill_development",
				Title:              fmt.Sprintf("提升%s技能", gap.SkillName),
				Description:        fmt.Sprintf("建议重点提升%s技能，当前水平%d，目标水平%d", gap.SkillName, gap.CurrentLevel, gap.RequiredLevel),
				Priority:           "high",
				Effort:             "medium",
				Timeline:           30,
				ExpectedOutcome:    fmt.Sprintf("将%s技能提升到%d级", gap.SkillName, gap.RequiredLevel),
				Resources:          gap.LearningPath,
				SuccessRate:        0.7,
				PersonalizedScore:  0.9,
			})
		}
	}

	return recommendations
}

// 生成职业推荐
func (m *PersonalizedAnalysisManager) generateCareerRecommendations(profile UserProfile) []Recommendation {
	var recommendations []Recommendation

	// 基于职业目标生成推荐
	for _, goal := range profile.CareerProfile.CareerGoals {
		if goal.Status == "active" && goal.Progress < 0.5 {
			recommendations = append(recommendations, Recommendation{
				RecommendationID:   fmt.Sprintf("career_%s", goal.GoalID),
				RecommendationType: "career_move",
				Title:              fmt.Sprintf("推进%s目标", goal.GoalName),
				Description:        fmt.Sprintf("建议采取具体行动推进%s目标的实现", goal.GoalName),
				Priority:           goal.Priority,
				Effort:             "high",
				Timeline:           goal.Timeline,
				ExpectedOutcome:    fmt.Sprintf("实现%s目标", goal.GoalName),
				SuccessRate:        0.6,
				PersonalizedScore:  0.8,
			})
		}
	}

	return recommendations
}

// 生成学习推荐
func (m *PersonalizedAnalysisManager) generateLearningRecommendations(profile UserProfile) []Recommendation {
	var recommendations []Recommendation

	// 基于学习目标生成推荐
	for _, goal := range profile.LearningProfile.LearningGoals {
		if goal.Status == "active" {
			recommendations = append(recommendations, Recommendation{
				RecommendationID:   fmt.Sprintf("learning_%s", goal.GoalID),
				RecommendationType: "learning",
				Title:              fmt.Sprintf("学习%s", goal.GoalName),
				Description:        fmt.Sprintf("建议继续学习%s，当前进度%.1f%%", goal.GoalName, goal.Progress*100),
				Priority:           goal.Priority,
				Effort:             "medium",
				Timeline:           goal.Timeline,
				ExpectedOutcome:    fmt.Sprintf("完成%s学习目标", goal.GoalName),
				SuccessRate:        0.7,
				PersonalizedScore:  0.8,
			})
		}
	}

	return recommendations
}

// 获取分析历史
func (m *PersonalizedAnalysisManager) GetAnalysisHistory(userID uint, limit int) ([]AnalysisResult, error) {
	history, exists := m.analysisHistory[userID]
	if !exists {
		return []AnalysisResult{}, nil
	}

	// 按时间倒序排列
	for i := 0; i < len(history); i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].CreatedAt.Before(history[j].CreatedAt) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}

	// 限制返回数量
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	return history, nil
}
