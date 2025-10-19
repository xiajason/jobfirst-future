package main

import (
	"time"

	"gorm.io/gorm"
)

// ==============================================
// 基础数据模型
// ==============================================

// CompanyInfo 公司信息模型（用于Job服务，不包含完整公司数据）
type CompanyInfo struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"size:255;not null"`
	ShortName string `json:"short_name" gorm:"size:100"`
	LogoURL   string `json:"logo_url" gorm:"size:500"`
	Industry  string `json:"industry" gorm:"size:100"`
	Location  string `json:"location" gorm:"size:200"`
	// 注意：这里不包含JobCount、ViewCount等统计字段，避免数据不一致
}

// User 用户模型
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"size:100;unique"`
	Email    string `json:"email" gorm:"size:255;unique"`
	Role     string `json:"role" gorm:"size:50"`
}

// ResumeMetadata 简历元数据模型
type ResumeMetadata struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	Title         string         `json:"title" gorm:"size:255;not null"`
	ParsingStatus string         `json:"parsing_status" gorm:"size:20"`
	SQLiteDBPath  string         `json:"sqlite_db_path" gorm:"size:500"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// ==============================================
// 职位数据模型
// ==============================================

// Job 职位表
type Job struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Title        string    `json:"title" gorm:"size:200;not null"`
	Description  string    `json:"description" gorm:"type:text"`
	Requirements string    `json:"requirements" gorm:"type:text"`
	CompanyID    uint      `json:"company_id" gorm:"not null"`
	Industry     string    `json:"industry" gorm:"size:100"`
	Location     string    `json:"location" gorm:"size:200"`
	SalaryMin    int       `json:"salary_min"`
	SalaryMax    int       `json:"salary_max"`
	Experience   string    `json:"experience" gorm:"size:50"`
	Education    string    `json:"education" gorm:"size:100"`
	JobType      string    `json:"job_type" gorm:"size:50"` // full-time, part-time, contract
	Status       string    `json:"status" gorm:"size:20;default:active"`
	ViewCount    int       `json:"view_count" gorm:"default:0"`
	ApplyCount   int       `json:"apply_count" gorm:"default:0"`
	CreatedBy    uint      `json:"created_by" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联
	Company      CompanyInfo      `json:"company" gorm:"foreignKey:CompanyID"`
	Applications []JobApplication `json:"applications,omitempty" gorm:"foreignKey:JobID"`
}

// JobApplication 职位申请表
type JobApplication struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	JobID       uint       `json:"job_id" gorm:"not null"`
	UserID      uint       `json:"user_id" gorm:"not null"`
	ResumeID    uint       `json:"resume_id" gorm:"not null"`
	Status      string     `json:"status" gorm:"size:20;default:pending"` // pending, reviewed, accepted, rejected
	CoverLetter string     `json:"cover_letter" gorm:"type:text"`
	AppliedAt   time.Time  `json:"applied_at"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 关联
	Job    Job            `json:"job,omitempty" gorm:"foreignKey:JobID"`
	User   User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Resume ResumeMetadata `json:"resume,omitempty" gorm:"foreignKey:ResumeID"`
}

// JobMatchingLog 职位匹配日志表
type JobMatchingLog struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"not null"`
	ResumeID       uint      `json:"resume_id" gorm:"not null"`
	MatchesCount   int       `json:"matches_count" gorm:"default:0"`
	FiltersApplied string    `json:"filters_applied" gorm:"type:json"`
	ProcessingTime int       `json:"processing_time"` // 毫秒
	CreatedAt      time.Time `json:"created_at"`

	// 关联
	User   User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Resume ResumeMetadata `json:"resume,omitempty" gorm:"foreignKey:ResumeID"`
}

// ==============================================
// 请求和响应模型
// ==============================================

// CreateJobRequest 创建职位请求
type CreateJobRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description" binding:"required"`
	Requirements string `json:"requirements"`
	CompanyID    uint   `json:"company_id" binding:"required"`
	Industry     string `json:"industry"`
	Location     string `json:"location"`
	SalaryMin    *int   `json:"salary_min"`
	SalaryMax    *int   `json:"salary_max"`
	Experience   string `json:"experience"`
	Education    string `json:"education"`
	JobType      string `json:"job_type"`
}

