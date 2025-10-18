package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Recovery 错误恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				log.Printf("❌ Panic recovered: %v", err)

				// 获取请求ID
				requestID := c.GetString("RequestID")
				if requestID == "" {
					requestID = "unknown"
				}

				// 记录详细错误信息
				log.Printf("Request ID: %s, Path: %s, Method: %s, Client IP: %s",
					requestID,
					c.Request.URL.Path,
					c.Request.Method,
					c.ClientIP(),
				)

				// 返回统一错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "内部服务器错误",
					"request_id": requestID,
					"timestamp":  fmt.Sprintf("%d", time.Now().Unix()),
				})

				// 中止请求处理
				c.Abort()
			}
		}()

		// 继续处理请求
		c.Next()
	}
}
