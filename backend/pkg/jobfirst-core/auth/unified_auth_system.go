package auth

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UnifiedAuthSystem 统一认证系统
type UnifiedAuthSystem struct {
	db         *sql.DB
	jwtSecret  string
	roleConfig *RoleConfig
}

// RoleConfig 角色配置
type RoleConfig struct {
	Roles map[string]*RoleInfo `json:"roles"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	Name        string   `json:"name"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description"`
}

// UserInfo 用户信息（统一结构）
type UserInfo struct {
	ID                 int        `json:"id" db:"id"`
	Username           string     `json:"username" db:"username"`
	Email              string     `json:"email" db:"email"`
	PasswordHash       string     `json:"-" db:"password_hash"`
	Role               string     `json:"role" db:"role"`
	Status             string     `json:"status" db:"status"`
	SubscriptionType   *string    `json:"subscription_type" db:"subscription_type"`
	SubscriptionExpiry *time.Time `json:"subscription_expiry" db:"subscription_expiry"`
	LastLogin          *time.Time `json:"last_login" db:"last_login"`
	CreatedAt          *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID      int      `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// AuthResult 认证结果
type AuthResult struct {
	Success     bool      `json:"success"`
	Token       string    `json:"token,omitempty"`
	User        *UserInfo `json:"user,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
	Error       string    `json:"error,omitempty"`
	ErrorCode   string    `json:"error_code,omitempty"`
}

// NewUnifiedAuthSystem 创建统一认证系统
func NewUnifiedAuthSystem(db *sql.DB, jwtSecret string) *UnifiedAuthSystem {
	roleConfig := &RoleConfig{
		Roles: map[string]*RoleInfo{
			"guest": {
				Name:        "guest",
				Level:       1,
				Permissions: []string{"read:public"},
				Description: "访客用户",
			},
			"user": {
				Name:        "user",
				Level:       2,
				Permissions: []string{"read:public", "read:own", "write:own"},
				Description: "普通用户",
			},
			"admin": {
				Name:        "admin",
				Level:       3,
				Permissions: []string{"read:public", "read:own", "write:own", "read:all", "write:all", "delete:own"},
				Description: "管理员",
			},
			"super_admin": {
				Name:        "super_admin",
				Level:       4,
				Permissions: []string{"*"},
				Description: "超级管理员",
			},
		},
	}

	return &UnifiedAuthSystem{
		db:         db,
		jwtSecret:  jwtSecret,
		roleConfig: roleConfig,
	}
}