// UpdateJobRequest 更新职位请求
type UpdateJobRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	Requirements *string `json:"requirements"`
	Industry     *string `json:"industry"`
	Location     *string `json:"location"`
	SalaryMin    *int    `json:"salary_min"`
	SalaryMax    *int    `json:"salary_max"`
	Experience   *string `json:"experience"`
	Education    *string `json:"education"`
	JobType      *string `json:"job_type"`
	Status       *string `json:"status"`
}

// JobListRequest 职位列表请求
type JobListRequest struct {
	Page       int    `form:"page,default=1"`
	Size       int    `form:"size,default=20"`
	CompanyID  uint   `form:"company_id"`
	Industry   string `form:"industry"`
	Location   string `form:"location"`
	Experience string `form:"experience"`
	Education  string `form:"education"`
	JobType    string `form:"job_type"`
	SalaryMin  *int   `form:"salary_min"`
	SalaryMax  *int   `form:"salary_max"`
	Keyword    string `form:"keyword"`
	SortBy     string `form:"sort_by,default=created_at"`
	SortOrder  string `form:"sort_order,default=desc"`
}

// JobListResponse 职位列表响应
type JobListResponse struct {
	Jobs  []Job `json:"jobs"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// JobDetailResponse 职位详情响应
type JobDetailResponse struct {
	Job          Job              `json:"job"`
	Company      CompanyInfo      `json:"company"`
	Applications []JobApplication `json:"applications,omitempty"`
}

// ApplyJobRequest 申请职位请求
type ApplyJobRequest struct {
	ResumeID    uint   `json:"resume_id" binding:"required"`
	CoverLetter string `json:"cover_letter"`
}

// JobMatchingRequest 职位匹配请求
type JobMatchingRequest struct {
	ResumeID uint                   `json:"resume_id" binding:"required"`
	Limit    int                    `json:"limit,default=10"`
	Filters  map[string]interface{} `json:"filters"`
}

// JobMatchingResponse 职位匹配响应
type JobMatchingResponse struct {
	Matches        []JobMatchResult       `json:"matches"`
	Total          int                    `json:"total"`
	ResumeID       uint                   `json:"resume_id"`
	UserID         uint                   `json:"user_id"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
	Timestamp      string                 `json:"timestamp"`
}

// JobMatchResult 职位匹配结果
type JobMatchResult struct {
	JobID       uint               `json:"job_id"`
	MatchScore  float64            `json:"match_score"`
	Breakdown   map[string]float64 `json:"breakdown"`
	Confidence  float64            `json:"confidence"`
	JobInfo     Job                `json:"job_info"`
	CompanyInfo CompanyInfo        `json:"company_info"`
}

// ==============================================
// 工具函数
// ==============================================

// TableName 方法
func (Job) TableName() string {
	return "jobs"
}

func (JobApplication) TableName() string {
	return "job_applications"
}

func (JobMatchingLog) TableName() string {
	return "job_matching_logs"
}

// 职位状态常量
const (
	JobStatusActive   = "active"
	JobStatusInactive = "inactive"
	JobStatusClosed   = "closed"
	JobStatusDraft    = "draft"
)

// 申请状态常量
const (
	ApplicationStatusPending  = "pending"
	ApplicationStatusReviewed = "reviewed"
	ApplicationStatusAccepted = "accepted"
	ApplicationStatusRejected = "rejected"
)

// 工作类型常量
const (
	JobTypeFullTime   = "full-time"
	JobTypePartTime   = "part-time"
	JobTypeContract   = "contract"
	JobTypeInternship = "internship"
)

// 经验要求常量
const (
	ExperienceEntry    = "entry"
	ExperienceJunior   = "junior"
	ExperienceMid      = "mid"
	ExperienceSenior   = "senior"
	ExperienceLead     = "lead"
	ExperienceManager  = "manager"
	ExperienceDirector = "director"
)

// 学历要求常量
const (
	EducationHighSchool = "high-school"
	EducationCollege    = "college"
	EducationBachelor   = "bachelor"
	EducationMaster     = "master"
	EducationPhD        = "phd"
)

// 行业常量
const (
	IndustryTechnology = "technology"
	IndustryFinance    = "finance"
	IndustryHealthcare = "healthcare"
	IndustryEducation  = "education"
	IndustryMarketing  = "marketing"
	IndustrySales      = "sales"
	IndustryHR         = "hr"
	IndustryDesign     = "design"
	IndustryMedia      = "media"
	IndustryOther      = "other"
)
