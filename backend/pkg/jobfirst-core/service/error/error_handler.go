package error

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Message   string `json:"message,omitempty"`
	Code      int    `json:"code,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// SuccessResponse 统一成功响应结构
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

// ErrorHandler 错误处理器
type ErrorHandler struct{}

// NewErrorHandler 创建错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleError 处理错误并返回JSON响应
func (eh *ErrorHandler) HandleError(c *gin.Context, err error, statusCode int) {
	response := ErrorResponse{
		Success: false,
		Error:   err.Error(),
		Code:    statusCode,
	}

	c.JSON(statusCode, response)
}

// HandleValidationError 处理验证错误
func (eh *ErrorHandler) HandleValidationError(c *gin.Context, err error) {
	eh.HandleError(c, err, http.StatusBadRequest)
}

// HandleUnauthorizedError 处理未授权错误
func (eh *ErrorHandler) HandleUnauthorizedError(c *gin.Context, message string) {
	response := ErrorResponse{
		Success: false,
		Error:   "Unauthorized",
		Message: message,
		Code:    http.StatusUnauthorized,
	}
	c.JSON(http.StatusUnauthorized, response)
}

// HandleForbiddenError 处理禁止访问错误
func (eh *ErrorHandler) HandleForbiddenError(c *gin.Context, message string) {
	response := ErrorResponse{
		Success: false,
		Error:   "Forbidden",
		Message: message,
		Code:    http.StatusForbidden,
	}
	c.JSON(http.StatusForbidden, response)
}

// HandleNotFoundError 处理未找到错误
func (eh *ErrorHandler) HandleNotFoundError(c *gin.Context, message string) {
	response := ErrorResponse{
		Success: false,
		Error:   "Not Found",
		Message: message,
		Code:    http.StatusNotFound,
	}
	c.JSON(http.StatusNotFound, response)
}

// HandleInternalServerError 处理内部服务器错误
func (eh *ErrorHandler) HandleInternalServerError(c *gin.Context, err error) {
	eh.HandleError(c, err, http.StatusInternalServerError)
}

// HandleSuccess 处理成功响应
func (eh *ErrorHandler) HandleSuccess(c *gin.Context, data interface{}, message string) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	c.JSON(http.StatusOK, response)
}

// HandleCreated 处理创建成功响应
func (eh *ErrorHandler) HandleCreated(c *gin.Context, data interface{}, message string) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	c.JSON(http.StatusCreated, response)
}

// HandlePaginatedResponse 处理分页响应
func (eh *ErrorHandler) HandlePaginatedResponse(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	response := gin.H{
		"success": true,
		"data": gin.H{
			"items":     data,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	}
	c.JSON(http.StatusOK, response)
}

// Middleware 错误处理中间件
func (eh *ErrorHandler) Middleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			eh.HandleInternalServerError(c, fmt.Errorf("panic: %s", err))
		} else {
			eh.HandleInternalServerError(c, fmt.Errorf("panic: %v", recovered))
		}
		c.Abort()
	})
}
