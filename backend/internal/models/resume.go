package models

import (
	"time"

	"gorm.io/gorm"
)

// ResumeV3 V3.0简历模型
type ResumeV3 struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UUID        string         `json:"uuid" gorm:"size:36;unique;not null"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Slug        string         `json:"slug" gorm:"size:255;unique"`
	Summary     string         `json:"summary" gorm:"type:text"`
	Content     string         `json:"content" gorm:"type:longtext"`
	Status      string         `json:"status" gorm:"size:20;default:'draft'"`       // draft, published, archived
	Visibility  string         `json:"visibility" gorm:"size:20;default:'private'"` // private, public, unlisted
	CanComment  bool           `json:"can_comment" gorm:"default:true"`
	TemplateID  uint           `json:"template_id"`
	Template    Template       `json:"template" gorm:"foreignKey:TemplateID"`
	Skills      []Skill        `json:"skills" gorm:"many2many:resume_skills;"`
	Experiences []Experience   `json:"experiences" gorm:"foreignKey:ResumeID"`
	Educations  []Education    `json:"educations" gorm:"foreignKey:ResumeID"`
	Projects    []Project      `json:"projects" gorm:"foreignKey:ResumeID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Template 简历模板模型
type Template struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Category    string         `json:"category" gorm:"size:100"`
	PreviewURL  string         `json:"preview_url" gorm:"size:500"`
	TemplateURL string         `json:"template_url" gorm:"size:500"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Skill 技能模型
type Skill struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null;unique"`
	Category    string         `json:"category" gorm:"size:50"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Experience 工作经验模型
type Experience struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ResumeID    uint           `json:"resume_id"`
	CompanyID   uint           `json:"company_id"`
	Company     Company        `json:"company" gorm:"foreignKey:CompanyID"`
	PositionID  uint           `json:"position_id"`
	Position    Position       `json:"position" gorm:"foreignKey:PositionID"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	IsCurrent   bool           `json:"is_current" gorm:"default:false"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Education 教育经历模型
type Education struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ResumeID    uint           `json:"resume_id"`
	School      string         `json:"school" gorm:"size:255;not null"`
	Degree      string         `json:"degree" gorm:"size:100"`
	Major       string         `json:"major" gorm:"size:100"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	IsCurrent   bool           `json:"is_current" gorm:"default:false"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Project 项目经历模型
type Project struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ResumeID     uint           `json:"resume_id"`
	Name         string         `json:"name" gorm:"size:255;not null"`
	Description  string         `json:"description" gorm:"type:text"`
	StartDate    time.Time      `json:"start_date"`
	EndDate      *time.Time     `json:"end_date"`
	IsCurrent    bool           `json:"is_current" gorm:"default:false"`
	Technologies string         `json:"technologies" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// Position 职位模型
type Position struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Category    string         `json:"category" gorm:"size:100"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// 响应模型
type ResumeListResponse struct {
	Resumes []ResumeV3 `json:"resumes"`
	Total   int64      `json:"total"`
	Page    int        `json:"page"`
	Size    int        `json:"size"`
}

type SkillListResponse struct {
	Skills []Skill `json:"skills"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Size   int     `json:"size"`
}

type CompanyListResponse struct {
	Companies []Company `json:"companies"`
	Total     int64     `json:"total"`
	Page      int       `json:"page"`
	Size      int       `json:"size"`
}

type PositionListResponse struct {
	Positions []Position `json:"positions"`
	Total     int64      `json:"total"`
	Page      int        `json:"page"`
	Size      int        `json:"size"`
}

// 请求模型
type CreateResumeRequest struct {
	Title      string `json:"title" binding:"required"`
	Summary    string `json:"summary"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	TemplateID uint   `json:"template_id"`
	SkillIDs   []uint `json:"skill_ids"`
}

// TableName 方法
func (ResumeV3) TableName() string {
	return "resumes_v3"
}

func (Template) TableName() string {
	return "templates"
}

func (Skill) TableName() string {
	return "skills"
}

func (Experience) TableName() string {
	return "experiences"
}

func (Education) TableName() string {
	return "educations"
}

func (Project) TableName() string {
	return "projects"
}

func (Position) TableName() string {
	return "positions"
}
