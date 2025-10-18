package middleware

import (
	"net/http"
	"strings"

	"resume-centre/shared/infrastructure/auth"

	"github.com/gin-gonic/gin"
)

// AuthConfig 认证配置
type AuthConfig struct {
	JWTConfig     *auth.JWTConfig `json:"jwt_config"`
	PublicPaths   []string        `json:"public_paths"`
	AdminPaths    []string        `json:"admin_paths"`
	RequiredRoles []string        `json:"required_roles"`
}

// AuthMiddleware 认证中间件
func AuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	jwtAuth := auth.NewJWTAuth(config.JWTConfig)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查是否为公开路径
		if isPublicPath(path, method, config.PublicPaths) {
			c.Next()
			return
		}

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization required",
				"code":    "AUTH_REQUIRED",
				"message": "请提供有效的认证信息",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"code":    "INVALID_AUTH_FORMAT",
				"message": "认证格式无效，请使用Bearer token",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		claims, err := jwtAuth.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"code":    "INVALID_TOKEN",
				"message": "认证token无效或已过期",
			})
			c.Abort()
			return
		}

		// 检查管理员权限
		if isAdminPath(path, method, config.AdminPaths) {
			if !claims.HasAnyRole("admin", "super_admin") {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Admin access required",
					"code":    "ADMIN_REQUIRED",
					"message": "需要管理员权限",
				})
				c.Abort()
				return
			}
		}

		// 检查角色权限
		if len(config.RequiredRoles) > 0 {
			if !claims.HasAnyRole(config.RequiredRoles...) {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Insufficient permissions",
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "权限不足",
				})
				c.Abort()
				return
			}
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("metadata", claims.Metadata)
		c.Set("claims", claims)

		c.Next()
	}
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	jwtAuth := auth.NewJWTAuth(config.JWTConfig)

	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization required",
				"code":    "AUTH_REQUIRED",
				"message": "请提供有效的认证信息",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"code":    "INVALID_AUTH_FORMAT",
				"message": "认证格式无效，请使用Bearer token",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		claims, err := jwtAuth.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"code":    "INVALID_TOKEN",
				"message": "认证token无效或已过期",
			})
			c.Abort()
			return
		}

		// 检查管理员权限
		if !claims.HasAnyRole("admin", "super_admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Admin access required",
				"code":    "ADMIN_REQUIRED",
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("metadata", claims.Metadata)
		c.Set("claims", claims)

		c.Next()
	}
}

// RoleAuthMiddleware 角色认证中间件
func RoleAuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not authenticated",
				"code":    "NOT_AUTHENTICATED",
				"message": "用户未认证",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Invalid user claims",
				"code":    "INVALID_CLAIMS",
				"message": "用户信息无效",
			})
			c.Abort()
			return
		}

		if !userClaims.HasAnyRole(requiredRoles...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"code":    "INSUFFICIENT_PERMISSIONS",
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPublicPath 检查是否为公开路径
func isPublicPath(path, method string, publicPaths []string) bool {
	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
		// 支持方法匹配
		if strings.Contains(publicPath, ":") {
			parts := strings.Split(publicPath, ":")
			if len(parts) == 2 {
				if parts[0] == method && strings.HasPrefix(path, parts[1]) {
					return true
				}
			}
		}
	}
	return false
}

// isAdminPath 检查是否为管理员路径
func isAdminPath(path, method string, adminPaths []string) bool {
	for _, adminPath := range adminPaths {
		if strings.HasPrefix(path, adminPath) {
			return true
		}
		// 支持方法匹配
		if strings.Contains(adminPath, ":") {
			parts := strings.Split(adminPath, ":")
			if len(parts) == 2 {
				if parts[0] == method && strings.HasPrefix(path, parts[1]) {
					return true
				}
			}
		}
	}
	return false
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUsername 从上下文中获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetUserRoles 从上下文中获取用户角色
func GetUserRoles(c *gin.Context) []string {
	if roles, exists := c.Get("roles"); exists {
		if userRoles, ok := roles.([]string); ok {
			return userRoles
		}
	}
	return []string{}
}

// GetUserClaims 从上下文中获取用户声明
func GetUserClaims(c *gin.Context) *auth.Claims {
	if claims, exists := c.Get("claims"); exists {
		if userClaims, ok := claims.(*auth.Claims); ok {
			return userClaims
		}
	}
	return nil
}
