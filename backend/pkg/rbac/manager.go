package rbac

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"gorm.io/gorm"
)

type Manager struct {
	enforcer *casbin.Enforcer
	db       *gorm.DB
}

func NewManager(db *gorm.DB) (*Manager, error) {
	// 创建Casbin模型
	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	// 创建enforcer
	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	manager := &Manager{
		enforcer: enforcer,
		db:       db,
	}

	// 初始化默认策略
	if err := manager.InitializeDefaultPolicies(); err != nil {
		return nil, fmt.Errorf("failed to initialize default policies: %w", err)
	}

	return manager, nil
}

func (m *Manager) HasPermission(sub, obj, act string) (bool, error) {
	return m.enforcer.Enforce(sub, obj, act)
}

func (m *Manager) HasRole(user, role string) (bool, error) {
	return m.enforcer.HasRoleForUser(user, role)
}

func (m *Manager) AddPolicy(sub, obj, act string) error {
	_, err := m.enforcer.AddPolicy(sub, obj, act)
	return err
}

func (m *Manager) AddGroupingPolicy(user, role string) error {
	_, err := m.enforcer.AddGroupingPolicy(user, role)
	return err
}

func (m *Manager) GetRolesForUser(user string) ([]string, error) {
	return m.enforcer.GetRolesForUser(user)
}

func (m *Manager) GetPermissionsForUser(user string) ([]string, error) {
	permissions, err := m.enforcer.GetPermissionsForUser(user)
	if err != nil {
		return nil, err
	}

	// 将 [][]string 转换为 []string
	var result []string
	for _, perm := range permissions {
		if len(perm) >= 3 {
			result = append(result, perm[1]+":"+perm[2]) // resource:action
		}
	}
	return result, nil
}

func (m *Manager) InitializeDefaultPolicies() error {
	// 添加角色
	roles := []string{"super_admin", "admin", "dev_team", "user"}
	for _, role := range roles {
		if err := m.AddGroupingPolicy(role, role); err != nil {
			return fmt.Errorf("failed to add role %s: %w", role, err)
		}
	}

	// 添加权限策略
	policies := [][]string{
		// 超级管理员权限
		{"super_admin", "user", "read"},
		{"super_admin", "user", "write"},
		{"super_admin", "user", "delete"},
		{"super_admin", "role", "read"},
		{"super_admin", "role", "write"},
		{"super_admin", "role", "delete"},
		{"super_admin", "permission", "read"},
		{"super_admin", "permission", "write"},
		{"super_admin", "permission", "delete"},
		{"super_admin", "system", "read"},
		{"super_admin", "system", "write"},
		{"super_admin", "system", "delete"},

		// 管理员权限
		{"admin", "user", "read"},
		{"admin", "user", "write"},
		{"admin", "role", "read"},
		{"admin", "permission", "read"},

		// 开发团队权限
		{"dev_team", "user", "read"},
		{"dev_team", "code", "read"},
		{"dev_team", "code", "write"},
		{"dev_team", "database", "read"},

		// 普通用户权限
		{"user", "profile", "read"},
		{"user", "profile", "write"},
		{"user", "resume", "read"},
		{"user", "resume", "write"},
	}

	for _, policy := range policies {
		if err := m.AddPolicy(policy[0], policy[1], policy[2]); err != nil {
			return fmt.Errorf("failed to add policy %v: %w", policy, err)
		}
	}

	return nil
}
