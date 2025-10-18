package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode 错误码类型
type ErrorCode int

// 错误码定义
const (
	// 成功
	ErrCodeSuccess ErrorCode = 0

	// 数据库错误 (1000-1999)
	ErrCodeDatabase    ErrorCode = 1000
	ErrCodeMySQL       ErrorCode = 1001
	ErrCodeRedis       ErrorCode = 1002
	ErrCodePostgreSQL  ErrorCode = 1003
	ErrCodeNeo4j       ErrorCode = 1004
	ErrCodeConnection  ErrorCode = 1005
	ErrCodeTransaction ErrorCode = 1006
	ErrCodeMigration   ErrorCode = 1007

	// 认证错误 (2000-2999)
	ErrCodeAuth             ErrorCode = 2000
	ErrCodeUnauthorized     ErrorCode = 2001
	ErrCodeForbidden        ErrorCode = 2002
	ErrCodeTokenExpired     ErrorCode = 2003
	ErrCodeInvalidToken     ErrorCode = 2004
	ErrCodeLoginFailed      ErrorCode = 2005
	ErrCodePasswordMismatch ErrorCode = 2006
	ErrCodeAccountLocked    ErrorCode = 2007

	// 验证错误 (3000-3999)
	ErrCodeValidation    ErrorCode = 3000
	ErrCodeInvalidInput  ErrorCode = 3001
	ErrCodeMissingField  ErrorCode = 3002
	ErrCodeInvalidFormat ErrorCode = 3003
	ErrCodeOutOfRange    ErrorCode = 3004
	ErrCodeDuplicate     ErrorCode = 3005

	// 服务错误 (4000-4999)
	ErrCodeService       ErrorCode = 4000
	ErrCodeInternal      ErrorCode = 4001
	ErrCodeNotFound      ErrorCode = 4002
	ErrCodeAlreadyExists ErrorCode = 4003
	ErrCodeTimeout       ErrorCode = 4004
	ErrCodeRateLimit     ErrorCode = 4005
	ErrCodeMaintenance   ErrorCode = 4006

	// 网络错误 (5000-5999)
	ErrCodeNetwork           ErrorCode = 5000
	ErrCodeConnectionTimeout ErrorCode = 5001
	ErrCodeRequestTimeout    ErrorCode = 5002
	ErrCodeDNS               ErrorCode = 5003
	ErrCodeProxy             ErrorCode = 5004

	// 配置错误 (6000-6999)
	ErrCodeConfig        ErrorCode = 6000
	ErrCodeMissingConfig ErrorCode = 6001
	ErrCodeInvalidConfig ErrorCode = 6002
	ErrCodeConfigLoad    ErrorCode = 6003

	// 文件错误 (7000-7999)
	ErrCodeFile           ErrorCode = 7000
	ErrCodeFileNotFound   ErrorCode = 7001
	ErrCodeFilePermission ErrorCode = 7002
	ErrCodeFileSize       ErrorCode = 7003
	ErrCodeFileFormat     ErrorCode = 7004

	// 业务错误 (8000-8999)
	ErrCodeBusiness        ErrorCode = 8000
	ErrCodeUserNotFound    ErrorCode = 8001
	ErrCodeResumeNotFound  ErrorCode = 8002
	ErrCodeJobNotFound     ErrorCode = 8003
	ErrCodeCompanyNotFound ErrorCode = 8004
	ErrCodeAIError         ErrorCode = 8005
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	Path      string    `json:"path,omitempty"`
	Method    string    `json:"method,omitempty"`
}

// JobFirstError 自定义错误类型
type JobFirstError struct {
	Code    ErrorCode
	Message string
	Details string
	Cause   error
}

// Error 实现error接口
func (e *JobFirstError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *JobFirstError) Unwrap() error {
	return e.Cause
}

