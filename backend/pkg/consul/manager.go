package consul

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/xiajason/zervi-basic/basic/backend/pkg/config"

	"github.com/hashicorp/consul/api"
)

// ServiceManager Consul服务管理器
type ServiceManager struct {
	client   *api.Client
	config   *config.ConsulConfig
	services map[string]*api.AgentServiceRegistration
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager(cfg *config.ConsulConfig) (*ServiceManager, error) {
	consulConfig := &api.Config{
		Address:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Scheme:     cfg.Scheme,
		Datacenter: cfg.Datacenter,
		Token:      cfg.Token,
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &ServiceManager{
		client:   client,
		config:   cfg,
		services: make(map[string]*api.AgentServiceRegistration),
		ctx:      ctx,
		cancel:   cancel,
	}

	return manager, nil
}

// RegisterService 注册服务到Consul
func (m *ServiceManager) RegisterService(serviceName, serviceID, address string, port int, tags []string) error {
	if !m.config.Enabled {
		log.Printf("Consul is disabled, skipping service registration for %s", serviceName)
		return nil
	}

	// 构建健康检查URL
	healthCheckURL := fmt.Sprintf("http://%s:%d%s", address, port, m.config.HealthCheckURL)

	// 解析健康检查间隔和超时（用于日志记录）
	_, err := time.ParseDuration(m.config.HealthCheckInterval)
	if err != nil {
		log.Printf("Invalid health check interval: %s, using default", m.config.HealthCheckInterval)
	}

	_, err = time.ParseDuration(m.config.HealthCheckTimeout)
	if err != nil {
		log.Printf("Invalid health check timeout: %s, using default", m.config.HealthCheckTimeout)
	}

	_, err = time.ParseDuration(m.config.DeregisterAfter)
	if err != nil {
		log.Printf("Invalid deregister after: %s, using default", m.config.DeregisterAfter)
	}

	// 创建服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:                           healthCheckURL,
			Interval:                       m.config.HealthCheckInterval,
			Timeout:                        m.config.HealthCheckTimeout,
			DeregisterCriticalServiceAfter: m.config.DeregisterAfter,
		},
	}

	// 注册服务
	err = m.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service %s: %v", serviceName, err)
	}

	// 保存到本地缓存
	m.services[serviceID] = registration

	log.Printf("Service %s registered successfully to Consul", serviceName)
	return nil
}

// DeregisterService 从Consul注销服务
func (m *ServiceManager) DeregisterService(serviceID string) error {
	if !m.config.Enabled {
		log.Printf("Consul is disabled, skipping service deregistration for %s", serviceID)
		return nil
	}

	err := m.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service %s: %v", serviceID, err)
	}

	// 从本地缓存移除
	delete(m.services, serviceID)

	log.Printf("Service %s deregistered successfully from Consul", serviceID)
	return nil
}

// DiscoverService 发现服务
func (m *ServiceManager) DiscoverService(serviceName string) ([]*api.ServiceEntry, error) {
	if !m.config.Enabled {
		return nil, fmt.Errorf("consul is disabled")
	}

	services, _, err := m.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %v", serviceName, err)
	}

	return services, nil
}

// GetServiceHealth 获取服务健康状态
func (m *ServiceManager) GetServiceHealth(serviceName string) (string, error) {
	if !m.config.Enabled {
		return "unknown", nil
	}

	services, _, err := m.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "error", err
	}

	if len(services) == 0 {
		return "not_found", nil
	}

	// 返回第一个服务的健康状态
	return services[0].Checks.AggregatedStatus(), nil
}

// ListServices 列出所有注册的服务
func (m *ServiceManager) ListServices() (map[string]*api.AgentService, error) {
	if !m.config.Enabled {
		return nil, fmt.Errorf("consul is disabled")
	}

	services, err := m.client.Agent().Services()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	return services, nil
}

// HealthCheck 执行健康检查
func (m *ServiceManager) HealthCheck(serviceID string) error {
	if !m.config.Enabled {
		return nil
	}

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found in local cache", serviceID)
	}

	// 构建健康检查URL
	healthCheckURL := fmt.Sprintf("http://%s:%d%s", service.Address, service.Port, m.config.HealthCheckURL)

	// 执行HTTP健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthCheckURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed for %s: %v", serviceID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed for %s: status %d", serviceID, resp.StatusCode)
	}

	log.Printf("Health check passed for service %s", serviceID)
	return nil
}

// StartHealthCheckLoop 启动健康检查循环
func (m *ServiceManager) StartHealthCheckLoop() {
	if !m.config.Enabled {
		log.Println("Consul is disabled, skipping health check loop")
		return
	}

	interval, err := time.ParseDuration(m.config.HealthCheckInterval)
	if err != nil {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting health check loop with interval %v", interval)

	for {
		select {
		case <-m.ctx.Done():
			log.Println("Health check loop stopped")
			return
		case <-ticker.C:
			// 对所有注册的服务执行健康检查
			for serviceID := range m.services {
				go func(id string) {
					if err := m.HealthCheck(id); err != nil {
						log.Printf("Health check failed for service %s: %v", id, err)
					}
				}(serviceID)
			}
		}
	}
}

// Close 关闭服务管理器
func (m *ServiceManager) Close() error {
	m.cancel()

	// 注销所有服务
	for serviceID := range m.services {
		if err := m.DeregisterService(serviceID); err != nil {
			log.Printf("Failed to deregister service %s: %v", serviceID, err)
		}
	}

	return nil
}

// IsHealthy 检查Consul连接是否健康
func (m *ServiceManager) IsHealthy() bool {
	if !m.config.Enabled {
		return false
	}

	leader, err := m.client.Status().Leader()
	if err != nil {
		return false
	}

	return leader != ""
}

// GetConsulStatus 获取Consul状态信息
func (m *ServiceManager) GetConsulStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled": m.config.Enabled,
		"healthy": m.IsHealthy(),
	}

	if m.config.Enabled {
		leader, err := m.client.Status().Leader()
		if err == nil {
			status["leader"] = leader
		}

		peers, err := m.client.Status().Peers()
		if err == nil {
			status["peers"] = peers
		}

		services, err := m.ListServices()
		if err == nil {
			status["registered_services"] = len(services)
		}
	}

	return status
}
