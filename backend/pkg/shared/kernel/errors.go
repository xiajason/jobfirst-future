package kernel

import "errors"

// 通用错误定义
var (
	ErrInvalidDateRange = errors.New("invalid date range: start date must be before end date")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidPhone     = errors.New("invalid phone number format")
	ErrEntityNotFound   = errors.New("entity not found")
	ErrInvalidInput     = errors.New("invalid input data")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrForbidden        = errors.New("forbidden access")
	ErrInternalError    = errors.New("internal server error")
)
