package system

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// TeamChecker 开发团队检查器
type TeamChecker struct {
	config *TeamConfig
	db     *sql.DB
}

// TeamConfig 团队配置
type TeamConfig struct {
	DatabaseConfig DatabaseConfig `json:"database_config"`
	RequiredRoles  []string       `json:"required_roles"`
	MinTeamSize    int            `json:"min_team_size"`
	MaxTeamSize    int            `json:"max_team_size"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// TeamStatus 团队状态
type TeamStatus struct {
	Timestamp        time.Time        `json:"timestamp"`
	OverallStatus    string           `json:"overall_status"` // success, warning, critical
	TeamComposition  TeamComposition  `json:"team_composition"`
	RoleDistribution RoleDistribution `json:"role_distribution"`
	PermissionMatrix PermissionMatrix `json:"permission_matrix"`
	Violations       []TeamViolation  `json:"violations"`
	Recommendations  []string         `json:"recommendations"`
}

// TeamComposition 团队组成
type TeamComposition struct {
	TotalMembers    int                    `json:"total_members"`
	ActiveMembers   int                    `json:"active_members"`
	InactiveMembers int                    `json:"inactive_members"`
	Members         []TeamMember           `json:"members"`
	TeamStructure   map[string]interface{} `json:"team_structure"`
}

// TeamMember 团队成员
type TeamMember struct {
	ID               int       `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	Permissions      []string  `json:"permissions"`
	Status           string    `json:"status"` // active, inactive, suspended
	JoinDate         time.Time `json:"join_date"`
	LastActive       time.Time `json:"last_active"`
	Department       string    `json:"department"`
	Responsibilities []string  `json:"responsibilities"`
}

