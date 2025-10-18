package middleware

import (
	"net/http"
	"strings"

	"github.com/xiajason/zervi-basic/basic/backend/internal/domain/auth"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/logger"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/rbac"

	"github.com/gin-gonic/gin"
)

func RequireRole(rbacManager *rbac.Manager, logger logger.Logger, requiredRoles ...auth.RoleName) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// 检查用户角色
		userRoles, err := rbacManager.GetRolesForUser(userIDStr)
		if err != nil {
			logger.Errorf("Failed to get user roles: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user roles"})
			c.Abort()
			return
		}

		// 检查是否有任一所需角色
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if string(requiredRole) == userRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireSuperAdmin(rbacManager *rbac.Manager, logger logger.Logger) gin.HandlerFunc {
	return RequireRole(rbacManager, logger, auth.RoleSuperAdmin)
}

func RequireAdmin(rbacManager *rbac.Manager, logger logger.Logger) gin.HandlerFunc {
	return RequireRole(rbacManager, logger, auth.RoleSuperAdmin, auth.RoleAdmin)
}

func RequireDevTeam(rbacManager *rbac.Manager, logger logger.Logger) gin.HandlerFunc {
	return RequireRole(rbacManager, logger, auth.RoleSuperAdmin, auth.RoleAdmin, auth.RoleDevTeam)
}

func RequirePermission(rbacManager *rbac.Manager, logger logger.Logger, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := rbacManager.HasPermission(userIDStr, resource, action)
		if err != nil {
			logger.Errorf("Failed to check permission: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func AutoPermissionCheck(rbacManager *rbac.Manager, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		// 根据路径自动确定资源和操作
		path := c.Request.URL.Path
		method := c.Request.Method

		var resource, action string

		// 解析资源
		pathParts := strings.Split(strings.Trim(path, "/"), "/")
		if len(pathParts) >= 3 {
			resource = pathParts[2] // 假设路径格式为 /api/v1/resource
		}

		// 解析操作
		switch method {
		case "GET":
			action = "read"
		case "POST":
			action = "write"
		case "PUT", "PATCH":
			action = "write"
		case "DELETE":
			action = "delete"
		default:
			action = "read"
		}

		// 检查权限
		if resource != "" && action != "" {
			hasPermission, err := rbacManager.HasPermission(userIDStr, resource, action)
			if err != nil {
				logger.Errorf("Failed to check auto permission: %v", err)
				c.Next()
				return
			}

			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
