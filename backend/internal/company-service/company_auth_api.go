package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// CompanyAuthAPI 企业认证API
type CompanyAuthAPI struct {
	core              *jobfirst.Core
	permissionManager *CompanyPermissionManager
	dataSyncService   *CompanyDataSyncService
}

// NewCompanyAuthAPI 创建企业认证API
func NewCompanyAuthAPI(core *jobfirst.Core, permissionManager *CompanyPermissionManager, dataSyncService *CompanyDataSyncService) *CompanyAuthAPI {
	return &CompanyAuthAPI{
		core:              core,
		permissionManager: permissionManager,
		dataSyncService:   dataSyncService,
	}
}

// SetupCompanyAuthRoutes 设置企业认证路由
func (api *CompanyAuthAPI) SetupCompanyAuthRoutes(r *gin.Engine) {
	// 需要认证的API路由
	authMiddleware := api.core.AuthMiddleware.RequireAuth()
	auth := r.Group("/api/v1/company/auth")
	auth.Use(authMiddleware)
	{
		// 企业授权管理API
		auth.POST("/users", api.addAuthorizedUser)
		auth.GET("/users/:company_id", api.getAuthorizedUsers)
		auth.DELETE("/users/:company_id/:user_id", api.removeAuthorizedUser)
		auth.PUT("/users/:company_id/:user_id", api.updateUserRole)
		auth.PUT("/legal-rep/:company_id", api.setLegalRepresentative)

		// 企业权限查询API
		auth.GET("/permissions/:user_id", api.getUserCompanyPermissions)
		auth.GET("/permissions/company/:company_id", api.getCompanyPermissions)

		// 企业认证信息API
		auth.PUT("/company/:company_id/auth-info", api.updateCompanyAuthInfo)
		auth.GET("/company/:company_id/auth-info", api.getCompanyAuthInfo)

		// 企业地理位置API
		auth.PUT("/company/:company_id/location", api.updateCompanyLocation)
		auth.GET("/company/:company_id/location", api.getCompanyLocation)

		// 数据同步API
		auth.POST("/sync/:company_id", api.syncCompanyData)
		auth.GET("/sync/:company_id/status", api.getSyncStatus)
		auth.POST("/sync/:company_id/check", api.checkDataConsistency)

		// 权限审计API
		auth.GET("/audit/:company_id", api.getPermissionAuditLogs)
	}
}

// addAuthorizedUser 添加授权用户
func (api *CompanyAuthAPI) addAuthorizedUser(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	var req struct {
		CompanyID   uint     `json:"company_id" binding:"required"`
		UserID      uint     `json:"user_id" binding:"required"`
		Role        string   `json:"role" binding:"required"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限：只有企业创建者或法定代表人可以添加授权用户
	if !api.permissionManager.CheckCompanyAccess(userID, req.CompanyID, "add_authorized_user", c) {
		return
	}

	// 添加授权用户
	if err := api.permissionManager.AddAuthorizedUser(req.CompanyID, req.UserID, CompanyRole(req.Role), req.Permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 触发数据同步
	go api.dataSyncService.SyncCompanyData(req.CompanyID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "授权用户添加成功",
	})
}

// getAuthorizedUsers 获取企业授权用户列表
func (api *CompanyAuthAPI) getAuthorizedUsers(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "view_authorized_users", c) {
		return
	}

	// 获取授权用户列表
	users, err := api.permissionManager.GetCompanyAuthorizedUsers(uint(companyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   users,
	})
}

// removeAuthorizedUser 移除授权用户
func (api *CompanyAuthAPI) removeAuthorizedUser(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "remove_authorized_user", c) {
		return
	}

	// 移除授权用户
	if err := api.permissionManager.RemoveAuthorizedUser(uint(companyID), uint(targetUserID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 触发数据同步
	go api.dataSyncService.SyncCompanyData(uint(companyID))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "授权用户移除成功",
	})
}

// updateUserRole 更新用户角色
func (api *CompanyAuthAPI) updateUserRole(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Role        string   `json:"role" binding:"required"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "update_user_role", c) {
		return
	}

	// 更新用户角色
	if err := api.permissionManager.UpdateUserRole(uint(companyID), uint(targetUserID), CompanyRole(req.Role), req.Permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 触发数据同步
	go api.dataSyncService.SyncCompanyData(uint(companyID))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "用户角色更新成功",
	})
}

