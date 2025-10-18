package system

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"superadmin/errors"
)

// Monitor 系统监控器
type Monitor struct {
	config *MonitorConfig
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	ConsulPort    int           `json:"consul_port"`
	ConsulHost    string        `json:"consul_host"`
	CheckInterval time.Duration `json:"check_interval"`
	Timeout       time.Duration `json:"timeout"`
}

// NewMonitor 创建系统监控器
func NewMonitor(config *MonitorConfig) *Monitor {
	return &Monitor{
		config: config,
	}
}

// SystemStatus 系统状态
type SystemStatus struct {
	Timestamp      time.Time            `json:"timestamp"`
	Infrastructure InfrastructureStatus `json:"infrastructure"`
	Microservices  MicroservicesStatus  `json:"microservices"`
	Health         HealthStatus         `json:"health"`
	Resources      ResourceStatus       `json:"resources"`
}

// InfrastructureStatus 基础设施状态
type InfrastructureStatus struct {
	Consul *ServiceStatus `json:"consul"`
	MySQL  *ServiceStatus `json:"mysql"`
	Redis  *ServiceStatus `json:"redis"`
	Neo4j  *ServiceStatus `json:"neo4j"`
}

// MicroservicesStatus 微服务状态
type MicroservicesStatus struct {
	APIGateway          *ServiceStatus `json:"api_gateway"`
	UserService         *ServiceStatus `json:"user_service"`
	ResumeService       *ServiceStatus `json:"resume_service"`
	BannerService       *ServiceStatus `json:"banner_service"`
	TemplateService     *ServiceStatus `json:"template_service"`
	NotificationService *ServiceStatus `json:"notification_service"`
	StatisticsService   *ServiceStatus `json:"statistics_service"`
	JobService          *ServiceStatus `json:"job_service"`
	AIService           *ServiceStatus `json:"ai_service"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Overall   string    `json:"overall"`
	Score     float64   `json:"score"`
	Issues    []string  `json:"issues"`
	LastCheck time.Time `json:"last_check"`
}

// ResourceStatus 资源状态
type ResourceStatus struct {
	CPU     CPUStatus     `json:"cpu"`
	Memory  MemoryStatus  `json:"memory"`
	Disk    DiskStatus    `json:"disk"`
	Network NetworkStatus `json:"network"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Port      int       `json:"port"`
	PID       int       `json:"pid,omitempty"`
	Uptime    string    `json:"uptime,omitempty"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
}

// CPUStatus CPU状态
type CPUStatus struct {
	Usage   float64   `json:"usage"`
	Cores   int       `json:"cores"`
	LoadAvg []float64 `json:"load_avg"`
}

// MemoryStatus 内存状态
type MemoryStatus struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

// DiskStatus 磁盘状态
type DiskStatus struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

// NetworkStatus 网络状态
type NetworkStatus struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// NetworkInterface 网络接口
type NetworkInterface struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Up   bool   `json:"up"`
}

// GetSystemStatus 获取系统整体状态
func (m *Monitor) GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{
		Timestamp: time.Now(),
	}

	// 获取基础设施状态
	infraStatus, err := m.getInfrastructureStatus()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取基础设施状态失败", err)
	}
	status.Infrastructure = *infraStatus

	// 获取微服务状态
	microStatus, err := m.getMicroservicesStatus()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取微服务状态失败", err)
	}
	status.Microservices = *microStatus

	// 计算健康状态
	healthStatus, err := m.calculateHealthStatus(&status.Infrastructure, &status.Microservices)
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "计算健康状态失败", err)
	}
	status.Health = *healthStatus

	// 获取资源状态
	resourceStatus, err := m.getResourceStatus()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取资源状态失败", err)
	}
	status.Resources = *resourceStatus

	return status, nil
}

// getInfrastructureStatus 获取基础设施状态
func (m *Monitor) getInfrastructureStatus() (*InfrastructureStatus, error) {
	status := &InfrastructureStatus{}

	// 检查Consul
	consulStatus, err := m.checkConsulService()
	if err != nil {
		status.Consul = &ServiceStatus{
			Name:      "consul",
			Status:    "error",
			Port:      m.config.ConsulPort,
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.Consul = consulStatus
	}

	// 检查MySQL
	mysqlStatus, err := m.checkService("mysql", 3306)
	if err != nil {
		status.MySQL = &ServiceStatus{
			Name:      "mysql",
			Status:    "error",
			Port:      3306,
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.MySQL = mysqlStatus
	}

	// 检查Redis
	redisStatus, err := m.checkService("redis", 6379)
	if err != nil {
		status.Redis = &ServiceStatus{
			Name:      "redis",
			Status:    "error",
			Port:      6379,
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.Redis = redisStatus
	}

	// 检查Neo4j
	neo4jStatus, err := m.checkService("neo4j", 7687)
	if err != nil {
		status.Neo4j = &ServiceStatus{
			Name:      "neo4j",
			Status:    "error",
			Port:      7687,
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.Neo4j = neo4jStatus
	}

	return status, nil
}

// checkConsulService 检查Consul服务
func (m *Monitor) checkConsulService() (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:      "consul",
		Port:      m.config.ConsulPort,
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", m.config.ConsulHost, m.config.ConsulPort), m.config.Timeout)
	if err != nil {
		status.Status = "down"
		status.Error = err.Error()
		return status, err
	}
	defer conn.Close()

	// 尝试获取Consul状态
	cmd := exec.Command("consul", "members")
	output, err := cmd.Output()
	if err != nil {
		status.Status = "running"
		status.Error = "无法获取Consul成员信息"
		return status, nil
	}

	// 解析输出
	lines := strings.Split(string(output), "\n")
	memberCount := 0
	for _, line := range lines {
		if strings.Contains(line, "alive") {
			memberCount++
		}
	}

	status.Status = "healthy"
	if memberCount > 0 {
		status.Status = "healthy"
	} else {
		status.Status = "warning"
		status.Error = "没有活跃的Consul成员"
	}

	return status, nil
}

// getMicroservicesStatus 获取微服务状态
func (m *Monitor) getMicroservicesStatus() (*MicroservicesStatus, error) {
	status := &MicroservicesStatus{}

	// 检查各个微服务 (v3.1.1 端口配置修正)
	services := map[string]int{
		"api_gateway":          8080,
		"user_service":         8081,
		"resume_service":       8082,
		"company_service":      8083,
		"notification_service": 8084,
		"template_service":     8085, // 修正：Template Service 端口 8085
		"statistics_service":   8086, // 保持：Statistics Service 端口 8086
		"banner_service":       8087, // 修正：Banner Service 端口 8087
		"dev_team_service":     8088,
		"job_service":          8089, // 新增：Job Service 端口 8089
		"ai_service":           8206,
	}

	for serviceName, port := range services {
		serviceStatus, err := m.checkService(serviceName, port)
		if err != nil {
			// 设置错误状态
			switch serviceName {
			case "api_gateway":
				status.APIGateway = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "user_service":
				status.UserService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "resume_service":
				status.ResumeService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "banner_service":
				status.BannerService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "template_service":
				status.TemplateService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "notification_service":
				status.NotificationService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "statistics_service":
				status.StatisticsService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "company_service":
				// 暂时跳过，因为MicroservicesStatus结构体中没有CompanyService字段
			case "dev_team_service":
				// 暂时跳过，因为MicroservicesStatus结构体中没有DevTeamService字段
			case "job_service":
				status.JobService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			case "ai_service":
				status.AIService = &ServiceStatus{
					Name: serviceName, Status: "error", Port: port,
					LastCheck: time.Now(), Error: err.Error(),
				}
			}
		} else {
			// 设置正常状态
			switch serviceName {
			case "api_gateway":
				status.APIGateway = serviceStatus
			case "user_service":
				status.UserService = serviceStatus
			case "resume_service":
				status.ResumeService = serviceStatus
			case "banner_service":
				status.BannerService = serviceStatus
			case "template_service":
				status.TemplateService = serviceStatus
			case "notification_service":
				status.NotificationService = serviceStatus
			case "statistics_service":
				status.StatisticsService = serviceStatus
			case "company_service":
				// 暂时跳过，因为MicroservicesStatus结构体中没有CompanyService字段
			case "dev_team_service":
				// 暂时跳过，因为MicroservicesStatus结构体中没有DevTeamService字段
			case "job_service":
				status.JobService = serviceStatus
			case "ai_service":
				status.AIService = serviceStatus
			}
		}
	}

	return status, nil
}

// checkService 检查服务状态
func (m *Monitor) checkService(serviceName string, port int) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name:      serviceName,
		Port:      port,
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), m.config.Timeout)
	if err != nil {
		status.Status = "down"
		status.Error = err.Error()
		return status, err
	}
	defer conn.Close()

	// 尝试获取进程信息
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		status.Status = "running"
		status.Error = "无法获取进程信息"
		return status, nil
	}

	// 解析进程信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) > 1 {
			status.PID = 0 // 这里可以解析PID
		}
	}

	status.Status = "healthy"
	return status, nil
}

// calculateHealthStatus 计算健康状态
func (m *Monitor) calculateHealthStatus(infra *InfrastructureStatus, micro *MicroservicesStatus) (*HealthStatus, error) {
	health := &HealthStatus{
		LastCheck: time.Now(),
		Issues:    []string{},
	}

	// 检查基础设施
	infraServices := []*ServiceStatus{
		infra.Consul, infra.MySQL, infra.Redis, infra.Neo4j,
	}

	healthyCount := 0
	totalCount := len(infraServices)

	for _, service := range infraServices {
		if service != nil {
			if service.Status == "healthy" {
				healthyCount++
			} else {
				health.Issues = append(health.Issues, fmt.Sprintf("%s: %s", service.Name, service.Error))
			}
		}
	}

	// 检查微服务
	microServices := []*ServiceStatus{
		micro.APIGateway, micro.UserService, micro.ResumeService,
		micro.BannerService, micro.TemplateService, micro.NotificationService,
		micro.StatisticsService, micro.AIService,
	}

	for _, service := range microServices {
		if service != nil {
			totalCount++
			if service.Status == "healthy" {
				healthyCount++
			} else {
				health.Issues = append(health.Issues, fmt.Sprintf("%s: %s", service.Name, service.Error))
			}
		}
	}

	// 计算健康分数
	if totalCount > 0 {
		health.Score = float64(healthyCount) / float64(totalCount) * 100
	}

	// 确定整体状态
	if health.Score >= 90 {
		health.Overall = "excellent"
	} else if health.Score >= 70 {
		health.Overall = "good"
	} else if health.Score >= 50 {
		health.Overall = "warning"
	} else {
		health.Overall = "critical"
	}

	return health, nil
}

// getResourceStatus 获取资源状态
func (m *Monitor) getResourceStatus() (*ResourceStatus, error) {
	status := &ResourceStatus{}

	// 获取CPU状态
	cpuStatus, err := m.getCPUStatus()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取CPU状态失败", err)
	}
	status.CPU = *cpuStatus

	// 获取内存状态
	memoryStatus, err := m.getMemoryStatus()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取内存状态失败", err)
	}
	status.Memory = *memoryStatus

	// 获取磁盘状态
	diskStatus, err := m.getDiskStatus()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取磁盘状态失败", err)
	}
	status.Disk = *diskStatus

	// 获取网络状态
	networkStatus, err := m.getNetworkStatus()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取网络状态失败", err)
	}
	status.Network = *networkStatus

	return status, nil
}

// getCPUStatus 获取CPU状态
func (m *Monitor) getCPUStatus() (*CPUStatus, error) {
	status := &CPUStatus{
		Cores: runtime.NumCPU(),
	}

	// 获取负载平均值
	cmd := exec.Command("uptime")
	output, err := cmd.Output()
	if err != nil {
		return status, err
	}

	// 解析负载平均值
	parts := strings.Split(string(output), "load average:")
	if len(parts) > 1 {
		loadStr := strings.TrimSpace(parts[1])
		loads := strings.Split(loadStr, ",")
		status.LoadAvg = make([]float64, len(loads))
		for i, load := range loads {
			// 这里可以解析负载值
			_ = load
			status.LoadAvg[i] = 0.0 // 简化处理
		}
	}

	return status, nil
}

// getMemoryStatus 获取内存状态
func (m *Monitor) getMemoryStatus() (*MemoryStatus, error) {
	status := &MemoryStatus{}

	// 检测操作系统并使用相应的命令
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// macOS使用vm_stat命令
		cmd = exec.Command("vm_stat")
	} else {
		// Linux使用free命令
		cmd = exec.Command("free", "-b")
	}

	output, err := cmd.Output()
	if err != nil {
		// 如果命令失败，返回默认值而不是错误
		status.Total = 0
		status.Used = 0
		status.Available = 0
		return status, nil
	}

	// 解析内存信息
	if runtime.GOOS == "darwin" {
		// macOS vm_stat解析
		status.Total = 0
		status.Used = 0
		status.Available = 0
		status.Usage = 0.0
	} else {
		// Linux free命令解析
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 7 {
				// 解析内存数据
				status.Total = 0
				status.Used = 0
				status.Available = 0
				status.Usage = 0.0
			}
		}
	}

	return status, nil
}

// getDiskStatus 获取磁盘状态
func (m *Monitor) getDiskStatus() (*DiskStatus, error) {
	status := &DiskStatus{}

	// 获取磁盘信息
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// macOS使用df命令
		cmd = exec.Command("df", "-k", "/")
	} else {
		// Linux使用df命令
		cmd = exec.Command("df", "-B1", "/")
	}

	output, err := cmd.Output()
	if err != nil {
		// 如果命令失败，返回默认值而不是错误
		status.Total = 0
		status.Used = 0
		status.Available = 0
		status.Usage = 0.0
		return status, nil
	}

	// 解析磁盘信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			// 解析磁盘数据
			// 这里需要根据实际输出格式解析
			status.Total = 0
			status.Used = 0
			status.Available = 0
			status.Usage = 0.0
		}
	}

	return status, nil
}

// getNetworkStatus 获取网络状态
func (m *Monitor) getNetworkStatus() (*NetworkStatus, error) {
	status := &NetworkStatus{
		Interfaces: []NetworkInterface{},
	}

	// 获取网络接口信息
	interfaces, err := net.Interfaces()
	if err != nil {
		return status, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		interfaceInfo := NetworkInterface{
			Name: iface.Name,
			Up:   iface.Flags&net.FlagUp != 0,
		}

		if len(addrs) > 0 {
			interfaceInfo.IP = addrs[0].String()
		}

		status.Interfaces = append(status.Interfaces, interfaceInfo)
	}

	return status, nil
}
