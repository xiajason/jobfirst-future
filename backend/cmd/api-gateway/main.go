package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/health"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/registry"
)

func main() {
	log.Println("🚀 启动JobFirst Future版 API Gateway...")

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 创建统一的服务注册器
	registryFactory := registry.NewRegistryFactory()
	serviceRegistry, err := registryFactory.CreateDefaultRegistry()
	if err != nil {
		log.Fatalf("❌ 创建服务注册器失败: %v", err)
	}

	// 创建健康检查器
	healthChecker, err := health.NewHealthChecker(serviceRegistry, 10*time.Second, 3*time.Second)
	if err != nil {
		log.Fatalf("❌ 创建健康检查器失败: %v", err)
	}

	// 启动健康检查器
	go func() {
		if err := healthChecker.Start(); err != nil {
			log.Printf("❌ 健康检查器启动失败: %v", err)
		}
	}()

	// 创建服务注册助手
	helper := registry.NewServiceRegistrationHelper()

	// 获取端口配置
	port := helper.GetPortFromEnv("API_GATEWAY_PORT", 7521)

	// 创建API Gateway服务信息
	serviceInfo, err := helper.CreateAPIGatewayService(port)
	if err != nil {
		log.Fatalf("❌ 创建服务信息失败: %v", err)
	}

	// 注册服务
	err = serviceRegistry.Register(serviceInfo)
	if err != nil {
		log.Printf("⚠️ 注册API Gateway失败: %v", err)
	} else {
		log.Println("✅ API Gateway已注册到服务注册中心")
	}

	// 创建健康检查处理器
	healthHandler := health.NewHealthHandler(healthChecker, serviceRegistry)

	// 设置路由
	setupRoutes(r, serviceRegistry, healthHandler)

	// 启动服务器
	srv := &http.Server{
		Addr:    ":" + fmt.Sprintf("%d", port),
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ API Gateway启动失败: %v", err)
		}
	}()

	log.Printf("✅ JobFirst Future版 API Gateway 已启动，端口: %d", port)
	log.Printf("🔍 健康检查端点: http://localhost:%d/health", port)
	log.Printf("📊 服务列表端点: http://localhost:%d/api/v1/services", port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭API Gateway...")

	// 停止健康检查器
	healthChecker.Stop()

	// 注销服务
	err = serviceRegistry.Deregister(serviceInfo.ID)
	if err != nil {
		log.Printf("⚠️ 注销API Gateway失败: %v", err)
	} else {
		log.Println("✅ API Gateway已从服务注册中心注销")
	}

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ API Gateway关闭失败: %v", err)
	} else {
		log.Println("✅ API Gateway已成功关闭")
	}
}

func setupRoutes(r *gin.Engine, serviceRegistry registry.ServiceRegistry, healthHandler *health.HealthHandler) {
	// 根路径
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "JobFirst Future版 API Gateway",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// 健康检查路由
	health := r.Group("/health")
	{
		health.GET("", healthHandler.Health)
		health.GET("/ready", healthHandler.Ready)
		health.GET("/live", healthHandler.Live)
		health.GET("/services", healthHandler.Services)
		health.GET("/services/:serviceId", healthHandler.ServiceHealth)
		health.GET("/healthy", healthHandler.HealthyServices)
		health.GET("/unhealthy", healthHandler.UnhealthyServices)
		health.GET("/check/:serviceName", healthHandler.CheckService)
	}

	// 服务列表
	r.GET("/api/v1/services", func(c *gin.Context) {
		services, err := serviceRegistry.ListServices()
		if err != nil {
			c.JSON(500, gin.H{"error": "获取服务列表失败"})
			return
		}

		c.JSON(200, gin.H{
			"services": services,
			"count":    len(services),
		})
	})

	// 服务发现
	r.GET("/api/v1/services/:serviceName", func(c *gin.Context) {
		serviceName := c.Param("serviceName")
		
		services, err := serviceRegistry.Discover(serviceName)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "服务发现失败",
				"details": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"service_name": serviceName,
			"instances":    services,
			"count":        len(services),
		})
	})

	// 服务代理
	r.Any("/api/v1/:serviceName/*path", func(c *gin.Context) {
		serviceName := c.Param("serviceName")
		path := c.Param("path")

		// 从服务注册中心发现服务
		serviceURL, err := serviceRegistry.GetHealthyServiceURL(serviceName)
		if err != nil {
			c.JSON(503, gin.H{
				"error":   "服务不可用",
				"service": serviceName,
				"details": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "服务代理功能",
			"service": serviceName,
			"target":  serviceURL + path,
			"note":    "这是简化版本，实际代理功能需要进一步实现",
		})
	})
}
