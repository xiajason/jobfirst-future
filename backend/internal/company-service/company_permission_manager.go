package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// CompanyPermissionManager 企业权限管理器
type CompanyPermissionManager struct {
	mysqlDB     *gorm.DB
	redisClient *redis.Client
	cacheTTL    time.Duration
}

// NewCompanyPermissionManager 创建企业权限管理器
func NewCompanyPermissionManager(mysqlDB *gorm.DB, redisClient *redis.Client) *CompanyPermissionManager {
	return &CompanyPermissionManager{
		mysqlDB:     mysqlDB,
		redisClient: redisClient,
		cacheTTL:    time.Hour, // 缓存1小时
	}
}

// CheckCompanyAccess 检查企业访问权限
func (cpm *CompanyPermissionManager) CheckCompanyAccess(userID uint, companyID uint, action string, c *gin.Context) bool {
	// 1. 尝试从缓存获取权限
	cacheKey := fmt.Sprintf("company_permission:%d:%d:%s", userID, companyID, action)
	if cpm.redisClient != nil {
		if cached, err := cpm.redisClient.Get(context.Background(), cacheKey).Result(); err == nil {
			if cached == "true" {
				cpm.logPermissionCheck(userID, companyID, action, true, c)
				return true
			}
			cpm.logPermissionCheck(userID, companyID, action, false, c)
			return false
		}
	}

	// 2. 检查系统管理员权限
	var user User
	if err := cpm.mysqlDB.First(&user, userID).Error; err == nil {
		if user.Role == "admin" || user.Role == "super_admin" {
			if cpm.redisClient != nil {
				cpm.redisClient.Set(context.Background(), cacheKey, "true", cpm.cacheTTL)
			}
			cpm.logPermissionCheck(userID, companyID, action, true, c)
			return true
		}
	}

	// 3. 检查企业权限
	var company EnhancedCompany
	if err := cpm.mysqlDB.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
		cpm.logPermissionCheck(userID, companyID, action, false, c)
		return false
	}

	// 4. 检查企业创建者权限
	if company.CreatedBy == userID {
		if cpm.redisClient != nil {
			cpm.redisClient.Set(context.Background(), cacheKey, "true", cpm.cacheTTL)
		}
		cpm.logPermissionCheck(userID, companyID, action, true, c)
		return true
	}

	// 5. 检查法定代表人权限
	if company.LegalRepUserID == userID {
		if cpm.redisClient != nil {
			cpm.redisClient.Set(context.Background(), cacheKey, "true", cpm.cacheTTL)
		}
		cpm.logPermissionCheck(userID, companyID, action, true, c)
		return true
	}

	// 6. 检查企业用户关联权限
	var companyUser CompanyUser
	if err := cpm.mysqlDB.Where("company_id = ? AND user_id = ? AND status = ?",
		companyID, userID, "active").First(&companyUser).Error; err == nil {
		if cpm.redisClient != nil {
			cpm.redisClient.Set(context.Background(), cacheKey, "true", cpm.cacheTTL)
		}
		cpm.logPermissionCheck(userID, companyID, action, true, c)
		return true
	}

	// 7. 检查JSON授权用户列表
	if company.AuthorizedUsers != "" {
		var authorizedUsers []uint
		if err := json.Unmarshal([]byte(company.AuthorizedUsers), &authorizedUsers); err == nil {
			for _, authorizedUserID := range authorizedUsers {
				if authorizedUserID == userID {
					if cpm.redisClient != nil {
						cpm.redisClient.Set(context.Background(), cacheKey, "true", cpm.cacheTTL)
					}
					cpm.logPermissionCheck(userID, companyID, action, true, c)
					return true
				}
			}
		}
	}

	// 8. 权限不足
	if cpm.redisClient != nil {
		cpm.redisClient.Set(context.Background(), cacheKey, "false", cpm.cacheTTL)
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，您没有访问该企业的权限"})
	cpm.logPermissionCheck(userID, companyID, action, false, c)
	return false
}

