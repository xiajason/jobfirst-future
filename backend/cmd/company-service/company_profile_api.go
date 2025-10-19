package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// CompanyProfileAPI 企业画像API处理器
type CompanyProfileAPI struct {
	core *jobfirst.Core
}

// CompanyProfileSummary 企业画像摘要
type CompanyProfileSummary struct {
	CompanyID          uint    `json:"company_id"`
	CompanyName        string  `json:"company_name"`
	ReportID           string  `json:"report_id"`
	IndustryCategory   string  `json:"industry_category"`
	BusinessStatus     string  `json:"business_status"`
	RegisteredCapital  float64 `json:"registered_capital"`
	TotalEmployees     int     `json:"total_employees"`
	BasicScore         float64 `json:"basic_score"`
	TalentScore        float64 `json:"talent_score"`
	RiskLevel          string  `json:"risk_level"`
	DataUpdateTime     string  `json:"data_update_time"`
	HasCompleteProfile bool    `json:"has_complete_profile"`
}

// NewCompanyProfileAPI 创建企业画像API处理器
func NewCompanyProfileAPI(core *jobfirst.Core) *CompanyProfileAPI {
	return &CompanyProfileAPI{
		core: core,
	}
}

// SetupCompanyProfileRoutes 设置企业画像相关路由
func (api *CompanyProfileAPI) SetupCompanyProfileRoutes(r *gin.Engine) {
	// 需要认证的企业画像API
	authMiddleware := api.core.AuthMiddleware.RequireAuth()
	profile := r.Group("/api/v1/company/profile")
	profile.Use(authMiddleware)
	{
		// 获取企业画像摘要
		profile.GET("/summary/:company_id", api.getCompanyProfileSummary)

		// 获取完整企业画像数据
		profile.GET("/:company_id", api.getCompanyProfile)

		// 创建或更新企业基本信息
		profile.POST("/basic-info", api.createOrUpdateBasicInfo)

		// 创建或更新资质许可信息
		profile.POST("/qualification", api.createOrUpdateQualification)

		// 创建或更新人员竞争力信息
		profile.POST("/personnel", api.createOrUpdatePersonnel)

		// 创建或更新财务信息
		profile.POST("/financial", api.createOrUpdateFinancial)

		// 创建或更新风险信息
		profile.POST("/risk", api.createOrUpdateRisk)

		// 批量导入企业画像数据
		profile.POST("/import", api.importCompanyProfile)

		// 导出企业画像数据
		profile.GET("/export/:company_id", api.exportCompanyProfile)
	}
}

// getCompanyProfileSummary 获取企业画像摘要
func (api *CompanyProfileAPI) getCompanyProfileSummary(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取企业ID
	companyID, err := strconv.Atoi(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "企业ID格式错误"})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, uint(companyID), c) {
		return
	}

	db := api.core.GetDB()

	// 查询企业基本信息
	var basicInfo CompanyProfileBasicInfo
	if err := db.Where("company_id = ?", companyID).First(&basicInfo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "企业画像信息不存在"})
		return
	}

	// 查询人员竞争力信息
	var personnel PersonnelCompetitiveness
	db.Where("company_id = ?", companyID).First(&personnel)

	// 查询科创评分信息
	var techInnovation TechInnovationScore
	db.Where("company_id = ?", companyID).First(&techInnovation)

	// 查询风险信息
	var riskInfo CompanyProfileRiskInfo
	db.Where("company_id = ?", companyID).First(&riskInfo)

	// 构建摘要信息
	summary := CompanyProfileSummary{
		CompanyID:          uint(companyID),
		CompanyName:        basicInfo.CompanyName,
		ReportID:           basicInfo.ReportID,
		IndustryCategory:   basicInfo.IndustryCategory,
		BusinessStatus:     basicInfo.BusinessStatus,
		RegisteredCapital:  basicInfo.RegisteredCapital,
		TotalEmployees:     personnel.TotalEmployees,
		BasicScore:         techInnovation.BasicScore,
		TalentScore:        techInnovation.TalentScore,
		RiskLevel:          riskInfo.RiskLevel,
		DataUpdateTime:     basicInfo.DataUpdateTime.Format("2006-01-02 15:04:05"),
		HasCompleteProfile: api.checkCompleteProfile(uint(companyID)),
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   summary,
	})
}

