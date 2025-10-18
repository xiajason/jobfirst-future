package superadmin

import (
	"fmt"
	"time"

	"superadmin/ai"
	"superadmin/cicd"
	configmanager "superadmin/config"
	"superadmin/database"
	"superadmin/system"
	"superadmin/user"
)

// Manager 模块化超级管理员管理器
type Manager struct {
	SystemMonitor   *system.Monitor
	UserManager     *user.Manager
	DatabaseManager *database.Manager
	AIManager       *ai.Manager
	ConfigManager   *configmanager.Manager
	CICDManager     *cicd.Manager
	config          *Config
}

// Config 超级管理员配置
type Config struct {
	System   system.MonitorConfig              `json:"system"`
	User     user.UserConfig                   `json:"user"`
	Database database.DatabaseConfig           `json:"database"`
	AI       ai.AIConfig                       `json:"ai"`
	Config   configmanager.ConfigManagerConfig `json:"config"`
	CICD     cicd.CICDConfig                   `json:"cicd"`
}

// NewManager 创建模块化超级管理员管理器
func NewManager(config *Config) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	// 初始化系统监控器
	systemMonitor := system.NewMonitor(&config.System)
	manager.SystemMonitor = systemMonitor

	// 初始化用户管理器
	userManager := user.NewManager(&config.User)
	manager.UserManager = userManager

	// 初始化数据库管理器
	databaseManager := database.NewManager(&config.Database)
	manager.DatabaseManager = databaseManager

	// 初始化AI管理器
	aiManager := ai.NewManager(&config.AI)
	manager.AIManager = aiManager

	// 初始化配置管理器
	configManager := configmanager.NewManager(&config.Config)
	manager.ConfigManager = configManager

	// 初始化CI/CD管理器
	cicdManager := cicd.NewManager(&config.CICD)
	manager.CICDManager = cicdManager

	return manager, nil
}

// GetSystemStatus 获取系统整体状态
func (m *Manager) GetSystemStatus() (*system.SystemStatus, error) {
	return m.SystemMonitor.GetSystemStatus()
}

// GetUsers 获取所有用户
func (m *Manager) GetUsers() ([]user.User, error) {
	return m.UserManager.GetUsers()
}

// CreateUser 创建用户
func (m *Manager) CreateUser(username, role, sshKey string) error {
	return m.UserManager.CreateUser(username, role, sshKey)
}

// DeleteUser 删除用户
func (m *Manager) DeleteUser(username string) error {
	return m.UserManager.DeleteUser(username)
}

// AssignRole 分配角色
func (m *Manager) AssignRole(username, role string) error {
	return m.UserManager.AssignRole(username, role)
}

// CheckUserPermission 检查用户权限
func (m *Manager) CheckUserPermission(username, permission string) (bool, error) {
	return m.UserManager.CheckUserPermission(username, permission)
}

// GetUserPermissions 获取用户权限
func (m *Manager) GetUserPermissions(username string) ([]string, error) {
	return m.UserManager.GetUserPermissions(username)
}

// ValidateAccess 验证访问权限
func (m *Manager) ValidateAccess(username, resource, action string) (bool, error) {
	return m.UserManager.ValidateAccess(username, resource, action)
}

// GetProjectMembers 获取项目成员
func (m *Manager) GetProjectMembers() ([]user.User, error) {
	return m.UserManager.GetProjectMembers()
}

// AddProjectMember 添加项目成员
func (m *Manager) AddProjectMember(username, role, department string) error {
	return m.UserManager.AddProjectMember(username, role, department)
}

// RemoveProjectMember 移除项目成员
func (m *Manager) RemoveProjectMember(username string) error {
	return m.UserManager.RemoveProjectMember(username)
}

// GetMemberActivity 获取成员活动
func (m *Manager) GetMemberActivity(username string) ([]string, error) {
	return m.UserManager.GetMemberActivity(username)
}

// GetRoles 获取所有角色
func (m *Manager) GetRoles() []user.Role {
	return m.UserManager.GetRoles()
}

// CheckPermissions 检查权限控制
func (m *Manager) CheckPermissions() (*user.AccessControl, error) {
	return m.UserManager.CheckPermissions()
}

// GetDatabaseStatus 获取数据库状态
func (m *Manager) GetDatabaseStatus() (*database.DatabaseStatus, error) {
	return m.DatabaseManager.GetDatabaseStatus()
}

// GetDatabaseInitStatus 获取数据库初始化状态
func (m *Manager) GetDatabaseInitStatus() (*database.DatabaseInitStatus, error) {
	return m.DatabaseManager.GetDatabaseInitStatus()
}

// InitializeDatabase 初始化数据库
func (m *Manager) InitializeDatabase(dbType string) error {
	return m.DatabaseManager.InitializeDatabase(dbType)
}

// GetAIServiceStatus 获取AI服务状态
func (m *Manager) GetAIServiceStatus() (*ai.AIServiceStatus, error) {
	return m.AIManager.GetAIServiceStatus()
}

// ConfigureAIService 配置AI服务
func (m *Manager) ConfigureAIService(provider, apiKey, baseURL, model string) error {
	return m.AIManager.ConfigureAIService(provider, apiKey, baseURL, model)
}

// TestAIService 测试AI服务
func (m *Manager) TestAIService() (*ai.AITestResult, error) {
	return m.AIManager.TestAIService()
}

// CollectAllConfigs 收集所有配置
func (m *Manager) CollectAllConfigs() error {
	return m.ConfigManager.CollectAllConfigs()
}

// BackupConfigs 备份配置
func (m *Manager) BackupConfigs() (*configmanager.ConfigBackup, error) {
	return m.ConfigManager.BackupConfigs()
}

