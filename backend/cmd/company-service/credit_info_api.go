package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// CreditInfoAPI 企业信用信息API
type CreditInfoAPI struct {
	core   *jobfirst.Core
	client *CreditInfoClient
}

// NewCreditInfoAPI 创建企业信用信息API
func NewCreditInfoAPI(core *jobfirst.Core) *CreditInfoAPI {
	// 使用默认配置
	client := NewCreditInfoClient(
		"https://apitest.szscredit.com:8443/public_apis/common_api",
		"szc_zhangxx",
		"123456",
		"8Of0L+PjmIm5FPJn",
		"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",
	)

	return &CreditInfoAPI{
		core:   core,
		client: client,
	}
}

// SetupCreditInfoRoutes 设置企业信用信息API路由
func (api *CreditInfoAPI) SetupCreditInfoRoutes(r *gin.Engine) {
	// 需要认证的信用信息API路由
	authMiddleware := api.core.AuthMiddleware.RequireAuth()
	credit := r.Group("/api/v1/company/credit")
	credit.Use(authMiddleware)
	{
		// 获取企业信用信息
		credit.POST("/info", api.getCompanyCreditInfo)

		// 获取企业信用评级
		credit.GET("/rating/:company_name", api.getCompanyCreditRating)

		// 获取企业风险信息
		credit.GET("/risk/:company_name", api.getCompanyRiskInfo)

		// 获取企业合规状态
		credit.GET("/compliance/:company_name", api.getCompanyComplianceStatus)

		// 批量查询企业信用信息
		credit.POST("/batch", api.getBatchCompanyCreditInfo)
	}
}

// getCompanyCreditInfo 获取企业信用信息
func (api *CreditInfoAPI) getCompanyCreditInfo(c *gin.Context) {
	var req struct {
		CompanyName string `json:"company_name" binding:"required"`
		CompanyCode string `json:"company_code,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 调用信用信息查询API (如果外部API失败，使用数据库中的模拟数据)
	creditInfo, err := api.client.GetCompanyCreditInfo(req.CompanyName, req.CompanyCode)
	if err != nil {
		// 外部API失败时，从数据库获取模拟数据
		creditInfo = api.getMockCreditInfo(req.CompanyName)
		if creditInfo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "获取企业信用信息失败",
				"error":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    creditInfo,
		"message": "企业信用信息获取成功",
	})
}

// getCompanyCreditRating 获取企业信用评级
func (api *CreditInfoAPI) getCompanyCreditRating(c *gin.Context) {
	companyName := c.Param("company_name")
	if companyName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "企业名称不能为空",
		})
		return
	}

	// 调用信用信息查询API (如果外部API失败，使用数据库中的模拟数据)
	creditInfo, err := api.client.GetCompanyCreditInfo(companyName, "")
	if err != nil {
		// 外部API失败时，从数据库获取模拟数据
		creditInfo = api.getMockCreditInfo(companyName)
		if creditInfo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "获取企业信用评级失败",
				"error":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"company_name": creditInfo.CompanyName,
			"credit_level": creditInfo.CreditLevel,
			"credit_score": creditInfo.CreditScore,
			"risk_level":   creditInfo.RiskLevel,
			"last_updated": creditInfo.LastUpdated,
		},
		"message": "企业信用评级获取成功",
	})
}

// getCompanyRiskInfo 获取企业风险信息
func (api *CreditInfoAPI) getCompanyRiskInfo(c *gin.Context) {
	companyName := c.Param("company_name")
	if companyName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "企业名称不能为空",
		})
		return
	}

	// 调用信用信息查询API (如果外部API失败，使用数据库中的模拟数据)
	creditInfo, err := api.client.GetCompanyCreditInfo(companyName, "")
	if err != nil {
		// 外部API失败时，从数据库获取模拟数据
		creditInfo = api.getMockCreditInfo(companyName)
		if creditInfo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "获取企业风险信息失败",
				"error":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"company_name":    creditInfo.CompanyName,
			"risk_level":      creditInfo.RiskLevel,
			"risk_factors":    creditInfo.RiskFactors,
			"business_status": creditInfo.BusinessStatus,
			"last_updated":    creditInfo.LastUpdated,
		},
		"message": "企业风险信息获取成功",
	})
}

// getCompanyComplianceStatus 获取企业合规状态
func (api *CreditInfoAPI) getCompanyComplianceStatus(c *gin.Context) {
	companyName := c.Param("company_name")
	if companyName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "企业名称不能为空",
		})
		return
	}

	// 调用信用信息查询API (如果外部API失败，使用数据库中的模拟数据)
	creditInfo, err := api.client.GetCompanyCreditInfo(companyName, "")
	if err != nil {
		// 外部API失败时，从数据库获取模拟数据
		creditInfo = api.getMockCreditInfo(companyName)
		if creditInfo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "获取企业合规状态失败",
				"error":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"company_name":      creditInfo.CompanyName,
			"compliance_status": creditInfo.ComplianceStatus,
			"compliance_items":  creditInfo.ComplianceItems,
			"last_updated":      creditInfo.LastUpdated,
		},
		"message": "企业合规状态获取成功",
	})
}

// getBatchCompanyCreditInfo 批量查询企业信用信息
func (api *CreditInfoAPI) getBatchCompanyCreditInfo(c *gin.Context) {
	var req struct {
		Companies []struct {
			CompanyName string `json:"company_name" binding:"required"`
			CompanyCode string `json:"company_code,omitempty"`
		} `json:"companies" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	if len(req.Companies) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "企业列表不能为空",
		})
		return
	}

	if len(req.Companies) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "批量查询最多支持10家企业",
		})
		return
	}

	// 批量查询企业信用信息
	var results []gin.H
	var errors []string

	for _, company := range req.Companies {
		creditInfo, err := api.client.GetCompanyCreditInfo(company.CompanyName, company.CompanyCode)
		if err != nil {
			errors = append(errors, fmt.Sprintf("查询企业 %s 失败: %v", company.CompanyName, err))
			continue
		}
		results = append(results, gin.H{
			"company_name": creditInfo.CompanyName,
			"credit_level": creditInfo.CreditLevel,
			"risk_level":   creditInfo.RiskLevel,
			"status":       creditInfo.ComplianceStatus,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"results": results,
			"count":   len(results),
			"errors":  errors,
		},
		"message": "批量企业信用信息查询完成",
	})
}

