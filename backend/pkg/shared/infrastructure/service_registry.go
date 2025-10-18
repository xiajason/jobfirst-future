package infrastructure

import (
	"context"
	"fmt"
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

// ServiceRegistry 服务注册接口
type ServiceRegistry interface {
	Register(service *ServiceInfo) error
	Deregister(serviceID string) error
	Discover(serviceName string) ([]*ServiceInfo, error)
	HealthCheck(serviceID string) error
	GetService(serviceID string) (*ServiceInfo, error)
	ListServices() ([]*ServiceInfo, error)
	Watch(serviceName string, callback func([]*ServiceInfo)) error
}

// ConsulRegistry Consul服务注册实现
type ConsulRegistry struct {
	client   *api.Client
	mu       sync.RWMutex
	services map[string]*ServiceInfo
	watchers map[string][]func([]*ServiceInfo)
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewConsulRegistry 创建Consul服务注册器
func NewConsulRegistry(config *api.Config) (*ConsulRegistry, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	registry := &ConsulRegistry{
		client:   client,
		services: make(map[string]*ServiceInfo),
		watchers: make(map[string][]func([]*ServiceInfo)),
		ctx:      ctx,
		cancel:   cancel,
	}

	return registry, nil
}

// Register 注册服务
func (r *ConsulRegistry) Register(service *ServiceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 创建Consul服务注册
	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Meta:    service.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:                           service.HealthCheck,
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	// 注册到Consul
	err := r.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}

	// 保存到本地缓存
	service.LastSeen = time.Now()
	r.services[service.ID] = service

	Info("Service registered successfully",
		Field{Key: "service_id", Value: service.ID},
		Field{Key: "service_name", Value: service.Name},
		Field{Key: "address", Value: service.Address},
		Field{Key: "port", Value: service.Port},
	)

	return nil
}

// Deregister 注销服务
func (r *ConsulRegistry) Deregister(serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 从Consul注销
	err := r.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %v", err)
	}

	// 从本地缓存移除
	delete(r.services, serviceID)

	Info("Service deregistered successfully",
		Field{Key: "service_id", Value: serviceID},
	)

	return nil
}

// Discover 发现服务
func (r *ConsulRegistry) Discover(serviceName string) ([]*ServiceInfo, error) {
	// 从Consul查询服务
	services, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %v", err)
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

	Info("Service discovery completed",
		Field{Key: "service_name", Value: serviceName},
		Field{Key: "count", Value: len(serviceInfos)},
	)

	return serviceInfos, nil
}

// HealthCheck 健康检查
func (r *ConsulRegistry) HealthCheck(serviceID string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	// 更新最后检查时间
	service.LastSeen = time.Now()

	// 这里可以添加自定义的健康检查逻辑
	// 例如：检查服务是否响应HTTP请求

	return nil
}

// GetService 获取服务信息
func (r *ConsulRegistry) GetService(serviceID string) (*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	return service, nil
}

// ListServices 列出所有服务
func (r *ConsulRegistry) ListServices() ([]*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []*ServiceInfo
	for _, service := range r.services {
		services = append(services, service)
	}

	return services, nil
}

// Watch 监听服务变化
func (r *ConsulRegistry) Watch(serviceName string, callback func([]*ServiceInfo)) error {
	r.mu.Lock()
	r.watchers[serviceName] = append(r.watchers[serviceName], callback)
	r.mu.Unlock()

	// 启动监听协程
	go r.watchService(serviceName)

	return nil
}

// watchService 监听服务变化
func (r *ConsulRegistry) watchService(serviceName string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			services, err := r.Discover(serviceName)
			if err != nil {
				Error("Failed to discover services during watch",
					Field{Key: "service_name", Value: serviceName},
					Field{Key: "error", Value: err.Error()},
				)
				continue
			}

			// 调用回调函数
			r.mu.RLock()
			callbacks := r.watchers[serviceName]
			r.mu.RUnlock()

			for _, callback := range callbacks {
				callback(services)
			}
		}
	}
}

// Close 关闭服务注册器
func (r *ConsulRegistry) Close() error {
	r.cancel()
	return nil
}

// InMemoryRegistry 内存服务注册实现（用于测试和开发）
type InMemoryRegistry struct {
	mu       sync.RWMutex
	services map[string]*ServiceInfo
	watchers map[string][]func([]*ServiceInfo)
}

// NewInMemoryRegistry 创建内存服务注册器
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		services: make(map[string]*ServiceInfo),
		watchers: make(map[string][]func([]*ServiceInfo)),
	}
}

// Register 注册服务
func (r *InMemoryRegistry) Register(service *ServiceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service.LastSeen = time.Now()
	service.Status = "passing"
	r.services[service.ID] = service

	Info("Service registered in memory",
		Field{Key: "service_id", Value: service.ID},
		Field{Key: "service_name", Value: service.Name},
	)

	return nil
}

// Deregister 注销服务
func (r *InMemoryRegistry) Deregister(serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.services, serviceID)

	Info("Service deregistered from memory",
		Field{Key: "service_id", Value: serviceID},
	)

	return nil
}

// Discover 发现服务
func (r *InMemoryRegistry) Discover(serviceName string) ([]*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []*ServiceInfo
	for _, service := range r.services {
		if service.Name == serviceName {
			services = append(services, service)
		}
	}

	return services, nil
}

// HealthCheck 健康检查
func (r *InMemoryRegistry) HealthCheck(serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	service.LastSeen = time.Now()
	return nil
}

// GetService 获取服务信息
func (r *InMemoryRegistry) GetService(serviceID string) (*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	return service, nil
}

// ListServices 列出所有服务
func (r *InMemoryRegistry) ListServices() ([]*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []*ServiceInfo
	for _, service := range r.services {
		services = append(services, service)
	}

	return services, nil
}

// Watch 监听服务变化
func (r *InMemoryRegistry) Watch(serviceName string, callback func([]*ServiceInfo)) error {
	r.mu.Lock()
	r.watchers[serviceName] = append(r.watchers[serviceName], callback)
	r.mu.Unlock()

	// 立即调用一次回调
	services, _ := r.Discover(serviceName)
	callback(services)

	return nil
}

// 全局服务注册器实例
var globalServiceRegistry ServiceRegistry

// InitGlobalServiceRegistry 初始化全局服务注册器
func InitGlobalServiceRegistry(registry ServiceRegistry) {
	globalServiceRegistry = registry
}

// GetServiceRegistry 获取全局服务注册器
func GetServiceRegistry() ServiceRegistry {
	return globalServiceRegistry
}