// NewError 创建新错误
func NewError(code ErrorCode, message string) *JobFirstError {
	return &JobFirstError{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithDetails 创建带详情的错误
func NewErrorWithDetails(code ErrorCode, message, details string) *JobFirstError {
	return &JobFirstError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WrapError 包装现有错误
func WrapError(code ErrorCode, message string, cause error) *JobFirstError {
	return &JobFirstError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// GetHTTPStatus 根据错误码获取HTTP状态码
func GetHTTPStatus(code ErrorCode) int {
	switch {
	case code == ErrCodeSuccess:
		return http.StatusOK
	case code >= ErrCodeDatabase && code < ErrCodeAuth:
		return http.StatusInternalServerError
	case code >= ErrCodeAuth && code < ErrCodeValidation:
		if code == ErrCodeUnauthorized {
			return http.StatusUnauthorized
		}
		if code == ErrCodeForbidden {
			return http.StatusForbidden
		}
		return http.StatusUnauthorized
	case code >= ErrCodeValidation && code < ErrCodeService:
		return http.StatusBadRequest
	case code >= ErrCodeService && code < ErrCodeNetwork:
		if code == ErrCodeNotFound {
			return http.StatusNotFound
		}
		if code == ErrCodeAlreadyExists {
			return http.StatusConflict
		}
		if code == ErrCodeTimeout {
			return http.StatusRequestTimeout
		}
		if code == ErrCodeRateLimit {
			return http.StatusTooManyRequests
		}
		return http.StatusInternalServerError
	case code >= ErrCodeNetwork && code < ErrCodeConfig:
		return http.StatusBadGateway
	case code >= ErrCodeConfig && code < ErrCodeFile:
		return http.StatusInternalServerError
	case code >= ErrCodeFile && code < ErrCodeBusiness:
		return http.StatusBadRequest
	case code >= ErrCodeBusiness:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code ErrorCode) string {
	messages := map[ErrorCode]string{
		ErrCodeSuccess: "操作成功",

		// 数据库错误
		ErrCodeDatabase:    "数据库错误",
		ErrCodeMySQL:       "MySQL数据库错误",
		ErrCodeRedis:       "Redis数据库错误",
		ErrCodePostgreSQL:  "PostgreSQL数据库错误",
		ErrCodeNeo4j:       "Neo4j数据库错误",
		ErrCodeConnection:  "数据库连接错误",
		ErrCodeTransaction: "数据库事务错误",
		ErrCodeMigration:   "数据库迁移错误",

		// 认证错误
		ErrCodeAuth:             "认证错误",
		ErrCodeUnauthorized:     "未授权访问",
		ErrCodeForbidden:        "禁止访问",
		ErrCodeTokenExpired:     "令牌已过期",
		ErrCodeInvalidToken:     "无效令牌",
		ErrCodeLoginFailed:      "登录失败",
		ErrCodePasswordMismatch: "密码不匹配",
		ErrCodeAccountLocked:    "账户已锁定",

		// 验证错误
		ErrCodeValidation:    "验证错误",
		ErrCodeInvalidInput:  "无效输入",
		ErrCodeMissingField:  "缺少必填字段",
		ErrCodeInvalidFormat: "格式错误",
		ErrCodeOutOfRange:    "超出范围",
		ErrCodeDuplicate:     "重复数据",

		// 服务错误
		ErrCodeService:       "服务错误",
		ErrCodeInternal:      "内部服务器错误",
		ErrCodeNotFound:      "资源未找到",
		ErrCodeAlreadyExists: "资源已存在",
		ErrCodeTimeout:       "请求超时",
		ErrCodeRateLimit:     "请求频率限制",
		ErrCodeMaintenance:   "服务维护中",

		// 网络错误
		ErrCodeNetwork:           "网络错误",
		ErrCodeConnectionTimeout: "连接超时",
		ErrCodeRequestTimeout:    "请求超时",
		ErrCodeDNS:               "DNS解析错误",
		ErrCodeProxy:             "代理错误",

		// 配置错误
		ErrCodeConfig:        "配置错误",
		ErrCodeMissingConfig: "缺少配置",
		ErrCodeInvalidConfig: "无效配置",
		ErrCodeConfigLoad:    "配置加载失败",

		// 文件错误
		ErrCodeFile:           "文件错误",
		ErrCodeFileNotFound:   "文件未找到",
		ErrCodeFilePermission: "文件权限错误",
		ErrCodeFileSize:       "文件大小错误",
		ErrCodeFileFormat:     "文件格式错误",

		// 业务错误
		ErrCodeBusiness:        "业务错误",
		ErrCodeUserNotFound:    "用户未找到",
		ErrCodeResumeNotFound:  "简历未找到",
		ErrCodeJobNotFound:     "职位未找到",
		ErrCodeCompanyNotFound: "公司未找到",
		ErrCodeAIError:         "AI服务错误",
	}

	if message, exists := messages[code]; exists {
		return message
	}
	return "未知错误"
}

// IsJobFirstError 检查是否为JobFirst错误
func IsJobFirstError(err error) bool {
	_, ok := err.(*JobFirstError)
	return ok
}

// GetErrorCode 从错误中获取错误码
func GetErrorCode(err error) ErrorCode {
	if jfErr, ok := err.(*JobFirstError); ok {
		return jfErr.Code
	}
	return ErrCodeInternal
}

// GetErrorDetails 从错误中获取详情
func GetErrorDetails(err error) string {
	if jfErr, ok := err.(*JobFirstError); ok {
		return jfErr.Details
	}
	return err.Error()
}