// RoleDistribution 角色分布
type RoleDistribution struct {
	Roles           []RoleInfo     `json:"roles"`
	Distribution    map[string]int `json:"distribution"`
	IsBalanced      bool           `json:"is_balanced"`
	Recommendations []string       `json:"recommendations"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	RequiredCount    int      `json:"required_count"`
	CurrentCount     int      `json:"current_count"`
	Permissions      []string `json:"permissions"`
	Responsibilities []string `json:"responsibilities"`
	IsCritical       bool     `json:"is_critical"`
}

// PermissionMatrix 权限矩阵
type PermissionMatrix struct {
	Matrix          map[string]map[string]bool `json:"matrix"` // role -> permission -> allowed
	Violations      []TeamPermissionViolation  `json:"violations"`
	Recommendations []string                   `json:"recommendations"`
}

// TeamPermissionViolation 团队权限违规
type TeamPermissionViolation struct {
	Type           string `json:"type"` // excessive, insufficient, conflict
	Role           string `json:"role"`
	Permission     string `json:"permission"`
	Message        string `json:"message"`
	Severity       string `json:"severity"` // high, medium, low
	Recommendation string `json:"recommendation"`
}

// TeamViolation 团队违规
type TeamViolation struct {
	Type           string `json:"type"` // size, role, permission, structure
	Message        string `json:"message"`
	Severity       string `json:"severity"` // high, medium, low
	Recommendation string `json:"recommendation"`
	AffectedRole   string `json:"affected_role,omitempty"`
}

// NewTeamChecker 创建团队检查器
func NewTeamChecker(config *TeamConfig) (*TeamChecker, error) {
	if config == nil {
		config = getDefaultTeamConfig()
	}

	// 连接数据库
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.DatabaseConfig.Username,
		config.DatabaseConfig.Password,
		config.DatabaseConfig.Host,
		config.DatabaseConfig.Port,
		config.DatabaseConfig.Database))
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	return &TeamChecker{
		config: config,
		db:     db,
	}, nil
}

// getDefaultTeamConfig 获取默认团队配置
func getDefaultTeamConfig() *TeamConfig {
	return &TeamConfig{
		DatabaseConfig: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Username: "jobfirst",
			Password: "jobfirst_password_2024",
			Database: "jobfirst",
		},
		RequiredRoles: []string{
			"super_admin",     // 超级管理员
			"tech_lead",       // 技术负责人
			"backend_dev",     // 后端开发
			"frontend_dev",    // 前端开发
			"devops_engineer", // DevOps工程师
			"qa_engineer",     // 测试工程师
			"product_manager", // 产品经理
			"ui_designer",     // UI设计师
		},
		MinTeamSize: 4,
		MaxTeamSize: 20,
	}
}

// CheckTeamStatus 检查团队状态
func (tc *TeamChecker) CheckTeamStatus() (*TeamStatus, error) {
	status := &TeamStatus{
		Timestamp:       time.Now(),
		OverallStatus:   "success",
		Violations:      []TeamViolation{},
		Recommendations: []string{},
	}

	// 检查团队组成
	composition, err := tc.checkTeamComposition()
	if err != nil {
		return nil, fmt.Errorf("检查团队组成失败: %w", err)
	}
	status.TeamComposition = *composition

	// 检查角色分布
	roleDistribution, err := tc.checkRoleDistribution()
	if err != nil {
		return nil, fmt.Errorf("检查角色分布失败: %w", err)
	}
	status.RoleDistribution = *roleDistribution

	// 检查权限矩阵
	permissionMatrix, err := tc.checkPermissionMatrix()
	if err != nil {
		return nil, fmt.Errorf("检查权限矩阵失败: %w", err)
	}
	status.PermissionMatrix = *permissionMatrix

	// 检查团队违规
	tc.checkTeamViolations(status)

	// 生成建议
	tc.generateRecommendations(status)

	// 确定整体状态
	tc.determineOverallStatus(status)

	return status, nil
}

// checkTeamComposition 检查团队组成
func (tc *TeamChecker) checkTeamComposition() (*TeamComposition, error) {
	composition := &TeamComposition{
		Members:       []TeamMember{},
		TeamStructure: make(map[string]interface{}),
	}

	// 查询团队成员
	query := `
		SELECT u.id, u.username, u.email, u.status, u.created_at, u.last_login,
		       r.name as role_name, r.description as role_description,
		       GROUP_CONCAT(p.name) as permissions
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		LEFT JOIN permissions p ON rp.permission_id = p.id
		WHERE r.name IN ('super_admin', 'tech_lead', 'backend_dev', 'frontend_dev', 
		                 'devops_engineer', 'qa_engineer', 'product_manager', 'ui_designer')
		GROUP BY u.id, u.username, u.email, u.status, u.created_at, u.last_login, r.name, r.description
		ORDER BY u.created_at DESC
	`

	rows, err := tc.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询团队成员失败: %w", err)
	}
	defer rows.Close()

	memberMap := make(map[int]*TeamMember)

	for rows.Next() {
		var member TeamMember
		var permissions sql.NullString
		var lastActive sql.NullTime

		err := rows.Scan(&member.ID, &member.Username, &member.Email, &member.Status,
			&member.JoinDate, &lastActive, &member.Role, &member.Department, &permissions)
		if err != nil {
			continue
		}

		if lastActive.Valid {
			member.LastActive = lastActive.Time
		}

		if permissions.Valid {
			// 解析权限字符串
			// 这里需要根据实际权限格式进行解析
		}

		if existingMember, exists := memberMap[member.ID]; exists {
			// 如果用户已存在，添加角色信息
			existingMember.Role += ", " + member.Role
		} else {
			memberMap[member.ID] = &member
		}
	}

	// 转换为切片
	for _, member := range memberMap {
		composition.Members = append(composition.Members, *member)
		if member.Status == "active" {
			composition.ActiveMembers++
		} else {
			composition.InactiveMembers++
		}
	}

	composition.TotalMembers = len(composition.Members)

	// 构建团队结构
	tc.buildTeamStructure(composition)

	return composition, nil
}

// buildTeamStructure 构建团队结构
func (tc *TeamChecker) buildTeamStructure(composition *TeamComposition) {
	structure := make(map[string]interface{})

	// 按角色分组
	roleGroups := make(map[string][]TeamMember)
	for _, member := range composition.Members {
		roleGroups[member.Role] = append(roleGroups[member.Role], member)
	}

	structure["role_groups"] = roleGroups
	structure["hierarchy"] = tc.getTeamHierarchy()
	structure["communication_channels"] = tc.getCommunicationChannels()

	composition.TeamStructure = structure
}

// getTeamHierarchy 获取团队层级
func (tc *TeamChecker) getTeamHierarchy() map[string]interface{} {
	return map[string]interface{}{
		"level_1": []string{"super_admin"},
		"level_2": []string{"tech_lead", "product_manager"},
		"level_3": []string{"backend_dev", "frontend_dev", "devops_engineer"},
		"level_4": []string{"qa_engineer", "ui_designer"},
	}
}

// getCommunicationChannels 获取沟通渠道
func (tc *TeamChecker) getCommunicationChannels() []string {
	return []string{
		"daily_standup",
		"tech_discussion",
		"product_planning",
		"code_review",
		"incident_response",
	}
}

// checkRoleDistribution 检查角色分布
func (tc *TeamChecker) checkRoleDistribution() (*RoleDistribution, error) {
	distribution := &RoleDistribution{
		Roles:           []RoleInfo{},
		Distribution:    make(map[string]int),
		Recommendations: []string{},
	}

	// 定义角色要求
	roleRequirements := map[string]RoleInfo{
		"super_admin": {
			Name:             "super_admin",
			Description:      "超级管理员",
			RequiredCount:    1,
			Permissions:      []string{"system_management", "user_management", "role_management", "permission_management"},
			Responsibilities: []string{"系统管理", "用户管理", "角色管理", "权限管理"},
			IsCritical:       true,
		},
		"tech_lead": {
			Name:             "tech_lead",
			Description:      "技术负责人",
			RequiredCount:    1,
			Permissions:      []string{"code_review", "architecture_decision", "team_management"},
			Responsibilities: []string{"技术架构", "代码审查", "团队管理"},
			IsCritical:       true,
		},
		"backend_dev": {
			Name:             "backend_dev",
			Description:      "后端开发",
			RequiredCount:    2,
			Permissions:      []string{"api_development", "database_management", "microservice_development"},
			Responsibilities: []string{"API开发", "数据库管理", "微服务开发"},
			IsCritical:       true,
		},
		"frontend_dev": {
			Name:             "frontend_dev",
			Description:      "前端开发",
			RequiredCount:    2,
			Permissions:      []string{"ui_development", "user_experience", "frontend_architecture"},
			Responsibilities: []string{"UI开发", "用户体验", "前端架构"},
			IsCritical:       true,
		},
		"devops_engineer": {
			Name:             "devops_engineer",
			Description:      "DevOps工程师",
			RequiredCount:    1,
			Permissions:      []string{"deployment", "monitoring", "infrastructure_management"},
			Responsibilities: []string{"部署管理", "系统监控", "基础设施管理"},
			IsCritical:       true,
		},
		"qa_engineer": {
			Name:             "qa_engineer",
			Description:      "测试工程师",
			RequiredCount:    1,
			Permissions:      []string{"testing", "quality_assurance", "bug_tracking"},
			Responsibilities: []string{"测试执行", "质量保证", "缺陷跟踪"},
			IsCritical:       false,
		},
		"product_manager": {
			Name:             "product_manager",
			Description:      "产品经理",
			RequiredCount:    1,
			Permissions:      []string{"product_planning", "requirement_analysis", "stakeholder_management"},
			Responsibilities: []string{"产品规划", "需求分析", "利益相关者管理"},
			IsCritical:       false,
		},
		"ui_designer": {
			Name:             "ui_designer",
			Description:      "UI设计师",
			RequiredCount:    1,
			Permissions:      []string{"ui_design", "user_research", "prototype_design"},
			Responsibilities: []string{"UI设计", "用户研究", "原型设计"},
			IsCritical:       false,
		},
	}

	// 统计当前角色分布
	for _, roleName := range tc.config.RequiredRoles {
		count, err := tc.getRoleMemberCount(roleName)
		if err != nil {
			continue
		}

		distribution.Distribution[roleName] = count

		if roleInfo, exists := roleRequirements[roleName]; exists {
			roleInfo.CurrentCount = count
			distribution.Roles = append(distribution.Roles, roleInfo)
		}
	}

	// 检查角色平衡
	tc.checkRoleBalance(distribution)

	return distribution, nil
}

// getRoleMemberCount 获取角色成员数量
func (tc *TeamChecker) getRoleMemberCount(roleName string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.name = ? AND u.status = 'active'
	`

	var count int
	err := tc.db.QueryRow(query, roleName).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// checkRoleBalance 检查角色平衡
func (tc *TeamChecker) checkRoleBalance(distribution *RoleDistribution) {
	isBalanced := true

	for _, role := range distribution.Roles {
		if role.CurrentCount < role.RequiredCount {
			isBalanced = false
			if role.IsCritical {
				distribution.Recommendations = append(distribution.Recommendations,
					fmt.Sprintf("关键角色 %s 人员不足，当前 %d 人，需要 %d 人", role.Name, role.CurrentCount, role.RequiredCount))
			} else {
				distribution.Recommendations = append(distribution.Recommendations,
					fmt.Sprintf("角色 %s 人员不足，当前 %d 人，建议 %d 人", role.Name, role.CurrentCount, role.RequiredCount))
			}
		} else if role.CurrentCount > role.RequiredCount*2 {
			distribution.Recommendations = append(distribution.Recommendations,
				fmt.Sprintf("角色 %s 人员过多，当前 %d 人，建议不超过 %d 人", role.Name, role.CurrentCount, role.RequiredCount*2))
		}
	}

	distribution.IsBalanced = isBalanced
}

// checkPermissionMatrix 检查权限矩阵
func (tc *TeamChecker) checkPermissionMatrix() (*PermissionMatrix, error) {
	matrix := &PermissionMatrix{
		Matrix:          make(map[string]map[string]bool),
		Violations:      []TeamPermissionViolation{},
		Recommendations: []string{},
	}

	// 定义标准权限矩阵
	standardMatrix := tc.getStandardPermissionMatrix()

	// 从数据库获取实际权限矩阵
	actualMatrix, err := tc.getActualPermissionMatrix()
	if err != nil {
		return nil, fmt.Errorf("获取实际权限矩阵失败: %w", err)
	}

	// 比较标准矩阵和实际矩阵
	tc.comparePermissionMatrices(standardMatrix, actualMatrix, matrix)

	matrix.Matrix = actualMatrix

	return matrix, nil
}

// getStandardPermissionMatrix 获取标准权限矩阵
func (tc *TeamChecker) getStandardPermissionMatrix() map[string]map[string]bool {
	return map[string]map[string]bool{
		"super_admin": {
			"system_management":     true,
			"user_management":       true,
			"role_management":       true,
			"permission_management": true,
			"code_review":           true,
			"deployment":            true,
			"monitoring":            true,
		},
		"tech_lead": {
			"code_review":           true,
			"architecture_decision": true,
			"team_management":       true,
			"deployment":            true,
			"monitoring":            true,
		},
		"backend_dev": {
			"api_development":          true,
			"database_management":      true,
			"microservice_development": true,
			"code_review":              false,
		},
		"frontend_dev": {
			"ui_development":        true,
			"user_experience":       true,
			"frontend_architecture": true,
			"code_review":           false,
		},
		"devops_engineer": {
			"deployment":                true,
			"monitoring":                true,
			"infrastructure_management": true,
			"system_management":         false,
		},
		"qa_engineer": {
			"testing":           true,
			"quality_assurance": true,
			"bug_tracking":      true,
			"deployment":        false,
		},
		"product_manager": {
			"product_planning":       true,
			"requirement_analysis":   true,
			"stakeholder_management": true,
			"code_review":            false,
		},
		"ui_designer": {
			"ui_design":        true,
			"user_research":    true,
			"prototype_design": true,
			"deployment":       false,
		},
	}
}

// getActualPermissionMatrix 获取实际权限矩阵
func (tc *TeamChecker) getActualPermissionMatrix() (map[string]map[string]bool, error) {
	matrix := make(map[string]map[string]bool)

	query := `
		SELECT r.name as role_name, p.name as permission_name
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE r.name IN ('super_admin', 'tech_lead', 'backend_dev', 'frontend_dev', 
		                 'devops_engineer', 'qa_engineer', 'product_manager', 'ui_designer')
	`

	rows, err := tc.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleName, permissionName string
		if err := rows.Scan(&roleName, &permissionName); err != nil {
			continue
		}

		if matrix[roleName] == nil {
			matrix[roleName] = make(map[string]bool)
		}
		matrix[roleName][permissionName] = true
	}

	return matrix, nil
}