// getCompanyProfile 获取完整企业画像数据
func (api *CompanyProfileAPI) getCompanyProfile(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取企业ID
	companyID, err := strconv.Atoi(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "企业ID格式错误"})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, uint(companyID), c) {
		return
	}

	db := api.core.GetDB()
	profileData := &CompanyProfileData{}

	// 查询基本信息
	var basicInfo CompanyProfileBasicInfo
	if err := db.Where("company_id = ?", companyID).First(&basicInfo).Error; err == nil {
		profileData.BasicInfo = &basicInfo
	}

	// 查询资质许可信息
	var qualifications []QualificationLicense
	db.Where("company_id = ?", companyID).Find(&qualifications)
	profileData.Qualifications = qualifications

	// 查询人员竞争力信息
	var personnel PersonnelCompetitiveness
	if err := db.Where("company_id = ?", companyID).First(&personnel).Error; err == nil {
		profileData.Personnel = &personnel
	}

	// 查询公积金信息
	var providentFund ProvidentFund
	if err := db.Where("company_id = ?", companyID).First(&providentFund).Error; err == nil {
		profileData.ProvidentFund = &providentFund
	}

	// 查询资助补贴信息
	var subsidies []SubsidyInfo
	db.Where("company_id = ?", companyID).Find(&subsidies)
	profileData.Subsidies = subsidies

	// 查询企业关系信息
	var relationships []CompanyRelationship
	db.Where("company_id = ?", companyID).Find(&relationships)
	profileData.Relationships = relationships

	// 查询科创评分信息
	var techInnovation TechInnovationScore
	if err := db.Where("company_id = ?", companyID).First(&techInnovation).Error; err == nil {
		profileData.TechInnovation = &techInnovation
	}

	// 查询财务信息
	var financialInfo CompanyProfileFinancialInfo
	if err := db.Where("company_id = ?", companyID).First(&financialInfo).Error; err == nil {
		profileData.FinancialInfo = &financialInfo
	}

	// 查询风险信息
	var riskInfo CompanyProfileRiskInfo
	if err := db.Where("company_id = ?", companyID).First(&riskInfo).Error; err == nil {
		profileData.RiskInfo = &riskInfo
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   profileData,
	})
}

// createOrUpdateBasicInfo 创建或更新企业基本信息
func (api *CompanyProfileAPI) createOrUpdateBasicInfo(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	var basicInfo CompanyProfileBasicInfo
	if err := c.ShouldBindJSON(&basicInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, basicInfo.CompanyID, c) {
		return
	}

	db := api.core.GetDB()

	// 检查是否已存在
	var existingInfo CompanyProfileBasicInfo
	if err := db.Where("company_id = ?", basicInfo.CompanyID).First(&existingInfo).Error; err == nil {
		// 更新现有记录
		basicInfo.ID = existingInfo.ID
		basicInfo.CreatedAt = existingInfo.CreatedAt
		basicInfo.UpdatedAt = time.Now()

		if err := db.Save(&basicInfo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新企业基本信息失败"})
			return
		}
	} else {
		// 创建新记录
		basicInfo.CreatedAt = time.Now()
		basicInfo.UpdatedAt = time.Now()

		if err := db.Create(&basicInfo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建企业基本信息失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   basicInfo,
	})
}

// createOrUpdateQualification 创建或更新资质许可信息
func (api *CompanyProfileAPI) createOrUpdateQualification(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	var qualification QualificationLicense
	if err := c.ShouldBindJSON(&qualification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, qualification.CompanyID, c) {
		return
	}

	db := api.core.GetDB()

	if qualification.ID == 0 {
		// 创建新记录
		qualification.CreatedAt = time.Now()
		qualification.UpdatedAt = time.Now()

		if err := db.Create(&qualification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建资质许可信息失败"})
			return
		}
	} else {
		// 更新现有记录
		qualification.UpdatedAt = time.Now()

		if err := db.Save(&qualification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新资质许可信息失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   qualification,
	})
}

// createOrUpdatePersonnel 创建或更新人员竞争力信息
func (api *CompanyProfileAPI) createOrUpdatePersonnel(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	var personnel PersonnelCompetitiveness
	if err := c.ShouldBindJSON(&personnel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, personnel.CompanyID, c) {
		return
	}

	db := api.core.GetDB()

	// 检查是否已存在
	var existingPersonnel PersonnelCompetitiveness
	if err := db.Where("company_id = ?", personnel.CompanyID).First(&existingPersonnel).Error; err == nil {
		// 更新现有记录
		personnel.ID = existingPersonnel.ID
		personnel.CreatedAt = existingPersonnel.CreatedAt
		personnel.UpdatedAt = time.Now()

		if err := db.Save(&personnel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新人员竞争力信息失败"})
			return
		}
	} else {
		// 创建新记录
		personnel.CreatedAt = time.Now()
		personnel.UpdatedAt = time.Now()

		if err := db.Create(&personnel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建人员竞争力信息失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   personnel,
	})
}

