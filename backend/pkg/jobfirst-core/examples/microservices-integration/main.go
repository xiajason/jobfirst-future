package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jobfirst/jobfirst-core/service/registry"
)

// MicroserviceSimulator 模拟微服务
type MicroserviceSimulator struct {
	Name     string
	Port     int
	Healthy  bool
	server   *http.Server
	registry *registry.SimpleServiceRegistry
}

// NewMicroserviceSimulator 创建微服务模拟器
func NewMicroserviceSimulator(name string, port int, registry *registry.SimpleServiceRegistry) *MicroserviceSimulator {
	return &MicroserviceSimulator{
		Name:     name,
		Port:     port,
		Healthy:  true,
		registry: registry,
	}
}

// Start 启动微服务
func (ms *MicroserviceSimulator) Start() error {
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if ms.Healthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy","service":"` + ms.Name + `"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","service":"` + ms.Name + `"}`))
		}
	})

	// 服务信息端点
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"name": "` + ms.Name + `",
			"port": ` + fmt.Sprintf("%d", ms.Port) + `,
			"healthy": ` + fmt.Sprintf("%t", ms.Healthy) + `,
			"timestamp": "` + time.Now().Format(time.RFC3339) + `"
		}`))
	})

	ms.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ms.Port),
		Handler: mux,
	}

	// 启动HTTP服务器
	go func() {
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start %s: %v", ms.Name, err)
		}
	}()

	// 注册到服务注册中心
	serviceInfo := &registry.ServiceInfo{
		ID:       ms.Name + "-" + fmt.Sprintf("%d", ms.Port),
		Name:     ms.Name,
		Version:  "1.0.0",
		Endpoint: fmt.Sprintf("localhost:%d", ms.Port),
		Tags:     []string{"microservice", "simulator"},
		Health: &registry.HealthStatus{
			Status:    "healthy",
			Message:   "service started successfully",
			Timestamp: time.Now(),
		},
		Metadata: map[string]string{
			"port":     fmt.Sprintf("%d", ms.Port),
			"type":     "simulator",
			"language": "go",
		},
	}

	if err := ms.registry.Register(serviceInfo); err != nil {
		return fmt.Errorf("failed to register %s: %v", ms.Name, err)
	}

	log.Printf("✅ %s started on port %d and registered to service registry", ms.Name, ms.Port)
	return nil
}

// Stop 停止微服务
func (ms *MicroserviceSimulator) Stop() error {
	// 从服务注册中心注销
	if err := ms.registry.Deregister(ms.Name + "-" + fmt.Sprintf("%d", ms.Port)); err != nil {
		log.Printf("Failed to deregister %s: %v", ms.Name, err)
	}

	// 停止HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ms.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown %s: %v", ms.Name, err)
	}

	log.Printf("✅ %s stopped and deregistered from service registry", ms.Name)
	return nil
}

// SetHealthy 设置健康状态
func (ms *MicroserviceSimulator) SetHealthy(healthy bool) {
	ms.Healthy = healthy
	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	// 更新服务注册中心的健康状态
	health := &registry.HealthStatus{
		Status:    status,
		Message:   fmt.Sprintf("health status changed to %s", status),
		Timestamp: time.Now(),
		Details: map[string]string{
			"port":    fmt.Sprintf("%d", ms.Port),
			"healthy": fmt.Sprintf("%t", healthy),
		},
	}

	if err := ms.registry.UpdateServiceHealth(ms.Name+"-"+fmt.Sprintf("%d", ms.Port), health); err != nil {
		log.Printf("Failed to update health for %s: %v", ms.Name, err)
	} else {
		log.Printf("✅ %s health status updated to %s", ms.Name, status)
	}
}

