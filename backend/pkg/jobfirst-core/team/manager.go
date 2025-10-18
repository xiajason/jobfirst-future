package team

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jobfirst/jobfirst-core/auth"
	"gorm.io/gorm"
)

// Manager 团队管理器
type Manager struct {
	db *gorm.DB
}

// NewManager 创建团队管理器
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// AddMember 添加团队成员
func (tm *Manager) AddMember(req AddMemberRequest) (*AddMemberResponse, error) {
	// 验证角色
	validRoles := []string{"system_admin", "dev_lead", "frontend_dev", "backend_dev", "qa_engineer", "guest"}
	if !contains(validRoles, req.TeamRole) {
		return nil, fmt.Errorf("无效的角色，支持的角色: %v", validRoles)
	}

	// 检查用户名和邮箱是否已存在
	var existingUser auth.User
	if err := tm.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("用户名或邮箱已存在")
	}

	// 创建用户
	user := auth.User{
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  hashPassword(req.Password),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := tm.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 根据角色设置权限
	permissions := getRolePermissions(req.TeamRole)

	// 创建开发团队成员记录
	devTeam := auth.DevTeamUser{
		UserID:                    user.ID,
		TeamRole:                  req.TeamRole,
		ServerAccessLevel:         permissions["server_access_level"],
		CodeAccessModules:         permissions["code_access_modules"],
		DatabaseAccess:            permissions["database_access"],
		ServiceRestartPermissions: permissions["service_restart_permissions"],
		Status:                    "active",
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	if err := tm.db.Create(&devTeam).Error; err != nil {
		// 如果创建开发团队成员记录失败，删除用户
		tm.db.Delete(&user)
		return nil, fmt.Errorf("创建团队成员记录失败: %w", err)
	}

	return &AddMemberResponse{
		Success: true,
		Message: "团队成员添加成功",
		Data: AddMemberData{
			User:    user,
			DevTeam: devTeam,
		},
	}, nil
}

// UpdateMember 更新团队成员
func (tm *Manager) UpdateMember(memberID uint, req UpdateMemberRequest) (*UpdateMemberResponse, error) {
	var member auth.DevTeamUser
	if err := tm.db.Where("id = ?", memberID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("团队成员不存在: %w", err)
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.TeamRole != "" {
		updates["team_role"] = req.TeamRole
	}
	if req.ServerAccessLevel != "" {
		updates["server_access_level"] = req.ServerAccessLevel
	}
	if req.CodeAccessModules != "" {
		updates["code_access_modules"] = req.CodeAccessModules
	}
	if req.DatabaseAccess != "" {
		updates["database_access"] = req.DatabaseAccess
	}
	if req.ServiceRestartPermissions != "" {
		updates["service_restart_permissions"] = req.ServiceRestartPermissions
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now()

	if err := tm.db.Model(&member).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新团队成员失败: %w", err)
	}

	return &UpdateMemberResponse{
		Success: true,
		Message: "团队成员更新成功",
		Data:    member,
	}, nil
}

// RemoveMember 移除团队成员
func (tm *Manager) RemoveMember(memberID uint) (*RemoveMemberResponse, error) {
	var member auth.DevTeamUser
	if err := tm.db.Where("id = ?", memberID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("团队成员不存在: %w", err)
	}

	// 软删除团队成员记录
	if err := tm.db.Model(&member).Update("deleted_at", time.Now()).Error; err != nil {
		return nil, fmt.Errorf("移除团队成员失败: %w", err)
	}

	return &RemoveMemberResponse{
		Success: true,
		Message: "团队成员移除成功",
	}, nil
}

// GetMembers 获取团队成员列表
func (tm *Manager) GetMembers(req GetMembersRequest) (*GetMembersResponse, error) {
	var members []auth.DevTeamUser

	query := tm.db.Preload("User").Where("deleted_at IS NULL")

	if req.Role != "" {
		query = query.Where("team_role = ?", req.Role)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	var total int64
	query.Model(&auth.DevTeamUser{}).Count(&total)

	// 获取分页数据
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&members).Error; err != nil {
		return nil, fmt.Errorf("获取团队成员列表失败: %w", err)
	}

	return &GetMembersResponse{
		Success: true,
		Data: GetMembersData{
			Members: members,
			Pagination: Pagination{
				Page:       req.Page,
				PageSize:   req.PageSize,
				Total:      total,
				TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
			},
		},
	}, nil
}

// GetStats 获取团队统计信息
func (tm *Manager) GetStats() (*GetStatsResponse, error) {
	var stats TeamStats

	// 总成员数
	tm.db.Model(&auth.DevTeamUser{}).Where("deleted_at IS NULL").Count(&stats.TotalMembers)

	// 活跃成员数
	tm.db.Model(&auth.DevTeamUser{}).Where("status = 'active' AND deleted_at IS NULL").Count(&stats.ActiveMembers)

	// 非活跃成员数
	tm.db.Model(&auth.DevTeamUser{}).Where("status != 'active' AND deleted_at IS NULL").Count(&stats.InactiveMembers)

	// 角色统计
	stats.RoleStats = make(map[string]int64)
	roles := []string{"super_admin", "system_admin", "dev_lead", "frontend_dev", "backend_dev", "qa_engineer", "guest"}
	for _, role := range roles {
		var count int64
		tm.db.Model(&auth.DevTeamUser{}).Where("team_role = ? AND deleted_at IS NULL", role).Count(&count)
		stats.RoleStats[role] = count
	}

	return &GetStatsResponse{
		Success: true,
		Data:    stats,
	}, nil
}

// GetOperationLogs 获取操作日志
func (tm *Manager) GetOperationLogs(req GetOperationLogsRequest) (*GetOperationLogsResponse, error) {
	var logs []auth.DevOperationLog

	query := tm.db.Preload("User")

	if req.UserID != 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.OperationType != "" {
		query = query.Where("operation_type = ?", req.OperationType)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	var total int64
	query.Model(&auth.DevOperationLog{}).Count(&total)

	// 获取分页数据
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("获取操作日志失败: %w", err)
	}

	return &GetOperationLogsResponse{
		Success: true,
		Data: GetOperationLogsData{
			Logs: logs,
			Pagination: Pagination{
				Page:       req.Page,
				PageSize:   req.PageSize,
				Total:      total,
				TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
			},
		},
	}, nil
}

// LogOperation 记录操作日志
func (tm *Manager) LogOperation(userID uint, operationType, operationTarget string, details map[string]interface{}, ipAddress, userAgent, status string) error {
	detailsJSON, _ := json.Marshal(details)
	log := auth.DevOperationLog{
		UserID:           userID,
		OperationType:    operationType,
		OperationTarget:  operationTarget,
		OperationDetails: string(detailsJSON),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Status:           status,
		CreatedAt:        time.Now(),
	}
	return tm.db.Create(&log).Error
}

// 辅助方法

// getRolePermissions 获取角色权限
func getRolePermissions(role string) map[string]string {
	permissions := map[string]map[string]string{
		"system_admin": {
			"server_access_level":         "full",
			"code_access_modules":         `["frontend", "backend", "database", "config"]`,
			"database_access":             `["system"]`,
			"service_restart_permissions": `["system"]`,
		},
		"dev_lead": {
			"server_access_level":         "limited",
			"code_access_modules":         `["frontend", "backend"]`,
			"database_access":             `["development"]`,
			"service_restart_permissions": `["backend"]`,
		},
		"frontend_dev": {
			"server_access_level":         "limited",
			"code_access_modules":         `["frontend"]`,
			"database_access":             `[]`,
			"service_restart_permissions": `[]`,
		},
		"backend_dev": {
			"server_access_level":         "limited",
			"code_access_modules":         `["backend"]`,
			"database_access":             `["development"]`,
			"service_restart_permissions": `["backend"]`,
		},
		"qa_engineer": {
			"server_access_level":         "limited",
			"code_access_modules":         `["test"]`,
			"database_access":             `["test"]`,
			"service_restart_permissions": `[]`,
		},
		"guest": {
			"server_access_level":         "readonly",
			"code_access_modules":         `[]`,
			"database_access":             `[]`,
			"service_restart_permissions": `[]`,
		},
	}

	if perm, exists := permissions[role]; exists {
		return perm
	}

	// 默认权限
	return map[string]string{
		"server_access_level":         "readonly",
		"code_access_modules":         `[]`,
		"database_access":             `[]`,
		"service_restart_permissions": `[]`,
	}
}

// hashPassword 哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
