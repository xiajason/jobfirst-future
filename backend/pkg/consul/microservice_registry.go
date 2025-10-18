package consul

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xiajason/zervi-basic/basic/backend/pkg/config"
)

// MicroserviceInfo 微服务信息
type MicroserviceInfo struct {
	Name        string            `json:"name"`
	ID          string            `json:"id"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	HealthCheck string            `json:"health_check"`
	Status      string            `json:"status"`
	LastSeen    time.Time         `json:"last_seen"`
	Metadata    map[string]string `json:"metadata"`
}

// MicroserviceRegistry 微服务注册器
type MicroserviceRegistry struct {
	consulManager *ServiceManager
	services      map[string]*MicroserviceInfo
	mu            sync.RWMutex
}

// NewMicroserviceRegistry 创建微服务注册器
func NewMicroserviceRegistry(consulManager *ServiceManager) *MicroserviceRegistry {
	return &MicroserviceRegistry{
		consulManager: consulManager,
		services:      make(map[string]*MicroserviceInfo),
	}
}

// RegisterMicroservice 注册微服务
func (r *MicroserviceRegistry) RegisterMicroservice(service *MicroserviceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 注册到Consul
	err := r.consulManager.RegisterService(
		service.Name,
		service.ID,
		service.Address,
		service.Port,
		service.Tags,
	)
	if err != nil {
		return fmt.Errorf("failed to register microservice %s: %v", service.Name, err)
	}

	// 保存到本地缓存
	service.LastSeen = time.Now()
	service.Status = "registered"
	r.services[service.ID] = service

	log.Printf("Microservice %s registered successfully", service.Name)
	return nil
}

// DeregisterMicroservice 注销微服务
func (r *MicroserviceRegistry) DeregisterMicroservice(serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 从Consul注销
	err := r.consulManager.DeregisterService(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister microservice %s: %v", serviceID, err)
	}

	// 从本地缓存移除
	delete(r.services, serviceID)

	log.Printf("Microservice %s deregistered successfully", serviceID)
	return nil
}

// GetMicroservice 获取微服务信息
func (r *MicroserviceRegistry) GetMicroservice(serviceID string) (*MicroserviceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("microservice not found: %s", serviceID)
	}

	return service, nil
}

// ListMicroservices 列出所有微服务
func (r *MicroserviceRegistry) ListMicroservices() []*MicroserviceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []*MicroserviceInfo
	for _, service := range r.services {
		services = append(services, service)
	}

	return services
}

// UpdateMicroserviceStatus 更新微服务状态
func (r *MicroserviceRegistry) UpdateMicroserviceStatus(serviceID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[serviceID]
	if !exists {
		return fmt.Errorf("microservice not found: %s", serviceID)
	}

	service.Status = status
	service.LastSeen = time.Now()

	log.Printf("Microservice %s status updated to: %s", serviceID, status)
	return nil
}

// HealthCheckAll 对所有微服务执行健康检查
func (r *MicroserviceRegistry) HealthCheckAll() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]string)

	for serviceID, service := range r.services {
		// 执行健康检查
		err := r.consulManager.HealthCheck(serviceID)
		if err != nil {
			results[serviceID] = "unhealthy"
			log.Printf("Health check failed for microservice %s: %v", serviceID, err)
		} else {
			results[serviceID] = "healthy"
			service.LastSeen = time.Now()
		}
	}

	return results
}

// GetServiceDiscoveryInfo 获取服务发现信息
func (r *MicroserviceRegistry) GetServiceDiscoveryInfo() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info := map[string]interface{}{
		"total_services": len(r.services),
		"consul_status":  r.consulManager.GetConsulStatus(),
		"services":       r.services,
	}

	return info
}

// StartPeriodicHealthCheck 启动定期健康检查
func (r *MicroserviceRegistry) StartPeriodicHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting periodic health check with interval %v", interval)

		for {
			select {
			case <-ticker.C:
				results := r.HealthCheckAll()
				log.Printf("Health check results: %v", results)
			}
		}
	}()
}

// RegisterDefaultServices 注册默认的微服务
func (r *MicroserviceRegistry) RegisterDefaultServices(cfg *config.Config) error {
	// 注册主后端服务（API网关）
	basicServerService := &MicroserviceInfo{
		Name:        "basic-server",
		ID:          "basic-server-1",
		Address:     "localhost",
		Port:        8080,
		Tags:        []string{"api-gateway", "auth", "user", "resume", "job", "banner", "statistics"},
		HealthCheck: "/health",
		Status:      "pending",
		Metadata: map[string]string{
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"mode":        cfg.Mode,
			"apis":        "auth/login,auth/register,auth/wechat-login,user/sendCode,user/wechatRegister,banner/list,job/recommend,statistics/market,users,resumes,jobs,points",
		},
	}

	if err := r.RegisterMicroservice(basicServerService); err != nil {
		log.Printf("Warning: Failed to register basic-server service: %v", err)
	}

	// 注册用户服务
	userService := &MicroserviceInfo{
		Name:        "user-service",
		ID:          "user-service-1",
		Address:     "localhost",
		Port:        8081,
		Tags:        []string{"user", "auth", "profile", "chat", "points", "notifications"},
		HealthCheck: "/health",
		Status:      "pending",
		Metadata: map[string]string{
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"mode":        cfg.Mode,
			"apis":        "auth/login,auth/register,auth/check,user/profile,user/logout,jobs,companies,banners,chat,points,notifications",
		},
	}

	if err := r.RegisterMicroservice(userService); err != nil {
		log.Printf("Warning: Failed to register user service: %v", err)
	}

	// 注册简历服务
	resumeService := &MicroserviceInfo{
		Name:        "resume-service",
		ID:          "resume-service-1",
		Address:     "localhost",
		Port:        8082,
		Tags:        []string{"resume", "template", "ai", "upload", "preview"},
		HealthCheck: "/health",
		Status:      "pending",
		Metadata: map[string]string{
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"mode":        cfg.Mode,
			"apis":        "resume/templates,resume/banners,resume/list,resume/detail,resume/create,resume/update,resume/delete,resume/upload,resume/preview,resume/auth,resume/blacklist",
		},
	}

	if err := r.RegisterMicroservice(resumeService); err != nil {
		log.Printf("Warning: Failed to register resume service: %v", err)
	}

	// 注册AI服务
	aiService := &MicroserviceInfo{
		Name:        "ai-service",
		ID:          "ai-service-1",
		Address:     "localhost",
		Port:        8206,
		Tags:        []string{"ai", "ml", "vector", "analysis", "chat"},
		HealthCheck: "/health",
		Status:      "pending",
		Metadata: map[string]string{
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"mode":        cfg.Mode,
			"apis":        "analyze/resume,vectors,vectors/search,chat,health",
		},
	}

	if err := r.RegisterMicroservice(aiService); err != nil {
		log.Printf("Warning: Failed to register AI service: %v", err)
	}

	// 注册Job服务
	jobService := &MicroserviceInfo{
		Name:        "job-service",
		ID:          "job-service-1",
		Address:     "localhost",
		Port:        8089,
		Tags:        []string{"job", "company", "application", "matching"},
		HealthCheck: "/health",
		Status:      "pending",
		Metadata: map[string]string{
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"mode":        cfg.Mode,
			"apis":        "job/public/jobs,job/public/companies,job/jobs,job/applications,job/admin",
		},
	}

	if err := r.RegisterMicroservice(jobService); err != nil {
		log.Printf("Warning: Failed to register job service: %v", err)
	}

	log.Printf("Default microservices registration completed")
	return nil
}

// DiscoverService 发现服务
func (r *MicroserviceRegistry) DiscoverService(serviceName string) ([]*MicroserviceInfo, error) {
	// 从Consul发现服务
	services, err := r.consulManager.DiscoverService(serviceName)
	if err != nil {
		return nil, err
	}

	var microservices []*MicroserviceInfo
	for _, service := range services {
		microservice := &MicroserviceInfo{
			Name:        service.Service.Service,
			ID:          service.Service.ID,
			Address:     service.Service.Address,
			Port:        service.Service.Port,
			Tags:        service.Service.Tags,
			HealthCheck: service.Service.Address,
			Status:      service.Checks.AggregatedStatus(),
			LastSeen:    time.Now(),
		}
		microservices = append(microservices, microservice)
	}

	return microservices, nil
}

// GetServiceEndpoints 获取服务端点信息
func (r *MicroserviceRegistry) GetServiceEndpoints() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	endpoints := make(map[string]string)
	for _, service := range r.services {
		endpoints[service.Name] = fmt.Sprintf("http://%s:%d", service.Address, service.Port)
	}

	return endpoints
}
