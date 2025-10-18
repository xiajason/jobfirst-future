package auth

import (
	"database/sql"
	"log"
	"time"

	"superadmin/errors"
)

// Manager 认证管理器
type Manager struct {
	config *AuthConfig
	db     *sql.DB
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret    string        `json:"jwt_secret"`
	TokenExpiry  time.Duration `json:"token_expiry"`
	DatabaseURL  string        `json:"database_url"`
	CacheEnabled bool          `json:"cache_enabled"`
	CacheTimeout time.Duration `json:"cache_timeout"`
}

// UserInfo 用户信息
type UserInfo struct {
	UserID             int       `json:"user_id"`
	Username           string    `json:"username"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	SubscriptionStatus string    `json:"subscription_status"`
	SubscriptionType   string    `json:"subscription_type"`
	ExpiresAt          time.Time `json:"expires_at"`
	IsActive           bool      `json:"is_active"`
	LastLogin          time.Time `json:"last_login"`
	CreatedAt          time.Time `json:"created_at"`
}

// PermissionInfo 权限信息
type PermissionInfo struct {
	PermissionID   int    `json:"permission_id"`
	PermissionName string `json:"permission_name"`
	Resource       string `json:"resource"`
	Action         string `json:"action"`
	IsAllowed      bool   `json:"is_allowed"`
}

// QuotaInfo 配额信息
type QuotaInfo struct {
	ResourceType   string    `json:"resource_type"`
	TotalQuota     int       `json:"total_quota"`
	UsedQuota      int       `json:"used_quota"`
	RemainingQuota int       `json:"remaining_quota"`
	ResetTime      time.Time `json:"reset_time"`
	IsUnlimited    bool      `json:"is_unlimited"`
}

// AuthResult 认证结果
type AuthResult struct {
	Success     bool             `json:"success"`
	User        *UserInfo        `json:"user,omitempty"`
	Permissions []PermissionInfo `json:"permissions,omitempty"`
	Quotas      []QuotaInfo      `json:"quotas,omitempty"`
	Error       string           `json:"error,omitempty"`
	ErrorCode   string           `json:"error_code,omitempty"`
}

// NewManager 创建认证管理器
func NewManager(config *AuthConfig, db *sql.DB) *Manager {
	return &Manager{
		config: config,
		db:     db,
	}
}

// ValidateJWT 验证JWT token
func (m *Manager) ValidateJWT(token string) (*AuthResult, error) {
	// TODO: 实现JWT token验证
	// 这里需要集成JWT库进行token解析和验证

	// 临时实现：从token中提取用户信息
	// 实际实现需要：
	// 1. 验证JWT签名
	// 2. 检查token过期时间
	// 3. 从payload中提取用户ID
	// 4. 查询数据库获取用户信息

	return &AuthResult{
		Success: true,
		User: &UserInfo{
			UserID:   1,
			Username: "test_user",
			Email:    "test@example.com",
			Role:     "user",
		},
	}, nil
}

// CheckPermission 检查用户权限
func (m *Manager) CheckPermission(userID int, permission string) (bool, error) {
	query := `
		SELECT COUNT(*) as count
		FROM user_permissions up
		JOIN permissions p ON up.permission_id = p.id
		WHERE up.user_id = ? AND p.name = ? AND up.is_active = 1
	`

	var count int
	err := m.db.QueryRow(query, userID, permission).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.WrapError(errors.ErrCodeDatabase, "权限检查失败", err)
	}

	return count > 0, nil
}

// GetUserPermissions 获取用户所有权限
func (m *Manager) GetUserPermissions(userID int) ([]PermissionInfo, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, up.is_active
		FROM user_permissions up
		JOIN permissions p ON up.permission_id = p.id
		WHERE up.user_id = ? AND up.is_active = 1
	`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeDatabase, "获取用户权限失败", err)
	}
	defer rows.Close()

	var permissions []PermissionInfo
	for rows.Next() {
		var perm PermissionInfo
		err := rows.Scan(&perm.PermissionID, &perm.PermissionName, &perm.Resource, &perm.Action, &perm.IsAllowed)
		if err != nil {
			return nil, errors.WrapError(errors.ErrCodeDatabase, "扫描权限数据失败", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CheckQuota 检查用户配额
func (m *Manager) CheckQuota(userID int, resourceType string) (*QuotaInfo, error) {
	query := `
		SELECT total_quota, used_quota, reset_time, is_unlimited
		FROM user_quotas
		WHERE user_id = ? AND resource_type = ? AND is_active = 1
	`

	var quota QuotaInfo
	var resetTimeStr string

	err := m.db.QueryRow(query, userID, resourceType).Scan(
		&quota.TotalQuota, &quota.UsedQuota, &resetTimeStr, &quota.IsUnlimited,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// 返回默认配额
			return &QuotaInfo{
				ResourceType:   resourceType,
				TotalQuota:     100, // 默认配额
				UsedQuota:      0,
				RemainingQuota: 100,
				ResetTime:      time.Now().AddDate(0, 0, 1), // 1天后重置
				IsUnlimited:    false,
			}, nil
		}
		return nil, errors.WrapError(errors.ErrCodeDatabase, "配额检查失败", err)
	}

	// 解析重置时间
	resetTime, err := time.Parse("2006-01-02 15:04:05", resetTimeStr)
	if err != nil {
		resetTime = time.Now().AddDate(0, 0, 1)
	}

	quota.ResourceType = resourceType
	quota.RemainingQuota = quota.TotalQuota - quota.UsedQuota
	quota.ResetTime = resetTime

	return &quota, nil
}

// ConsumeQuota 消耗用户配额
func (m *Manager) ConsumeQuota(userID int, resourceType string, amount int) error {
	query := `
		UPDATE user_quotas 
		SET used_quota = used_quota + ?, updated_at = NOW()
		WHERE user_id = ? AND resource_type = ? AND is_active = 1
	`

	result, err := m.db.Exec(query, amount, userID, resourceType)
	if err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "配额消耗失败", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "获取影响行数失败", err)
	}

	if rowsAffected == 0 {
		// 创建新的配额记录
		insertQuery := `
			INSERT INTO user_quotas (user_id, resource_type, total_quota, used_quota, reset_time, is_unlimited, is_active, created_at, updated_at)
			VALUES (?, ?, 100, ?, NOW() + INTERVAL 1 DAY, 0, 1, NOW(), NOW())
		`
		_, err = m.db.Exec(insertQuery, userID, resourceType, amount)
		if err != nil {
			return errors.WrapError(errors.ErrCodeDatabase, "创建配额记录失败", err)
		}
	}

	return nil
}

