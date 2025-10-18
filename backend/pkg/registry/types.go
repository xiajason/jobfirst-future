package registry

import "time"

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	Meta        map[string]string `json:"meta"`
	HealthCheck string            `json:"health_check"`
	Status      string            `json:"status"`
	LastSeen    time.Time         `json:"last_seen"`
}

// ServiceRegistry 服务注册接口
type ServiceRegistry interface {
	// 服务注册和注销
	Register(service *ServiceInfo) error
	Deregister(serviceID string) error

	// 服务发现
	Discover(serviceName string) ([]*ServiceInfo, error)
	GetService(serviceID string) (*ServiceInfo, error)
	ListServices() ([]*ServiceInfo, error)

	// 健康检查
	HealthCheck(serviceID string) error
	GetHealthyServiceURL(serviceName string) (string, error)

	// 服务监听
	Watch(serviceName string, callback func([]*ServiceInfo)) error
}

// ServiceRegistration 服务注册信息
type ServiceRegistration struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	Meta        map[string]string `json:"meta"`
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	ServiceID    string    `json:"service_id"`
	ServiceName  string    `json:"service_name"`
	Status       string    `json:"status"`
	LastCheck    time.Time `json:"last_check"`
	ResponseTime int64     `json:"response_time_ms"`
	Error        string    `json:"error,omitempty"`
}

// ServiceDiscoveryResult 服务发现结果
type ServiceDiscoveryResult struct {
	Services  []*ServiceInfo `json:"services"`
	Count     int            `json:"count"`
	Healthy   int            `json:"healthy"`
	Unhealthy int            `json:"unhealthy"`
}

// ServiceMetrics 服务指标
type ServiceMetrics struct {
	TotalServices     int `json:"total_services"`
	HealthyServices   int `json:"healthy_services"`
	UnhealthyServices int `json:"unhealthy_services"`
	UnknownServices   int `json:"unknown_services"`
}
