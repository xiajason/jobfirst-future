package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core/auth"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	authManager *auth.AuthManager
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(authManager *auth.AuthManager) *AuthMiddleware {
	return &AuthMiddleware{
		authManager: authManager,
	}
}

// RequireAuth 需要登录的中间件
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("DEBUG: 认证中间件 - 开始处理请求: %s %s", c.Request.Method, c.Request.URL.Path)
		
		token := am.extractToken(c)
		log.Printf("DEBUG: 认证中间件 - 提取到的token: %s", func() string {
			if token == "" {
				return "空token"
			}
			if len(token) > 50 {
				return token[:50] + "..."
			}
			return token
		}())
		
		if token == "" {
			log.Printf("DEBUG: 认证中间件 - token为空，返回未登录")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		log.Printf("DEBUG: 认证中间件 - 开始验证token")
		claims, err := am.authManager.ValidateToken(token)
		if err != nil {
			log.Printf("DEBUG: 认证中间件 - token验证失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "无效的token",
			})
			c.Abort()
			return
		}

		log.Printf("DEBUG: 认证中间件 - token验证成功，用户ID: %d, 用户名: %s, 角色: %s", claims.UserID, claims.Username, claims.Role)
		
		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		log.Printf("DEBUG: 认证中间件 - 用户信息已设置到上下文，继续处理请求")
		c.Next()
	}
}

// RequireDevTeam 需要开发团队权限的中间件
func (am *AuthMiddleware) RequireDevTeam() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		devTeam, err := am.authManager.GetDevTeamUser(userID.(uint))
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "需要开发团队成员权限",
			})
			c.Abort()
			return
		}

		c.Set("dev_team", devTeam)
		c.Next()
	}
}

// RequireSuperAdmin 需要超级管理员权限的中间件
func (am *AuthMiddleware) RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		hasPermission, err := am.authManager.CheckPermission(userID.(uint), "super_admin")
		if err != nil || !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "需要超级管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 需要管理员权限的中间件
func (am *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		// 检查是否为管理员（超级管理员或系统管理员）
		hasPermission, err := am.authManager.CheckPermission(userID.(uint), "super_admin")
		if err == nil && hasPermission {
			c.Next()
			return
		}

		hasPermission, err = am.authManager.CheckPermission(userID.(uint), "system_admin")
		if err != nil || !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (am *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		hasPermission, err := am.authManager.CheckPermission(userID.(uint), requiredRole)
		if err != nil || !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 需要任意一个角色的中间件
func (am *AuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未登录",
			})
			c.Abort()
			return
		}

		for _, role := range roles {
			hasPermission, err := am.authManager.CheckPermission(userID.(uint), role)
			if err == nil && hasPermission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "权限不足",
		})
		c.Abort()
	}
}

// extractToken 从请求中提取token
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	log.Printf("DEBUG: extractToken - 开始提取token")
	
	// 从Authorization头获取
	authHeader := c.GetHeader("Authorization")
	log.Printf("DEBUG: extractToken - Authorization header: '%s'", authHeader)
	
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		log.Printf("DEBUG: extractToken - 分割后的parts: %v, 长度: %d", parts, len(parts))
		
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			if len(token) > 50 {
				log.Printf("DEBUG: extractToken - 找到Bearer token: %s...", token[:50])
			} else {
				log.Printf("DEBUG: extractToken - 找到Bearer token: %s", token)
			}
			return token
		} else {
			log.Printf("DEBUG: extractToken - Bearer格式不正确或缺失")
		}
	} else {
		log.Printf("DEBUG: extractToken - Authorization header为空")
	}

	// 从查询参数获取
	token := c.Query("token")
	log.Printf("DEBUG: extractToken - 查询参数token: '%s'", token)
	if token != "" {
		return token
	}

	// 从Cookie获取
	cookie, err := c.Cookie("token")
	log.Printf("DEBUG: extractToken - Cookie token: '%s', 错误: %v", cookie, err)
	if err == nil && cookie != "" {
		return cookie
	}

	log.Printf("DEBUG: extractToken - 未找到任何token")
	return ""
}
