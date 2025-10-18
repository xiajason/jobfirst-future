package infrastructure

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// SecurityMiddleware 安全中间件接口
type SecurityMiddleware interface {
	AuthMiddleware() gin.HandlerFunc
	RateLimitMiddleware() gin.HandlerFunc
	CORSMiddleware() gin.HandlerFunc
	JWTMiddleware() gin.HandlerFunc
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret           string        `yaml:"jwt_secret" json:"jwt_secret"`
	JWTExpiresIn        time.Duration `yaml:"jwt_expires_in" json:"jwt_expires_in"`
	JWTRefreshExpiresIn time.Duration `yaml:"jwt_refresh_expires_in" json:"jwt_refresh_expires_in"`

	CORSAllowedOrigins   []string `yaml:"cors_allowed_origins" json:"cors_allowed_origins"`
	CORSAllowedMethods   []string `yaml:"cors_allowed_methods" json:"cors_allowed_methods"`
	CORSAllowedHeaders   []string `yaml:"cors_allowed_headers" json:"cors_allowed_headers"`
	CORSAllowCredentials bool     `yaml:"cors_allow_credentials" json:"cors_allow_credentials"`

	RateLimitEnabled  bool          `yaml:"rate_limit_enabled" json:"rate_limit_enabled"`
	RateLimitRequests int           `yaml:"rate_limit_requests" json:"rate_limit_requests"`
	RateLimitWindow   time.Duration `yaml:"rate_limit_window" json:"rate_limit_window"`
	RateLimitBurst    int           `yaml:"rate_limit_burst" json:"rate_limit_burst"`
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// SecurityManager 安全管理器
type SecurityManager struct {
	config *SecurityConfig
	cache  Cache
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager(config *SecurityConfig, cache Cache) *SecurityManager {
	return &SecurityManager{
		config: config,
		cache:  cache,
	}
}

// AuthMiddleware 认证中间件
func (sm *SecurityManager) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		claims, err := sm.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// JWTMiddleware JWT中间件（可选认证）
func (sm *SecurityManager) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有token，继续处理（可选认证）
			c.Next()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		claims, err := sm.ValidateJWT(tokenString)
		if err != nil {
			// token无效，但不阻止请求继续
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RateLimitMiddleware 限流中间件
func (sm *SecurityManager) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !sm.config.RateLimitEnabled {
			c.Next()
			return
		}

		// 获取客户端IP
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// 检查当前请求数
		current, err := sm.getCurrentRequests(key)
		if err != nil {
			Error("Rate limit check failed",
				Field{Key: "client_ip", Value: clientIP},
				Field{Key: "error", Value: err.Error()},
			)
			c.Next()
			return
		}

		// 检查是否超过限制
		if current >= sm.config.RateLimitRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too Many Requests",
				"message":     "Rate limit exceeded",
				"retry_after": sm.config.RateLimitWindow.Seconds(),
			})
			c.Abort()
			return
		}

		// 增加请求计数
		err = sm.incrementRequests(key)
		if err != nil {
			Error("Rate limit increment failed",
				Field{Key: "client_ip", Value: clientIP},
				Field{Key: "error", Value: err.Error()},
			)
		}

		c.Next()
	}
}

// CORSMiddleware CORS中间件
func (sm *SecurityManager) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该来源
		allowed := false
		for _, allowedOrigin := range sm.config.CORSAllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.CORSAllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.CORSAllowedHeaders, ", "))

		if sm.config.CORSAllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// GenerateJWT 生成JWT token
func (sm *SecurityManager) GenerateJWT(userID uint, username, role string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(sm.config.JWTExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "jobfirst",
			Subject:   strconv.FormatUint(uint64(userID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(sm.config.JWTSecret))
}

// GenerateRefreshToken 生成刷新token
func (sm *SecurityManager) GenerateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(sm.config.JWTRefreshExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "jobfirst",
			Subject:   strconv.FormatUint(uint64(userID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(sm.config.JWTSecret))
}

// ValidateJWT 验证JWT token
func (sm *SecurityManager) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(sm.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashPassword 哈希密码
func (sm *SecurityManager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 检查密码
func (sm *SecurityManager) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateHMAC 生成HMAC签名
func (sm *SecurityManager) GenerateHMAC(data string) string {
	h := hmac.New(sha256.New, []byte(sm.config.JWTSecret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifyHMAC 验证HMAC签名
func (sm *SecurityManager) VerifyHMAC(data, signature string) bool {
	expectedSignature := sm.GenerateHMAC(data)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// getCurrentRequests 获取当前请求数
func (sm *SecurityManager) getCurrentRequests(key string) (int, error) {
	if sm.cache == nil {
		return 0, nil
	}

	data, err := sm.cache.Get(context.Background(), key)
	if err != nil {
		return 0, nil
	}

	var count int
	err = json.Unmarshal(data, &count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// incrementRequests 增加请求计数
func (sm *SecurityManager) incrementRequests(key string) error {
	if sm.cache == nil {
		return nil
	}

	current, err := sm.getCurrentRequests(key)
	if err != nil {
		current = 0
	}

	current++

	data, err := json.Marshal(current)
	if err != nil {
		return err
	}

	return sm.cache.Set(context.Background(), key, data, sm.config.RateLimitWindow)
}

// CreateDefaultSecurityConfig 创建默认安全配置
func CreateDefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		JWTSecret:           "your-secret-key-change-in-production",
		JWTExpiresIn:        24 * time.Hour,
		JWTRefreshExpiresIn: 7 * 24 * time.Hour,

		CORSAllowedOrigins:   []string{"*"},
		CORSAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSAllowedHeaders:   []string{"*"},
		CORSAllowCredentials: true,

		RateLimitEnabled:  true,
		RateLimitRequests: 100,
		RateLimitWindow:   1 * time.Minute,
		RateLimitBurst:    20,
	}
}

// 全局安全管理器实例
var globalSecurityManager *SecurityManager

// InitGlobalSecurityManager 初始化全局安全管理器
func InitGlobalSecurityManager(config *SecurityConfig, cache Cache) {
	globalSecurityManager = NewSecurityManager(config, cache)
}

// GetSecurityManager 获取全局安全管理器
func GetSecurityManager() *SecurityManager {
	return globalSecurityManager
}
