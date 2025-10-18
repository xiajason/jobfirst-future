package system

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// ServiceDefinition 服务定义
type ServiceDefinition struct {
	Name        string `json:"name"`
	Port        int    `json:"port"`
	HealthPath  string `json:"health_path"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Version     string `json:"version"`
}

// GetServiceDefinitions 获取所有服务定义
func GetServiceDefinitions() map[string]ServiceDefinition {
	return map[string]ServiceDefinition{
		// 基础设施服务
		"mysql": {
			Name:        "mysql",
			Port:        3306,
			HealthPath:  "/",
			Description: "MySQL数据库服务",
			Category:    "infrastructure",
			Version:     "8.0",
		},
		"redis": {
			Name:        "redis",
			Port:        6379,
			HealthPath:  "/",
			Description: "Redis缓存服务",
			Category:    "infrastructure",
			Version:     "7.0",
		},
		"consul": {
			Name:        "consul",
			Port:        8500,
			HealthPath:  "/v1/status/leader",
			Description: "Consul服务发现",
			Category:    "infrastructure",
			Version:     "1.9",
		},
		"nginx": {
			Name:        "nginx",
			Port:        80,
			HealthPath:  "/",
			Description: "Nginx反向代理",
			Category:    "infrastructure",
			Version:     "1.18",
		},

		// 核心微服务 (重构后)
		"basic-server": {
			Name:        "basic-server",
			Port:        8080,
			HealthPath:  "/health",
			Description: "基础服务器 - API网关",
			Category:    "core",
			Version:     "3.1.0",
		},
		"user-service": {
			Name:        "user-service",
			Port:        8081,
			HealthPath:  "/health",
			Description: "用户管理服务",
			Category:    "core",
			Version:     "3.1.0",
		},
		"company-service": {
			Name:        "company-service",
			Port:        8083,
			HealthPath:  "/health",
			Description: "公司管理服务",
			Category:    "core",
			Version:     "3.1.0",
		},
		"dev-team-service": {
			Name:        "dev-team-service",
			Port:        8088,
			HealthPath:  "/health",
			Description: "开发团队管理服务",
			Category:    "core",
			Version:     "3.1.0",
		},

		// 重构后的微服务 (v3.1.1 端口配置修正)
		"template-service": {
			Name:        "template-service",
			Port:        8085, // 修正：Template Service 端口 8085
			HealthPath:  "/health",
			Description: "模板管理服务 - 支持评分、搜索、统计",
			Category:    "refactored",
			Version:     "3.1.1",
		},
		"statistics-service": {
			Name:        "statistics-service",
			Port:        8086, // 保持：Statistics Service 端口 8086
			HealthPath:  "/health",
			Description: "数据统计服务 - 系统分析和趋势监控",
			Category:    "refactored",
			Version:     "3.1.1",
		},
		"banner-service": {
			Name:        "banner-service",
			Port:        8087, // 修正：Banner Service 端口 8087
			HealthPath:  "/health",
			Description: "内容管理服务 - Banner、Markdown、评论",
			Category:    "refactored",
			Version:     "3.1.1",
		},

		// 其他微服务
		"resume-service": {
			Name:        "resume-service",
			Port:        8082,
			HealthPath:  "/health",
			Description: "简历管理服务",
			Category:    "business",
			Version:     "3.1.0",
		},
		"notification-service": {
			Name:        "notification-service",
			Port:        8084,
			HealthPath:  "/health",
			Description: "通知服务",
			Category:    "business",
			Version:     "3.1.0",
		},
		"ai-service": {
			Name:        "ai-service",
			Port:        8206,
			HealthPath:  "/health",
			Description: "AI服务 - Python/Sanic",
			Category:    "ai",
			Version:     "3.1.0",
		},
	}
}

// CheckServiceHealth 检查服务健康状态
func CheckServiceHealth(service ServiceDefinition) *ServiceStatus {
	status := &ServiceStatus{
		Name:      service.Name,
		Port:      service.Port,
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	if !isPortOpen("localhost", service.Port) {
		status.Status = "error"
		status.Error = fmt.Sprintf("端口 %d 未开放", service.Port)
		return status
	}

	// 对于HTTP服务，检查健康端点
	if service.HealthPath != "/" {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		url := fmt.Sprintf("http://localhost:%d%s", service.Port, service.HealthPath)
		resp, err := client.Get(url)

		if err != nil {
			status.Status = "error"
			status.Error = fmt.Sprintf("健康检查失败: %v", err)
			return status
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status.Status = "active"
		} else {
			status.Status = "warning"
			status.Error = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
		}
	} else {
		// 对于非HTTP服务，仅检查端口
		status.Status = "active"
	}

	return status
}

// GetServicesByCategory 按分类获取服务
func GetServicesByCategory(category string) map[string]ServiceDefinition {
	allServices := GetServiceDefinitions()
	categoryServices := make(map[string]ServiceDefinition)

	for name, service := range allServices {
		if service.Category == category {
			categoryServices[name] = service
		}
	}

	return categoryServices
}

// isPortOpen 检查端口是否开放
func isPortOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GetRefactoredServices 获取重构后的服务
func GetRefactoredServices() map[string]ServiceDefinition {
	return GetServicesByCategory("refactored")
}

// GetCoreServices 获取核心服务
func GetCoreServices() map[string]ServiceDefinition {
	return GetServicesByCategory("core")
}

// GetInfrastructureServices 获取基础设施服务
func GetInfrastructureServices() map[string]ServiceDefinition {
	return GetServicesByCategory("infrastructure")
}
