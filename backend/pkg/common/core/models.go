package core

import (
	"time"
)

// BaseResponse 基础响应模型
type BaseResponse struct {
	Code    int         `json:"code"`              // 状态码
	Message string      `json:"message,omitempty"` // 消息
	Data    interface{} `json:"data,omitempty"`    // 数据
	Time    string      `json:"time,omitempty"`    // 时间戳
}

// PageResponse 分页响应模型
type PageResponse struct {
	BaseResponse
	Data PageData `json:"data"`
}

// PageData 分页数据
type PageData struct {
	List     interface{} `json:"list"`      // 数据列表
	Total    int64       `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页
	PageSize int         `json:"page_size"` // 每页大小
	Pages    int         `json:"pages"`     // 总页数
}

// LoginRequest 登录请求模型
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
	Captcha  string `json:"captcha,omitempty"`           // 验证码
	Remember bool   `json:"remember,omitempty"`          // 记住我
}

// LoginResponse 登录响应模型
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`  // 访问令牌
	RefreshToken string   `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64    `json:"expires_in"`    // 过期时间（秒）
	TokenType    string   `json:"token_type"`    // 令牌类型
	User         UserInfo `json:"user"`          // 用户信息
}

// UserInfo 用户信息模型
type UserInfo struct {
	ID       int64  `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
	Email    string `json:"email"`    // 邮箱
	Phone    string `json:"phone"`    // 手机号
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像
	Role     string `json:"role"`     // 角色
	Status   int    `json:"status"`   // 状态
	Created  string `json:"created"`  // 创建时间
	Updated  string `json:"updated"`  // 更新时间
}

// RegisterRequest 注册请求模型
type RegisterRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
	Email    string `json:"email" binding:"required"`    // 邮箱
	Phone    string `json:"phone,omitempty"`             // 手机号
	Nickname string `json:"nickname,omitempty"`          // 昵称
	Captcha  string `json:"captcha,omitempty"`           // 验证码
}

// PasswordChangeRequest 修改密码请求模型
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"` // 旧密码
	NewPassword string `json:"new_password" binding:"required"` // 新密码
}

// PasswordResetRequest 重置密码请求模型
type PasswordResetRequest struct {
	Email       string `json:"email" binding:"required"`        // 邮箱
	Code        string `json:"code" binding:"required"`         // 验证码
	NewPassword string `json:"new_password" binding:"required"` // 新密码
}

// CaptchaResponse 验证码响应模型
type CaptchaResponse struct {
	CaptchaID   string `json:"captcha_id"`   // 验证码ID
	CaptchaData string `json:"captcha_data"` // 验证码数据（Base64图片）
}

// FileUploadResponse 文件上传响应模型
type FileUploadResponse struct {
	FileName   string `json:"file_name"`   // 文件名
	FileSize   int64  `json:"file_size"`   // 文件大小
	FileType   string `json:"file_type"`   // 文件类型
	FileURL    string `json:"file_url"`    // 文件URL
	FileHash   string `json:"file_hash"`   // 文件哈希
	UploadTime string `json:"upload_time"` // 上传时间
}

// ErrorResponse 错误响应模型
type ErrorResponse struct {
	Code    int    `json:"code"`              // 错误码
	Message string `json:"message"`           // 错误消息
	Details string `json:"details,omitempty"` // 错误详情
	Time    string `json:"time"`              // 时间戳
}

// HealthResponse 健康检查响应模型
type HealthResponse struct {
	Status   string            `json:"status"`   // 状态
	Time     string            `json:"time"`     // 时间戳
	Services map[string]string `json:"services"` // 服务状态
	Version  string            `json:"version"`  // 版本
	Uptime   string            `json:"uptime"`   // 运行时间
}

// ConfigResponse 配置响应模型
type ConfigResponse struct {
	Key   string `json:"key"`   // 配置键
	Value string `json:"value"` // 配置值
	Type  string `json:"type"`  // 配置类型
}

// AuditLog 审计日志模型
type AuditLog struct {
	ID         int64     `json:"id"`          // 日志ID
	UserID     int64     `json:"user_id"`     // 用户ID
	Username   string    `json:"username"`    // 用户名
	Action     string    `json:"action"`      // 操作
	Resource   string    `json:"resource"`    // 资源
	ResourceID string    `json:"resource_id"` // 资源ID
	IP         string    `json:"ip"`          // IP地址
	UserAgent  string    `json:"user_agent"`  // 用户代理
	Status     int       `json:"status"`      // 状态
	Message    string    `json:"message"`     // 消息
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
}

// Notification 通知模型
type Notification struct {
	ID        int64     `json:"id"`         // 通知ID
	UserID    int64     `json:"user_id"`    // 用户ID
	Type      string    `json:"type"`       // 通知类型
	Title     string    `json:"title"`      // 标题
	Content   string    `json:"content"`    // 内容
	IsRead    bool      `json:"is_read"`    // 是否已读
	CreatedAt time.Time `json:"created_at"` // 创建时间
	ReadAt    time.Time `json:"read_at"`    // 阅读时间
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) BaseResponse {
	return BaseResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
		Time:    time.Now().Format(TimeFormatDefault),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) BaseResponse {
	return BaseResponse{
		Code:    code,
		Message: message,
		Time:    time.Now().Format(TimeFormatDefault),
	}
}

// NewPageResponse 创建分页响应
func NewPageResponse(list interface{}, total int64, page, pageSize int) PageResponse {
	pages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return PageResponse{
		BaseResponse: BaseResponse{
			Code:    CodeSuccess,
			Message: "success",
			Time:    time.Now().Format(TimeFormatDefault),
		},
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Pages:    pages,
		},
	}
}
