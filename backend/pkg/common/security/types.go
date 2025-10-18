package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Token       string   `json:"token"`
	jwt.RegisteredClaims
}

// UserContext 用户上下文
type UserContext struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// TokenInfo 令牌信息
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	SecretKey     string        `json:"secret_key"`
	TokenExpiry   time.Duration `json:"token_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
	Whitelist     []string      `json:"whitelist"`
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		SecretKey:     "jobfirst-secret-key",
		TokenExpiry:   24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
		Whitelist: []string{
			"/health",
			"/version",
			"/v2/api-docs",
			"/swagger/",
			"/metrics",
			"/utils/",
			"/monitor/",
			"/config/",
			"POST:/api/v1/user/auth/login",
			"POST:/api/v1/user/auth/register",
			"GET:/api/v1/user/public/",
		},
	}
}
