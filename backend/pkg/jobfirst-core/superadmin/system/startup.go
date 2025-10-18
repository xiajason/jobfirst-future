package system

import (
	"fmt"
	"net"
	"time"
)

// StartupChecker 系统启动顺序检查器
type StartupChecker struct {
	config *StartupConfig
}

// StartupConfig 启动配置
type StartupConfig struct {
	RequiredServices []ServiceStartupOrder `json:"required_services"`
	Timeout          time.Duration         `json:"timeout"`
	RetryInterval    time.Duration         `json:"retry_interval"`
}

// ServiceStartupOrder 服务启动顺序
type ServiceStartupOrder struct {
	Name         string   `json:"name"`
	Port         int      `json:"port"`
	Priority     int      `json:"priority"`     // 优先级，数字越小优先级越高
	Dependencies []string `json:"dependencies"` // 依赖的服务
	HealthCheck  string   `json:"health_check"` // 健康检查端点
	Description  string   `json:"description"`  // 服务描述
}

// StartupStatus 启动状态
type StartupStatus struct {
	Timestamp       time.Time              `json:"timestamp"`
	OverallStatus   string                 `json:"overall_status"` // success, warning, critical
	Services        []ServiceStartupStatus `json:"services"`
	Violations      []StartupViolation     `json:"violations"`
	Recommendations []string               `json:"recommendations"`
}

// ServiceStartupStatus 单个服务启动状态
type ServiceStartupStatus struct {
	Name          string    `json:"name"`
	Port          int       `json:"port"`
	Status        string    `json:"status"`         // active, inactive, error
	StartupOrder  int       `json:"startup_order"`  // 实际启动顺序
	ExpectedOrder int       `json:"expected_order"` // 期望启动顺序
	StartTime     time.Time `json:"start_time"`
	HealthStatus  string    `json:"health_status"`
	Dependencies  []string  `json:"dependencies"`
}

// StartupViolation 启动违规
type StartupViolation struct {
	Type           string `json:"type"` // dependency, order, timeout
	Service        string `json:"service"`
	Message        string `json:"message"`
	Severity       string `json:"severity"` // high, medium, low
	Recommendation string `json:"recommendation"`
}

// NewStartupChecker 创建启动检查器
func NewStartupChecker(config *StartupConfig) *StartupChecker {
	if config == nil {
		config = getDefaultStartupConfig()
	}
	return &StartupChecker{
		config: config,
	}
}

// getDefaultStartupConfig 获取默认启动配置
func getDefaultStartupConfig() *StartupConfig {
	return &StartupConfig{
		RequiredServices: []ServiceStartupOrder{
			// 第一层：基础设施服务
			{
				Name:         "consul",
				Port:         8500,
				Priority:     1,
				Dependencies: []string{},
				HealthCheck:  "/v1/status/leader",
				Description:  "服务发现和配置中心",
			},
			{
				Name:         "mysql",
				Port:         3306,
				Priority:     2,
				Dependencies: []string{},
				HealthCheck:  "/health",
				Description:  "主数据库",
			},
			{
				Name:         "redis",
				Port:         6379,
				Priority:     3,
				Dependencies: []string{},
				HealthCheck:  "/ping",
				Description:  "缓存服务",
			},
			{
				Name:         "postgresql",
				Port:         5432,
				Priority:     4,
				Dependencies: []string{},
				HealthCheck:  "/health",
				Description:  "AI服务数据库",
			},
			{
				Name:         "nginx",
				Port:         80,
				Priority:     5,
				Dependencies: []string{"consul"},
				HealthCheck:  "/nginx-health",
				Description:  "反向代理",
			},
			// 第二层：核心微服务
			{
				Name:         "api_gateway",
				Port:         8080,
				Priority:     10,
				Dependencies: []string{"consul", "mysql", "redis"},
				HealthCheck:  "/health",
				Description:  "API网关",
			},
			{
				Name:         "user_service",
				Port:         8081,
				Priority:     11,
				Dependencies: []string{"consul", "mysql", "redis"},
				HealthCheck:  "/health",
				Description:  "用户管理服务",
			},
			{
				Name:         "resume_service",
				Port:         8082,
				Priority:     12,
				Dependencies: []string{"consul", "mysql", "redis"},
				HealthCheck:  "/health",
				Description:  "简历管理服务",
			},
			{
				Name:         "company_service",
				Port:         8083,
				Priority:     13,
				Dependencies: []string{"consul", "mysql", "redis"},
				HealthCheck:  "/health",
				Description:  "公司管理服务",
			},
			{
				Name:         "notification_service",
				Port:         8084,
				Priority:     14,
				Dependencies: []string{"consul", "mysql", "redis"},
				HealthCheck:  "/health",
				Description:  "通知服务",
			},
			// 第三层：业务服务（依赖前端启动）
			{
				Name:         "template_service",
				Port:         8085,
				Priority:     20,
				Dependencies: []string{"consul", "api_gateway"},
				HealthCheck:  "/health",
				Description:  "模板管理服务",
			},
			{
				Name:         "statistics_service",
				Port:         8086,
				Priority:     21,
				Dependencies: []string{"consul", "api_gateway"},
				HealthCheck:  "/health",
				Description:  "数据统计服务",
			},
			{
				Name:         "banner_service",
				Port:         8087,
				Priority:     22,
				Dependencies: []string{"consul", "api_gateway"},
				HealthCheck:  "/health",
				Description:  "内容管理服务",
			},
			{
				Name:         "dev_team_service",
				Port:         8088,
				Priority:     23,
				Dependencies: []string{"consul", "api_gateway"},
				HealthCheck:  "/health",
				Description:  "开发团队管理服务",
			},
			// 第四层：AI服务
			{
				Name:         "ai_service",
				Port:         8206,
				Priority:     30,
				Dependencies: []string{"consul", "postgresql"},
				HealthCheck:  "/health",
				Description:  "AI服务",
			},
		},
		Timeout:       30 * time.Second,
		RetryInterval: 5 * time.Second,
	}
}

