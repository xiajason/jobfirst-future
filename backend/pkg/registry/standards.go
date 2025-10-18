package registry

import (
	"fmt"
	"strings"
)

// ServiceRegistryStandards 服务注册规范
type ServiceRegistryStandards struct{}

// NewServiceRegistryStandards 创建服务注册规范
func NewServiceRegistryStandards() *ServiceRegistryStandards {
	return &ServiceRegistryStandards{}
}

// ValidateServiceName 验证服务名称
func (srs *ServiceRegistryStandards) ValidateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("服务名称不能为空")
	}

	// 服务名称格式: {domain}-service (API Gateway除外)
	if !strings.HasSuffix(name, "-service") && name != "api-gateway" {
		return fmt.Errorf("服务名称必须以 '-service' 结尾（API Gateway除外）")
	}

	// 检查是否包含非法字符
	if strings.ContainsAny(name, " \t\n\r") {
		return fmt.Errorf("服务名称不能包含空白字符")
	}

	return nil
}

// ValidateServiceID 验证服务ID
func (srs *ServiceRegistryStandards) ValidateServiceID(id string) error {
	if id == "" {
		return fmt.Errorf("服务ID不能为空")
	}

	// 服务ID格式: {service-name}-{instance-id}
	parts := strings.Split(id, "-")
	if len(parts) < 2 {
		return fmt.Errorf("服务ID格式错误，应为: {service-name}-{instance-id}")
	}

	return nil
}

// ValidateTags 验证标签
func (srs *ServiceRegistryStandards) ValidateTags(tags []string) error {
	requiredTags := []string{"service_type", "version"}

	for _, requiredTag := range requiredTags {
		found := false
		for _, tag := range tags {
			if strings.HasPrefix(tag, requiredTag+":") {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("缺少必需标签: %s", requiredTag)
		}
	}

	return nil
}

// ValidateMetadata 验证元数据
func (srs *ServiceRegistryStandards) ValidateMetadata(meta map[string]string) error {
	requiredKeys := []string{"version", "type", "environment"}

	for _, key := range requiredKeys {
		if _, exists := meta[key]; !exists {
			return fmt.Errorf("缺少必需元数据: %s", key)
		}
	}

	// 验证版本号格式
	if version, exists := meta["version"]; exists {
		if !srs.isValidVersion(version) {
			return fmt.Errorf("版本号格式错误: %s，应为 v1.0.0 格式", version)
		}
	}

	// 验证服务类型
	if serviceType, exists := meta["type"]; exists {
		validTypes := []string{"api-gateway", "microservice", "database", "cache", "queue"}
		if !srs.isValidServiceType(serviceType, validTypes) {
			return fmt.Errorf("无效的服务类型: %s，有效类型: %v", serviceType, validTypes)
		}
	}

	// 验证环境标识
	if environment, exists := meta["environment"]; exists {
		validEnvironments := []string{"development", "staging", "production"}
		if !srs.isValidEnvironment(environment, validEnvironments) {
			return fmt.Errorf("无效的环境标识: %s，有效环境: %v", environment, validEnvironments)
		}
	}

	return nil
}

// isValidVersion 验证版本号格式
func (srs *ServiceRegistryStandards) isValidVersion(version string) bool {
	// 简单验证版本号格式 v1.0.0
	return strings.HasPrefix(version, "v") && len(strings.Split(version, ".")) == 3
}

// isValidServiceType 验证服务类型
func (srs *ServiceRegistryStandards) isValidServiceType(serviceType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if serviceType == validType {
			return true
		}
	}
	return false
}

// isValidEnvironment 验证环境标识
func (srs *ServiceRegistryStandards) isValidEnvironment(environment string, validEnvironments []string) bool {
	for _, validEnv := range validEnvironments {
		if environment == validEnv {
			return true
		}
	}
	return false
}

// GenerateServiceID 生成服务ID
func (srs *ServiceRegistryStandards) GenerateServiceID(serviceName, instanceID string) string {
	return fmt.Sprintf("%s-%s", serviceName, instanceID)
}

// GenerateInstanceID 生成实例ID
func (srs *ServiceRegistryStandards) GenerateInstanceID(hostname string, port int) string {
	return fmt.Sprintf("%s-%d", hostname, port)
}

// GetDefaultTags 获取默认标签
func (srs *ServiceRegistryStandards) GetDefaultTags(serviceType, version string) []string {
	return []string{
		fmt.Sprintf("service_type:%s", serviceType),
		fmt.Sprintf("version:%s", version),
		"jobfirst",
		"future",
	}
}

// GetDefaultMetadata 获取默认元数据
func (srs *ServiceRegistryStandards) GetDefaultMetadata(serviceType, version, environment string) map[string]string {
	return map[string]string{
		"version":     version,
		"type":        serviceType,
		"environment": environment,
		"mode":        "future",
		"framework":   "gin",
		"language":    "go",
	}
}

// GetHealthCheckConfig 获取健康检查配置
func (srs *ServiceRegistryStandards) GetHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		HTTPPath:                       "/health",
		Interval:                       "10s",
		Timeout:                        "3s",
		DeregisterCriticalServiceAfter: "30s",
	}
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	HTTPPath                       string
	Interval                       string
	Timeout                        string
	DeregisterCriticalServiceAfter string
}
