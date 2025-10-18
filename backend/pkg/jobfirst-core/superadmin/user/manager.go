package user

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"superadmin/errors"
)

// Manager 用户管理器
type Manager struct {
	config *UserConfig
}

// UserConfig 用户配置
type UserConfig struct {
	SSHKeyPath   string `json:"ssh_key_path"`
	UserHomePath string `json:"user_home_path"`
	DefaultShell string `json:"default_shell"`
	ProjectPath  string `json:"project_path"`
}

// NewManager 创建用户管理器
func NewManager(config *UserConfig) *Manager {
	return &Manager{
		config: config,
	}
}

// User 用户信息
type User struct {
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	SSHKey     string    `json:"ssh_key"`
	CreatedAt  time.Time `json:"created_at"`
	LastLogin  time.Time `json:"last_login"`
	IsActive   bool      `json:"is_active"`
	Department string    `json:"department,omitempty"`
}

// Role 角色信息
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	Level       int      `json:"level"`
}

// Permission 权限信息
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// GetUsers 获取所有用户
func (m *Manager) GetUsers() ([]User, error) {
	// 获取系统用户列表
	cmd := exec.Command("cut", "-d:", "-f1", "/etc/passwd")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取用户列表失败", err)
	}

	users := []User{}
	lines := strings.Split(string(output), "\n")

	for _, username := range lines {
		username = strings.TrimSpace(username)
		if username == "" || username == "root" || username == "nobody" {
			continue
		}

		// 检查用户是否存在于项目中
		if m.isProjectUser(username) {
			user := User{
				Username:  username,
				Role:      m.determineUserRole(username),
				CreatedAt: time.Now(), // 简化处理
				IsActive:  true,
			}
			users = append(users, user)
		}
	}

	return users, nil
}

// isProjectUser 检查是否为项目用户
func (m *Manager) isProjectUser(username string) bool {
	// 检查用户是否有项目目录访问权限
	cmd := exec.Command("test", "-d", fmt.Sprintf("%s/%s", m.config.ProjectPath, username))
	return cmd.Run() == nil
}

// determineUserRole 确定用户角色
func (m *Manager) determineUserRole(username string) string {
	// 检查用户组
	cmd := exec.Command("groups", username)
	output, err := cmd.Output()
	if err != nil {
		return "developer"
	}

	groups := string(output)
	if strings.Contains(groups, "admin") {
		return "admin"
	} else if strings.Contains(groups, "manager") {
		return "manager"
	} else if strings.Contains(groups, "developer") {
		return "developer"
	}

	return "developer"
}

// CreateUser 创建用户
func (m *Manager) CreateUser(username, role, sshKey string) error {
	// 检查用户是否已存在
	cmd := exec.Command("id", username)
	if cmd.Run() == nil {
		return errors.NewError(errors.ErrCodeAlreadyExists, "用户已存在")
	}

	// 创建用户
	cmd = exec.Command("useradd", "-m", "-s", m.config.DefaultShell, username)
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeService, "创建用户失败", err)
	}

	// 设置SSH密钥
	if sshKey != "" {
		if err := m.setSSHKey(username, sshKey); err != nil {
			return errors.WrapError(errors.ErrCodeService, "设置SSH密钥失败", err)
		}
	}

	// 分配角色权限
	if err := m.assignRolePermissions(username, role); err != nil {
		return errors.WrapError(errors.ErrCodeService, "分配角色权限失败", err)
	}

	// 创建项目目录
	if err := m.createProjectDirectory(username); err != nil {
		return errors.WrapError(errors.ErrCodeService, "创建项目目录失败", err)
	}

	return nil
}

// setSSHKey 设置SSH密钥
func (m *Manager) setSSHKey(username, sshKey string) error {
	// 创建.ssh目录
	sshDir := fmt.Sprintf("%s/%s/.ssh", m.config.UserHomePath, username)
	cmd := exec.Command("mkdir", "-p", sshDir)
	if err := cmd.Run(); err != nil {
		return err
	}

	// 设置权限
	cmd = exec.Command("chmod", "700", sshDir)
	if err := cmd.Run(); err != nil {
		return err
	}

	// 写入公钥
	authorizedKeys := fmt.Sprintf("%s/authorized_keys", sshDir)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("echo '%s' >> %s", sshKey, authorizedKeys))
	if err := cmd.Run(); err != nil {
		return err
	}

	// 设置权限
	cmd = exec.Command("chmod", "600", authorizedKeys)
	if err := cmd.Run(); err != nil {
		return err
	}

	// 设置所有者
	cmd = exec.Command("chown", "-R", fmt.Sprintf("%s:%s", username, username), sshDir)
	return cmd.Run()
}

