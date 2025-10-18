package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader("X-Request-ID")

		// 如果没有请求ID，生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到上下文中
		c.Set("RequestID", requestID)

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		// 继续处理请求
		c.Next()
	}
}
