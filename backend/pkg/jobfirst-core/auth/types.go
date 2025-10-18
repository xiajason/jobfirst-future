package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User 用户基础信息（与现有users表兼容）
type User struct {
	ID                    uint       `json:"id" gorm:"primaryKey"`
	Username              string     `json:"username" gorm:"type:varchar(100);uniqueIndex"`
	Email                 string     `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	PasswordHash          string     `json:"-" gorm:"column:password_hash;type:varchar(255)"`
	Role                  string     `json:"role" gorm:"type:enum('super_admin','system_admin','dev_lead','frontend_dev','backend_dev','qa_engineer','guest');default:guest"`
	Status                string     `json:"status" gorm:"type:enum('active','inactive','suspended');default:active"`
	CreatedAt             time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	LastLogin             *time.Time `json:"last_login"`
	UUID                  string     `json:"uuid" gorm:"type:varchar(36);uniqueIndex"`
	FirstName             string     `json:"first_name" gorm:"type:varchar(100)"`
	LastName              string     `json:"last_name" gorm:"type:varchar(100)"`
	Phone                 string     `json:"phone" gorm:"type:varchar(20)"`
	AvatarURL             string     `json:"avatar_url" gorm:"type:varchar(500)"`
	EmailVerified         bool       `json:"email_verified" gorm:"default:false"`
	PhoneVerified         bool       `json:"phone_verified" gorm:"default:false"`
	LastLoginAt           *time.Time `json:"last_login_at"`
	DeletedAt             *time.Time `json:"deleted_at" gorm:"index"`
	SubscriptionStatus    string     `json:"subscription_status" gorm:"type:enum('free','trial','premium','enterprise');default:free"`
	SubscriptionType      *string    `json:"subscription_type" gorm:"type:enum('monthly','yearly','lifetime')"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at"`
	SubscriptionFeatures  *string    `json:"subscription_features" gorm:"type:json"`
}

// DevTeamUser 开发团队成员
type DevTeamUser struct {
	ID                        uint       `json:"id" gorm:"primaryKey"`
	UserID                    uint       `json:"user_id" gorm:"uniqueIndex"`
	TeamRole                  string     `json:"team_role" gorm:"type:enum('super_admin','system_admin','dev_lead','frontend_dev','backend_dev','qa_engineer','guest');default:guest"`
	SSHPublicKey              string     `json:"ssh_public_key" gorm:"type:text"`
	ServerAccessLevel         string     `json:"server_access_level" gorm:"type:enum('full','limited','readonly','none');default:limited"`
	CodeAccessModules         string     `json:"code_access_modules" gorm:"type:json"`
	DatabaseAccess            string     `json:"database_access" gorm:"type:json"`
	ServiceRestartPermissions string     `json:"service_restart_permissions" gorm:"type:json"`
	LastLoginAt               *time.Time `json:"last_login_at"`
	Status                    string     `json:"status" gorm:"type:enum('active','inactive','suspended');default:active"`
	CreatedAt                 time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt                 *time.Time `json:"deleted_at" gorm:"index"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// DevOperationLog 开发操作日志
type DevOperationLog struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id"`
	OperationType    string    `json:"operation_type" gorm:"type:varchar(100)"`
	OperationTarget  string    `json:"operation_target" gorm:"type:varchar(255)"`
	OperationDetails string    `json:"operation_details" gorm:"type:json"`
	IPAddress        string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent        string    `json:"user_agent" gorm:"type:text"`
	Status           string    `json:"status" gorm:"type:enum('success','failed','blocked');default:success"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// JWT Claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success   bool        `json:"success"`
	Token     string      `json:"token"`
	User      User        `json:"user"`
	DevTeam   DevTeamUser `json:"dev_team,omitempty"`
	ExpiresAt string      `json:"expires_at"`
	Message   string      `json:"message"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    User   `json:"user,omitempty"`
}

// Role 角色定义
type Role struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description"`
}

// Permission 权限定义
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Scope    string `json:"scope"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret        string        `json:"jwt_secret"`
	TokenExpiry      time.Duration `json:"token_expiry"`
	RefreshExpiry    time.Duration `json:"refresh_expiry"`
	PasswordMin      int           `json:"password_min_length"`
	MaxLoginAttempts int           `json:"max_login_attempts"`
	LockoutDuration  time.Duration `json:"lockout_duration"`
}
