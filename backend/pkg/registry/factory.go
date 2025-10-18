package registry

import (
	"fmt"
	"os"

	"github.com/hashicorp/consul/api"
)

// RegistryFactory 注册器工厂
type RegistryFactory struct{}

// NewRegistryFactory 创建注册器工厂
func NewRegistryFactory() *RegistryFactory {
	return &RegistryFactory{}
}

// CreateConsulRegistry 创建Consul注册器
func (rf *RegistryFactory) CreateConsulRegistry() (ServiceRegistry, error) {
	// 获取Consul配置
	config := rf.getConsulConfig()

	// 创建Consul注册器
	registry, err := NewConsulRegistry(config)
	if err != nil {
		return nil, fmt.Errorf("创建Consul注册器失败: %v", err)
	}

	return registry, nil
}

// getConsulConfig 获取Consul配置
func (rf *RegistryFactory) getConsulConfig() *api.Config {
	config := api.DefaultConfig()

	// 从环境变量获取Consul地址
	if consulAddr := os.Getenv("CONSUL_ADDR"); consulAddr != "" {
		config.Address = consulAddr
	} else {
		config.Address = "localhost:8500" // 默认地址
	}

	// 从环境变量获取其他配置
	if consulToken := os.Getenv("CONSUL_TOKEN"); consulToken != "" {
		config.Token = consulToken
	}

	if consulScheme := os.Getenv("CONSUL_SCHEME"); consulScheme != "" {
		config.Scheme = consulScheme
	}

	return config
}

// CreateDefaultRegistry 创建默认注册器
func (rf *RegistryFactory) CreateDefaultRegistry() (ServiceRegistry, error) {
	// 默认使用Consul注册器
	return rf.CreateConsulRegistry()
}

// GetRegistryType 获取注册器类型
func (rf *RegistryFactory) GetRegistryType() string {
	registryType := os.Getenv("REGISTRY_TYPE")
	if registryType == "" {
		registryType = "consul" // 默认类型
	}
	return registryType
}
