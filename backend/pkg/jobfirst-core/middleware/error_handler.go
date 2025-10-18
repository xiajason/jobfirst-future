package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core/errors"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// 创建错误响应
			response := &errors.ErrorResponse{
				Code:      errors.GetErrorCode(err),
				Message:   errors.GetErrorMessage(errors.GetErrorCode(err)),
				Details:   errors.GetErrorDetails(err),
				Timestamp: time.Now(),
				RequestID: c.GetString("request_id"),
				Path:      c.Request.URL.Path,
				Method:    c.Request.Method,
			}

			// 获取HTTP状态码
			status := errors.GetHTTPStatus(response.Code)

			// 记录错误日志
			if status >= 500 {
				// 服务器错误，记录详细日志
				c.Error(err)
			}

			// 返回错误响应
			c.JSON(status, response)
		}
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 创建内部服务器错误
		err := errors.NewError(errors.ErrCodeInternal, "内部服务器错误")

		// 创建错误响应
		response := &errors.ErrorResponse{
			Code:      err.Code,
			Message:   err.Message,
			Details:   "系统发生未预期的错误",
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		}

		// 记录panic日志
		c.Error(err)

		// 返回错误响应
		c.JSON(http.StatusInternalServerError, response)
	})
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成新的请求ID
			requestID = generateRequestID()
		}

		// 设置到上下文
		c.Set("request_id", requestID)

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ValidateRequest 请求验证中间件
func ValidateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求方法
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// 检查Content-Type
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				c.Error(errors.NewError(errors.ErrCodeValidation, "缺少Content-Type头"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RateLimit 限流中间件
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	// 简单的内存限流器
	requests := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// 清理过期记录
		if clientRequests, exists := requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range clientRequests {
				if now.Sub(reqTime) < window {
					validRequests = append(validRequests, reqTime)
				}
			}
			requests[clientIP] = validRequests
		}

		// 检查请求数量
		if len(requests[clientIP]) >= maxRequests {
			c.Error(errors.NewError(errors.ErrCodeRateLimit, "请求频率过高"))
			c.Abort()
			return
		}

		// 记录当前请求
		requests[clientIP] = append(requests[clientIP], now)

		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 设置CORS头
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Security 安全中间件
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
