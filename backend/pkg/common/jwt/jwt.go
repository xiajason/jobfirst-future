package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager JWT管理器
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(config *JWTConfig) *JWTManager {
	if config == nil {
		config = DefaultJWTConfig()
	}
	return &JWTManager{
		config: config,
	}
}

// GenerateToken 生成令牌标识
func (j *JWTManager) GenerateToken(prefix string) string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	return prefix + "-" + hex.EncodeToString(randomBytes)
}

// CreateAccessToken 创建访问令牌
func (j *JWTManager) CreateAccessToken(req *TokenRequest) (string, error) {
	claims := &JWTClaims{
		UserID:      req.UserID,
		Username:    req.Username,
		TenantType:  req.TenantType,
		Role:        req.Role,
		Permissions: req.Permissions,
		Token:       j.GenerateToken("access"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("%d", req.UserID),
			Audience:  []string{j.config.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// CreateRefreshToken 创建刷新令牌
func (j *JWTManager) CreateRefreshToken(userID int64) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Token:  j.GenerateToken("refresh"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{j.config.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// CreateTokenPair 创建令牌对（访问令牌+刷新令牌）
func (j *JWTManager) CreateTokenPair(req *TokenRequest) (*TokenResponse, error) {
	// 创建访问令牌
	accessToken, err := j.CreateAccessToken(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %v", err)
	}

	// 创建刷新令牌
	refreshToken, err := j.CreateRefreshToken(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %v", err)
	}

	// 计算过期时间
	expiresAt := time.Now().Add(j.config.TokenExpiry)

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.config.TokenExpiry.Seconds()),
		ExpiresAt:    expiresAt,
		User: UserInfo{
			UserID:      req.UserID,
			Username:    req.Username,
			TenantType:  req.TenantType,
			Role:        req.Role,
			Permissions: req.Permissions,
		},
	}, nil
}

// ParseToken 解析令牌
func (j *JWTManager) ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateToken 验证令牌
func (j *JWTManager) ValidateToken(tokenString string) bool {
	_, err := j.ParseToken(tokenString)
	return err == nil
}

// RefreshAccessToken 刷新访问令牌
func (j *JWTManager) RefreshAccessToken(refreshToken string, userInfo *UserInfo) (string, error) {
	// 解析刷新令牌
	claims, err := j.ParseToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %v", err)
	}

	// 检查是否为刷新令牌
	if !strings.Contains(claims.Token, "refresh") {
		return "", fmt.Errorf("not a refresh token")
	}

	// 创建新的访问令牌
	req := &TokenRequest{
		UserID:      claims.UserID,
		Username:    userInfo.Username,
		TenantType:  userInfo.TenantType,
		Role:        userInfo.Role,
		Permissions: userInfo.Permissions,
	}

	return j.CreateAccessToken(req)
}

// ExtractUserInfo 从令牌中提取用户信息
func (j *JWTManager) ExtractUserInfo(tokenString string) (*UserInfo, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		UserID:      claims.UserID,
		Username:    claims.Username,
		TenantType:  claims.TenantType,
		Role:        claims.Role,
		Permissions: claims.Permissions,
	}, nil
}

// IsTokenExpired 检查令牌是否过期
func (j *JWTManager) IsTokenExpired(tokenString string) (bool, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return true, err
	}

	return time.Now().Unix() > claims.ExpiresAt.Unix(), nil
}

// GetTokenExpiration 获取令牌过期时间
func (j *JWTManager) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.ExpiresAt.Time, nil
}

// CreateAdminToken 创建管理员令牌
func (j *JWTManager) CreateAdminToken(userID int64, username string) (*TokenResponse, error) {
	req := &TokenRequest{
		UserID:      userID,
		Username:    username,
		TenantType:  TenantAdmin,
		Role:        "admin",
		Permissions: []string{"admin", "user:manage", "system:manage"},
	}

	return j.CreateTokenPair(req)
}

// CreatePersonalToken 创建个人用户令牌
func (j *JWTManager) CreatePersonalToken(userID int64, username string) (*TokenResponse, error) {
	req := &TokenRequest{
		UserID:      userID,
		Username:    username,
		TenantType:  TenantPersonal,
		Role:        "user",
		Permissions: []string{"user:read", "resume:manage", "job:apply"},
	}

	return j.CreateTokenPair(req)
}

// CreateEnterpriseToken 创建企业用户令牌
func (j *JWTManager) CreateEnterpriseToken(userID int64, username string) (*TokenResponse, error) {
	req := &TokenRequest{
		UserID:      userID,
		Username:    username,
		TenantType:  TenantEnterprise,
		Role:        "enterprise",
		Permissions: []string{"enterprise:read", "job:manage", "candidate:view"},
	}

	return j.CreateTokenPair(req)
}

// ValidateTenantType 验证租户类型
func (j *JWTManager) ValidateTenantType(tenantType TenantType) bool {
	switch tenantType {
	case TenantAdmin, TenantPersonal, TenantEnterprise:
		return true
	default:
		return false
	}
}

// GetTenantPermissions 获取租户默认权限
func (j *JWTManager) GetTenantPermissions(tenantType TenantType) []string {
	switch tenantType {
	case TenantAdmin:
		return []string{"admin", "user:manage", "system:manage", "enterprise:manage"}
	case TenantPersonal:
		return []string{"user:read", "resume:manage", "job:apply", "profile:manage"}
	case TenantEnterprise:
		return []string{"enterprise:read", "job:manage", "candidate:view", "company:manage"}
	default:
		return []string{"user:read"}
	}
}