// setLegalRepresentative 设置法定代表人
func (api *CompanyAuthAPI) setLegalRepresentative(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限：只有企业创建者可以设置法定代表人
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "set_legal_representative", c) {
		return
	}

	// 设置法定代表人
	if err := api.permissionManager.SetLegalRepresentative(uint(companyID), req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 触发数据同步
	go api.dataSyncService.SyncCompanyData(uint(companyID))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "法定代表人设置成功",
	})
}

// getUserCompanyPermissions 获取用户的企业权限列表
func (api *CompanyAuthAPI) getUserCompanyPermissions(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 检查权限：只能查看自己的权限，或者系统管理员可以查看所有权限
	role := c.GetString("role")
	if userID != uint(targetUserID) && role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	// 获取用户企业权限
	permissions, err := api.permissionManager.GetUserCompanyPermissions(uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   permissions,
	})
}

// getCompanyPermissions 获取企业权限信息
func (api *CompanyAuthAPI) getCompanyPermissions(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "view_company_permissions", c) {
		return
	}

	// 获取企业权限信息
	users, err := api.permissionManager.GetCompanyAuthorizedUsers(uint(companyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   users,
	})
}

// updateCompanyAuthInfo 更新企业认证信息
func (api *CompanyAuthAPI) updateCompanyAuthInfo(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	var authInfo CompanyAuthInfo
	if err := c.ShouldBindJSON(&authInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "update_auth_info", c) {
		return
	}

	// 更新企业认证信息
	db := api.core.GetDB()
	var company EnhancedCompany
	if err := db.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
		return
	}

	company.SetAuthInfo(&authInfo)
	company.UpdatedAt = time.Now()

	if err := db.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新企业认证信息失败"})
		return
	}

	// 触发数据同步
	go api.dataSyncService.SyncCompanyData(uint(companyID))

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   company.GetAuthInfo(),
	})
}

// getCompanyAuthInfo 获取企业认证信息
func (api *CompanyAuthAPI) getCompanyAuthInfo(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "view_auth_info", c) {
		return
	}

	// 获取企业认证信息
	db := api.core.GetDB()
	var company EnhancedCompany
	if err := db.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   company.GetAuthInfo(),
	})
}

// updateCompanyLocation 更新企业地理位置
func (api *CompanyAuthAPI) updateCompanyLocation(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	var locationInfo CompanyLocationInfo
	if err := c.ShouldBindJSON(&locationInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "update_location", c) {
		return
	}

	// 更新企业地理位置
	db := api.core.GetDB()
	var company EnhancedCompany
	if err := db.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
		return
	}

	company.SetLocationInfo(&locationInfo)
	company.UpdatedAt = time.Now()

	if err := db.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新企业地理位置失败"})
		return
	}

	// 触发数据同步
	go api.dataSyncService.SyncCompanyData(uint(companyID))

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   company.GetLocationInfo(),
	})
}

// getCompanyLocation 获取企业地理位置
func (api *CompanyAuthAPI) getCompanyLocation(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "view_location", c) {
		return
	}

	// 获取企业地理位置
	db := api.core.GetDB()
	var company EnhancedCompany
	if err := db.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   company.GetLocationInfo(),
	})
}

// syncCompanyData 同步企业数据
func (api *CompanyAuthAPI) syncCompanyData(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "sync_data", c) {
		return
	}

	// 执行数据同步
	if err := api.dataSyncService.SyncCompanyData(uint(companyID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "企业数据同步成功",
	})
}

// getSyncStatus 获取同步状态
func (api *CompanyAuthAPI) getSyncStatus(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "view_sync_status", c) {
		return
	}

	// 获取同步状态
	syncStatus, err := api.dataSyncService.GetSyncStatus(uint(companyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   syncStatus,
	})
}

// checkDataConsistency 检查数据一致性
func (api *CompanyAuthAPI) checkDataConsistency(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "check_consistency", c) {
		return
	}

	// 检查数据一致性
	if err := api.dataSyncService.CheckDataConsistency(uint(companyID)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "inconsistent",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "consistent",
		"message": "数据一致性检查通过",
	})
}

// getPermissionAuditLogs 获取权限审计日志
func (api *CompanyAuthAPI) getPermissionAuditLogs(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
		return
	}

	// 检查权限
	if !api.permissionManager.CheckCompanyAccess(userID, uint(companyID), "view_audit_logs", c) {
		return
	}

	// 获取查询参数
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// 获取权限审计日志
	logs, err := api.permissionManager.GetPermissionAuditLogs(uint(companyID), 0, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   logs,
	})
}
