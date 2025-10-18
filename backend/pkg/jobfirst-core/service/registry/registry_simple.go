package registry

import (
	"context"
	"sync"
	"time"

	"github.com/jobfirst/jobfirst-core/errors"
)

// SimpleServiceRegistry 简化的服务注册中心
type SimpleServiceRegistry struct {
	services map[string]*ServiceInfo
	mutex    sync.RWMutex
	config   *RegistryConfig
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSimpleServiceRegistry 创建简化的服务注册中心
func NewSimpleServiceRegistry(config *RegistryConfig) (*SimpleServiceRegistry, error) {
	if config == nil {
		return nil, errors.NewError(errors.ErrCodeValidation, "registry config cannot be nil")
	}

	// 设置默认值
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.TTL == 0 {
		config.TTL = 60 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	registry := &SimpleServiceRegistry{
		services: make(map[string]*ServiceInfo),
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
	}

	return registry, nil
}

// Register 注册服务
func (sr *SimpleServiceRegistry) Register(service *ServiceInfo) error {
	if service == nil {
		return errors.NewError(errors.ErrCodeValidation, "service cannot be nil")
	}

	if service.ID == "" || service.Name == "" || service.Endpoint == "" {
		return errors.NewError(errors.ErrCodeValidation, "service ID, name and endpoint are required")
	}

	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	// 设置注册时间
	service.RegisteredAt = time.Now()
	service.LastCheck = time.Now()

	// 注册到内存
	sr.services[service.ID] = service

	return nil
}

// Deregister 注销服务
func (sr *SimpleServiceRegistry) Deregister(serviceID string) error {
	if serviceID == "" {
		return errors.NewError(errors.ErrCodeValidation, "service ID cannot be empty")
	}

	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	// 从内存中移除
	delete(sr.services, serviceID)

	return nil
}

// GetService 获取服务信息
func (sr *SimpleServiceRegistry) GetService(serviceID string) (*ServiceInfo, error) {
	if serviceID == "" {
		return nil, errors.NewError(errors.ErrCodeValidation, "service ID cannot be empty")
	}

	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	service, exists := sr.services[serviceID]
	if !exists {
		return nil, errors.NewError(errors.ErrCodeNotFound, "service not found")
	}

	return service, nil
}

// GetServices 获取所有服务
func (sr *SimpleServiceRegistry) GetServices() []*ServiceInfo {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	services := make([]*ServiceInfo, 0, len(sr.services))
	for _, service := range sr.services {
		services = append(services, service)
	}

	return services
}

// GetServicesByName 根据服务名获取服务列表
func (sr *SimpleServiceRegistry) GetServicesByName(serviceName string) ([]*ServiceInfo, error) {
	if serviceName == "" {
		return nil, errors.NewError(errors.ErrCodeValidation, "service name cannot be empty")
	}

	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	var services []*ServiceInfo
	for _, service := range sr.services {
		if service.Name == serviceName {
			services = append(services, service)
		}
	}

	return services, nil
}

// SelectService 选择服务（简单轮询）
func (sr *SimpleServiceRegistry) SelectService(serviceName string) (*ServiceInfo, error) {
	services, err := sr.GetServicesByName(serviceName)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, errors.NewError(errors.ErrCodeNotFound, "no services found")
	}

	// 过滤健康的服务
	healthyServices := make([]*ServiceInfo, 0)
	for _, service := range services {
		if service.Health != nil && service.Health.Status == "healthy" {
			healthyServices = append(healthyServices, service)
		}
	}

	if len(healthyServices) == 0 {
		return nil, errors.NewError(errors.ErrCodeService, "no healthy services available")
	}

	// 简单轮询选择
	return healthyServices[0], nil
}

// UpdateServiceHealth 更新服务健康状态
func (sr *SimpleServiceRegistry) UpdateServiceHealth(serviceID string, health *HealthStatus) error {
	if serviceID == "" {
		return errors.NewError(errors.ErrCodeValidation, "service ID cannot be empty")
	}

	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	service, exists := sr.services[serviceID]
	if !exists {
		return errors.NewError(errors.ErrCodeNotFound, "service not found")
	}

	service.Health = health
	service.LastCheck = time.Now()

	return nil
}

// GetRegistryStatus 获取注册中心状态
func (sr *SimpleServiceRegistry) GetRegistryStatus() map[string]interface{} {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	status := map[string]interface{}{
		"total_services": len(sr.services),
		"config":         sr.config,
		"services":       sr.services,
	}

	// 统计健康状态
	healthyCount := 0
	unhealthyCount := 0
	for _, service := range sr.services {
		if service.Health != nil {
			if service.Health.Status == "healthy" {
				healthyCount++
			} else {
				unhealthyCount++
			}
		}
	}

	status["healthy_services"] = healthyCount
	status["unhealthy_services"] = unhealthyCount

	return status
}

// Close 关闭注册中心
func (sr *SimpleServiceRegistry) Close() error {
	sr.cancel()
	return nil
}
