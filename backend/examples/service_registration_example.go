package main

import (
	"log"
	"time"

	"github.com/xiajason/zervi-basic/basic/backend/pkg/registry"
)

func main() {
	log.Println("🚀 服务注册规范使用示例")

	// 创建服务注册器
	registryFactory := registry.NewRegistryFactory()
	serviceRegistry, err := registryFactory.CreateDefaultRegistry()
	if err != nil {
		log.Fatalf("❌ 创建服务注册器失败: %v", err)
	}

	// 创建服务注册助手
	helper := registry.NewServiceRegistrationHelper()

	// 创建用户服务
	userService, err := helper.CreateUserService(7530)
	if err != nil {
		log.Fatalf("❌ 创建用户服务失败: %v", err)
	}

	// 验证服务信息
	if err := helper.ValidateServiceInfo(userService); err != nil {
		log.Fatalf("❌ 服务信息验证失败: %v", err)
	}

	// 注册服务
	err = serviceRegistry.Register(userService)
	if err != nil {
		log.Fatalf("❌ 注册用户服务失败: %v", err)
	}

	log.Printf("✅ 用户服务已注册: %s", userService.ID)

	// 等待一段时间
	time.Sleep(2 * time.Second)

	// 发现服务
	services, err := serviceRegistry.Discover("user-service")
	if err != nil {
		log.Fatalf("❌ 发现用户服务失败: %v", err)
	}

	log.Printf("✅ 发现 %d 个用户服务实例", len(services))

	// 注销服务
	err = serviceRegistry.Deregister(userService.ID)
	if err != nil {
		log.Fatalf("❌ 注销用户服务失败: %v", err)
	}

	log.Println("✅ 用户服务已注销")
}
