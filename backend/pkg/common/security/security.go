package security

import (
	"net/http"
	"strings"

	"resume-centre/common/core"
	"resume-centre/common/utils"

	"github.com/gin-gonic/gin"
)

// SecurityFilter 安全过滤器
type SecurityFilter struct {
	whitelist []string
}

// NewSecurityFilter 创建安全过滤器
func NewSecurityFilter(whitelist []string) *SecurityFilter {
	return &SecurityFilter{
		whitelist: whitelist,
	}
}

// Filter 过滤器中间件
func (s *SecurityFilter) Filter() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查是否为白名单路径
		if s.isWhitelistPath(path, method) {
			c.Next()
			return
		}

		// 获取令牌
		token := s.extractToken(c)
		if token == "" {
			s.unauthorized(c, "Token is required")
			return
		}

		// 验证令牌
		if !utils.ValidateToken(token) {
			s.unauthorized(c, "Invalid token")
			return
		}

		// 设置用户信息到上下文
		s.setUserContext(c, token)

		c.Next()
	}
}

// isWhitelistPath 检查是否为白名单路径
func (s *SecurityFilter) isWhitelistPath(path, method string) bool {
	for _, whitePath := range s.whitelist {
		// 支持通配符匹配
		if strings.HasPrefix(path, whitePath) || whitePath == "*" {
			return true
		}

		// 支持方法+路径匹配 (如: GET:/health)
		if strings.Contains(whitePath, ":") {
			parts := strings.Split(whitePath, ":")
			if len(parts) == 2 && parts[0] == method && strings.HasPrefix(path, parts[1]) {
				return true
			}
		}
	}
	return false
}

// extractToken 提取令牌
func (s *SecurityFilter) extractToken(c *gin.Context) string {
	// 从Authorization头获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return authHeader[7:]
		}
		return authHeader
	}

	// 从accessToken头获取
	token := c.GetHeader("accessToken")
	if token != "" {
		return token
	}

	// 从查询参数获取
	token = c.Query("token")
	if token != "" {
		return token
	}

	// 从Cookie获取
	token, _ = c.Cookie("access_token")
	return token
}

// setUserContext 设置用户信息到上下文
func (s *SecurityFilter) setUserContext(c *gin.Context, token string) {
	userID := utils.ExtractUserID(token)
	c.Set("userID", userID)
	c.Set("accessToken", token)
}

// unauthorized 未授权响应
func (s *SecurityFilter) unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, core.NewErrorResponse(core.CodeUnauthorized, message))
	c.Abort()
}

// GenerateToken 生成令牌
func (s *SecurityFilter) GenerateToken(prefix string) string {
	return utils.GenerateToken(prefix)
}

// GetUserFromContext 从上下文获取用户信息
func (s *SecurityFilter) GetUserFromContext(c *gin.Context) *UserContext {
	userID, _ := c.Get("userID")

	userIDStr := ""
	switch v := userID.(type) {
	case string:
		userIDStr = v
	case int64:
		userIDStr = string(v)
	case int:
		userIDStr = string(v)
	}

	return &UserContext{
		UserID:      userIDStr,
		Username:    "user_" + userIDStr,
		Role:        core.RoleUser,
		Permissions: []string{core.PermissionRead},
	}
}

// HasPermission 检查用户是否有指定权限
func (s *SecurityFilter) HasPermission(c *gin.Context, requiredPermission string) bool {
	userCtx := s.GetUserFromContext(c)

	// 超级管理员拥有所有权限
	if userCtx.Role == core.RoleSuper {
		return true
	}

	// 检查用户权限
	for _, permission := range userCtx.Permissions {
		if permission == requiredPermission || permission == core.PermissionAdmin {
			return true
		}
	}

	return false
}

// RequirePermission 要求指定权限的中间件
func (s *SecurityFilter) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.HasPermission(c, permission) {
			c.JSON(http.StatusForbidden, core.NewErrorResponse(core.CodeForbidden, "Insufficient permissions"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole 要求指定角色的中间件
func (s *SecurityFilter) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx := s.GetUserFromContext(c)

		// 角色权限检查
		roleHierarchy := map[string]int{
			core.RoleGuest:     1,
			core.RoleUser:      2,
			core.RoleVip:       3,
			core.RoleModerator: 4,
			core.RoleAdmin:     5,
			core.RoleSuper:     6,
		}

		userLevel := roleHierarchy[userCtx.Role]
		requiredLevel := roleHierarchy[role]

		if userLevel < requiredLevel {
			c.JSON(http.StatusForbidden, core.NewErrorResponse(core.CodeForbidden, "Insufficient role level"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// DefaultWhitelist 获取默认白名单
func DefaultWhitelist() []string {
	return []string{
		"/health",
		"/version",
		"/v2/api-docs",
		"/swagger/",
		"/metrics",
		"/utils/",
		"/monitor/",
		"/config/",
		"POST:/api/v1/user/auth/login",
		"POST:/api/v1/user/auth/register",
		"GET:/api/v1/user/public/",
	}
}
