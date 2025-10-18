package registry

import (
	"time"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Address      string            `json:"address"`
	Port         int               `json:"port"`
	Endpoint     string            `json:"endpoint"`
	Health       *HealthStatus     `json:"health"`
	Metadata     map[string]string `json:"metadata"`
	LastCheck    time.Time         `json:"last_check"`
	RegisteredAt time.Time         `json:"registered_at"`
	Tags         []string          `json:"tags"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string            `json:"status"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Details   map[string]string `json:"details"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	ConsulHost    string        `json:"consul_host"`
	ConsulPort    int           `json:"consul_port"`
	CheckInterval time.Duration `json:"check_interval"`
	Timeout       time.Duration `json:"timeout"`
	TTL           time.Duration `json:"ttl"`
}
