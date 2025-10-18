package main

import (
	"encoding/json"
	"time"
)

// EnhancedCompany 增强的企业模型 - 支持认证机制和地理位置
type EnhancedCompany struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:200;not null"`
	ShortName   string `json:"short_name" gorm:"size:100"`
	LogoURL     string `json:"logo_url" gorm:"size:500"`
	Industry    string `json:"industry" gorm:"size:100"`
	CompanySize string `json:"company_size" gorm:"size:50"`
	Location    string `json:"location" gorm:"size:200"`
	Website     string `json:"website" gorm:"size:200"`
	Description string `json:"description" gorm:"type:text"`
	FoundedYear int    `json:"founded_year"`

	// 企业认证信息
	UnifiedSocialCreditCode string `json:"unified_social_credit_code" gorm:"size:50;uniqueIndex"`
	LegalRepresentative     string `json:"legal_representative" gorm:"size:100"`
	LegalRepresentativeID   string `json:"legal_representative_id" gorm:"size:50"` // 身份证号

	// 权限管理字段
	CreatedBy       uint   `json:"created_by" gorm:"not null"`        // 创建者
	LegalRepUserID  uint   `json:"legal_rep_user_id"`                 // 法定代表人用户ID
	AuthorizedUsers string `json:"authorized_users" gorm:"type:json"` // 授权用户列表

	// 北斗地理位置信息
	BDLatitude  *float64 `json:"bd_latitude" gorm:"type:decimal(10,8)"`  // 北斗纬度
	BDLongitude *float64 `json:"bd_longitude" gorm:"type:decimal(11,8)"` // 北斗经度
	BDAltitude  *float64 `json:"bd_altitude" gorm:"type:decimal(8,2)"`   // 北斗海拔
	BDAccuracy  *float64 `json:"bd_accuracy" gorm:"type:decimal(6,2)"`   // 定位精度(米)
	BDTimestamp *int64   `json:"bd_timestamp"`                           // 定位时间戳

	// 解析后的地址信息
	Address    string `json:"address" gorm:"size:500"`    // 详细地址
	City       string `json:"city" gorm:"size:100"`       // 城市
	District   string `json:"district" gorm:"size:100"`   // 区县
	Area       string `json:"area" gorm:"size:100"`       // 区域/街道
	PostalCode string `json:"postal_code" gorm:"size:20"` // 邮政编码

	// 地理位置层级编码
	CityCode     string `json:"city_code" gorm:"size:20"`     // 城市编码
	DistrictCode string `json:"district_code" gorm:"size:20"` // 区县编码
	AreaCode     string `json:"area_code" gorm:"size:20"`     // 区域编码

	Status            string    `json:"status" gorm:"size:20;default:pending"`
	VerificationLevel string    `json:"verification_level" gorm:"size:20;default:unverified"`
	JobCount          int       `json:"job_count" gorm:"default:0"`
	ViewCount         int       `json:"view_count" gorm:"default:0"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// 关联数据
	CompanyUsers []CompanyUser `json:"company_users,omitempty" gorm:"foreignKey:CompanyID"`
}

// CompanyUser 企业用户关联模型
type CompanyUser struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CompanyID   uint      `json:"company_id" gorm:"not null"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	Role        string    `json:"role" gorm:"size:50;not null"`         // legal_rep, authorized_user, admin
	Status      string    `json:"status" gorm:"size:20;default:active"` // active, inactive, pending
	Permissions string    `json:"permissions" gorm:"type:json"`         // 权限列表
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 外键关联
	Company EnhancedCompany `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	User    User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// CompanyPermissionAuditLog 企业权限审计日志
type CompanyPermissionAuditLog struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	CompanyID        uint      `json:"company_id" gorm:"not null"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	Action           string    `json:"action" gorm:"size:100;not null"`       // 操作类型
	ResourceType     string    `json:"resource_type" gorm:"size:50;not null"` // 资源类型
	ResourceID       *uint     `json:"resource_id"`                           // 资源ID
	PermissionResult bool      `json:"permission_result" gorm:"not null"`     // 权限检查结果
	IPAddress        string    `json:"ip_address" gorm:"size:45"`             // IP地址
	UserAgent        string    `json:"user_agent" gorm:"type:text"`           // 用户代理
	CreatedAt        time.Time `json:"created_at"`

	// 外键关联
	Company EnhancedCompany `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	User    User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// CompanyDataSyncStatus 企业数据同步状态
type CompanyDataSyncStatus struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	CompanyID    uint       `json:"company_id" gorm:"not null"`
	SyncTarget   string     `json:"sync_target" gorm:"size:50;not null"`        // postgresql, neo4j, redis
	SyncStatus   string     `json:"sync_status" gorm:"size:20;default:pending"` // pending, syncing, success, failed
	LastSyncTime *time.Time `json:"last_sync_time"`                             // 最后同步时间
	SyncError    string     `json:"sync_error" gorm:"type:text"`                // 同步错误信息
	RetryCount   int        `json:"retry_count" gorm:"default:0"`               // 重试次数
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 外键关联
	Company EnhancedCompany `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

