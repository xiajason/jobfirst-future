package registry

import (
	"fmt"
	"os"
	"strconv"
)

// ServiceRegistrationHelper 服务注册助手
type ServiceRegistrationHelper struct {
	standards *ServiceRegistryStandards
}

// NewServiceRegistrationHelper 创建服务注册助手
func NewServiceRegistrationHelper() *ServiceRegistrationHelper {
	return &ServiceRegistrationHelper{
		standards: NewServiceRegistryStandards(),
	}
}

// CreateServiceInfo 创建服务信息
func (srh *ServiceRegistrationHelper) CreateServiceInfo(
	serviceName, serviceType, version, environment string,
	port int,
) (*ServiceInfo, error) {
	// 生成服务ID
	hostname, _ := os.Hostname()
	instanceID := srh.standards.GenerateInstanceID(hostname, port)
	serviceID := srh.standards.GenerateServiceID(serviceName, instanceID)

	// 获取默认标签
	tags := srh.standards.GetDefaultTags(serviceType, version)

	// 获取默认元数据
	meta := srh.standards.GetDefaultMetadata(serviceType, version, environment)

	// 获取健康检查配置
	healthConfig := srh.standards.GetHealthCheckConfig()

	// 创建服务信息
	serviceInfo := &ServiceInfo{
		ID:          serviceID,
		Name:        serviceName,
		Address:     "localhost",
		Port:        port,
		Tags:        tags,
		Meta:        meta,
		HealthCheck: healthConfig.HTTPPath,
		Status:      "pending",
	}

	return serviceInfo, nil
}

// CreateAPIGatewayService 创建API Gateway服务
func (srh *ServiceRegistrationHelper) CreateAPIGatewayService(port int) (*ServiceInfo, error) {
	return srh.CreateServiceInfo(
		"api-gateway",
		"api-gateway",
		"v1.0.0",
		"development",
		port,
	)
}

// CreateUserService 创建用户服务
func (srh *ServiceRegistrationHelper) CreateUserService(port int) (*ServiceInfo, error) {
	return srh.CreateServiceInfo(
		"user-service",
		"microservice",
		"v1.0.0",
		"development",
		port,
	)
}

// CreateResumeService 创建简历服务
func (srh *ServiceRegistrationHelper) CreateResumeService(port int) (*ServiceInfo, error) {
	return srh.CreateServiceInfo(
		"resume-service",
		"microservice",
		"v1.0.0",
		"development",
		port,
	)
}

// CreateCompanyService 创建企业服务
func (srh *ServiceRegistrationHelper) CreateCompanyService(port int) (*ServiceInfo, error) {
	return srh.CreateServiceInfo(
		"company-service",
		"microservice",
		"v1.0.0",
		"development",
		port,
	)
}

// CreateJobService 创建职位服务
func (srh *ServiceRegistrationHelper) CreateJobService(port int) (*ServiceInfo, error) {
	return srh.CreateServiceInfo(
		"job-service",
		"microservice",
		"v1.0.0",
		"development",
		port,
	)
}

// CreateNotificationService 创建通知服务
func (srh *ServiceRegistrationHelper) CreateNotificationService(port int) (*ServiceInfo, error) {
	return srh.CreateServiceInfo(
		"notification-service",
		"microservice",
		"v1.0.0",
		"development",
		port,
	)
}

// CreateDevTeamService 创建开发团队服务
func (srh *ServiceRegistrationHelper) CreateDevTeamService(port int) (*ServiceInfo, error) {
	return srh.CreateServiceInfo(
		"dev-team-service",
		"microservice",
		"v1.0.0",
		"development",
		port,
	)
}

// GetPortFromEnv 从环境变量获取端口
func (srh *ServiceRegistrationHelper) GetPortFromEnv(envVar string, defaultValue int) int {
	if portStr := os.Getenv(envVar); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port
		}
	}
	return defaultValue
}

// GetEnvironment 获取环境标识
func (srh *ServiceRegistrationHelper) GetEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	return env
}

// GetVersion 获取版本号
func (srh *ServiceRegistrationHelper) GetVersion() string {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "v1.0.0"
	}
	return version
}

// ValidateServiceInfo 验证服务信息
func (srh *ServiceRegistrationHelper) ValidateServiceInfo(service *ServiceInfo) error {
	if err := srh.standards.ValidateServiceName(service.Name); err != nil {
		return fmt.Errorf("服务名称验证失败: %v", err)
	}

	if err := srh.standards.ValidateServiceID(service.ID); err != nil {
		return fmt.Errorf("服务ID验证失败: %v", err)
	}

	if err := srh.standards.ValidateTags(service.Tags); err != nil {
		return fmt.Errorf("服务标签验证失败: %v", err)
	}

	if err := srh.standards.ValidateMetadata(service.Meta); err != nil {
		return fmt.Errorf("服务元数据验证失败: %v", err)
	}

	return nil
}
