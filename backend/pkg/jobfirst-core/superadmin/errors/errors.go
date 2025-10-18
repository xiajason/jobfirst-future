package errors

import "fmt"

// SuperAdminError represents a super admin specific error
type SuperAdminError struct {
	Code    string
	Message string
	Err     error
}

func (e *SuperAdminError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New creates a new SuperAdminError
func New(code, message string) *SuperAdminError {
	return &SuperAdminError{
		Code:    code,
		Message: message,
	}
}

// NewError creates a new SuperAdminError (alias for New)
func NewError(code, message string) *SuperAdminError {
	return &SuperAdminError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string) *SuperAdminError {
	return &SuperAdminError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WrapError wraps an existing error with additional context (alias for Wrap)
func WrapError(code, message string, err error) *SuperAdminError {
	return &SuperAdminError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error codes
const (
	ErrCodeConfigNotFound     = "CONFIG_NOT_FOUND"
	ErrCodeDatabaseConnection = "DATABASE_CONNECTION"
	ErrCodeServiceNotFound    = "SERVICE_NOT_FOUND"
	ErrCodePermissionDenied   = "PERMISSION_DENIED"
	ErrCodeInvalidOperation   = "INVALID_OPERATION"
	ErrCodeService            = "SERVICE_ERROR"
	ErrCodeFile               = "FILE_ERROR"
	ErrCodeDatabase           = "DATABASE_ERROR"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeFileNotFound       = "FILE_NOT_FOUND"
	ErrCodeAlreadyExists      = "ALREADY_EXISTS"
	ErrCodeNotFound           = "NOT_FOUND"
)