// User 用户模型（简化版，实际应该从用户服务获取）
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"size:100;not null"`
	Email    string `json:"email" gorm:"size:200;not null"`
	Role     string `json:"role" gorm:"size:50;default:user"`
}

// CompanyPermissionLevel 企业权限级别
type CompanyPermissionLevel string

const (
	PermissionSystemAdmin         CompanyPermissionLevel = "system_admin"
	PermissionCompanyOwner        CompanyPermissionLevel = "company_owner"
	PermissionLegalRepresentative CompanyPermissionLevel = "legal_representative"
	PermissionAuthorizedUser      CompanyPermissionLevel = "authorized_user"
	PermissionNoAccess            CompanyPermissionLevel = "no_access"
)

// CompanyRole 企业角色
type CompanyRole string

const (
	RoleLegalRepresentative CompanyRole = "legal_rep"
	RoleAuthorizedUser      CompanyRole = "authorized_user"
	RoleAdmin               CompanyRole = "admin"
)

// CompanyStatus 企业状态
type CompanyStatus string

const (
	StatusPending  CompanyStatus = "pending"
	StatusActive   CompanyStatus = "active"
	StatusInactive CompanyStatus = "inactive"
	StatusRejected CompanyStatus = "rejected"
)

// VerificationLevel 验证级别
type VerificationLevel string

const (
	VerificationUnverified VerificationLevel = "unverified"
	VerificationVerified   VerificationLevel = "verified"
	VerificationPremium    VerificationLevel = "premium"
)

// SyncTarget 同步目标
type SyncTarget string

const (
	SyncTargetPostgreSQL SyncTarget = "postgresql"
	SyncTargetNeo4j      SyncTarget = "neo4j"
	SyncTargetRedis      SyncTarget = "redis"
)

// SyncStatus 同步状态
type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSyncing SyncStatus = "syncing"
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
)

// CompanyLocationInfo 企业地理位置信息
type CompanyLocationInfo struct {
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Altitude     float64 `json:"altitude"`
	Accuracy     float64 `json:"accuracy"`
	Timestamp    int64   `json:"timestamp"`
	Address      string  `json:"address"`
	City         string  `json:"city"`
	District     string  `json:"district"`
	Area         string  `json:"area"`
	PostalCode   string  `json:"postal_code"`
	CityCode     string  `json:"city_code"`
	DistrictCode string  `json:"district_code"`
	AreaCode     string  `json:"area_code"`
}

// CompanyAuthInfo 企业认证信息
type CompanyAuthInfo struct {
	UnifiedSocialCreditCode string `json:"unified_social_credit_code"`
	LegalRepresentative     string `json:"legal_representative"`
	LegalRepresentativeID   string `json:"legal_representative_id"`
	LegalRepUserID          uint   `json:"legal_rep_user_id"`
	AuthorizedUsers         []uint `json:"authorized_users"`
}

// CompanyPermissionInfo 企业权限信息
type CompanyPermissionInfo struct {
	UserID                   uint                   `json:"user_id"`
	CompanyID                uint                   `json:"company_id"`
	EffectivePermissionLevel CompanyPermissionLevel `json:"effective_permission_level"`
	Role                     string                 `json:"role"`
	Status                   string                 `json:"status"`
	Permissions              []string               `json:"permissions"`
}

// CompanySyncInfo 企业同步信息
type CompanySyncInfo struct {
	CompanyID    uint       `json:"company_id"`
	SyncTarget   SyncTarget `json:"sync_target"`
	SyncStatus   SyncStatus `json:"sync_status"`
	LastSyncTime *time.Time `json:"last_sync_time"`
	SyncError    string     `json:"sync_error"`
	RetryCount   int        `json:"retry_count"`
}

// 方法定义

// GetLocationInfo 获取企业地理位置信息
func (c *EnhancedCompany) GetLocationInfo() *CompanyLocationInfo {
	if c.BDLatitude == nil || c.BDLongitude == nil {
		return nil
	}

	return &CompanyLocationInfo{
		Latitude:     *c.BDLatitude,
		Longitude:    *c.BDLongitude,
		Altitude:     getFloat64Value(c.BDAltitude),
		Accuracy:     getFloat64Value(c.BDAccuracy),
		Timestamp:    getInt64Value(c.BDTimestamp),
		Address:      c.Address,
		City:         c.City,
		District:     c.District,
		Area:         c.Area,
		PostalCode:   c.PostalCode,
		CityCode:     c.CityCode,
		DistrictCode: c.DistrictCode,
		AreaCode:     c.AreaCode,
	}
}

// SetLocationInfo 设置企业地理位置信息
func (c *EnhancedCompany) SetLocationInfo(location *CompanyLocationInfo) {
	if location == nil {
		return
	}

	c.BDLatitude = &location.Latitude
	c.BDLongitude = &location.Longitude
	c.BDAltitude = &location.Altitude
	c.BDAccuracy = &location.Accuracy
	c.BDTimestamp = &location.Timestamp
	c.Address = location.Address
	c.City = location.City
	c.District = location.District
	c.Area = location.Area
	c.PostalCode = location.PostalCode
	c.CityCode = location.CityCode
	c.DistrictCode = location.DistrictCode
	c.AreaCode = location.AreaCode
}

// GetAuthInfo 获取企业认证信息
func (c *EnhancedCompany) GetAuthInfo() *CompanyAuthInfo {
	var authorizedUsers []uint
	if c.AuthorizedUsers != "" {
		json.Unmarshal([]byte(c.AuthorizedUsers), &authorizedUsers)
	}

	return &CompanyAuthInfo{
		UnifiedSocialCreditCode: c.UnifiedSocialCreditCode,
		LegalRepresentative:     c.LegalRepresentative,
		LegalRepresentativeID:   c.LegalRepresentativeID,
		LegalRepUserID:          c.LegalRepUserID,
		AuthorizedUsers:         authorizedUsers,
	}
}

// SetAuthInfo 设置企业认证信息
func (c *EnhancedCompany) SetAuthInfo(auth *CompanyAuthInfo) {
	if auth == nil {
		return
	}

	c.UnifiedSocialCreditCode = auth.UnifiedSocialCreditCode
	c.LegalRepresentative = auth.LegalRepresentative
	c.LegalRepresentativeID = auth.LegalRepresentativeID
	c.LegalRepUserID = auth.LegalRepUserID

	if len(auth.AuthorizedUsers) > 0 {
		authorizedUsersJSON, _ := json.Marshal(auth.AuthorizedUsers)
		c.AuthorizedUsers = string(authorizedUsersJSON)
	}
}

// GetPermissions 获取用户权限列表
func (cu *CompanyUser) GetPermissions() []string {
	var permissions []string
	if cu.Permissions != "" {
		json.Unmarshal([]byte(cu.Permissions), &permissions)
	}
	return permissions
}

// SetPermissions 设置用户权限列表
func (cu *CompanyUser) SetPermissions(permissions []string) {
	if len(permissions) > 0 {
		permissionsJSON, _ := json.Marshal(permissions)
		cu.Permissions = string(permissionsJSON)
	}
}

// 辅助函数（移动到company_data_sync_service.go中）

// TableName 方法定义表名
func (EnhancedCompany) TableName() string {
	return "companies"
}

func (CompanyUser) TableName() string {
	return "company_users"
}

func (CompanyPermissionAuditLog) TableName() string {
	return "company_permission_audit_logs"
}

func (CompanyDataSyncStatus) TableName() string {
	return "company_data_sync_status"
}

func (User) TableName() string {
	return "users"
}
