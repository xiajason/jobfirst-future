package core

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BaseController 基础控制器
type BaseController struct {
	ServiceName string
}

// NewBaseController 创建基础控制器
func NewBaseController(serviceName string) *BaseController {
	return &BaseController{
		ServiceName: serviceName,
	}
}

// Success 成功响应
func (bc *BaseController) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, NewSuccessResponse(data))
}

// Error 错误响应
func (bc *BaseController) Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, NewErrorResponse(code, message))
}

// BadRequest 请求参数错误
func (bc *BaseController) BadRequest(c *gin.Context, message string) {
	bc.Error(c, CodeBadRequest, message)
}

// Unauthorized 未授权
func (bc *BaseController) Unauthorized(c *gin.Context, message string) {
	bc.Error(c, CodeUnauthorized, message)
}

// Forbidden 禁止访问
func (bc *BaseController) Forbidden(c *gin.Context, message string) {
	bc.Error(c, CodeForbidden, message)
}

// NotFound 资源不存在
func (bc *BaseController) NotFound(c *gin.Context, message string) {
	bc.Error(c, CodeNotFound, message)
}

// InternalError 内部错误
func (bc *BaseController) InternalError(c *gin.Context, message string) {
	bc.Error(c, CodeInternalError, message)
}

// GetUserID 获取用户ID
func (bc *BaseController) GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}

	switch v := userID.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			return id
		}
	}
	return 0
}

// GetPageParams 获取分页参数
func (bc *BaseController) GetPageParams(c *gin.Context) (page, pageSize int) {
	pageStr := c.DefaultQuery("page", strconv.Itoa(DefaultPage))
	pageSizeStr := c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize))

	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		pageSize = DefaultPageSize
	}

	return page, pageSize
}

// GetClientIP 获取客户端IP
func (bc *BaseController) GetClientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}

// GetUserAgent 获取用户代理
func (bc *BaseController) GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// ValidatePermission 验证权限
func (bc *BaseController) ValidatePermission(c *gin.Context, requiredPermission string) bool {
	// 这里可以实现具体的权限验证逻辑
	// 目前返回true，表示有权限
	return true
}

// LogAudit 记录审计日志
func (bc *BaseController) LogAudit(c *gin.Context, action, resource, resourceID, message string) {
	// 这里可以实现审计日志记录逻辑
	// 目前只是占位符
	_ = action
	_ = resource
	_ = resourceID
	_ = message
}