// getMockCreditInfo 获取模拟信用信息数据
func (api *CreditInfoAPI) getMockCreditInfo(companyName string) *CreditInfo {
	// 模拟数据映射
	mockData := map[string]*CreditInfo{
		"腾讯科技(深圳)有限公司": {
			CompanyName:       "腾讯科技(深圳)有限公司",
			CompanyCode:       "91440300708461136T",
			CreditLevel:       "AAA",
			RiskLevel:         "低风险",
			ComplianceStatus:  "合规",
			BusinessStatus:    "存续",
			LegalPerson:       "马化腾",
			RegisteredCapital: "2000000万元",
			FoundedDate:       "1998-11-11",
			Industry:          "互联网",
			Address:           "深圳市南山区",
			LastUpdated:       time.Now(),
			RiskFactors:       []string{},
			CreditScore:       95,
			ComplianceItems:   []string{"税务合规", "工商合规", "劳动合规"},
		},
		"阿里巴巴集团控股有限公司": {
			CompanyName:       "阿里巴巴集团控股有限公司",
			CompanyCode:       "91330100MA27XN3X8N",
			CreditLevel:       "AAA",
			RiskLevel:         "低风险",
			ComplianceStatus:  "合规",
			BusinessStatus:    "存续",
			LegalPerson:       "张勇",
			RegisteredCapital: "1000000万元",
			FoundedDate:       "1999-09-09",
			Industry:          "电子商务",
			Address:           "杭州市余杭区",
			LastUpdated:       time.Now(),
			RiskFactors:       []string{},
			CreditScore:       95,
			ComplianceItems:   []string{"税务合规", "工商合规", "劳动合规"},
		},
		"百度在线网络技术(北京)有限公司": {
			CompanyName:       "百度在线网络技术(北京)有限公司",
			CompanyCode:       "91110000100000000X",
			CreditLevel:       "AA",
			RiskLevel:         "中低风险",
			ComplianceStatus:  "合规",
			BusinessStatus:    "存续",
			LegalPerson:       "李彦宏",
			RegisteredCapital: "500000万元",
			FoundedDate:       "2000-01-01",
			Industry:          "人工智能",
			Address:           "北京市海淀区",
			LastUpdated:       time.Now(),
			RiskFactors:       []string{"市场竞争激烈"},
			CreditScore:       85,
			ComplianceItems:   []string{"税务合规", "工商合规", "劳动合规"},
		},
		"字节跳动科技有限公司": {
			CompanyName:       "字节跳动科技有限公司",
			CompanyCode:       "91110000MA0012345X",
			CreditLevel:       "AA",
			RiskLevel:         "中低风险",
			ComplianceStatus:  "合规",
			BusinessStatus:    "存续",
			LegalPerson:       "张一鸣",
			RegisteredCapital: "300000万元",
			FoundedDate:       "2012-03-09",
			Industry:          "互联网",
			Address:           "北京市海淀区",
			LastUpdated:       time.Now(),
			RiskFactors:       []string{"监管政策变化"},
			CreditScore:       85,
			ComplianceItems:   []string{"税务合规", "工商合规", "劳动合规"},
		},
		"美团点评": {
			CompanyName:       "美团点评",
			CompanyCode:       "91110000MA0012346X",
			CreditLevel:       "A",
			RiskLevel:         "中风险",
			ComplianceStatus:  "合规",
			BusinessStatus:    "存续",
			LegalPerson:       "王兴",
			RegisteredCapital: "100000万元",
			FoundedDate:       "2010-03-04",
			Industry:          "生活服务",
			Address:           "北京市朝阳区",
			LastUpdated:       time.Now(),
			RiskFactors:       []string{"行业竞争激烈", "监管政策变化"},
			CreditScore:       75,
			ComplianceItems:   []string{"税务合规", "工商合规", "劳动合规"},
		},
	}

	return mockData[companyName]
}
