package core

// 用户权限类型
const (
	// 基础权限
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionDelete = "delete"
	PermissionAdmin  = "admin"

	// 业务权限
	PermissionUserManage   = "user:manage"
	PermissionResumeManage = "resume:manage"
	PermissionJobManage    = "job:manage"
	PermissionPointsManage = "points:manage"
	PermissionStatsView    = "stats:view"
	PermissionConfigManage = "config:manage"
)

// 用户角色类型
const (
	RoleGuest     = "guest"     // 访客
	RoleUser      = "user"      // 普通用户
	RoleVip       = "vip"       // VIP用户
	RoleModerator = "moderator" // 版主
	RoleAdmin     = "admin"     // 管理员
	RoleSuper     = "super"     // 超级管理员
)

// 用户状态
const (
	UserStatusInactive  = 0 // 未激活
	UserStatusActive    = 1 // 已激活
	UserStatusSuspended = 2 // 已暂停
	UserStatusDeleted   = 3 // 已删除
)

// 业务状态码
const (
	// 成功状态码
	CodeSuccess = 0
	CodeOK      = 200

	// 客户端错误状态码
	CodeBadRequest       = 400
	CodeUnauthorized     = 401
	CodeForbidden        = 403
	CodeNotFound         = 404
	CodeMethodNotAllowed = 405
	CodeConflict         = 409
	CodeTooManyRequests  = 429

	// 服务器错误状态码
	CodeInternalError      = 500
	CodeNotImplemented     = 501
	CodeServiceUnavailable = 503
)

// 业务类型
const (
	// 用户相关
	BusinessTypeUserRegister = "user_register"
	BusinessTypeUserLogin    = "user_login"
	BusinessTypeUserLogout   = "user_logout"
	BusinessTypeUserUpdate   = "user_update"

	// 简历相关
	BusinessTypeResumeCreate = "resume_create"
	BusinessTypeResumeUpdate = "resume_update"
	BusinessTypeResumeDelete = "resume_delete"
	BusinessTypeResumeView   = "resume_view"

	// 职位相关
	BusinessTypeJobCreate = "job_create"
	BusinessTypeJobUpdate = "job_update"
	BusinessTypeJobDelete = "job_delete"
	BusinessTypeJobApply  = "job_apply"

	// 积分相关
	BusinessTypePointsEarn   = "points_earn"
	BusinessTypePointsSpend  = "points_spend"
	BusinessTypePointsRefund = "points_refund"
)

// 消息类型
const (
	MessageTypeSuccess = "success"
	MessageTypeError   = "error"
	MessageTypeWarning = "warning"
	MessageTypeInfo    = "info"
)

// 时间格式
const (
	TimeFormatDefault = "2006-01-02 15:04:05"
	TimeFormatDate    = "2006-01-02"
	TimeFormatTime    = "15:04:05"
	TimeFormatISO     = "2006-01-02T15:04:05Z07:00"
)

// 分页默认值
const (
	DefaultPageSize = 10
	MaxPageSize     = 100
	DefaultPage     = 1
)

// 缓存键前缀
const (
	CachePrefixUser    = "user:"
	CachePrefixResume  = "resume:"
	CachePrefixJob     = "job:"
	CachePrefixPoints  = "points:"
	CachePrefixSession = "session:"
	CachePrefixToken   = "token:"
)

// 文件相关
const (
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
	AllowedImageExt = ".jpg,.jpeg,.png,.gif,.webp"
	AllowedDocExt   = ".pdf,.doc,.docx,.txt"
)

// 验证规则
const (
	MinPasswordLength = 6
	MaxPasswordLength = 20
	MinUsernameLength = 3
	MaxUsernameLength = 20
	MaxEmailLength    = 100
	MaxPhoneLength    = 20
)

// 系统配置键
const (
	ConfigKeySystemName    = "system.name"
	ConfigKeySystemVersion = "system.version"
	ConfigKeySystemEnv     = "system.env"
	ConfigKeyJWTSecret     = "jwt.secret"
	ConfigKeyJWTExpiration = "jwt.expiration"
)