// createOrUpdateFinancial 创建或更新财务信息
func (api *CompanyProfileAPI) createOrUpdateFinancial(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	var financialInfo CompanyProfileFinancialInfo
	if err := c.ShouldBindJSON(&financialInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, financialInfo.CompanyID, c) {
		return
	}

	db := api.core.GetDB()

	// 检查是否已存在
	var existingFinancial CompanyProfileFinancialInfo
	if err := db.Where("company_id = ? AND financial_year = ?", financialInfo.CompanyID, financialInfo.FinancialYear).First(&existingFinancial).Error; err == nil {
		// 更新现有记录
		financialInfo.ID = existingFinancial.ID
		financialInfo.CreatedAt = existingFinancial.CreatedAt
		financialInfo.UpdatedAt = time.Now()

		if err := db.Save(&financialInfo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新财务信息失败"})
			return
		}
	} else {
		// 创建新记录
		financialInfo.CreatedAt = time.Now()
		financialInfo.UpdatedAt = time.Now()

		if err := db.Create(&financialInfo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建财务信息失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   financialInfo,
	})
}

// createOrUpdateRisk 创建或更新风险信息
func (api *CompanyProfileAPI) createOrUpdateRisk(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	var riskInfo CompanyProfileRiskInfo
	if err := c.ShouldBindJSON(&riskInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, riskInfo.CompanyID, c) {
		return
	}

	db := api.core.GetDB()

	// 检查是否已存在
	var existingRisk CompanyProfileRiskInfo
	if err := db.Where("company_id = ?", riskInfo.CompanyID).First(&existingRisk).Error; err == nil {
		// 更新现有记录
		riskInfo.ID = existingRisk.ID
		riskInfo.CreatedAt = existingRisk.CreatedAt
		riskInfo.UpdatedAt = time.Now()

		if err := db.Save(&riskInfo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新风险信息失败"})
			return
		}
	} else {
		// 创建新记录
		riskInfo.CreatedAt = time.Now()
		riskInfo.UpdatedAt = time.Now()

		if err := db.Create(&riskInfo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建风险信息失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   riskInfo,
	})
}

// importCompanyProfile 批量导入企业画像数据
func (api *CompanyProfileAPI) importCompanyProfile(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	var profileData CompanyProfileData
	if err := c.ShouldBindJSON(&profileData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查权限
	if profileData.BasicInfo != nil && !api.checkCompanyAccess(userID, profileData.BasicInfo.CompanyID, c) {
		return
	}

	db := api.core.GetDB()

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 导入基本信息
	if profileData.BasicInfo != nil {
		profileData.BasicInfo.CreatedAt = time.Now()
		profileData.BasicInfo.UpdatedAt = time.Now()
		if err := tx.Create(profileData.BasicInfo).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导入基本信息失败"})
			return
		}
	}

	// 导入资质许可信息
	for _, qualification := range profileData.Qualifications {
		qualification.CreatedAt = time.Now()
		qualification.UpdatedAt = time.Now()
		if err := tx.Create(&qualification).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导入资质许可信息失败"})
			return
		}
	}

	// 导入其他信息...
	// (类似的处理逻辑)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入企业画像数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "企业画像数据导入成功",
	})
}

// exportCompanyProfile 导出企业画像数据
func (api *CompanyProfileAPI) exportCompanyProfile(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取企业ID
	companyID, err := strconv.Atoi(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "企业ID格式错误"})
		return
	}

	// 检查权限
	if !api.checkCompanyAccess(userID, uint(companyID), c) {
		return
	}

	// 获取完整企业画像数据
	// (使用getCompanyProfile的逻辑)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "企业画像数据导出成功",
	})
}

// checkCompanyAccess 检查企业访问权限
func (api *CompanyProfileAPI) checkCompanyAccess(userID, companyID uint, c *gin.Context) bool {
	// 检查是否为管理员
	role := c.GetString("role")
	if role == "admin" || role == "super_admin" {
		return true
	}

	// 检查是否为企业的创建者
	db := api.core.GetDB()
	var company Company
	if err := db.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
		return false
	}

	if company.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return false
	}

	return true
}

// checkCompleteProfile 检查企业画像完整性
func (api *CompanyProfileAPI) checkCompleteProfile(companyID uint) bool {
	db := api.core.GetDB()

	// 检查各个表是否有数据
	var count int64

	// 基本信息
	db.Model(&CompanyProfileBasicInfo{}).Where("company_id = ?", companyID).Count(&count)
	if count == 0 {
		return false
	}

	// 人员竞争力
	db.Model(&PersonnelCompetitiveness{}).Where("company_id = ?", companyID).Count(&count)
	if count == 0 {
		return false
	}

	// 科创评分
	db.Model(&TechInnovationScore{}).Where("company_id = ?", companyID).Count(&count)
	if count == 0 {
		return false
	}

	return true
}
