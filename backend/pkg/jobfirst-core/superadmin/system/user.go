package system

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// UserChecker 用户权限和订阅状态检查器
type UserChecker struct {
	config *UserConfig
	db     *sql.DB
}

// UserConfig 用户配置
type UserConfig struct {
	DatabaseConfig    DatabaseConfig    `json:"database_config"`
	SubscriptionRules SubscriptionRules `json:"subscription_rules"`
	PermissionRules   PermissionRules   `json:"permission_rules"`
}

// SubscriptionRules 订阅规则
type SubscriptionRules struct {
	FreeTrialDays     int                `json:"free_trial_days"`
	SubscriptionPlans []SubscriptionPlan `json:"subscription_plans"`
	GracePeriodDays   int                `json:"grace_period_days"`
}

// SubscriptionPlan 订阅计划
type SubscriptionPlan struct {
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Duration    int      `json:"duration"` // 天数
	Permissions []string `json:"permissions"`
	Features    []string `json:"features"`
	MaxUsers    int      `json:"max_users"`
	IsActive    bool     `json:"is_active"`
}

// PermissionRules 权限规则
type PermissionRules struct {
	DefaultPermissions []string            `json:"default_permissions"`
	RolePermissions    map[string][]string `json:"role_permissions"`
	TestUserRoles      []string            `json:"test_user_roles"`
}

// UserStatus 用户状态
type UserStatus struct {
	Timestamp          time.Time          `json:"timestamp"`
	OverallStatus      string             `json:"overall_status"` // success, warning, critical
	UserStatistics     UserStatistics     `json:"user_statistics"`
	SubscriptionStatus SubscriptionStatus `json:"subscription_status"`
	PermissionStatus   PermissionStatus   `json:"permission_status"`
	TestUserStatus     TestUserStatus     `json:"test_user_status"`
	Violations         []UserViolation    `json:"violations"`
	Recommendations    []string           `json:"recommendations"`
}

// UserStatistics 用户统计
type UserStatistics struct {
	TotalUsers       int              `json:"total_users"`
	ActiveUsers      int              `json:"active_users"`
	SubscribedUsers  int              `json:"subscribed_users"`
	TestUsers        int              `json:"test_users"`
	ExpiredUsers     int              `json:"expired_users"`
	UserGrowth       []UserGrowthData `json:"user_growth"`
	UserDistribution map[string]int   `json:"user_distribution"`
}

// UserGrowthData 用户增长数据
type UserGrowthData struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

// SubscriptionStatus 订阅状态
type SubscriptionStatus struct {
	ActiveSubscriptions  []SubscriptionInfo      `json:"active_subscriptions"`
	ExpiredSubscriptions []SubscriptionInfo      `json:"expired_subscriptions"`
	TrialUsers           []TrialUserInfo         `json:"trial_users"`
	SubscriptionRevenue  SubscriptionRevenue     `json:"subscription_revenue"`
	Violations           []SubscriptionViolation `json:"violations"`
}

// SubscriptionInfo 订阅信息
type SubscriptionInfo struct {
	UserID        int       `json:"user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PlanName      string    `json:"plan_name"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Status        string    `json:"status"` // active, expired, cancelled
	AutoRenew     bool      `json:"auto_renew"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
}

// TrialUserInfo 试用用户信息
type TrialUserInfo struct {
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	TrialStart time.Time `json:"trial_start"`
	TrialEnd   time.Time `json:"trial_end"`
	DaysLeft   int       `json:"days_left"`
	IsExpired  bool      `json:"is_expired"`
}

// SubscriptionRevenue 订阅收入
type SubscriptionRevenue struct {
	MonthlyRevenue      float64 `json:"monthly_revenue"`
	YearlyRevenue       float64 `json:"yearly_revenue"`
	TotalRevenue        float64 `json:"total_revenue"`
	ActiveSubscriptions int     `json:"active_subscriptions"`
	ChurnRate           float64 `json:"churn_rate"`
}

// SubscriptionViolation 订阅违规
type SubscriptionViolation struct {
	Type           string `json:"type"` // expired, unauthorized, payment_failed
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	Message        string `json:"message"`
	Severity       string `json:"severity"` // high, medium, low
	Recommendation string `json:"recommendation"`
}

// PermissionStatus 权限状态
type PermissionStatus struct {
	ValidPermissions   []PermissionInfo      `json:"valid_permissions"`
	InvalidPermissions []PermissionViolation `json:"invalid_permissions"`
	PermissionMatrix   map[string][]string   `json:"permission_matrix"`
	Recommendations    []string              `json:"recommendations"`
}

// PermissionInfo 权限信息
type PermissionInfo struct {
	UserID      int       `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	IsValid     bool      `json:"is_valid"`
	LastUpdated time.Time `json:"last_updated"`
}

