package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/registry"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	healthChecker   *HealthChecker
	serviceRegistry registry.ServiceRegistry
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(healthChecker *HealthChecker, serviceRegistry registry.ServiceRegistry) *HealthHandler {
	return &HealthHandler{
		healthChecker:   healthChecker,
		serviceRegistry: serviceRegistry,
	}
}

// Health 综合健康检查
func (hh *HealthHandler) Health(c *gin.Context) {
	// 获取服务健康摘要
	summary := hh.healthChecker.GetServiceHealthSummary()

	// 检查整体健康状态
	total := summary["total"].(int)
	healthy := summary["healthy"].(int)

	status := "healthy"
	if total > 0 && healthy < total {
		status = "degraded"
	}
	if healthy == 0 && total > 0 {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"timestamp": time.Now(),
		"summary":   summary,
		"message":   "系统健康检查",
	})
}

// Ready 就绪检查
func (hh *HealthHandler) Ready(c *gin.Context) {
	// 检查关键服务是否就绪
	criticalServices := []string{"user-service", "api-gateway"}

	ready := true
	var notReadyServices []string

	for _, serviceName := range criticalServices {
		services, err := hh.serviceRegistry.Discover(serviceName)
		if err != nil || len(services) == 0 {
			ready = false
			notReadyServices = append(notReadyServices, serviceName)
		}
	}

	if ready {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"message": "系统已就绪",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":           "not_ready",
			"message":          "系统未就绪",
			"missing_services": notReadyServices,
		})
	}
}

// Live 存活检查
func (hh *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"message":   "系统存活",
		"timestamp": time.Now(),
	})
}

// Services 服务列表和状态
func (hh *HealthHandler) Services(c *gin.Context) {
	// 获取所有服务
	services, err := hh.serviceRegistry.ListServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取服务列表失败",
		})
		return
	}

	// 获取健康状态
	healthStatuses := hh.healthChecker.GetAllHealthStatuses()

	// 合并服务信息和健康状态
	var serviceList []gin.H
	for _, service := range services {
		serviceInfo := gin.H{
			"id":        service.ID,
			"name":      service.Name,
			"address":   service.Address,
			"port":      service.Port,
			"tags":      service.Tags,
			"meta":      service.Meta,
			"last_seen": service.LastSeen,
		}

		// 添加健康状态
		if healthStatus, exists := healthStatuses[service.ID]; exists {
			serviceInfo["health_status"] = healthStatus.Status
			serviceInfo["last_check"] = healthStatus.LastCheck
			serviceInfo["response_time_ms"] = healthStatus.ResponseTime
			if healthStatus.Error != "" {
				serviceInfo["error"] = healthStatus.Error
			}
		} else {
			serviceInfo["health_status"] = "unknown"
		}

		serviceList = append(serviceList, serviceInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"services": serviceList,
		"count":    len(serviceList),
		"summary":  hh.healthChecker.GetServiceHealthSummary(),
	})
}

// ServiceHealth 单个服务健康检查
func (hh *HealthHandler) ServiceHealth(c *gin.Context) {
	serviceID := c.Param("serviceId")

	healthStatus, err := hh.healthChecker.GetHealthStatus(serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "服务健康状态不存在",
		})
		return
	}

	c.JSON(http.StatusOK, healthStatus)
}

// HealthyServices 获取健康的服务
func (hh *HealthHandler) HealthyServices(c *gin.Context) {
	healthyServices := hh.healthChecker.GetHealthyServices()

	c.JSON(http.StatusOK, gin.H{
		"healthy_services": healthyServices,
		"count":            len(healthyServices),
	})
}

// UnhealthyServices 获取不健康的服务
func (hh *HealthHandler) UnhealthyServices(c *gin.Context) {
	unhealthyServices := hh.healthChecker.GetUnhealthyServices()

	c.JSON(http.StatusOK, gin.H{
		"unhealthy_services": unhealthyServices,
		"count":              len(unhealthyServices),
	})
}

// CheckService 检查指定服务
func (hh *HealthHandler) CheckService(c *gin.Context) {
	serviceName := c.Param("serviceName")

	healthStatus, err := hh.healthChecker.CheckServiceHealth(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "服务不存在或检查失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, healthStatus)
}
