package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xiajason/zervi-basic/basic/backend/internal/api-gateway/handlers"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, proxyHandler *handlers.ProxyHandler, healthHandler *handlers.HealthHandler) {
	// API版本组
	v1 := r.Group("/api/v1")
	{
		// 服务代理路由
		v1.Any("/:serviceName/*path", proxyHandler.ServiceProxy)

		// 服务信息路由
		v1.GET("/services", proxyHandler.ListServices)
		v1.GET("/services/:serviceName", proxyHandler.GetServiceInfo)
	}

	// 健康检查路由
	health := r.Group("/health")
	{
		health.GET("", healthHandler.Health)
		health.GET("/ready", healthHandler.Ready)
		health.GET("/live", healthHandler.Live)
		health.GET("/services", healthHandler.Services)
		health.GET("/services/:serviceId", healthHandler.ServiceHealth)
		health.GET("/healthy", healthHandler.HealthyServices)
		health.GET("/unhealthy", healthHandler.UnhealthyServices)
	}

	// 根路径
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "JobFirst Future版 API Gateway",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// 服务列表路由（兼容性）
	r.GET("/api/v1/services", proxyHandler.ListServices)
}
