package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 结构化日志格式
		return log.Sprintf("[%s] %s %s %d %s %s %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.ClientIP,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.Path,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CustomLogger 自定义日志中间件
func CustomLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 构建查询字符串
		if raw != "" {
			path = path + "?" + raw
		}

		// 获取请求ID
		requestID := c.GetString("RequestID")
		if requestID == "" {
			requestID = "unknown"
		}

		// 记录日志
		log.Printf("[%s] %s %s %d %s %s %s %s",
			time.Now().Format("2006-01-02 15:04:05"),
			c.ClientIP(),
			c.Request.Method,
			c.Writer.Status(),
			latency,
			path,
			c.Request.UserAgent(),
			requestID,
		)
	}
}