// assignRolePermissions 分配角色权限
func (m *Manager) assignRolePermissions(username, role string) error {
	// 根据角色添加到相应组
	groups := m.getRoleGroups(role)
	for _, group := range groups {
		cmd := exec.Command("usermod", "-a", "-G", group, username)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// getRoleGroups 获取角色对应的组
func (m *Manager) getRoleGroups(role string) []string {
	switch role {
	case "admin":
		return []string{"admin", "sudo", "docker"}
	case "manager":
		return []string{"manager", "docker"}
	case "developer":
		return []string{"developer"}
	default:
		return []string{"developer"}
	}
}

// createProjectDirectory 创建项目目录
func (m *Manager) createProjectDirectory(username string) error {
	projectDir := fmt.Sprintf("%s/%s", m.config.ProjectPath, username)

	// 创建目录
	cmd := exec.Command("mkdir", "-p", projectDir)
	if err := cmd.Run(); err != nil {
		return err
	}

	// 设置权限
	cmd = exec.Command("chown", "-R", fmt.Sprintf("%s:%s", username, username), projectDir)
	if err := cmd.Run(); err != nil {
		return err
	}

	// 设置权限
	cmd = exec.Command("chmod", "755", projectDir)
	return cmd.Run()
}

// DeleteUser 删除用户
func (m *Manager) DeleteUser(username string) error {
	// 删除用户
	cmd := exec.Command("userdel", "-r", username)
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeService, "删除用户失败", err)
	}

	// 删除项目目录
	projectDir := fmt.Sprintf("%s/%s", m.config.ProjectPath, username)
	cmd = exec.Command("rm", "-rf", projectDir)
	cmd.Run() // 忽略错误，目录可能不存在

	return nil
}

// AssignRole 分配角色
func (m *Manager) AssignRole(username, role string) error {
	// 检查用户是否存在
	cmd := exec.Command("id", username)
	if cmd.Run() != nil {
		return errors.NewError(errors.ErrCodeNotFound, "用户不存在")
	}

	// 分配角色权限
	return m.assignRolePermissions(username, role)
}

// CheckUserPermission 检查用户权限
func (m *Manager) CheckUserPermission(username, permission string) (bool, error) {
	// 获取用户组
	cmd := exec.Command("groups", username)
	output, err := cmd.Output()
	if err != nil {
		return false, errors.WrapError(errors.ErrCodeService, "获取用户组失败", err)
	}

	_ = string(output)
	userRole := m.determineUserRole(username)

	// 检查权限
	return m.hasPermission(userRole, permission), nil
}

// hasPermission 检查角色是否有权限
func (m *Manager) hasPermission(role, permission string) bool {
	rolePermissions := m.getRolePermissions(role)
	for _, perm := range rolePermissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// getRolePermissions 获取角色权限
func (m *Manager) getRolePermissions(role string) []string {
	switch role {
	case "admin":
		return []string{
			"system.manage", "user.manage", "database.manage",
			"service.manage", "config.manage", "deploy.manage",
			"monitor.view", "log.view", "backup.manage",
		}
	case "manager":
		return []string{
			"user.view", "database.view", "service.view",
			"config.view", "deploy.manage", "monitor.view",
			"log.view", "backup.view",
		}
	case "developer":
		return []string{
			"service.view", "config.view", "monitor.view",
			"log.view",
		}
	default:
		return []string{}
	}
}

// GetUserPermissions 获取用户权限
func (m *Manager) GetUserPermissions(username string) ([]string, error) {
	// 检查用户是否存在
	cmd := exec.Command("id", username)
	if cmd.Run() != nil {
		return nil, errors.NewError(errors.ErrCodeNotFound, "用户不存在")
	}

	userRole := m.determineUserRole(username)
	return m.getRolePermissions(userRole), nil
}

// ValidateAccess 验证访问权限
func (m *Manager) ValidateAccess(username, resource, action string) (bool, error) {
	permission := fmt.Sprintf("%s.%s", resource, action)
	return m.CheckUserPermission(username, permission)
}

// GetProjectMembers 获取项目成员
func (m *Manager) GetProjectMembers() ([]User, error) {
	users, err := m.GetUsers()
	if err != nil {
		return nil, err
	}

	// 过滤项目成员
	members := []User{}
	for _, user := range users {
		if m.isProjectUser(user.Username) {
			members = append(members, user)
		}
	}

	return members, nil
}

// AddProjectMember 添加项目成员
func (m *Manager) AddProjectMember(username, role, department string) error {
	// 检查用户是否存在
	cmd := exec.Command("id", username)
	if cmd.Run() != nil {
		return errors.NewError(errors.ErrCodeNotFound, "用户不存在")
	}

	// 分配角色
	if err := m.AssignRole(username, role); err != nil {
		return err
	}

	// 创建项目目录
	if err := m.createProjectDirectory(username); err != nil {
		return err
	}

	// 设置部门信息（这里可以存储到数据库）
	// 简化处理，实际应该存储到数据库

	return nil
}

// RemoveProjectMember 移除项目成员
func (m *Manager) RemoveProjectMember(username string) error {
	// 删除项目目录
	projectDir := fmt.Sprintf("%s/%s", m.config.ProjectPath, username)
	cmd := exec.Command("rm", "-rf", projectDir)
	cmd.Run() // 忽略错误

	// 从项目组中移除
	cmd = exec.Command("gpasswd", "-d", username, "developer")
	cmd.Run() // 忽略错误

	return nil
}

// GetMemberActivity 获取成员活动
func (m *Manager) GetMemberActivity(username string) ([]string, error) {
	// 获取用户最近的活动
	activities := []string{}

	// 检查登录记录
	cmd := exec.Command("last", "-n", "10", username)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, username) {
				activities = append(activities, fmt.Sprintf("登录: %s", line))
			}
		}
	}

	// 检查项目目录修改
	projectDir := fmt.Sprintf("%s/%s", m.config.ProjectPath, username)
	cmd = exec.Command("find", projectDir, "-type", "f", "-mtime", "-1")
	output, err = cmd.Output()
	if err == nil {
		files := strings.Split(string(output), "\n")
		if len(files) > 1 {
			activities = append(activities, fmt.Sprintf("修改了 %d 个文件", len(files)-1))
		}
	}

	return activities, nil
}