func main() {
	fmt.Println("=== JobFirst Core Microservices Integration Test ===")

	// 创建服务注册中心
	config := &registry.RegistryConfig{
		CheckInterval: 10 * time.Second,
		Timeout:       5 * time.Second,
		TTL:           30 * time.Second,
	}

	serviceRegistry, err := registry.NewSimpleServiceRegistry(config)
	if err != nil {
		log.Fatalf("Failed to create service registry: %v", err)
	}
	defer serviceRegistry.Close()

	// 创建微服务模拟器
	services := []*MicroserviceSimulator{
		NewMicroserviceSimulator("api-gateway", 8080, serviceRegistry),
		NewMicroserviceSimulator("user-service", 8081, serviceRegistry),
		NewMicroserviceSimulator("resume-service", 8082, serviceRegistry),
		NewMicroserviceSimulator("company-service", 8083, serviceRegistry),
		NewMicroserviceSimulator("banner-service", 8084, serviceRegistry),
		NewMicroserviceSimulator("template-service", 8085, serviceRegistry),
		NewMicroserviceSimulator("notification-service", 8086, serviceRegistry),
		NewMicroserviceSimulator("statistics-service", 8087, serviceRegistry),
	}

	// 启动所有微服务
	fmt.Println("\n1. Starting microservices...")
	for _, service := range services {
		if err := service.Start(); err != nil {
			log.Printf("Failed to start %s: %v", service.Name, err)
		}
		time.Sleep(100 * time.Millisecond) // 避免端口冲突
	}

	// 等待服务启动
	time.Sleep(2 * time.Second)

	// 测试服务发现
	fmt.Println("\n2. Testing service discovery...")
	allServices := serviceRegistry.GetServices()
	fmt.Printf("   Total registered services: %d\n", len(allServices))

	// 测试特定服务发现
	apiServices, err := serviceRegistry.GetServicesByName("api-gateway")
	if err != nil {
		log.Printf("Failed to get api-gateway services: %v", err)
	} else {
		fmt.Printf("   API Gateway services: %d\n", len(apiServices))
	}

	// 测试负载均衡
	fmt.Println("\n3. Testing load balancing...")
	for i := 0; i < 5; i++ {
		selectedService, err := serviceRegistry.SelectService("user-service")
		if err != nil {
			log.Printf("Failed to select user-service: %v", err)
		} else {
			fmt.Printf("   Selected service: %s (%s)\n", selectedService.ID, selectedService.Endpoint)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 测试健康状态变化
	fmt.Println("\n4. Testing health status changes...")

	// 模拟服务故障
	services[2].SetHealthy(false) // resume-service
	time.Sleep(1 * time.Second)

	// 检查注册中心状态
	status := serviceRegistry.GetRegistryStatus()
	fmt.Printf("   Registry status after failure:\n")
	fmt.Printf("   - Total services: %v\n", status["total_services"])
	fmt.Printf("   - Healthy services: %v\n", status["healthy_services"])
	fmt.Printf("   - Unhealthy services: %v\n", status["unhealthy_services"])

	// 尝试选择故障服务
	selectedService, err := serviceRegistry.SelectService("resume-service")
	if err != nil {
		fmt.Printf("   ✅ Correctly failed to select unhealthy resume-service: %v\n", err)
	} else {
		fmt.Printf("   ❌ Unexpectedly selected unhealthy service: %s\n", selectedService.ID)
	}

	// 恢复服务
	fmt.Println("\n5. Testing service recovery...")
	services[2].SetHealthy(true) // resume-service
	time.Sleep(1 * time.Second)

	// 再次尝试选择服务
	selectedService, err = serviceRegistry.SelectService("resume-service")
	if err != nil {
		fmt.Printf("   ❌ Failed to select recovered resume-service: %v\n", err)
	} else {
		fmt.Printf("   ✅ Successfully selected recovered service: %s (%s)\n", selectedService.ID, selectedService.Endpoint)
	}

	// 最终状态
	finalStatus := serviceRegistry.GetRegistryStatus()
	fmt.Printf("\n6. Final registry status:\n")
	fmt.Printf("   - Total services: %v\n", finalStatus["total_services"])
	fmt.Printf("   - Healthy services: %v\n", finalStatus["healthy_services"])
	fmt.Printf("   - Unhealthy services: %v\n", finalStatus["unhealthy_services"])

	// 测试HTTP健康检查
	fmt.Println("\n7. Testing HTTP health checks...")
	for _, service := range services {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", service.Port))
		if err != nil {
			fmt.Printf("   ❌ %s health check failed: %v\n", service.Name, err)
		} else {
			resp.Body.Close()
			fmt.Printf("   ✅ %s health check passed (status: %d)\n", service.Name, resp.StatusCode)
		}
	}

	// 停止所有服务
	fmt.Println("\n8. Stopping all services...")
	for _, service := range services {
		if err := service.Stop(); err != nil {
			log.Printf("Failed to stop %s: %v", service.Name, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 最终检查
	finalCheck := serviceRegistry.GetServices()
	fmt.Printf("\n9. Final service count: %d\n", len(finalCheck))

	fmt.Println("\n=== Microservices Integration Test Completed ===")
}