// GetUserInfo 获取用户信息
func (m *Manager) GetUserInfo(userID int) (*UserInfo, error) {
	query := `
		SELECT id, username, email, role, subscription_status, subscription_type, 
		       subscription_expires_at, status, last_login, created_at
		FROM users
		WHERE id = ? AND status = 'active'
	`

	var user UserInfo
	var expiresAtStr sql.NullString
	var lastLoginStr sql.NullString
	var statusStr string
	var createdAtStr sql.NullString

	err := m.db.QueryRow(query, userID).Scan(
		&user.UserID, &user.Username, &user.Email, &user.Role,
		&user.SubscriptionStatus, &user.SubscriptionType, &expiresAtStr,
		&statusStr, &lastLoginStr, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewError(errors.ErrCodeNotFound, "用户不存在")
		}
		return nil, errors.WrapError(errors.ErrCodeDatabase, "获取用户信息失败", err)
	}

	// 设置用户状态
	user.IsActive = (statusStr == "active")

	// 解析创建时间
	if createdAtStr.Valid {
		user.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr.String)
		if err != nil {
			log.Printf("解析创建时间失败: %v", err)
		}
	}

	// 解析过期时间
	if expiresAtStr.Valid {
		user.ExpiresAt, err = time.Parse("2006-01-02 15:04:05", expiresAtStr.String)
		if err != nil {
			log.Printf("解析过期时间失败: %v", err)
		}
	}

	// 解析最后登录时间
	if lastLoginStr.Valid {
		user.LastLogin, err = time.Parse("2006-01-02 15:04:05", lastLoginStr.String)
		if err != nil {
			log.Printf("解析最后登录时间失败: %v", err)
		}
	}

	return &user, nil
}

// ValidateUserAccess 验证用户访问权限（用于AI服务）
func (m *Manager) ValidateUserAccess(userID int, resource string) (*AuthResult, error) {
	// 获取用户信息
	user, err := m.GetUserInfo(userID)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "用户不存在或已被禁用",
			ErrorCode: "USER_NOT_FOUND",
		}, nil
	}

	// 检查用户是否活跃
	if !user.IsActive {
		return &AuthResult{
			Success:   false,
			Error:     "用户账户已被禁用",
			ErrorCode: "USER_DISABLED",
		}, nil
	}

	// 检查订阅状态
	if user.SubscriptionStatus == "expired" {
		return &AuthResult{
			Success:   false,
			Error:     "用户订阅已过期",
			ErrorCode: "SUBSCRIPTION_EXPIRED",
		}, nil
	}

	// 获取用户权限
	permissions, err := m.GetUserPermissions(userID)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "获取用户权限失败",
			ErrorCode: "PERMISSION_ERROR",
		}, nil
	}

	// 获取用户配额
	quotas := make([]QuotaInfo, 0)
	if resource == "ai_service" {
		aiQuota, err := m.CheckQuota(userID, "ai_requests")
		if err != nil {
			log.Printf("获取AI配额失败: %v", err)
		} else {
			quotas = append(quotas, *aiQuota)
		}
	}

	return &AuthResult{
		Success:     true,
		User:        user,
		Permissions: permissions,
		Quotas:      quotas,
	}, nil
}

// LogAccess 记录访问日志
func (m *Manager) LogAccess(userID int, action string, resource string, result string, ipAddress string, userAgent string) error {
	query := `
		INSERT INTO access_logs (user_id, action, resource, result, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`

	_, err := m.db.Exec(query, userID, action, resource, result, ipAddress, userAgent)
	if err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "记录访问日志失败", err)
	}

	return nil
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
