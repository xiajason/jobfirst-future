package auth

import "time"

// 用户实体
type User struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	UUID          string     `json:"uuid" gorm:"uniqueIndex;size:36"`
	Username      string     `json:"username" gorm:"uniqueIndex;size:100"`
	Email         string     `json:"email" gorm:"uniqueIndex;size:255"`
	PasswordHash  string     `json:"-" gorm:"size:255"`
	FirstName     string     `json:"first_name" gorm:"size:100"`
	LastName      string     `json:"last_name" gorm:"size:100"`
	Phone         string     `json:"phone" gorm:"size:20"`
	Status        Status     `json:"status" gorm:"type:enum('active','inactive','suspended');default:'active'"`
	EmailVerified bool       `json:"email_verified" gorm:"default:false"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"-" gorm:"index"`
}

// 角色实体
type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        RoleName  `json:"name" gorm:"uniqueIndex;size:50"`
	DisplayName string    `json:"display_name" gorm:"size:100"`
	Description string    `json:"description" gorm:"size:255"`
	Status      Status    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 权限实体
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;size:100"`
	Resource    string    `json:"resource" gorm:"size:100"`
	Action      string    `json:"action" gorm:"size:50"`
	Description string    `json:"description" gorm:"size:255"`
	Status      Status    `json:"status" gorm:"type:enum('active','inactive');default:'active'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 用户角色关联
type UserRole struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"not null"`
	RoleID uint `json:"role_id" gorm:"not null"`
	User   User `json:"user" gorm:"foreignKey:UserID"`
	Role   Role `json:"role" gorm:"foreignKey:RoleID"`
}

// 角色权限关联
type RolePermission struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	RoleID       uint       `json:"role_id" gorm:"not null"`
	PermissionID uint       `json:"permission_id" gorm:"not null"`
	Role         Role       `json:"role" gorm:"foreignKey:RoleID"`
	Permission   Permission `json:"permission" gorm:"foreignKey:PermissionID"`
}

// 状态枚举
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
)

// 角色名称枚举
type RoleName string

const (
	RoleSuperAdmin RoleName = "super_admin"
	RoleAdmin      RoleName = "admin"
	RoleDevTeam    RoleName = "dev_team"
	RoleUser       RoleName = "user"
)

// 请求和响应结构体
type InitializeSuperAdminRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type InitializeSuperAdminResponse struct {
	Message string `json:"message"`
	UserID  uint   `json:"user_id"`
}

type CheckSuperAdminStatusRequest struct{}

type CheckSuperAdminStatusResponse struct {
	Exists bool   `json:"exists"`
	UserID *uint  `json:"user_id,omitempty"`
	Status string `json:"status"`
}

type ResetSuperAdminPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

type RegisterResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

// 错误类型
type ErrSuperAdminNotFound struct{}

func (e *ErrSuperAdminNotFound) Error() string {
	return "super admin not found"
}
