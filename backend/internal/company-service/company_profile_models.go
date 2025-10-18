package main

import (
	"time"
)

// CompanyProfileBasicInfo 企业画像基本信息表
type CompanyProfileBasicInfo struct {
	ID                      uint       `json:"id" gorm:"primaryKey"`
	CompanyID               uint       `json:"company_id" gorm:"not null"`
	ReportID                string     `json:"report_id" gorm:"size:50;uniqueIndex"`
	CompanyName             string     `json:"company_name" gorm:"size:255;not null"`
	UsedName                string     `json:"used_name" gorm:"size:255"`
	UnifiedSocialCreditCode string     `json:"unified_social_credit_code" gorm:"size:50"`
	RegistrationDate        *time.Time `json:"registration_date"`
	LegalRepresentative     string     `json:"legal_representative" gorm:"size:100"`
	BusinessStatus          string     `json:"business_status" gorm:"size:50"`
	RegisteredCapital       float64    `json:"registered_capital" gorm:"type:decimal(18,2)"`
	Currency                string     `json:"currency" gorm:"size:20;default:CNY"`
	InsuredCount            int        `json:"insured_count"`
	IndustryCategory        string     `json:"industry_category" gorm:"size:100"`
	RegistrationAuthority   string     `json:"registration_authority" gorm:"size:255"`
	BusinessScope           string     `json:"business_scope" gorm:"type:text"`
	Tags                    string     `json:"tags" gorm:"type:json"` // JSON数组
	DataSource              string     `json:"data_source" gorm:"size:100"`
	DataUpdateTime          time.Time  `json:"data_update_time" gorm:"autoUpdateTime"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// QualificationLicense 资质许可表
type QualificationLicense struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	CompanyID         uint       `json:"company_id" gorm:"not null"`
	ReportID          string     `json:"report_id" gorm:"size:50"`
	Type              string     `json:"type" gorm:"type:enum('资质','许可','备案');not null"`
	Name              string     `json:"name" gorm:"size:255;not null"`
	Status            string     `json:"status" gorm:"size:20;default:有效"`
	CertificateNumber string     `json:"certificate_number" gorm:"size:100"`
	IssueDate         *time.Time `json:"issue_date"`
	IssuingAuthority  string     `json:"issuing_authority" gorm:"size:255"`
	ValidityPeriod    *time.Time `json:"validity_period"`
	Content           string     `json:"content" gorm:"type:text"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// PersonnelCompetitiveness 人员竞争力表
type PersonnelCompetitiveness struct {
	ID                   uint       `json:"id" gorm:"primaryKey"`
	CompanyID            uint       `json:"company_id" gorm:"not null"`
	ReportID             string     `json:"report_id" gorm:"size:50"`
	DataUpdateDate       *time.Time `json:"data_update_date"`
	TotalEmployees       int        `json:"total_employees"`
	IndustryRanking      string     `json:"industry_ranking" gorm:"size:50"`
	IndustryAvgEmployees int        `json:"industry_avg_employees"`
	TurnoverRate         float64    `json:"turnover_rate" gorm:"type:decimal(5,2)"`
	EntryRate            float64    `json:"entry_rate" gorm:"type:decimal(5,2)"`
	TenureDistribution   string     `json:"tenure_distribution" gorm:"type:json"` // JSON格式
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// ProvidentFund 公积金信息表
type ProvidentFund struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	CompanyID        uint       `json:"company_id" gorm:"not null"`
	ReportID         string     `json:"report_id" gorm:"size:50"`
	UnitNature       string     `json:"unit_nature" gorm:"size:100"`
	OpeningDate      *time.Time `json:"opening_date"`
	LastPaymentMonth *time.Time `json:"last_payment_month"`
	TotalPayment     string     `json:"total_payment" gorm:"size:50"`
	PaymentRecords   string     `json:"payment_records" gorm:"type:json"` // JSON数组
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// SubsidyInfo 资助补贴表
type SubsidyInfo struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CompanyID   uint      `json:"company_id" gorm:"not null"`
	ReportID    string    `json:"report_id" gorm:"size:50"`
	SubsidyYear int       `json:"subsidy_year"`
	Amount      float64   `json:"amount" gorm:"type:decimal(18,2)"`
	Count       int       `json:"count"`
	Source      string    `json:"source" gorm:"size:255"`
	SubsidyList string    `json:"subsidy_list" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CompanyRelationship 企业关系图谱表
type CompanyRelationship struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	CompanyID          uint      `json:"company_id" gorm:"not null"`
	ReportID           string    `json:"report_id" gorm:"size:50"`
	RelatedCompanyName string    `json:"related_company_name" gorm:"size:255"`
	RelationshipType   string    `json:"relationship_type" gorm:"type:enum('投资','任职','合作','控股','参股');not null"`
	InvestmentAmount   float64   `json:"investment_amount" gorm:"type:decimal(18,2)"`
	InvestmentRatio    float64   `json:"investment_ratio" gorm:"type:decimal(5,2)"`
	Position           string    `json:"position" gorm:"size:100"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TechInnovationScore 科创评分表
type TechInnovationScore struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	CompanyID            uint      `json:"company_id" gorm:"not null"`
	ReportID             string    `json:"report_id" gorm:"size:50"`
	BasicScore           float64   `json:"basic_score" gorm:"type:decimal(5,2)"`
	TalentScore          float64   `json:"talent_score" gorm:"type:decimal(5,2)"`
	IndustryRanking      string    `json:"industry_ranking" gorm:"size:50"`
	StrategicIndustry    string    `json:"strategic_industry" gorm:"size:100"`
	IntellectualProperty string    `json:"intellectual_property" gorm:"type:json"` // JSON格式
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CompanyProfileFinancialInfo 企业画像财务信息表
type CompanyProfileFinancialInfo struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	CompanyID        uint      `json:"company_id" gorm:"not null"`
	ReportID         string    `json:"report_id" gorm:"size:50"`
	AnnualRevenue    float64   `json:"annual_revenue" gorm:"type:decimal(18,2)"`
	NetProfit        float64   `json:"net_profit" gorm:"type:decimal(18,2)"`
	TotalAssets      float64   `json:"total_assets" gorm:"type:decimal(18,2)"`
	TotalLiabilities float64   `json:"total_liabilities" gorm:"type:decimal(18,2)"`
	Equity           float64   `json:"equity" gorm:"type:decimal(18,2)"`
	CashFlow         float64   `json:"cash_flow" gorm:"type:decimal(18,2)"`
	ROE              float64   `json:"roe" gorm:"type:decimal(5,2)"`
	ROA              float64   `json:"roa" gorm:"type:decimal(5,2)"`
	DebtRatio        float64   `json:"debt_ratio" gorm:"type:decimal(5,2)"`
	CurrentRatio     float64   `json:"current_ratio" gorm:"type:decimal(5,2)"`
	QuickRatio       float64   `json:"quick_ratio" gorm:"type:decimal(5,2)"`
	FinancingStatus  string    `json:"financing_status" gorm:"size:100"`
	ListingStatus    string    `json:"listing_status" gorm:"size:50"`
	FinancialYear    int       `json:"financial_year"`
	DataSource       string    `json:"data_source" gorm:"size:100"`
	DataUpdateTime   time.Time `json:"data_update_time" gorm:"autoUpdateTime"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CompanyRiskInfo 企业风险信息表
type CompanyRiskInfo struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	CompanyID        uint      `json:"company_id" gorm:"not null"`
	ReportID         string    `json:"report_id" gorm:"size:50"`
	RiskLevel        string    `json:"risk_level" gorm:"type:enum('低风险','中风险','高风险');default:低风险"`
	LegalRisks       string    `json:"legal_risks" gorm:"type:json"`       // JSON格式
	FinancialRisks   string    `json:"financial_risks" gorm:"type:json"`   // JSON格式
	OperationalRisks string    `json:"operational_risks" gorm:"type:json"` // JSON格式
	CreditRating     string    `json:"credit_rating" gorm:"size:20"`
	RiskFactors      string    `json:"risk_factors" gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CompanyProfileData 企业画像数据聚合结构
type CompanyProfileData struct {
	BasicInfo      *CompanyProfileBasicInfo     `json:"basic_info,omitempty"`
	Qualifications []QualificationLicense       `json:"qualifications,omitempty"`
	Personnel      *PersonnelCompetitiveness    `json:"personnel,omitempty"`
	ProvidentFund  *ProvidentFund               `json:"provident_fund,omitempty"`
	Subsidies      []SubsidyInfo                `json:"subsidies,omitempty"`
	Relationships  []CompanyRelationship        `json:"relationships,omitempty"`
	TechInnovation *TechInnovationScore         `json:"tech_innovation,omitempty"`
	FinancialInfo  *CompanyProfileFinancialInfo `json:"financial_info,omitempty"`
	RiskInfo       *CompanyProfileRiskInfo      `json:"risk_info,omitempty"`
}

// CompanyProfileSummary 企业画像摘要信息

// TableName 方法定义表名
func (CompanyBasicInfo) TableName() string {
	return "company_basic_info"
}

func (QualificationLicense) TableName() string {
	return "qualification_license"
}

func (PersonnelCompetitiveness) TableName() string {
	return "personnel_competitiveness"
}

func (ProvidentFund) TableName() string {
	return "provident_fund"
}

func (SubsidyInfo) TableName() string {
	return "subsidy_info"
}

func (CompanyRelationship) TableName() string {
	return "company_relationships"
}

func (TechInnovationScore) TableName() string {
	return "tech_innovation_score"
}

func (CompanyFinancialInfo) TableName() string {
	return "company_financial_info"
}

func (CompanyRiskInfo) TableName() string {
	return "company_risk_info"
}

// CompanyProfileRiskInfo 企业风险信息表
type CompanyProfileRiskInfo struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	CompanyID        uint      `json:"company_id" gorm:"not null"`
	ReportID         string    `json:"report_id" gorm:"size:50"`
	RiskLevel        string    `json:"risk_level" gorm:"size:20;default:低风险"`
	RiskFactors      string    `json:"risk_factors" gorm:"type:json"` // JSON格式
	CreditRating     string    `json:"credit_rating" gorm:"size:20"`
	LegalDisputes    string    `json:"legal_disputes" gorm:"type:json"` // JSON格式
	FinancialHealth  string    `json:"financial_health" gorm:"size:20"`
	OperationalRisk  string    `json:"operational_risk" gorm:"size:20"`
	MarketRisk       string    `json:"market_risk" gorm:"size:20"`
	ComplianceStatus string    `json:"compliance_status" gorm:"size:20"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (CompanyProfileRiskInfo) TableName() string {
	return "company_profile_risk_info"
}