// comparePermissionMatrices 比较权限矩阵
func (tc *TeamChecker) comparePermissionMatrices(standard, actual map[string]map[string]bool, result *PermissionMatrix) {
	for role, permissions := range standard {
		if actual[role] == nil {
			result.Violations = append(result.Violations, TeamPermissionViolation{
				Type:           "insufficient",
				Role:           role,
				Permission:     "all",
				Message:        fmt.Sprintf("角色 %s 缺少所有权限配置", role),
				Severity:       "high",
				Recommendation: fmt.Sprintf("为角色 %s 配置必要的权限", role),
			})
			continue
		}

		for permission, allowed := range permissions {
			actualAllowed := actual[role][permission]

			if allowed && !actualAllowed {
				result.Violations = append(result.Violations, TeamPermissionViolation{
					Type:           "insufficient",
					Role:           role,
					Permission:     permission,
					Message:        fmt.Sprintf("角色 %s 缺少必要权限 %s", role, permission),
					Severity:       "medium",
					Recommendation: fmt.Sprintf("为角色 %s 添加权限 %s", role, permission),
				})
			} else if !allowed && actualAllowed {
				result.Violations = append(result.Violations, TeamPermissionViolation{
					Type:           "excessive",
					Role:           role,
					Permission:     permission,
					Message:        fmt.Sprintf("角色 %s 拥有过多权限 %s", role, permission),
					Severity:       "low",
					Recommendation: fmt.Sprintf("考虑移除角色 %s 的权限 %s", role, permission),
				})
			}
		}
	}
}

