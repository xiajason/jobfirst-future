package authHandler

import (
	"fmt"
	"net/http"

	authApp "github.com/xiajason/zervi-basic/basic/backend/internal/app/auth"
	authDomain "github.com/xiajason/zervi-basic/basic/backend/internal/domain/auth"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/logger"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/rbac"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService *authApp.Service
	rbacManager *rbac.Manager
	logger      logger.Logger
}

func NewHandler(authService *authApp.Service, rbacManager *rbac.Manager, logger logger.Logger) *Handler {
	return &Handler{
		authService: authService,
		rbacManager: rbacManager,
		logger:      logger,
	}
}

func (h *Handler) InitializeSuperAdmin(c *gin.Context) {
	var req authDomain.InitializeSuperAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.InitializeSuperAdmin(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": resp.Message,
		"data": gin.H{
			"user_id": resp.UserID,
		},
	})
}

func (h *Handler) CheckSuperAdminStatus(c *gin.Context) {
	var req authDomain.CheckSuperAdminStatusRequest
	resp, err := h.authService.CheckSuperAdminStatus(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

func (h *Handler) ResetSuperAdminPassword(c *gin.Context) {
	var req authDomain.ResetSuperAdminPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.ResetSuperAdminPassword(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Super admin password reset successfully",
	})
}

func (h *Handler) CheckPermission(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 处理不同类型的userID
	var userIDStr string
	switch v := userID.(type) {
	case string:
		userIDStr = v
	case float64:
		userIDStr = fmt.Sprintf("%.0f", v)
	case int:
		userIDStr = fmt.Sprintf("%d", v)
	case uint:
		userIDStr = fmt.Sprintf("%d", v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	resource := c.Query("resource")
	action := c.Query("action")

	if resource == "" || action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource and action parameters are required"})
		return
	}

	hasPermission, err := h.rbacManager.HasPermission(userIDStr, resource, action)
	if err != nil {
		h.logger.Errorf("Failed to check permission: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user_id":        userIDStr,
			"resource":       resource,
			"action":         action,
			"has_permission": hasPermission,
		},
	})
}

func (h *Handler) GetUserRoles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 处理不同类型的userID
	var userIDStr string
	switch v := userID.(type) {
	case string:
		userIDStr = v
	case float64:
		userIDStr = fmt.Sprintf("%.0f", v)
	case int:
		userIDStr = fmt.Sprintf("%d", v)
	case uint:
		userIDStr = fmt.Sprintf("%d", v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	roles, err := h.rbacManager.GetRolesForUser(userIDStr)
	if err != nil {
		h.logger.Errorf("Failed to get user roles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user_id": userIDStr,
			"roles":   roles,
		},
	})
}

func (h *Handler) GetUserPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 处理不同类型的userID
	var userIDStr string
	switch v := userID.(type) {
	case string:
		userIDStr = v
	case float64:
		userIDStr = fmt.Sprintf("%.0f", v)
	case int:
		userIDStr = fmt.Sprintf("%d", v)
	case uint:
		userIDStr = fmt.Sprintf("%d", v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	permissions, err := h.rbacManager.GetPermissionsForUser(userIDStr)
	if err != nil {
		h.logger.Errorf("Failed to get user permissions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user_id":     userIDStr,
			"permissions": permissions,
		},
	})
}