// CheckStartupOrder 检查系统启动顺序
func (sc *StartupChecker) CheckStartupOrder() *StartupStatus {
	status := &StartupStatus{
		Timestamp:       time.Now(),
		OverallStatus:   "success",
		Services:        []ServiceStartupStatus{},
		Violations:      []StartupViolation{},
		Recommendations: []string{},
	}

	// 检查所有服务的状态
	serviceStatuses := make(map[string]*ServiceStartupStatus)

	for _, service := range sc.config.RequiredServices {
		serviceStatus := sc.checkServiceStatus(service)
		serviceStatuses[service.Name] = serviceStatus
		status.Services = append(status.Services, *serviceStatus)
	}

	// 检查依赖关系
	sc.checkDependencies(serviceStatuses, status)

	// 检查启动顺序
	sc.checkStartupOrder(serviceStatuses, status)

	// 生成建议
	sc.generateRecommendations(status)

	// 确定整体状态
	sc.determineOverallStatus(status)

	return status
}

// checkServiceStatus 检查单个服务状态
func (sc *StartupChecker) checkServiceStatus(service ServiceStartupOrder) *ServiceStartupStatus {
	status := &ServiceStartupStatus{
		Name:          service.Name,
		Port:          service.Port,
		ExpectedOrder: service.Priority,
		Dependencies:  service.Dependencies,
		Status:        "inactive",
		HealthStatus:  "unknown",
	}

	// 检查端口是否开放
	if sc.isPortOpen(service.Port) {
		status.Status = "active"
		status.StartTime = time.Now() // 简化处理，实际应该获取真实启动时间

		// 检查健康状态
		if sc.checkHealth(service) {
			status.HealthStatus = "healthy"
		} else {
			status.HealthStatus = "unhealthy"
		}
	}

	return status
}

// isPortOpen 检查端口是否开放
func (sc *StartupChecker) isPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// checkHealth 检查服务健康状态
func (sc *StartupChecker) checkHealth(service ServiceStartupOrder) bool {
	// 这里应该实现实际的健康检查逻辑
	// 简化处理，假设端口开放就是健康
	return sc.isPortOpen(service.Port)
}

// checkDependencies 检查依赖关系
func (sc *StartupChecker) checkDependencies(serviceStatuses map[string]*ServiceStartupStatus, status *StartupStatus) {
	for serviceName, serviceStatus := range serviceStatuses {
		for _, dependency := range serviceStatus.Dependencies {
			if depStatus, exists := serviceStatuses[dependency]; exists {
				if depStatus.Status != "active" || depStatus.HealthStatus != "healthy" {
					violation := StartupViolation{
						Type:           "dependency",
						Service:        serviceName,
						Message:        fmt.Sprintf("服务 %s 的依赖服务 %s 未正常运行", serviceName, dependency),
						Severity:       "high",
						Recommendation: fmt.Sprintf("请先启动依赖服务 %s", dependency),
					}
					status.Violations = append(status.Violations, violation)
				}
			} else {
				violation := StartupViolation{
					Type:           "dependency",
					Service:        serviceName,
					Message:        fmt.Sprintf("服务 %s 的依赖服务 %s 不存在", serviceName, dependency),
					Severity:       "medium",
					Recommendation: fmt.Sprintf("检查服务配置，确保依赖服务 %s 已定义", dependency),
				}
				status.Violations = append(status.Violations, violation)
			}
		}
	}
}