// GetRoles 获取所有角色
func (m *Manager) GetRoles() []Role {
	return []Role{
		{
			Name:        "admin",
			Description: "系统管理员",
			Permissions: m.getRolePermissions("admin"),
			Level:       3,
		},
		{
			Name:        "manager",
			Description: "项目经理",
			Permissions: m.getRolePermissions("manager"),
			Level:       2,
		},
		{
			Name:        "developer",
			Description: "开发人员",
			Permissions: m.getRolePermissions("developer"),
			Level:       1,
		},
	}
}

// CheckPermissions 检查权限控制
func (m *Manager) CheckPermissions() (*AccessControl, error) {
	control := &AccessControl{
		SSH:      &SSHControl{},
		Ports:    &PortsControl{},
		Firewall: &FirewallControl{},
	}

	// 检查SSH权限
	sshControl, err := m.checkSSHPermissions()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "检查SSH权限失败", err)
	}
	control.SSH = sshControl

	// 检查端口权限
	portsControl, err := m.checkPortsPermissions()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "检查端口权限失败", err)
	}
	control.Ports = portsControl

	// 检查防火墙权限
	firewallControl, err := m.checkFirewallPermissions()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "检查防火墙权限失败", err)
	}
	control.Firewall = firewallControl

	return control, nil
}

// AccessControl 访问控制
type AccessControl struct {
	SSH      *SSHControl      `json:"ssh"`
	Ports    *PortsControl    `json:"ports"`
	Firewall *FirewallControl `json:"firewall"`
}

// SSHControl SSH控制
type SSHControl struct {
	Enabled      bool     `json:"enabled"`
	Port         int      `json:"port"`
	AllowedUsers []string `json:"allowed_users"`
	KeyAuthOnly  bool     `json:"key_auth_only"`
}

// PortsControl 端口控制
type PortsControl struct {
	OpenPorts   []int `json:"open_ports"`
	ClosedPorts []int `json:"closed_ports"`
}

// FirewallControl 防火墙控制
type FirewallControl struct {
	Enabled    bool     `json:"enabled"`
	Rules      []string `json:"rules"`
	BlockedIPs []string `json:"blocked_ips"`
}

// checkSSHPermissions 检查SSH权限
func (m *Manager) checkSSHPermissions() (*SSHControl, error) {
	control := &SSHControl{}

	// 检查SSH服务状态
	cmd := exec.Command("systemctl", "is-active", "ssh")
	output, err := cmd.Output()
	if err == nil {
		control.Enabled = strings.TrimSpace(string(output)) == "active"
	}

	// 检查SSH配置
	cmd = exec.Command("grep", "Port", "/etc/ssh/sshd_config")
	output, err = cmd.Output()
	if err == nil {
		// 解析端口配置
		control.Port = 22 // 默认端口
	}

	return control, nil
}

// checkPortsPermissions 检查端口权限
func (m *Manager) checkPortsPermissions() (*PortsControl, error) {
	control := &PortsControl{
		OpenPorts:   []int{},
		ClosedPorts: []int{},
	}

	// 检查开放端口
	cmd := exec.Command("netstat", "-tlnp")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "LISTEN") {
				// 解析端口信息
				// 简化处理
			}
		}
	}

	return control, nil
}

// checkFirewallPermissions 检查防火墙权限
func (m *Manager) checkFirewallPermissions() (*FirewallControl, error) {
	control := &FirewallControl{}

	// 检查防火墙状态
	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err == nil {
		control.Enabled = strings.Contains(string(output), "active")
	}

	return control, nil
}
