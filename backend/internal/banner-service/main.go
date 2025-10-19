package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/jobfirst/jobfirst-core"
)

func main() {
	// 从环境变量获取端口，默认为8087
	port := os.Getenv("BANNER_SERVICE_PORT")
	if port == "" {
		port = "8087"
	}
	portInt, _ := strconv.Atoi(port)

	// 设置进程名称
	if len(os.Args) > 0 {
		os.Args[0] = "banner-service"
	}

	// 初始化JobFirst核心包
	core, err := jobfirst.NewCore("../../configs/jobfirst-core-config.yaml")
	if err != nil {
		log.Fatalf("初始化JobFirst核心包失败: %v", err)
	}
	defer core.Close()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.Default()

	// 添加健康检查端点
	r.GET("/health", healthCheck)
	r.GET("/info", serviceInfo)

	// 注册到Consul
	registerToConsul("banner-service", "127.0.0.1", portInt)

	// 启动服务
	log.Printf("Starting Banner Service with jobfirst-core on 0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Banner Service启动失败: %v", err)
	}
}

func registerToConsul(serviceName, serviceHost string, servicePort int) {
	// 创建Consul客户端
	config := api.DefaultConfig()
	config.Address = "localhost:8500"
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("创建Consul客户端失败: %v", err)
		return
	}

	// 服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%d", serviceName, servicePort),
		Name:    serviceName,
		Address: serviceHost,
		Port:    servicePort,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", serviceHost, servicePort),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Tags: []string{"banner", "microservice"},
		Meta: map[string]string{
			"version":     "3.1.0",
			"environment": "production",
			"port":        "7535",
		},
	}

	// 注册服务
	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Printf("注册服务到Consul失败: %v", err)
	} else {
		log.Printf("服务 %s 已注册到Consul: %s:%d", serviceName, serviceHost, servicePort)
	}
}

// 健康检查端点
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "banner-service",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "3.1.0",
	})
}

// 服务信息端点
func serviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "banner-service",
		"version":   "3.1.0",
		"port":      7535,
		"status":    "running",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