// PermissionViolation 权限违规
type PermissionViolation struct {
	Type           string `json:"type"` // excessive, insufficient, unauthorized
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	Permission     string `json:"permission"`
	Message        string `json:"message"`
	Severity       string `json:"severity"` // high, medium, low
	Recommendation string `json:"recommendation"`
}

// TestUserStatus 测试用户状态
type TestUserStatus struct {
	TestUsers       []TestUserInfo       `json:"test_users"`
	TestPermissions []TestPermissionInfo `json:"test_permissions"`
	Violations      []TestUserViolation  `json:"violations"`
	Recommendations []string             `json:"recommendations"`
}

// TestUserInfo 测试用户信息
type TestUserInfo struct {
	UserID      int        `json:"user_id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	InvitedBy   string     `json:"invited_by"`
	InviteDate  time.Time  `json:"invite_date"`
	ExpiryDate  time.Time  `json:"expiry_date"`
	Status      string     `json:"status"` // active, expired, revoked
	Permissions []string   `json:"permissions"`
	UsageStats  UsageStats `json:"usage_stats"`
	DaysLeft    int        `json:"days_left"`
	IsExpired   bool       `json:"is_expired"`
}

// TestPermissionInfo 测试权限信息
type TestPermissionInfo struct {
	UserID      int      `json:"user_id"`
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	IsValid     bool     `json:"is_valid"`
	Reason      string   `json:"reason"`
}

// UsageStats 使用统计
type UsageStats struct {
	LoginCount   int            `json:"login_count"`
	LastLogin    time.Time      `json:"last_login"`
	APICalls     int            `json:"api_calls"`
	DataUsage    int64          `json:"data_usage"` // bytes
	FeatureUsage map[string]int `json:"feature_usage"`
}

// TestUserViolation 测试用户违规
type TestUserViolation struct {
	Type           string `json:"type"` // expired, unauthorized, excessive_usage
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	Message        string `json:"message"`
	Severity       string `json:"severity"` // high, medium, low
	Recommendation string `json:"recommendation"`
}

// UserViolation 用户违规
type UserViolation struct {
	Type           string `json:"type"` // subscription, permission, usage, security
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	Message        string `json:"message"`
	Severity       string `json:"severity"` // high, medium, low
	Recommendation string `json:"recommendation"`
}

// NewUserChecker 创建用户检查器
func NewUserChecker(config *UserConfig) (*UserChecker, error) {
	if config == nil {
		config = getDefaultUserConfig()
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

	return &UserChecker{
		config: config,
		db:     db,
	}, nil
}

// getDefaultUserConfig 获取默认用户配置
func getDefaultUserConfig() *UserConfig {
	return &UserConfig{
		DatabaseConfig: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Username: "jobfirst",
			Password: "jobfirst_password_2024",
			Database: "jobfirst",
		},
		SubscriptionRules: SubscriptionRules{
			FreeTrialDays:   14,
			GracePeriodDays: 3,
			SubscriptionPlans: []SubscriptionPlan{
				{
					Name:        "basic",
					Price:       29.99,
					Duration:    30,
					Permissions: []string{"basic_features", "api_access", "support"},
					Features:    []string{"简历管理", "基础模板", "邮件支持"},
					MaxUsers:    5,
					IsActive:    true,
				},
				{
					Name:        "professional",
					Price:       79.99,
					Duration:    30,
					Permissions: []string{"advanced_features", "api_access", "priority_support", "custom_templates"},
					Features:    []string{"高级模板", "数据分析", "优先支持", "自定义模板"},
					MaxUsers:    20,
					IsActive:    true,
				},
				{
					Name:        "enterprise",
					Price:       199.99,
					Duration:    30,
					Permissions: []string{"all_features", "api_access", "dedicated_support", "white_label"},
					Features:    []string{"所有功能", "白标服务", "专属支持", "无限用户"},
					MaxUsers:    999,
					IsActive:    true,
				},
			},
		},
		PermissionRules: PermissionRules{
			DefaultPermissions: []string{"basic_access", "profile_management"},
			RolePermissions: map[string][]string{
				"subscriber":      {"basic_features", "api_access", "support"},
				"premium_user":    {"advanced_features", "api_access", "priority_support"},
				"enterprise_user": {"all_features", "api_access", "dedicated_support"},
				"test_user":       {"limited_access", "basic_features"},
				"admin":           {"system_management", "user_management"},
			},
			TestUserRoles: []string{"test_user", "beta_tester", "demo_user"},
		},
	}
}

// CheckUserStatus 检查用户状态
func (uc *UserChecker) CheckUserStatus() (*UserStatus, error) {
	status := &UserStatus{
		Timestamp:       time.Now(),
		OverallStatus:   "success",
		Violations:      []UserViolation{},
		Recommendations: []string{},
	}

	// 检查用户统计
	statistics, err := uc.checkUserStatistics()
	if err != nil {
		return nil, fmt.Errorf("检查用户统计失败: %w", err)
	}
	status.UserStatistics = *statistics

	// 检查订阅状态
	subscriptionStatus, err := uc.checkSubscriptionStatus()
	if err != nil {
		return nil, fmt.Errorf("检查订阅状态失败: %w", err)
	}
	status.SubscriptionStatus = *subscriptionStatus

	// 检查权限状态
	permissionStatus, err := uc.checkPermissionStatus()
	if err != nil {
		return nil, fmt.Errorf("检查权限状态失败: %w", err)
	}
	status.PermissionStatus = *permissionStatus

	// 检查测试用户状态
	testUserStatus, err := uc.checkTestUserStatus()
	if err != nil {
		return nil, fmt.Errorf("检查测试用户状态失败: %w", err)
	}
	status.TestUserStatus = *testUserStatus

	// 检查用户违规
	uc.checkUserViolations(status)

	// 生成建议
	uc.generateRecommendations(status)

	// 确定整体状态
	uc.determineOverallStatus(status)

	return status, nil
}

// checkUserStatistics 检查用户统计
func (uc *UserChecker) checkUserStatistics() (*UserStatistics, error) {
	statistics := &UserStatistics{
		UserDistribution: make(map[string]int),
	}

	// 查询用户总数
	err := uc.db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&statistics.ActiveUsers)
	if err != nil {
		return nil, err
	}

	err = uc.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&statistics.TotalUsers)
	if err != nil {
		return nil, err
	}

	// 查询订阅用户数
	err = uc.db.QueryRow(`
		SELECT COUNT(DISTINCT u.id) 
		FROM users u 
		JOIN subscriptions s ON u.id = s.user_id 
		WHERE s.status = 'active' AND s.end_date > NOW()
	`).Scan(&statistics.SubscribedUsers)
	if err != nil {
		// 如果订阅表不存在，设置为0
		statistics.SubscribedUsers = 0
	}

	// 查询测试用户数
	err = uc.db.QueryRow(`
		SELECT COUNT(*) 
		FROM users u 
		JOIN user_roles ur ON u.id = ur.user_id 
		JOIN roles r ON ur.role_id = r.id 
		WHERE r.name IN ('test_user', 'beta_tester', 'demo_user')
	`).Scan(&statistics.TestUsers)
	if err != nil {
		statistics.TestUsers = 0
	}

	// 查询过期用户数
	err = uc.db.QueryRow(`
		SELECT COUNT(DISTINCT u.id) 
		FROM users u 
		JOIN subscriptions s ON u.id = s.user_id 
		WHERE s.status = 'expired' OR s.end_date < NOW()
	`).Scan(&statistics.ExpiredUsers)
	if err != nil {
		statistics.ExpiredUsers = 0
	}

	// 查询用户增长数据
	uc.getUserGrowthData(statistics)

	// 查询用户分布
	uc.getUserDistribution(statistics)

	return statistics, nil
}

// getUserGrowthData 获取用户增长数据
func (uc *UserChecker) getUserGrowthData(statistics *UserStatistics) {
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	rows, err := uc.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var growth UserGrowthData
		err := rows.Scan(&growth.Date, &growth.Count)
		if err != nil {
			continue
		}
		statistics.UserGrowth = append(statistics.UserGrowth, growth)
	}
}

// getUserDistribution 获取用户分布
func (uc *UserChecker) getUserDistribution(statistics *UserStatistics) {
	query := `
		SELECT r.name as role_name, COUNT(*) as count
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		GROUP BY r.name
	`

	rows, err := uc.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var roleName string
		var count int
		err := rows.Scan(&roleName, &count)
		if err != nil {
			continue
		}
		statistics.UserDistribution[roleName] = count
	}
}

// checkSubscriptionStatus 检查订阅状态
func (uc *UserChecker) checkSubscriptionStatus() (*SubscriptionStatus, error) {
	status := &SubscriptionStatus{
		ActiveSubscriptions:  []SubscriptionInfo{},
		ExpiredSubscriptions: []SubscriptionInfo{},
		TrialUsers:           []TrialUserInfo{},
		Violations:           []SubscriptionViolation{},
	}

	// 查询活跃订阅
	activeQuery := `
		SELECT u.id, u.username, u.email, s.plan_name, s.start_date, s.end_date, 
		       s.status, s.auto_renew, s.amount, s.payment_method
		FROM users u
		JOIN subscriptions s ON u.id = s.user_id
		WHERE s.status = 'active' AND s.end_date > NOW()
		ORDER BY s.end_date ASC
	`

	rows, err := uc.db.Query(activeQuery)
	if err != nil {
		// 如果订阅表不存在，返回空状态
		return status, nil
	}
	defer rows.Close()

	for rows.Next() {
		var subscription SubscriptionInfo
		err := rows.Scan(&subscription.UserID, &subscription.Username, &subscription.Email,
			&subscription.PlanName, &subscription.StartDate, &subscription.EndDate,
			&subscription.Status, &subscription.AutoRenew, &subscription.Amount,
			&subscription.PaymentMethod)
		if err != nil {
			continue
		}
		status.ActiveSubscriptions = append(status.ActiveSubscriptions, subscription)
	}

	// 查询过期订阅
	expiredQuery := `
		SELECT u.id, u.username, u.email, s.plan_name, s.start_date, s.end_date, 
		       s.status, s.auto_renew, s.amount, s.payment_method
		FROM users u
		JOIN subscriptions s ON u.id = s.user_id
		WHERE s.status = 'expired' OR s.end_date < NOW()
		ORDER BY s.end_date DESC
	`

	rows, err = uc.db.Query(expiredQuery)
	if err != nil {
		return status, nil
	}
	defer rows.Close()

	for rows.Next() {
		var subscription SubscriptionInfo
		err := rows.Scan(&subscription.UserID, &subscription.Username, &subscription.Email,
			&subscription.PlanName, &subscription.StartDate, &subscription.EndDate,
			&subscription.Status, &subscription.AutoRenew, &subscription.Amount,
			&subscription.PaymentMethod)
		if err != nil {
			continue
		}
		status.ExpiredSubscriptions = append(status.ExpiredSubscriptions, subscription)
	}

	// 计算订阅收入
	uc.calculateSubscriptionRevenue(status)

	// 检查订阅违规
	uc.checkSubscriptionViolations(status)

	return status, nil
}

// calculateSubscriptionRevenue 计算订阅收入
func (uc *UserChecker) calculateSubscriptionRevenue(status *SubscriptionStatus) {
	revenue := &SubscriptionRevenue{}

	// 计算月度收入
	monthlyQuery := `
		SELECT COALESCE(SUM(amount), 0) as monthly_revenue
		FROM subscriptions
		WHERE status = 'active' 
		AND start_date >= DATE_SUB(NOW(), INTERVAL 1 MONTH)
		AND end_date > NOW()
	`

	err := uc.db.QueryRow(monthlyQuery).Scan(&revenue.MonthlyRevenue)
	if err != nil {
		revenue.MonthlyRevenue = 0
	}

	// 计算年度收入
	yearlyQuery := `
		SELECT COALESCE(SUM(amount), 0) as yearly_revenue
		FROM subscriptions
		WHERE status = 'active' 
		AND start_date >= DATE_SUB(NOW(), INTERVAL 1 YEAR)
		AND end_date > NOW()
	`

	err = uc.db.QueryRow(yearlyQuery).Scan(&revenue.YearlyRevenue)
	if err != nil {
		revenue.YearlyRevenue = 0
	}

	// 计算总收入和活跃订阅数
	revenue.TotalRevenue = revenue.MonthlyRevenue + revenue.YearlyRevenue
	revenue.ActiveSubscriptions = len(status.ActiveSubscriptions)

	// 计算流失率
	churnQuery := `
		SELECT COUNT(*) as churned_users
		FROM subscriptions
		WHERE status = 'expired' 
		AND end_date >= DATE_SUB(NOW(), INTERVAL 1 MONTH)
	`

	var churnedUsers int
	err = uc.db.QueryRow(churnQuery).Scan(&churnedUsers)
	if err != nil {
		churnedUsers = 0
	}

	if revenue.ActiveSubscriptions > 0 {
		revenue.ChurnRate = float64(churnedUsers) / float64(revenue.ActiveSubscriptions) * 100
	}

	status.SubscriptionRevenue = *revenue
}

// checkSubscriptionViolations 检查订阅违规
func (uc *UserChecker) checkSubscriptionViolations(status *SubscriptionStatus) {
	// 检查过期但仍在使用的用户
	for _, subscription := range status.ExpiredSubscriptions {
		// 检查用户是否仍在活跃使用系统
		var lastLogin time.Time
		query := "SELECT last_login FROM users WHERE id = ?"
		err := uc.db.QueryRow(query, subscription.UserID).Scan(&lastLogin)
		if err != nil {
			continue
		}

		// 如果过期后仍在登录，标记为违规
		if lastLogin.After(subscription.EndDate) {
			violation := SubscriptionViolation{
				Type:           "expired",
				UserID:         subscription.UserID,
				Username:       subscription.Username,
				Message:        fmt.Sprintf("用户 %s 订阅已过期但仍在使用系统", subscription.Username),
				Severity:       "high",
				Recommendation: fmt.Sprintf("立即限制用户 %s 的访问权限", subscription.Username),
			}
			status.Violations = append(status.Violations, violation)
		}
	}
}

// checkPermissionStatus 检查权限状态
func (uc *UserChecker) checkPermissionStatus() (*PermissionStatus, error) {
	status := &PermissionStatus{
		ValidPermissions:   []PermissionInfo{},
		InvalidPermissions: []PermissionViolation{},
		PermissionMatrix:   make(map[string][]string),
		Recommendations:    []string{},
	}

	// 查询用户权限
	query := `
		SELECT u.id, u.username, u.email, r.name as role_name, 
		       GROUP_CONCAT(p.name) as permissions, u.updated_at
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		LEFT JOIN permissions p ON rp.permission_id = p.id
		WHERE u.status = 'active'
		GROUP BY u.id, u.username, u.email, r.name, u.updated_at
		ORDER BY u.id
	`

	rows, err := uc.db.Query(query)
	if err != nil {
		return status, nil
	}
	defer rows.Close()

	for rows.Next() {
		var permissionInfo PermissionInfo
		var permissions sql.NullString
		var lastUpdated sql.NullTime

		err := rows.Scan(&permissionInfo.UserID, &permissionInfo.Username, &permissionInfo.Email,
			&permissionInfo.Role, &permissions, &lastUpdated)
		if err != nil {
			continue
		}

		if lastUpdated.Valid {
			permissionInfo.LastUpdated = lastUpdated.Time
		}

		if permissions.Valid {
			// 解析权限字符串
			// 这里需要根据实际权限格式进行解析
		}

		// 验证权限
		uc.validateUserPermissions(&permissionInfo, status)

		status.ValidPermissions = append(status.ValidPermissions, permissionInfo)
	}

	return status, nil
}

// validateUserPermissions 验证用户权限
func (uc *UserChecker) validateUserPermissions(permissionInfo *PermissionInfo, status *PermissionStatus) {
	// 检查角色是否存在
	if permissionInfo.Role == "" {
		violation := PermissionViolation{
			Type:           "insufficient",
			UserID:         permissionInfo.UserID,
			Username:       permissionInfo.Username,
			Permission:     "role",
			Message:        fmt.Sprintf("用户 %s 没有分配角色", permissionInfo.Username),
			Severity:       "high",
			Recommendation: fmt.Sprintf("为用户 %s 分配适当的角色", permissionInfo.Username),
		}
		status.InvalidPermissions = append(status.InvalidPermissions, violation)
		return
	}

	// 检查权限是否与角色匹配
	expectedPermissions, exists := uc.config.PermissionRules.RolePermissions[permissionInfo.Role]
	if !exists {
		violation := PermissionViolation{
			Type:           "unauthorized",
			UserID:         permissionInfo.UserID,
			Username:       permissionInfo.Username,
			Permission:     "role",
			Message:        fmt.Sprintf("用户 %s 的角色 %s 未定义", permissionInfo.Username, permissionInfo.Role),
			Severity:       "high",
			Recommendation: fmt.Sprintf("检查角色 %s 的配置", permissionInfo.Role),
		}
		status.InvalidPermissions = append(status.InvalidPermissions, violation)
		return
	}

	// 检查权限是否充足
	for _, expectedPerm := range expectedPermissions {
		hasPermission := false
		for _, userPerm := range permissionInfo.Permissions {
			if userPerm == expectedPerm {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			violation := PermissionViolation{
				Type:           "insufficient",
				UserID:         permissionInfo.UserID,
				Username:       permissionInfo.Username,
				Permission:     expectedPerm,
				Message:        fmt.Sprintf("用户 %s 缺少必要权限 %s", permissionInfo.Username, expectedPerm),
				Severity:       "medium",
				Recommendation: fmt.Sprintf("为用户 %s 添加权限 %s", permissionInfo.Username, expectedPerm),
			}
			status.InvalidPermissions = append(status.InvalidPermissions, violation)
		}
	}

	permissionInfo.IsValid = len(status.InvalidPermissions) == 0
}

// checkTestUserStatus 检查测试用户状态
func (uc *UserChecker) checkTestUserStatus() (*TestUserStatus, error) {
	status := &TestUserStatus{
		TestUsers:       []TestUserInfo{},
		TestPermissions: []TestPermissionInfo{},
		Violations:      []TestUserViolation{},
		Recommendations: []string{},
	}

	// 查询测试用户
	query := `
		SELECT u.id, u.username, u.email, u.created_at,
		       GROUP_CONCAT(p.name) as permissions
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		LEFT JOIN permissions p ON rp.permission_id = p.id
		WHERE r.name IN ('test_user', 'beta_tester', 'demo_user')
		GROUP BY u.id, u.username, u.email, u.created_at
		ORDER BY u.created_at DESC
	`

	rows, err := uc.db.Query(query)
	if err != nil {
		return status, nil
	}
	defer rows.Close()

	for rows.Next() {
		var testUser TestUserInfo
		var permissions sql.NullString

		err := rows.Scan(&testUser.UserID, &testUser.Username, &testUser.Email,
			&testUser.InviteDate, &permissions)
		if err != nil {
			continue
		}

		if permissions.Valid {
			// 解析权限字符串
		}

		// 计算过期时间（假设测试用户有30天试用期）
		testUser.ExpiryDate = testUser.InviteDate.AddDate(0, 0, 30)
		daysLeft := int(time.Until(testUser.ExpiryDate).Hours() / 24)
		testUser.DaysLeft = daysLeft
		testUser.IsExpired = daysLeft <= 0

		if testUser.IsExpired {
			testUser.Status = "expired"
		} else {
			testUser.Status = "active"
		}

		// 获取使用统计
		uc.getTestUserUsageStats(&testUser)

		status.TestUsers = append(status.TestUsers, testUser)
	}

	// 检查测试用户违规
	uc.checkTestUserViolations(status)

	return status, nil
}

// getTestUserUsageStats 获取测试用户使用统计
func (uc *UserChecker) getTestUserUsageStats(testUser *TestUserInfo) {
	// 查询登录次数
	var loginCount int
	query := "SELECT COUNT(*) FROM user_sessions WHERE user_id = ?"
	err := uc.db.QueryRow(query, testUser.UserID).Scan(&loginCount)
	if err != nil {
		loginCount = 0
	}
	testUser.UsageStats.LoginCount = loginCount

	// 查询最后登录时间
	var lastLogin time.Time
	query = "SELECT MAX(created_at) FROM user_sessions WHERE user_id = ?"
	err = uc.db.QueryRow(query, testUser.UserID).Scan(&lastLogin)
	if err == nil {
		testUser.UsageStats.LastLogin = lastLogin
	}
}

// checkTestUserViolations 检查测试用户违规
func (uc *UserChecker) checkTestUserViolations(status *TestUserStatus) {
	for _, testUser := range status.TestUsers {
		// 检查过期用户是否仍在活跃使用
		if testUser.IsExpired && testUser.UsageStats.LastLogin.After(testUser.ExpiryDate) {
			violation := TestUserViolation{
				Type:           "expired",
				UserID:         testUser.UserID,
				Username:       testUser.Username,
				Message:        fmt.Sprintf("测试用户 %s 已过期但仍在使用系统", testUser.Username),
				Severity:       "high",
				Recommendation: fmt.Sprintf("立即限制测试用户 %s 的访问权限", testUser.Username),
			}
			status.Violations = append(status.Violations, violation)
		}

		// 检查使用量是否过大
		if testUser.UsageStats.LoginCount > 100 { // 假设测试用户登录次数不应超过100次
			violation := TestUserViolation{
				Type:           "excessive_usage",
				UserID:         testUser.UserID,
				Username:       testUser.Username,
				Message:        fmt.Sprintf("测试用户 %s 使用量过大，登录 %d 次", testUser.Username, testUser.UsageStats.LoginCount),
				Severity:       "medium",
				Recommendation: fmt.Sprintf("考虑将测试用户 %s 转换为正式用户", testUser.Username),
			}
			status.Violations = append(status.Violations, violation)
		}
	}
}

// checkUserViolations 检查用户违规
func (uc *UserChecker) checkUserViolations(status *UserStatus) {
	// 合并所有违规
	for _, violation := range status.SubscriptionStatus.Violations {
		status.Violations = append(status.Violations, UserViolation{
			Type:           "subscription",
			UserID:         violation.UserID,
			Username:       violation.Username,
			Message:        violation.Message,
			Severity:       violation.Severity,
			Recommendation: violation.Recommendation,
		})
	}

	for _, violation := range status.PermissionStatus.InvalidPermissions {
		status.Violations = append(status.Violations, UserViolation{
			Type:           "permission",
			UserID:         violation.UserID,
			Username:       violation.Username,
			Message:        violation.Message,
			Severity:       violation.Severity,
			Recommendation: violation.Recommendation,
		})
	}

	for _, violation := range status.TestUserStatus.Violations {
		status.Violations = append(status.Violations, UserViolation{
			Type:           "usage",
			UserID:         violation.UserID,
			Username:       violation.Username,
			Message:        violation.Message,
			Severity:       violation.Severity,
			Recommendation: violation.Recommendation,
		})
	}
}

// generateRecommendations 生成建议
func (uc *UserChecker) generateRecommendations(status *UserStatus) {
	if len(status.Violations) == 0 {
		status.Recommendations = append(status.Recommendations, "用户权限和订阅状态良好，所有用户都有适当的访问权限")
		return
	}

	status.Recommendations = append(status.Recommendations, "用户管理需要优化，建议采取以下措施：")

	for _, violation := range status.Violations {
		status.Recommendations = append(status.Recommendations, fmt.Sprintf("- %s", violation.Recommendation))
	}

	// 添加通用建议
	status.Recommendations = append(status.Recommendations, "")
	status.Recommendations = append(status.Recommendations, "通用建议：")
	status.Recommendations = append(status.Recommendations, "- 定期审查用户权限和订阅状态")
	status.Recommendations = append(status.Recommendations, "- 建立用户权限审计机制")
	status.Recommendations = append(status.Recommendations, "- 监控异常用户行为")
	status.Recommendations = append(status.Recommendations, "- 及时处理过期订阅和测试用户")
}

// determineOverallStatus 确定整体状态
func (uc *UserChecker) determineOverallStatus(status *UserStatus) {
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
func (uc *UserChecker) Close() error {
	if uc.db != nil {
		return uc.db.Close()
	}
	return nil
}