// GetUserCompanyPermissions 获取用户的企业权限列表
func (cpm *CompanyPermissionManager) GetUserCompanyPermissions(userID uint) ([]CompanyPermissionInfo, error) {
	var permissions []CompanyPermissionInfo

	// 1. 获取用户信息
	var user User
	if err := cpm.mysqlDB.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 2. 如果是系统管理员，返回所有企业权限
	if user.Role == "admin" || user.Role == "super_admin" {
		var companies []EnhancedCompany
		if err := cpm.mysqlDB.Find(&companies).Error; err != nil {
			return nil, err
		}

		for _, company := range companies {
			permissions = append(permissions, CompanyPermissionInfo{
				UserID:                   userID,
				CompanyID:                company.ID,
				EffectivePermissionLevel: PermissionSystemAdmin,
				Role:                     user.Role,
				Status:                   "active",
				Permissions:              []string{"*"}, // 系统管理员拥有所有权限
			})
		}
		return permissions, nil
	}

	// 3. 获取用户创建的企业
	var createdCompanies []EnhancedCompany
	if err := cpm.mysqlDB.Where("created_by = ?", userID).Find(&createdCompanies).Error; err != nil {
		return nil, err
	}

	for _, company := range createdCompanies {
		permissions = append(permissions, CompanyPermissionInfo{
			UserID:                   userID,
			CompanyID:                company.ID,
			EffectivePermissionLevel: PermissionCompanyOwner,
			Role:                     "company_owner",
			Status:                   "active",
			Permissions:              []string{"*"}, // 企业创建者拥有所有权限
		})
	}

	// 4. 获取用户作为法定代表人的企业
	var legalRepCompanies []EnhancedCompany
	if err := cpm.mysqlDB.Where("legal_rep_user_id = ?", userID).Find(&legalRepCompanies).Error; err != nil {
		return nil, err
	}

	for _, company := range legalRepCompanies {
		permissions = append(permissions, CompanyPermissionInfo{
			UserID:                   userID,
			CompanyID:                company.ID,
			EffectivePermissionLevel: PermissionLegalRepresentative,
			Role:                     string(RoleLegalRepresentative),
			Status:                   "active",
			Permissions:              []string{"read", "write", "manage_users"},
		})
	}

	// 5. 获取用户作为授权用户的企业
	var companyUsers []CompanyUser
	if err := cpm.mysqlDB.Preload("Company").Where("user_id = ? AND status = ?", userID, "active").Find(&companyUsers).Error; err != nil {
		return nil, err
	}

	for _, companyUser := range companyUsers {
		permissions = append(permissions, CompanyPermissionInfo{
			UserID:                   userID,
			CompanyID:                companyUser.CompanyID,
			EffectivePermissionLevel: PermissionAuthorizedUser,
			Role:                     companyUser.Role,
			Status:                   companyUser.Status,
			Permissions:              companyUser.GetPermissions(),
		})
	}

	return permissions, nil
}

// AddAuthorizedUser 添加授权用户
func (cpm *CompanyPermissionManager) AddAuthorizedUser(companyID uint, userID uint, role CompanyRole, permissions []string) error {
	// 检查企业是否存在
	var company EnhancedCompany
	if err := cpm.mysqlDB.First(&company, companyID).Error; err != nil {
		return fmt.Errorf("企业不存在: %v", err)
	}

	// 检查用户是否存在
	var user User
	if err := cpm.mysqlDB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %v", err)
	}

	// 创建企业用户关联
	companyUser := CompanyUser{
		CompanyID: companyID,
		UserID:    userID,
		Role:      string(role),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	companyUser.SetPermissions(permissions)

	if err := cpm.mysqlDB.Create(&companyUser).Error; err != nil {
		return fmt.Errorf("添加授权用户失败: %v", err)
	}

	// 清除相关缓存
	cpm.clearCompanyPermissionCache(companyID)

	return nil
}

// RemoveAuthorizedUser 移除授权用户
func (cpm *CompanyPermissionManager) RemoveAuthorizedUser(companyID uint, userID uint) error {
	// 删除企业用户关联
	if err := cpm.mysqlDB.Where("company_id = ? AND user_id = ?", companyID, userID).Delete(&CompanyUser{}).Error; err != nil {
		return fmt.Errorf("移除授权用户失败: %v", err)
	}

	// 清除相关缓存
	cpm.clearCompanyPermissionCache(companyID)

	return nil
}

// UpdateUserRole 更新用户角色
func (cpm *CompanyPermissionManager) UpdateUserRole(companyID uint, userID uint, role CompanyRole, permissions []string) error {
	var companyUser CompanyUser
	if err := cpm.mysqlDB.Where("company_id = ? AND user_id = ?", companyID, userID).First(&companyUser).Error; err != nil {
		return fmt.Errorf("企业用户关联不存在: %v", err)
	}

	companyUser.Role = string(role)
	companyUser.SetPermissions(permissions)
	companyUser.UpdatedAt = time.Now()

	if err := cpm.mysqlDB.Save(&companyUser).Error; err != nil {
		return fmt.Errorf("更新用户角色失败: %v", err)
	}

	// 清除相关缓存
	cpm.clearCompanyPermissionCache(companyID)

	return nil
}