// checkTeamViolations 检查团队违规
func (tc *TeamChecker) checkTeamViolations(status *TeamStatus) {
	// 检查团队规模
	if status.TeamComposition.TotalMembers < tc.config.MinTeamSize {
		status.Violations = append(status.Violations, TeamViolation{
			Type:           "size",
			Message:        fmt.Sprintf("团队规模过小，当前 %d 人，最少需要 %d 人", status.TeamComposition.TotalMembers, tc.config.MinTeamSize),
			Severity:       "high",
			Recommendation: "招聘更多团队成员",
		})
	}

	if status.TeamComposition.TotalMembers > tc.config.MaxTeamSize {
		status.Violations = append(status.Violations, TeamViolation{
			Type:           "size",
			Message:        fmt.Sprintf("团队规模过大，当前 %d 人，建议不超过 %d 人", status.TeamComposition.TotalMembers, tc.config.MaxTeamSize),
			Severity:       "medium",
			Recommendation: "考虑团队拆分或优化管理结构",
		})
	}

	// 检查关键角色
	for _, role := range status.RoleDistribution.Roles {
		if role.IsCritical && role.CurrentCount == 0 {
			status.Violations = append(status.Violations, TeamViolation{
				Type:           "role",
				Message:        fmt.Sprintf("缺少关键角色 %s", role.Name),
				Severity:       "high",
				Recommendation: fmt.Sprintf("立即招聘或指定 %s 角色", role.Name),
				AffectedRole:   role.Name,
			})
		}
	}

	// 检查权限违规
	for _, violation := range status.PermissionMatrix.Violations {
		if violation.Severity == "high" {
			status.Violations = append(status.Violations, TeamViolation{
				Type:           "permission",
				Message:        violation.Message,
				Severity:       violation.Severity,
				Recommendation: violation.Recommendation,
			})
		}
	}
}

