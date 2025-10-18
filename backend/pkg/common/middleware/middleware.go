package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"resume-centre/common/utils"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(whitelist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 检查是否为白名单路径
		if utils.IsWhitelistPath(path, whitelist) {
			c.Next()
			return
		}

		// 获取认证头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 尝试从accessToken头获取
			authHeader = c.GetHeader("accessToken")
		}

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Authorization header required",
			})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if strings.HasPrefix(authHeader, "Bearer ") {
			authHeader = authHeader[7:]
		}

		// 验证令牌
		if !utils.ValidateToken(authHeader) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Invalid token",
			})
			c.Abort()
			return
		}

		// 提取用户ID并设置到上下文
		userID := utils.ExtractUserID(authHeader)
		c.Set("userID", userID)
		c.Set("accessToken", authHeader)

		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, accessToken")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return gin.Logger()
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateToken("req")
		}
		c.Header("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		c.Next()
	}
}

// RateLimitMiddleware 限流中间件（简化版本）
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现更复杂的限流逻辑
		// 目前只是简单的通过
		c.Next()
	}
}

// MetricsMiddleware 指标中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// 记录请求指标
		duration := time.Since(start)
		status := c.Writer.Status()

		// 这里可以记录到Prometheus或其他监控系统
		_ = duration
		_ = status
	}
}