// SetLegalRepresentative 设置法定代表人
func (cpm *CompanyPermissionManager) SetLegalRepresentative(companyID uint, userID uint) error {
	// 检查企业是否存在
	var company EnhancedCompany
	if err := cpm.mysqlDB.First(&company, companyID).Error; err != nil {
		return fmt.Errorf("企业不存在: %v", err)
	}

	// 检查用户是否存在
	var user User
	if err := cpm.mysqlDB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %v", err)
	}

	// 更新企业法定代表人
	company.LegalRepUserID = userID
	company.UpdatedAt = time.Now()

	if err := cpm.mysqlDB.Save(&company).Error; err != nil {
		return fmt.Errorf("设置法定代表人失败: %v", err)
	}

	// 确保用户在企业用户关联表中
	var companyUser CompanyUser
	if err := cpm.mysqlDB.Where("company_id = ? AND user_id = ?", companyID, userID).First(&companyUser).Error; err != nil {
		// 如果不存在，创建关联
		companyUser = CompanyUser{
			CompanyID: companyID,
			UserID:    userID,
			Role:      string(RoleLegalRepresentative),
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		companyUser.SetPermissions([]string{"read", "write", "manage_users"})

		if err := cpm.mysqlDB.Create(&companyUser).Error; err != nil {
			return fmt.Errorf("创建企业用户关联失败: %v", err)
		}
	} else {
		// 如果存在，更新角色
		companyUser.Role = string(RoleLegalRepresentative)
		companyUser.SetPermissions([]string{"read", "write", "manage_users"})
		companyUser.UpdatedAt = time.Now()

		if err := cpm.mysqlDB.Save(&companyUser).Error; err != nil {
			return fmt.Errorf("更新企业用户关联失败: %v", err)
		}
	}

	// 清除相关缓存
	cpm.clearCompanyPermissionCache(companyID)

	return nil
}

// GetCompanyAuthorizedUsers 获取企业授权用户列表
func (cpm *CompanyPermissionManager) GetCompanyAuthorizedUsers(companyID uint) ([]CompanyUser, error) {
	var companyUsers []CompanyUser
	if err := cpm.mysqlDB.Preload("User").Where("company_id = ?", companyID).Find(&companyUsers).Error; err != nil {
		return nil, fmt.Errorf("获取企业授权用户失败: %v", err)
	}

	return companyUsers, nil
}

// logPermissionCheck 记录权限检查日志
func (cpm *CompanyPermissionManager) logPermissionCheck(userID uint, companyID uint, action string, result bool, c *gin.Context) {
	auditLog := CompanyPermissionAuditLog{
		CompanyID:        companyID,
		UserID:           userID,
		Action:           action,
		ResourceType:     "company",
		ResourceID:       &companyID,
		PermissionResult: result,
		IPAddress:        c.ClientIP(),
		UserAgent:        c.GetHeader("User-Agent"),
		CreatedAt:        time.Now(),
	}

	if err := cpm.mysqlDB.Create(&auditLog).Error; err != nil {
		log.Printf("记录权限检查日志失败: %v", err)
	}
}

// clearCompanyPermissionCache 清除企业权限缓存
func (cpm *CompanyPermissionManager) clearCompanyPermissionCache(companyID uint) {
	if cpm.redisClient == nil {
		return
	}

	// 清除该企业相关的所有权限缓存
	pattern := fmt.Sprintf("company_permission:*:%d:*", companyID)
	keys, err := cpm.redisClient.Keys(context.Background(), pattern).Result()
	if err != nil {
		log.Printf("获取权限缓存键失败: %v", err)
		return
	}

	if len(keys) > 0 {
		if err := cpm.redisClient.Del(context.Background(), keys...).Err(); err != nil {
			log.Printf("清除权限缓存失败: %v", err)
		}
	}
}

// GetPermissionAuditLogs 获取权限审计日志
func (cpm *CompanyPermissionManager) GetPermissionAuditLogs(companyID uint, userID uint, limit int) ([]CompanyPermissionAuditLog, error) {
	var logs []CompanyPermissionAuditLog
	query := cpm.mysqlDB.Preload("User").Order("created_at DESC")

	if companyID > 0 {
		query = query.Where("company_id = ?", companyID)
	}

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("获取权限审计日志失败: %v", err)
	}

	return logs, nil
}
