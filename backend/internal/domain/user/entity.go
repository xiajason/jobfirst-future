package user

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

// 状态枚举
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
)

// 请求和响应结构体
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

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email" binding:"omitempty,email"`
}

type UpdateProfileResponse struct {
	Message string `json:"message"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type VerifyPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type ListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Search   string `json:"search" form:"search"`
	Status   Status `json:"status" form:"status"`
}

type ListResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

// 错误类型
type ErrUsernameExists struct {
	Username string
}

func (e *ErrUsernameExists) Error() string {
	return "username already exists: " + e.Username
}

type ErrEmailExists struct {
	Email string
}

func (e *ErrEmailExists) Error() string {
	return "email already exists: " + e.Email
}

type ErrInvalidCredentials struct{}

func (e *ErrInvalidCredentials) Error() string {
	return "invalid credentials"
}

type ErrUserInactive struct{}

func (e *ErrUserInactive) Error() string {
	return "user account is inactive"
}
