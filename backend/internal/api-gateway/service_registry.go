package apigateway

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

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

// ServiceRegistry 服务注册管理器
type ServiceRegistry struct {
	consulClient *api.Client
	services     map[string]*ServiceInfo
	mutex        sync.RWMutex
}

// NewServiceRegistry 创建服务注册管理器
func NewServiceRegistry() (*ServiceRegistry, error) {
	// 创建Consul客户端
	config := api.DefaultConfig()
	config.Address = "localhost:8500" // Consul地址

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建Consul客户端失败: %v", err)
	}

	registry := &ServiceRegistry{
		consulClient: client,
		services:     make(map[string]*ServiceInfo),
	}

	// 注册API Gateway自身
	if err := registry.registerAPIGateway(); err != nil {
		log.Printf("⚠️ 注册API Gateway失败: %v", err)
	}

	return registry, nil
}

// registerAPIGateway 注册API Gateway自身
func (sr *ServiceRegistry) registerAPIGateway() error {
	serviceInfo := &ServiceInfo{
		ID:          "api-gateway-future-1",
		Name:        "api-gateway",
		Address:     "localhost",
		Port:        7521,
		Tags:        []string{"api-gateway", "future", "gateway"},
		HealthCheck: "/health",
		Status:      "healthy",
		LastSeen:    time.Now(),
		Meta: map[string]string{
			"version":     "1.0.0",
			"type":        "api-gateway",
			"environment": "development",
			"mode":        "future",
		},
	}

	return sr.RegisterService(serviceInfo)
}

// RegisterService 注册服务
func (sr *ServiceRegistry) RegisterService(serviceInfo *ServiceInfo) error {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	// 创建Consul服务注册
	registration := &api.AgentServiceRegistration{
		ID:      serviceInfo.ID,
		Name:    serviceInfo.Name,
		Address: serviceInfo.Address,
		Port:    serviceInfo.Port,
		Tags:    serviceInfo.Tags,
		Meta:    serviceInfo.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d%s", serviceInfo.Address, serviceInfo.Port, serviceInfo.HealthCheck),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	// 注册到Consul
	err := sr.consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("注册服务失败: %v", err)
	}

	// 保存到本地缓存
	serviceInfo.LastSeen = time.Now()
	sr.services[serviceInfo.ID] = serviceInfo

	log.Printf("✅ 服务 %s 已成功注册到Consul", serviceInfo.Name)
	return nil
}

// DiscoverService 发现服务
func (sr *ServiceRegistry) DiscoverService(serviceName string) ([]*ServiceInfo, error) {
	// 从Consul查询服务
	services, _, err := sr.consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("发现服务失败: %v", err)
	}

	var serviceInfos []*ServiceInfo
	for _, service := range services {
		serviceInfo := &ServiceInfo{
			ID:          service.Service.ID,
			Name:        service.Service.Service,
			Address:     service.Service.Address,
			Port:        service.Service.Port,
			Tags:        service.Service.Tags,
			Meta:        service.Service.Meta,
			HealthCheck: service.Service.Address,
			Status:      service.Checks.AggregatedStatus(),
			LastSeen:    time.Now(),
		}
		serviceInfos = append(serviceInfos, serviceInfo)
	}

	return serviceInfos, nil
}

// GetHealthyServiceURL 获取健康服务的URL
func (sr *ServiceRegistry) GetHealthyServiceURL(serviceName string) (string, error) {
	services, err := sr.DiscoverService(serviceName)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("未找到服务: %s", serviceName)
	}

	// 选择第一个健康的服务
	for _, service := range services {
		if service.Status == "passing" {
			return fmt.Sprintf("http://%s:%d", service.Address, service.Port), nil
		}
	}

	return "", fmt.Errorf("未找到健康的服务实例: %s", serviceName)
}

// ListServices 列出所有服务
func (sr *ServiceRegistry) ListServices() ([]*ServiceInfo, error) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	var services []*ServiceInfo
	for _, service := range sr.services {
		services = append(services, service)
	}

	return services, nil
}

// GetService 获取服务信息
func (sr *ServiceRegistry) GetService(serviceID string) (*ServiceInfo, error) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	service, exists := sr.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("服务不存在: %s", serviceID)
	}

	return service, nil
}
