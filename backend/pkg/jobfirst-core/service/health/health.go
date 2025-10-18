package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheckerInterface 健康检查器接口
type HealthCheckerInterface interface {
	Check() *HealthStatus
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Timestamp string                 `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ServiceHealth 服务健康检查器
type ServiceHealth struct {
	serviceName string
	version     string
	checkers    []HealthCheckerInterface
}

// NewServiceHealth 创建服务健康检查器
func NewServiceHealth(serviceName, version string) *ServiceHealth {
	return &ServiceHealth{
		serviceName: serviceName,
		version:     version,
		checkers:    make([]HealthCheckerInterface, 0),
	}
}

// AddChecker 添加健康检查器
func (sh *ServiceHealth) AddChecker(checker HealthCheckerInterface) {
	sh.checkers = append(sh.checkers, checker)
}

// Check 执行健康检查
func (sh *ServiceHealth) Check() *HealthStatus {
	status := "healthy"
	message := "Service is healthy"
	details := make(map[string]interface{})

	// 执行所有检查器
	for _, checker := range sh.checkers {
		checkResult := checker.Check()
		if checkResult.Status != "healthy" {
			status = "unhealthy"
			message = checkResult.Message
		}
		details[checker.(*ComponentHealth).componentName] = checkResult
	}

	return &HealthStatus{
		Status:    status,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
		Details:   details,
	}
}

// ComponentHealth 组件健康检查器
type ComponentHealth struct {
	componentName string
	checkFunc     func() error
}

// NewComponentHealth 创建组件健康检查器
func NewComponentHealth(componentName string, checkFunc func() error) *ComponentHealth {
	return &ComponentHealth{
		componentName: componentName,
		checkFunc:     checkFunc,
	}
}

// Check 检查组件健康状态
func (ch *ComponentHealth) Check() *HealthStatus {
	err := ch.checkFunc()
	if err != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Message:   err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return &HealthStatus{
		Status:    "healthy",
		Message:   "Component is healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// HealthHandler 健康检查处理器
func HealthHandler(serviceHealth *ServiceHealth) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := serviceHealth.Check()

		statusCode := http.StatusOK
		if health.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		response := gin.H{
			"service":   serviceHealth.serviceName,
			"version":   serviceHealth.version,
			"status":    health.Status,
			"message":   health.Message,
			"timestamp": health.Timestamp,
		}

		if len(health.Details) > 0 {
			response["details"] = health.Details
		}

		c.JSON(statusCode, response)
	}
}