// ValidateConfigs 验证配置
func (m *Manager) ValidateConfigs() (*configmanager.ConfigValidation, error) {
	return m.ConfigManager.ValidateConfigs()
}

// GetEnvironments 获取环境配置
func (m *Manager) GetEnvironments() ([]configmanager.Environment, error) {
	return m.ConfigManager.GetEnvironments()
}

// GetCICDStatus 获取CI/CD状态
func (m *Manager) GetCICDStatus() (*cicd.CICDStatus, error) {
	return m.CICDManager.GetCICDStatus()
}

// GetCICDPipelines 获取CI/CD流水线
func (m *Manager) GetCICDPipelines() ([]cicd.CICDPipeline, error) {
	return m.CICDManager.GetCICDPipelines()
}

// TriggerCICDDeploy 触发CI/CD部署
func (m *Manager) TriggerCICDDeploy(environment string) error {
	return m.CICDManager.TriggerCICDDeploy(environment)
}

// GetCICDWebhooks 获取CI/CD Webhook
func (m *Manager) GetCICDWebhooks() ([]cicd.CICDWebhook, error) {
	return m.CICDManager.GetCICDWebhooks()
}

// GetCICDRepositories 获取CI/CD仓库
func (m *Manager) GetCICDRepositories() ([]cicd.CICDRepository, error) {
	return m.CICDManager.GetCICDRepositories()
}

// GetCICDLogs 获取CI/CD日志
func (m *Manager) GetCICDLogs(pipelineID string) ([]string, error) {
	return m.CICDManager.GetCICDLogs(pipelineID)
}

// GetHealthStatus 获取整体健康状态
func (m *Manager) GetHealthStatus() (map[string]interface{}, error) {
	health := make(map[string]interface{})

	// 获取系统状态
	systemStatus, err := m.GetSystemStatus()
	if err != nil {
		health["system"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["system"] = systemStatus.Health
	}

	// 获取数据库状态
	dbStatus, err := m.GetDatabaseStatus()
	if err != nil {
		health["database"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["database"] = dbStatus
	}

	// 获取AI服务状态
	aiStatus, err := m.GetAIServiceStatus()
	if err != nil {
		health["ai"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["ai"] = aiStatus.Health
	}

	// 获取CI/CD状态
	cicdStatus, err := m.GetCICDStatus()
	if err != nil {
		health["cicd"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["cicd"] = map[string]interface{}{
			"health": cicdStatus.Health,
		}
	}

	// 计算整体健康状态
	overallHealth := "healthy"
	for _, status := range health {
		if statusMap, ok := status.(map[string]interface{}); ok {
			if status, exists := statusMap["status"]; exists && status != "healthy" {
				overallHealth = "warning"
				break
			}
		}
	}

	health["overall"] = overallHealth
	return health, nil
}

// 向后兼容的方法（待实现）
func (m *Manager) RestartInfrastructure() error {
	return fmt.Errorf("重启基础设施功能待实现")
}

func (m *Manager) CreateBackup(backupType string) (*Backup, error) {
	return nil, fmt.Errorf("创建备份功能待实现")
}

func (m *Manager) GetAlerts() ([]Alert, error) {
	return nil, fmt.Errorf("获取告警功能待实现")
}

func (m *Manager) GetLogs(limit int) ([]LogEntry, error) {
	return nil, fmt.Errorf("获取日志功能待实现")
}

func (m *Manager) GetFrontendStatus() (*FrontendStatus, error) {
	return nil, fmt.Errorf("获取前端状态功能待实现")
}

func (m *Manager) StartFrontendDevServer() error {
	return fmt.Errorf("启动前端开发服务器功能待实现")
}

func (m *Manager) StopFrontendDevServer() error {
	return fmt.Errorf("停止前端开发服务器功能待实现")
}

func (m *Manager) RestartFrontendDevServer() error {
	return fmt.Errorf("重启前端开发服务器功能待实现")
}

func (m *Manager) BuildFrontendProduction() error {
	return fmt.Errorf("构建前端生产版本功能待实现")
}

func (m *Manager) SyncFrontendSource() error {
	return fmt.Errorf("同步前端源码功能待实现")
}

func (m *Manager) InstallFrontendDependencies() error {
	return fmt.Errorf("安装前端依赖功能待实现")
}

// 向后兼容的类型定义
type Backup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
}

type Alert struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	Severity   string    `json:"severity"`
	CreatedAt  time.Time `json:"created_at"`
	IsResolved bool      `json:"is_resolved"`
}

type LogEntry struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
}

type FrontendStatus struct {
	DevServer    *FrontendDevServer    `json:"dev_server"`
	Production   *FrontendProduction   `json:"production"`
	Dependencies *FrontendDependencies `json:"dependencies"`
	SourceCode   *FrontendSourceCode   `json:"source_code"`
}

type FrontendDevServer struct {
	Status    string `json:"status"`
	Port      int    `json:"port"`
	URL       string `json:"url"`
	Uptime    string `json:"uptime"`
	Processes int    `json:"processes"`
}

type FrontendProduction struct {
	Status     string `json:"status"`
	BuildTime  string `json:"build_time"`
	Size       string `json:"size"`
	LastDeploy string `json:"last_deploy"`
}

type FrontendDependencies struct {
	Status          string `json:"status"`
	Count           int    `json:"count"`
	LastUpdate      string `json:"last_update"`
	Vulnerabilities int    `json:"vulnerabilities"`
}

type FrontendSourceCode struct {
	Status    string `json:"status"`
	LastSync  string `json:"last_sync"`
	Conflicts int    `json:"conflicts"`
	Branch    string `json:"branch"`
}
