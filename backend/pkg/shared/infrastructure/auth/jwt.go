package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string        `json:"secret_key"`
	Issuer        string        `json:"issuer"`
	Audience      string        `json:"audience"`
	ExpireTime    time.Duration `json:"expire_time"`
	RefreshTime   time.Duration `json:"refresh_time"`
	RefreshSecret string        `json:"refresh_secret"`
}

// Claims JWT声明
type Claims struct {
	UserID   string            `json:"user_id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Roles    []string          `json:"roles"`
	Metadata map[string]string `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证管理器
type JWTAuth struct {
	config *JWTConfig
}

// NewJWTAuth 创建JWT认证管理器
func NewJWTAuth(config *JWTConfig) *JWTAuth {
	if config.SecretKey == "" {
		config.SecretKey = generateSecretKey()
	}
	if config.RefreshSecret == "" {
		config.RefreshSecret = generateSecretKey()
	}
	if config.ExpireTime == 0 {
		config.ExpireTime = 24 * time.Hour // 默认24小时
	}
	if config.RefreshTime == 0 {
		config.RefreshTime = 7 * 24 * time.Hour // 默认7天
	}

	return &JWTAuth{
		config: config,
	}
}

// GenerateToken 生成访问token
func (j *JWTAuth) GenerateToken(userID, username, email string, roles []string, metadata map[string]string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		Metadata: metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Audience:  []string{j.config.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// GenerateRefreshToken 生成刷新token
func (j *JWTAuth) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Audience:  []string{j.config.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.RefreshSecret))
}

// ValidateToken 验证访问token
func (j *JWTAuth) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken 验证刷新token
func (j *JWTAuth) ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.RefreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

// RefreshToken 刷新token
func (j *JWTAuth) RefreshToken(refreshToken string, username, email string, roles []string, metadata map[string]string) (string, string, error) {
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	// 生成新的访问token
	accessToken, err := j.GenerateToken(claims.UserID, username, email, roles, metadata)
	if err != nil {
		return "", "", err
	}

	// 生成新的刷新token
	newRefreshToken, err := j.GenerateRefreshToken(claims.UserID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// HasRole 检查用户是否有指定角色
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole 检查用户是否有任意指定角色
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return true
}

// HasAllRoles 检查用户是否有所有指定角色
func (c *Claims) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}

// generateSecretKey 生成随机密钥
func generateSecretKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}
