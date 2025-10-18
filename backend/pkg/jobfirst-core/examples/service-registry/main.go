package main

import (
	"fmt"
	"time"

	"github.com/jobfirst/jobfirst-core/service/registry"
)

func main() {
	fmt.Println("=== JobFirst Core Service Registry Example ===")

	// 创建配置
	config := &registry.RegistryConfig{
		CheckInterval: 30 * time.Second,
		Timeout:       5 * time.Second,
		TTL:           60 * time.Second,
	}

	// 创建服务注册中心
	serviceRegistry, err := registry.NewSimpleServiceRegistry(config)
	if err != nil {
		fmt.Printf("Failed to create registry: %v\n", err)
		return
	}
	defer serviceRegistry.Close()

	// 注册服务
	services := []*registry.ServiceInfo{
		{
			ID:       "api-gateway-1",
			Name:     "api-gateway",
			Version:  "1.0.0",
			Endpoint: "localhost:8080",
			Tags:     []string{"gateway", "api"},
			Health: &registry.HealthStatus{
				Status:    "healthy",
				Message:   "all checks passed",
				Timestamp: time.Now(),
			},
		},
		{
			ID:       "user-service-1",
			Name:     "user-service",
			Version:  "1.0.0",
			Endpoint: "localhost:8081",
			Tags:     []string{"user", "auth"},
			Health: &registry.HealthStatus{
				Status:    "healthy",
				Message:   "all checks passed",
				Timestamp: time.Now(),
			},
		},
		{
			ID:       "resume-service-1",
			Name:     "resume-service",
			Version:  "1.0.0",
			Endpoint: "localhost:8082",
			Tags:     []string{"resume", "document"},
			Health: &registry.HealthStatus{
				Status:    "unhealthy",
				Message:   "database connection failed",
				Timestamp: time.Now(),
			},
		},
	}

	// 注册所有服务
	fmt.Println("\n1. Registering services...")
	for _, service := range services {
		err = serviceRegistry.Register(service)
		if err != nil {
			fmt.Printf("   ❌ Failed to register service %s: %v\n", service.ID, err)
		} else {
			fmt.Printf("   ✅ Successfully registered service: %s\n", service.ID)
		}
	}

	// 获取所有服务
	allServices := serviceRegistry.GetServices()
	fmt.Printf("\n2. Total services registered: %d\n", len(allServices))

	// 根据名称获取服务
	apiServices, err := serviceRegistry.GetServicesByName("api-gateway")
	if err != nil {
		fmt.Printf("   ❌ Failed to get api-gateway services: %v\n", err)
	} else {
		fmt.Printf("   ✅ API Gateway services: %d\n", len(apiServices))
	}

	// 选择健康的服务
	fmt.Println("\n3. Service selection...")
	selectedService, err := serviceRegistry.SelectService("user-service")
	if err != nil {
		fmt.Printf("   ❌ Failed to select user-service: %v\n", err)
	} else {
		fmt.Printf("   ✅ Selected service: %s (%s)\n", selectedService.ID, selectedService.Endpoint)
	}

	// 获取注册中心状态
	status := serviceRegistry.GetRegistryStatus()
	fmt.Printf("\n4. Registry Status:\n")
	fmt.Printf("   - Total services: %v\n", status["total_services"])
	fmt.Printf("   - Healthy services: %v\n", status["healthy_services"])
	fmt.Printf("   - Unhealthy services: %v\n", status["unhealthy_services"])

	// 更新服务健康状态
	fmt.Println("\n5. Updating service health...")
	health := &registry.HealthStatus{
		Status:    "healthy",
		Message:   "recovered from database issue",
		Timestamp: time.Now(),
		Details: map[string]string{
			"database": "connected",
			"memory":   "normal",
		},
	}

	err = serviceRegistry.UpdateServiceHealth("resume-service-1", health)
	if err != nil {
		fmt.Printf("   ❌ Failed to update health: %v\n", err)
	} else {
		fmt.Printf("   ✅ Successfully updated health for resume-service-1\n")
	}

	// 再次选择服务
	selectedService, err = serviceRegistry.SelectService("resume-service")
	if err != nil {
		fmt.Printf("   ❌ Failed to select resume-service: %v\n", err)
	} else {
		fmt.Printf("   ✅ Selected service after health update: %s (%s)\n", selectedService.ID, selectedService.Endpoint)
	}

	// 注销服务
	fmt.Println("\n6. Deregistering service...")
	err = serviceRegistry.Deregister("api-gateway-1")
	if err != nil {
		fmt.Printf("   ❌ Failed to deregister service: %v\n", err)
	} else {
		fmt.Printf("   ✅ Successfully deregistered api-gateway-1\n")
	}

	// 最终状态
	finalStatus := serviceRegistry.GetRegistryStatus()
	fmt.Printf("\n7. Final Registry Status:\n")
	fmt.Printf("   - Total services: %v\n", finalStatus["total_services"])
	fmt.Printf("   - Healthy services: %v\n", finalStatus["healthy_services"])
	fmt.Printf("   - Unhealthy services: %v\n", finalStatus["unhealthy_services"])

	fmt.Println("\n=== Service Registry Example Completed ===")
}
