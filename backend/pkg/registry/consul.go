package registry

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulRegistry Consul服务注册实现
type ConsulRegistry struct {
	client    *api.Client
	mu        sync.RWMutex
	services  map[string]*ServiceInfo
	standards *ServiceRegistryStandards
}

// NewConsulRegistry 创建Consul服务注册器
func NewConsulRegistry(config *api.Config) (*ConsulRegistry, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建Consul客户端失败: %v", err)
	}

	registry := &ConsulRegistry{
		client:    client,
		services:  make(map[string]*ServiceInfo),
		standards: NewServiceRegistryStandards(),
	}

	return registry, nil
}

// Register 注册服务
func (cr *ConsulRegistry) Register(service *ServiceInfo) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// 验证服务信息
	if err := cr.standards.ValidateServiceName(service.Name); err != nil {
		return fmt.Errorf("服务名称验证失败: %v", err)
	}

	if err := cr.standards.ValidateServiceID(service.ID); err != nil {
		return fmt.Errorf("服务ID验证失败: %v", err)
	}

	if err := cr.standards.ValidateTags(service.Tags); err != nil {
		return fmt.Errorf("服务标签验证失败: %v", err)
	}

	if err := cr.standards.ValidateMetadata(service.Meta); err != nil {
		return fmt.Errorf("服务元数据验证失败: %v", err)
	}

	// 获取健康检查配置
	healthConfig := cr.standards.GetHealthCheckConfig()

	// 创建Consul服务注册
	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Meta:    service.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d%s", service.Address, service.Port, service.HealthCheck),
			Interval:                       healthConfig.Interval,
			Timeout:                        healthConfig.Timeout,
			DeregisterCriticalServiceAfter: healthConfig.DeregisterCriticalServiceAfter,
		},
	}

	// 注册到Consul
	err := cr.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("注册服务失败: %v", err)
	}

	// 保存到本地缓存
	service.LastSeen = time.Now()
	cr.services[service.ID] = service

	log.Printf("✅ 服务 %s 已成功注册到Consul", service.Name)
	return nil
}

// Deregister 注销服务
func (cr *ConsulRegistry) Deregister(serviceID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// 从Consul注销
	err := cr.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("注销服务失败: %v", err)
	}

	// 从本地缓存移除
	delete(cr.services, serviceID)

	log.Printf("✅ 服务 %s 已从Consul注销", serviceID)
	return nil
}

// Discover 发现服务
func (cr *ConsulRegistry) Discover(serviceName string) ([]*ServiceInfo, error) {
	// 从Consul查询服务
	services, _, err := cr.client.Health().Service(serviceName, "", true, nil)
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

// GetService 获取服务信息
func (cr *ConsulRegistry) GetService(serviceID string) (*ServiceInfo, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	service, exists := cr.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("服务不存在: %s", serviceID)
	}

	return service, nil
}

// ListServices 列出所有服务
func (cr *ConsulRegistry) ListServices() ([]*ServiceInfo, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var services []*ServiceInfo
	for _, service := range cr.services {
		services = append(services, service)
	}

	return services, nil
}

// HealthCheck 健康检查
func (cr *ConsulRegistry) HealthCheck(serviceID string) error {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	service, exists := cr.services[serviceID]
	if !exists {
		return fmt.Errorf("服务不存在: %s", serviceID)
	}

	// 更新最后检查时间
	service.LastSeen = time.Now()

	return nil
}

// GetHealthyServiceURL 获取健康服务的URL
func (cr *ConsulRegistry) GetHealthyServiceURL(serviceName string) (string, error) {
	services, err := cr.Discover(serviceName)
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

// Watch 监听服务变化
func (cr *ConsulRegistry) Watch(serviceName string, callback func([]*ServiceInfo)) error {
	// 实现服务监听逻辑
	// 这里可以添加Consul的watch功能
	return fmt.Errorf("服务监听功能待实现")
}
