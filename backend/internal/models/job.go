package models

import (
	"time"

	"gorm.io/gorm"
)

// Company 公司模型
type Company struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"size:255;not null"`
	ShortName string `json:"short_name" gorm:"size:100"`
	LogoURL   string `json:"logo_url" gorm:"size:500"`
}

// Job 职位模型
type Job struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Title              string         `json:"title" gorm:"size:255;not null"`
	CompanyID          uint           `json:"company_id"`
	Company            Company        `json:"company" gorm:"foreignKey:CompanyID"`
	Location           string         `json:"location" gorm:"size:255"`
	Description        string         `json:"description" gorm:"type:text"`
	Requirements       string         `json:"requirements" gorm:"type:text"`
	SalaryMin          int            `json:"salary_min"`
	SalaryMax          int            `json:"salary_max"`
	SalaryType         string         `json:"salary_type" gorm:"size:50"` // monthly, yearly, hourly
	ExperienceRequired string         `json:"experience_required" gorm:"size:100"`
	EducationRequired  string         `json:"education_required" gorm:"size:100"`
	JobType            string         `json:"job_type" gorm:"size:50"` // full-time, part-time, contract
	Benefits           string         `json:"benefits" gorm:"type:text"`
	Skills             string         `json:"skills" gorm:"type:text"`
	Tags               string         `json:"tags" gorm:"type:text"`
	Status             string         `json:"status" gorm:"size:20;default:'active'"` // active, inactive, closed
	ViewCount          int            `json:"view_count" gorm:"default:0"`
	ApplicationCount   int            `json:"application_count" gorm:"default:0"`
	FavoriteCount      int            `json:"favorite_count" gorm:"default:0"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Job) TableName() string {
	return "jobs"
}