// checkStartupOrder 检查启动顺序
func (sc *StartupChecker) checkStartupOrder(serviceStatuses map[string]*ServiceStartupStatus, status *StartupStatus) {
	// 检查优先级顺序
	for serviceName, serviceStatus := range serviceStatuses {
		if serviceStatus.Status == "active" {
			// 检查是否有高优先级服务未启动
			for otherName, otherStatus := range serviceStatuses {
				if otherStatus.ExpectedOrder < serviceStatus.ExpectedOrder &&
					otherStatus.Status != "active" {
					violation := StartupViolation{
						Type:           "order",
						Service:        serviceName,
						Message:        fmt.Sprintf("服务 %s 在更高优先级服务 %s 之前启动", serviceName, otherName),
						Severity:       "medium",
						Recommendation: fmt.Sprintf("建议按优先级顺序重启服务，先启动 %s", otherName),
					}
					status.Violations = append(status.Violations, violation)
				}
			}
		}
	}
}

// generateRecommendations 生成建议
func (sc *StartupChecker) generateRecommendations(status *StartupStatus) {
	if len(status.Violations) == 0 {
		status.Recommendations = append(status.Recommendations, "系统启动顺序正确，所有服务按预期顺序运行")
		return
	}

	// 根据违规类型生成建议
	hasDependencyIssues := false
	hasOrderIssues := false

	for _, violation := range status.Violations {
		if violation.Type == "dependency" {
			hasDependencyIssues = true
		}
		if violation.Type == "order" {
			hasOrderIssues = true
		}
	}

	if hasDependencyIssues {
		status.Recommendations = append(status.Recommendations, "发现依赖问题，请按以下顺序重启服务：")
		status.Recommendations = append(status.Recommendations, "1. 基础设施服务 (Consul, MySQL, Redis, PostgreSQL, Nginx)")
		status.Recommendations = append(status.Recommendations, "2. 核心微服务 (API Gateway, User Service, Resume Service, Company Service, Notification Service)")
		status.Recommendations = append(status.Recommendations, "3. 业务服务 (Template Service, Statistics Service, Banner Service, Dev Team Service)")
		status.Recommendations = append(status.Recommendations, "4. AI服务 (AI Service)")
	}

	if hasOrderIssues {
		status.Recommendations = append(status.Recommendations, "发现启动顺序问题，建议使用 ZerviGo 的自动启动功能")
	}
}

// determineOverallStatus 确定整体状态
func (sc *StartupChecker) determineOverallStatus(status *StartupStatus) {
	highSeverityCount := 0
	mediumSeverityCount := 0

	for _, violation := range status.Violations {
		if violation.Severity == "high" {
			highSeverityCount++
		} else if violation.Severity == "medium" {
			mediumSeverityCount++
		}
	}

	if highSeverityCount > 0 {
		status.OverallStatus = "critical"
	} else if mediumSeverityCount > 0 {
		status.OverallStatus = "warning"
	} else {
		status.OverallStatus = "success"
	}
}

// GetStartupScript 获取启动脚本
func (sc *StartupChecker) GetStartupScript() string {
	script := `#!/bin/bash
# ZerviGo 系统启动脚本
# 按正确顺序启动所有服务

echo "🚀 开始启动 JobFirst 微服务系统..."

# 第一层：基础设施服务
echo "📦 启动基础设施服务..."
consul agent -dev -config-dir=consul/config/ &
sleep 5

# 等待基础设施服务启动
echo "⏳ 等待基础设施服务启动..."
sleep 10

# 第二层：核心微服务
echo "🔧 启动核心微服务..."
cd backend/cmd/basic-server && go run main.go &
cd backend/internal/user && go run main.go &
cd backend/internal/resume && go run main.go &
cd backend/internal/company-service && go run main.go &
cd backend/internal/notification-service && go run main.go &

sleep 10

# 第三层：业务服务
echo "💼 启动业务服务..."
cd backend/internal/template-service && go run main.go &
cd backend/internal/statistics-service && go run main.go &
cd backend/internal/banner-service && go run main.go &
cd backend/internal/dev-team-service && go run main.go &

sleep 10

# 第四层：AI服务
echo "🤖 启动AI服务..."
cd backend/internal/ai-service && python main.py &

echo "✅ 系统启动完成！"
echo "🔍 运行 ZerviGo 检查系统状态..."
./backend/pkg/jobfirst-core/superadmin/zervigo
`

	return script
}