// InitializeDatabase 初始化数据库表结构
func (uas *UnifiedAuthSystem) InitializeDatabase() error {
	// 创建用户表（如果不存在）
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) DEFAULT 'user',
			status VARCHAR(20) DEFAULT 'active',
			subscription_type VARCHAR(50) DEFAULT NULL,
			subscription_expiry TIMESTAMP NULL DEFAULT NULL,
			last_login TIMESTAMP NULL DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_username (username),
			INDEX idx_email (email),
			INDEX idx_role (role),
			INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建权限表（如果不存在）
	createPermissionsTable := `
		CREATE TABLE IF NOT EXISTS permissions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			resource VARCHAR(100) NOT NULL,
			action VARCHAR(50) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_name (name),
			INDEX idx_resource (resource),
			INDEX idx_action (action)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建角色权限关联表（如果不存在）
	createRolePermissionsTable := `
		CREATE TABLE IF NOT EXISTS role_permissions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			role VARCHAR(20) NOT NULL,
			permission VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_role_permission (role, permission),
			INDEX idx_role (role),
			INDEX idx_permission (permission)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 创建访问日志表（如果不存在）
	createAccessLogsTable := `
		CREATE TABLE IF NOT EXISTS access_logs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT,
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(100) NOT NULL,
			result VARCHAR(20) NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_user_id (user_id),
			INDEX idx_action (action),
			INDEX idx_resource (resource),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 执行表创建
	tables := []string{
		createUsersTable,
		createPermissionsTable,
		createRolePermissionsTable,
		createAccessLogsTable,
	}

	for _, table := range tables {
		if _, err := uas.db.Exec(table); err != nil {
			return fmt.Errorf("创建表失败: %w", err)
		}
	}

	// 初始化权限数据
	if err := uas.initializePermissions(); err != nil {
		return fmt.Errorf("初始化权限数据失败: %w", err)
	}

	// 创建默认超级管理员
	if err := uas.createDefaultSuperAdmin(); err != nil {
		return fmt.Errorf("创建默认超级管理员失败: %w", err)
	}

	return nil
}

// initializePermissions 初始化权限数据
func (uas *UnifiedAuthSystem) initializePermissions() error {
	// 定义所有权限
	permissions := []struct {
		name        string
		resource    string
		action      string
		description string
	}{
		{"read:public", "public", "read", "读取公开内容"},
		{"read:own", "own", "read", "读取自己的内容"},
		{"write:own", "own", "write", "修改自己的内容"},
		{"read:all", "all", "read", "读取所有内容"},
		{"write:all", "all", "write", "修改所有内容"},
		{"delete:own", "own", "delete", "删除自己的内容"},
		{"delete:all", "all", "delete", "删除所有内容"},
		{"admin:users", "users", "admin", "用户管理"},
		{"admin:system", "system", "admin", "系统管理"},
	}

	// 插入权限
	for _, perm := range permissions {
		_, err := uas.db.Exec(`
			INSERT IGNORE INTO permissions (name, resource, action, description)
			VALUES (?, ?, ?, ?)
		`, perm.name, perm.resource, perm.action, perm.description)
		if err != nil {
			return fmt.Errorf("插入权限失败 %s: %w", perm.name, err)
		}
	}

	// 为每个角色分配权限
	for roleName, roleInfo := range uas.roleConfig.Roles {
		for _, permission := range roleInfo.Permissions {
			if permission == "*" {
				// 超级管理员拥有所有权限
				rows, err := uas.db.Query("SELECT name FROM permissions")
				if err != nil {
					continue
				}
				defer rows.Close()

				for rows.Next() {
					var permName string
					if err := rows.Scan(&permName); err != nil {
						continue
					}
					uas.db.Exec(`
						INSERT IGNORE INTO role_permissions (role, permission)
						VALUES (?, ?)
					`, roleName, permName)
				}
			} else {
				uas.db.Exec(`
					INSERT IGNORE INTO role_permissions (role, permission)
					VALUES (?, ?)
				`, roleName, permission)
			}
		}
	}

	return nil
}

// createDefaultSuperAdmin 创建默认超级管理员
func (uas *UnifiedAuthSystem) createDefaultSuperAdmin() error {
	// 检查是否已存在超级管理员
	var count int
	err := uas.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'super_admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 已存在超级管理员
	}

	// 创建默认超级管理员
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = uas.db.Exec(`
		INSERT INTO users (username, email, password_hash, role, status)
		VALUES (?, ?, ?, ?, ?)
	`, "admin", "admin@jobfirst.com", string(hashedPassword), "super_admin", "active")

	if err != nil {
		return err
	}

	log.Printf("默认超级管理员已创建: username=admin, password=%s", password)
	return nil
}

// Authenticate 用户认证
func (uas *UnifiedAuthSystem) Authenticate(username, password string) (*AuthResult, error) {
	// 查询用户信息
	user, err := uas.getUserByUsername(username)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "用户不存在",
			ErrorCode: "USER_NOT_FOUND",
		}, nil
	}

	// 检查用户状态
	if user.Status != "active" {
		return &AuthResult{
			Success:   false,
			Error:     "用户账户已被禁用",
			ErrorCode: "USER_DISABLED",
		}, nil
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "密码错误",
			ErrorCode: "INVALID_PASSWORD",
		}, nil
	}

	// 获取用户权限
	permissions, err := uas.getUserPermissions(user.Role)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "获取用户权限失败",
			ErrorCode: "PERMISSION_ERROR",
		}, nil
	}

	// 生成JWT token
	token, err := uas.generateJWT(user, permissions)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "生成token失败",
			ErrorCode: "TOKEN_ERROR",
		}, nil
	}

	// 更新最后登录时间
	uas.updateLastLogin(user.ID)

	// 记录访问日志
	uas.logAccess(user.ID, "login", "auth", "success", "", "")

	return &AuthResult{
		Success:     true,
		Token:       token,
		User:        user,
		Permissions: permissions,
	}, nil
}

// ValidateJWT 验证JWT token
func (uas *UnifiedAuthSystem) ValidateJWT(tokenString string) (*AuthResult, error) {
	// 解析JWT token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(uas.jwtSecret), nil
	})

	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "无效的token",
			ErrorCode: "INVALID_TOKEN",
		}, nil
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return &AuthResult{
			Success:   false,
			Error:     "token验证失败",
			ErrorCode: "TOKEN_VALIDATION_FAILED",
		}, nil
	}

	// 检查token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return &AuthResult{
			Success:   false,
			Error:     "token已过期",
			ErrorCode: "TOKEN_EXPIRED",
		}, nil
	}

	// 获取用户信息
	user, err := uas.getUserByID(claims.UserID)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "用户不存在",
			ErrorCode: "USER_NOT_FOUND",
		}, nil
	}

	// 检查用户状态
	if user.Status != "active" {
		return &AuthResult{
			Success:   false,
			Error:     "用户账户已被禁用",
			ErrorCode: "USER_DISABLED",
		}, nil
	}

	return &AuthResult{
		Success:     true,
		User:        user,
		Permissions: claims.Permissions,
	}, nil
}

// CheckPermission 检查用户权限
func (uas *UnifiedAuthSystem) CheckPermission(userID int, permission string) (bool, error) {
	// 获取用户信息
	user, err := uas.getUserByID(userID)
	if err != nil {
		return false, err
	}

	// 超级管理员拥有所有权限
	if user.Role == "super_admin" {
		return true, nil
	}

	// 检查角色权限
	permissions, err := uas.getUserPermissions(user.Role)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == permission || perm == "*" {
			return true, nil
		}
	}

	return false, nil
}

// getUserByUsername 根据用户名获取用户信息
func (uas *UnifiedAuthSystem) getUserByUsername(username string) (*UserInfo, error) {
	query := `
		SELECT id, username, email, password_hash, role, status,
		       subscription_type, subscription_expires_at, last_login,
		       created_at, updated_at
		FROM users
		WHERE username = ? AND status = 'active' AND deleted_at IS NULL
	`

	var user UserInfo
	var subscriptionType, lastLogin sql.NullString
	var subscriptionExpiry, createdAt, updatedAt sql.NullTime

	err := uas.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Status, &subscriptionType, &subscriptionExpiry,
		&lastLogin, &createdAt, &updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// 处理NULL值
	if subscriptionType.Valid {
		user.SubscriptionType = &subscriptionType.String
	}
	if subscriptionExpiry.Valid {
		user.SubscriptionExpiry = &subscriptionExpiry.Time
	}
	if lastLogin.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", lastLogin.String); err == nil {
			user.LastLogin = &t
		}
	}
	if createdAt.Valid {
		user.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return &user, nil
}

// getUserByID 根据ID获取用户信息
func (uas *UnifiedAuthSystem) getUserByID(userID int) (*UserInfo, error) {
	query := `
		SELECT id, username, email, password_hash, role, status,
		       subscription_type, subscription_expires_at, last_login,
		       created_at, updated_at
		FROM users
		WHERE id = ? AND status = 'active' AND deleted_at IS NULL
	`

	var user UserInfo
	var subscriptionType, lastLogin sql.NullString
	var subscriptionExpiry, createdAt, updatedAt sql.NullTime

	err := uas.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Status, &subscriptionType, &subscriptionExpiry,
		&lastLogin, &createdAt, &updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// 处理NULL值
	if subscriptionType.Valid {
		user.SubscriptionType = &subscriptionType.String
	}
	if subscriptionExpiry.Valid {
		user.SubscriptionExpiry = &subscriptionExpiry.Time
	}
	if lastLogin.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", lastLogin.String); err == nil {
			user.LastLogin = &t
		}
	}
	if createdAt.Valid {
		user.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return &user, nil
}

// getUserPermissions 获取用户权限
func (uas *UnifiedAuthSystem) getUserPermissions(role string) ([]string, error) {
	// 首先从配置中获取
	if roleInfo, exists := uas.roleConfig.Roles[role]; exists {
		return roleInfo.Permissions, nil
	}

	// 从数据库获取
	query := `
		SELECT p.name
		FROM role_permissions rp
		JOIN permissions p ON rp.permission = p.name
		WHERE rp.role = ?
	`

	rows, err := uas.db.Query(query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			continue
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// generateJWT 生成JWT token
func (uas *UnifiedAuthSystem) generateJWT(user *UserInfo, permissions []string) (string, error) {
	roleInfo, exists := uas.roleConfig.Roles[user.Role]
	if !exists {
		roleInfo = &RoleInfo{Level: 1}
	}

	claims := &JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Level:       roleInfo.Level,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)), // 7天，适配测试需要
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "jobfirst-auth",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uas.jwtSecret))
}

// updateLastLogin 更新最后登录时间
func (uas *UnifiedAuthSystem) updateLastLogin(userID int) {
	uas.db.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", userID)
}

// logAccess 记录访问日志
func (uas *UnifiedAuthSystem) logAccess(userID int, action, resource, result, ipAddress, userAgent string) {
	uas.db.Exec(`
		INSERT INTO access_logs (user_id, action, resource, result, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, action, resource, result, ipAddress, userAgent)
}

// Close 关闭数据库连接
func (uas *UnifiedAuthSystem) Close() error {
	if uas.db != nil {
		return uas.db.Close()
	}
	return nil
}