// generateRecommendations 生成建议
func (tc *TeamChecker) generateRecommendations(status *TeamStatus) {
	if len(status.Violations) == 0 {
		status.Recommendations = append(status.Recommendations, "团队配置良好，所有关键角色和权限都已正确设置")
		return
	}

	// 根据违规类型生成建议
	status.Recommendations = append(status.Recommendations, "团队配置需要优化，建议采取以下措施：")

	for _, violation := range status.Violations {
		status.Recommendations = append(status.Recommendations, fmt.Sprintf("- %s", violation.Recommendation))
	}

	// 添加通用建议
	status.Recommendations = append(status.Recommendations, "")
	status.Recommendations = append(status.Recommendations, "通用建议：")
	status.Recommendations = append(status.Recommendations, "- 定期审查团队成员角色和权限")
	status.Recommendations = append(status.Recommendations, "- 确保关键角色有备份人员")
	status.Recommendations = append(status.Recommendations, "- 建立清晰的职责分工和沟通机制")
}

// determineOverallStatus 确定整体状态
func (tc *TeamChecker) determineOverallStatus(status *TeamStatus) {
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

// Close 关闭数据库连接
func (tc *TeamChecker) Close() error {
	if tc.db != nil {
		return tc.db.Close()
	}
	return nil
}
