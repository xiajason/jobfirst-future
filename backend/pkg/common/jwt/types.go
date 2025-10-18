package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TenantType 租户类型
type TenantType string

const (
	TenantAdmin      TenantType = "ADMIN"      // 管理员
	TenantPersonal   TenantType = "PERSONAL"   // 个人用户
	TenantEnterprise TenantType = "ENTERPRISE" // 企业用户
)

// JWTClaims JWT令牌主体结构
type JWTClaims struct {
	UserID      int64      `json:"user_id"`     // 用户ID
	Username    string     `json:"username"`    // 用户名
	TenantType  TenantType `json:"tenant_type"` // 租户类型
	Role        string     `json:"role"`        // 角色
	Permissions []string   `json:"permissions"` // 权限列表
	Token       string     `json:"token"`       // 令牌标识
	jwt.RegisteredClaims
}

// TokenInfo 令牌信息
type TokenInfo struct {
	AccessToken  string     `json:"access_token"`  // 访问令牌
	RefreshToken string     `json:"refresh_token"` // 刷新令牌
	TokenType    string     `json:"token_type"`    // 令牌类型
	ExpiresIn    int64      `json:"expires_in"`    // 过期时间（秒）
	ExpiresAt    time.Time  `json:"expires_at"`    // 过期时间点
	UserID       int64      `json:"user_id"`       // 用户ID
	Username     string     `json:"username"`      // 用户名
	TenantType   TenantType `json:"tenant_type"`   // 租户类型
	Role         string     `json:"role"`          // 角色
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string        `json:"secret_key"`     // 密钥
	TokenExpiry   time.Duration `json:"token_expiry"`   // 访问令牌过期时间
	RefreshExpiry time.Duration `json:"refresh_expiry"` // 刷新令牌过期时间
	Issuer        string        `json:"issuer"`         // 签发者
	Audience      string        `json:"audience"`       // 受众
}

// DefaultJWTConfig 默认JWT配置
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:     "jobfirst-jwt-secret-key",
		TokenExpiry:   24 * time.Hour,     // 访问令牌24小时过期
		RefreshExpiry: 7 * 24 * time.Hour, // 刷新令牌7天过期
		Issuer:        "jobfirst",
		Audience:      "jobfirst-users",
	}
}

// UserInfo 用户信息
type UserInfo struct {
	UserID      int64      `json:"user_id"`
	Username    string     `json:"username"`
	TenantType  TenantType `json:"tenant_type"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
}

// TokenRequest 令牌请求
type TokenRequest struct {
	UserID      int64      `json:"user_id"`
	Username    string     `json:"username"`
	TenantType  TenantType `json:"tenant_type"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}
