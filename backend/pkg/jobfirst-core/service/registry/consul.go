package registry

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulRegistry Consul服务注册器
type ConsulRegistry struct {
	client *api.Client
}

// NewConsulRegistry 创建Consul注册器
func NewConsulRegistry(address string) (*ConsulRegistry, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建Consul客户端失败: %v", err)
	}

	return &ConsulRegistry{
		client: client,
	}, nil
}

// Register 注册服务到Consul
func (cr *ConsulRegistry) Register(serviceInfo *ServiceInfo) error {
	registration := &api.AgentServiceRegistration{
		ID:      serviceInfo.ID,
		Name:    serviceInfo.Name,
		Tags:    []string{"jobfirst", "microservice"},
		Port:    serviceInfo.Port,
		Address: serviceInfo.Address,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceInfo.Address, serviceInfo.Port),
			Timeout:                        "3s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err := cr.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("注册服务失败: %v", err)
	}

	log.Printf("✅ 服务 %s 已成功注册到Consul", serviceInfo.Name)
	return nil
}

// Deregister 从Consul注销服务
func (cr *ConsulRegistry) Deregister(serviceID string) error {
	err := cr.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("注销服务失败: %v", err)
	}

	log.Printf("✅ 服务 %s 已从Consul注销", serviceID)
	return nil
}

// GetService 获取服务信息
func (cr *ConsulRegistry) GetService(serviceName string) ([]*api.ServiceEntry, error) {
	services, _, err := cr.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("获取服务信息失败: %v", err)
	}

	return services, nil
}

// UpdateHealth 更新服务健康状态
func (cr *ConsulRegistry) UpdateHealth(serviceID string, health *HealthStatus) error {
	checkID := fmt.Sprintf("service:%s", serviceID)

	var status string
	switch health.Status {
	case "healthy":
		status = api.HealthPassing
	case "unhealthy":
		status = api.HealthCritical
	default:
		status = api.HealthWarning
	}

	err := cr.client.Agent().UpdateTTL(checkID, health.Message, status)
	if err != nil {
		return fmt.Errorf("更新健康状态失败: %v", err)
	}

	return nil
}

// WatchServices 监听服务变化
func (cr *ConsulRegistry) WatchServices(serviceName string, callback func([]*api.ServiceEntry)) error {
	queryOptions := &api.QueryOptions{
		WaitTime: 10 * time.Second,
	}

	for {
		services, meta, err := cr.client.Health().Service(serviceName, "", true, queryOptions)
		if err != nil {
			log.Printf("监听服务变化失败: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		queryOptions.WaitIndex = meta.LastIndex
		callback(services)
	}
}
