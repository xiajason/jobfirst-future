package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// CompanyDocumentParser 企业文档解析器
type CompanyDocumentParser struct {
	mineruClient *MinerUClient
}

// NewCompanyDocumentParser 创建企业文档解析器
func NewCompanyDocumentParser(mineruClient *MinerUClient) *CompanyDocumentParser {
	return &CompanyDocumentParser{
		mineruClient: mineruClient,
	}
}

// CompanyBasicInfo 企业基本信息
type CompanyBasicInfo struct {
	Name        string `json:"name"`
	ShortName   string `json:"short_name"`
	FoundedYear int    `json:"founded_year"`
	CompanySize string `json:"company_size"`
	Industry    string `json:"industry"`
	Location    string `json:"location"`
	Website     string `json:"website"`
}

// CompanyBusinessInfo 企业业务信息
type CompanyBusinessInfo struct {
	MainBusiness         string `json:"main_business"`
	Products             string `json:"products"`
	TargetCustomers      string `json:"target_customers"`
	CompetitiveAdvantage string `json:"competitive_advantage"`
}

// CompanyOrganizationInfo 企业组织信息
type CompanyOrganizationInfo struct {
	OrganizationStructure string `json:"organization_structure"`
	Departments           string `json:"departments"`
	PersonnelScale        string `json:"personnel_scale"`
	ManagementInfo        string `json:"management_info"`
}

// CompanyFinancialInfo 企业财务信息
type CompanyFinancialInfo struct {
	RegisteredCapital string `json:"registered_capital"`
	AnnualRevenue     string `json:"annual_revenue"`
	FinancingStatus   string `json:"financing_status"`
	ListingStatus     string `json:"listing_status"`
}

// CompanyStructuredData 企业结构化数据
type CompanyStructuredData struct {
	BasicInfo        CompanyBasicInfo        `json:"basic_info"`
	BusinessInfo     CompanyBusinessInfo     `json:"business_info"`
	OrganizationInfo CompanyOrganizationInfo `json:"organization_info"`
	FinancialInfo    CompanyFinancialInfo    `json:"financial_info"`
	Confidence       float64                 `json:"confidence"`
	ParsingVersion   string                  `json:"parsing_version"`
}

// ParseCompanyDocument 解析企业文档
func (p *CompanyDocumentParser) ParseCompanyDocument(filePath string, userID int) (*CompanyStructuredData, error) {
	// 调用MinerU服务解析文档
	documentInfo, err := p.mineruClient.UploadAndParse(filePath, userID)
	if err != nil {
		return nil, fmt.Errorf("MinerU解析失败: %v", err)
	}

	// 从解析结果中提取企业信息
	structuredData, err := p.extractCompanyInfo(documentInfo)
	if err != nil {
		return nil, fmt.Errorf("提取企业信息失败: %v", err)
	}

	return structuredData, nil
}

// extractCompanyInfo 从文档中提取企业信息
func (p *CompanyDocumentParser) extractCompanyInfo(documentInfo *CompanyDocumentInfo) (*CompanyStructuredData, error) {
	// 如果MinerU返回了结构化数据，直接使用
	if documentInfo.BusinessType == "company" {
		return p.extractFromStructuredData(documentInfo)
	}

	// 否则从文本内容中提取
	content := documentInfo.Content

	// 提取基本信息
	basicInfo := p.extractBasicInfo(content)

	// 提取业务信息
	businessInfo := p.extractBusinessInfo(content)

	// 提取组织信息
	organizationInfo := p.extractOrganizationInfo(content)

	// 提取财务信息
	financialInfo := p.extractFinancialInfo(content)

	// 计算置信度
	confidence := p.calculateConfidence(basicInfo, businessInfo, organizationInfo, financialInfo)

	return &CompanyStructuredData{
		BasicInfo:        basicInfo,
		BusinessInfo:     businessInfo,
		OrganizationInfo: organizationInfo,
		FinancialInfo:    financialInfo,
		Confidence:       confidence,
		ParsingVersion:   "mineru-v1.0",
	}, nil
}

// extractFromStructuredData 从MinerU返回的结构化数据中提取企业信息
func (p *CompanyDocumentParser) extractFromStructuredData(documentInfo *CompanyDocumentInfo) (*CompanyStructuredData, error) {
	// 从MinerU返回的企业画像数据中提取信息
	basicInfo := CompanyBasicInfo{
		Name:        documentInfo.CompanyName,
		Industry:    documentInfo.Industry,
		Location:    documentInfo.Location,
		FoundedYear: documentInfo.FoundedYear,
	}

	// 根据员工数量设置公司规模
	if documentInfo.EmployeeCount > 0 {
		if documentInfo.EmployeeCount < 50 {
			basicInfo.CompanySize = "小型企业"
		} else if documentInfo.EmployeeCount < 200 {
			basicInfo.CompanySize = "中型企业"
		} else {
			basicInfo.CompanySize = "大型企业"
		}
	}

	// 提取财务信息
	financialInfo := CompanyFinancialInfo{
		AnnualRevenue: documentInfo.Revenue,
	}

	// 使用MinerU返回的置信度
	confidence := documentInfo.Confidence
	if confidence == 0 {
		confidence = 0.88 // 默认置信度
	}

	return &CompanyStructuredData{
		BasicInfo:        basicInfo,
		BusinessInfo:     CompanyBusinessInfo{},     // 暂时为空，可以后续扩展
		OrganizationInfo: CompanyOrganizationInfo{}, // 暂时为空，可以后续扩展
		FinancialInfo:    financialInfo,
		Confidence:       confidence,
		ParsingVersion:   "mineru-v1.0",
	}, nil
}

// extractBasicInfo 提取基本信息
func (p *CompanyDocumentParser) extractBasicInfo(content string) CompanyBasicInfo {
	basicInfo := CompanyBasicInfo{}

	// 提取企业名称
	if name := p.extractField(content, []string{"公司名称", "企业名称", "公司", "企业"}); name != "" {
		basicInfo.Name = name
	}

	// 提取简称
	if shortName := p.extractField(content, []string{"简称", "公司简称", "企业简称"}); shortName != "" {
		basicInfo.ShortName = shortName
	}

	// 提取成立年份
	if year := p.extractYear(content); year > 0 {
		basicInfo.FoundedYear = year
	}

	// 提取公司规模
	if size := p.extractField(content, []string{"公司规模", "人员规模", "员工数量", "规模"}); size != "" {
		basicInfo.CompanySize = size
	}

	// 提取行业
	if industry := p.extractField(content, []string{"行业", "所属行业", "行业类别"}); industry != "" {
		basicInfo.Industry = industry
	}

	// 提取地址
	if location := p.extractField(content, []string{"地址", "公司地址", "注册地址", "办公地址"}); location != "" {
		basicInfo.Location = location
	}

	// 提取网站
	if website := p.extractWebsite(content); website != "" {
		basicInfo.Website = website
	}

	return basicInfo
}

// extractBusinessInfo 提取业务信息
func (p *CompanyDocumentParser) extractBusinessInfo(content string) CompanyBusinessInfo {
	businessInfo := CompanyBusinessInfo{}

	// 提取主营业务
	if mainBusiness := p.extractField(content, []string{"主营业务", "主要业务", "业务范围", "经营范围"}); mainBusiness != "" {
		businessInfo.MainBusiness = mainBusiness
	}

	// 提取产品服务
	if products := p.extractField(content, []string{"产品", "服务", "产品服务", "主要产品"}); products != "" {
		businessInfo.Products = products
	}

	// 提取目标客户
	if targetCustomers := p.extractField(content, []string{"目标客户", "客户群体", "服务对象"}); targetCustomers != "" {
		businessInfo.TargetCustomers = targetCustomers
	}

	// 提取竞争优势
	if competitiveAdvantage := p.extractField(content, []string{"竞争优势", "核心竞争力", "优势"}); competitiveAdvantage != "" {
		businessInfo.CompetitiveAdvantage = competitiveAdvantage
	}

	return businessInfo
}

// extractOrganizationInfo 提取组织信息
func (p *CompanyDocumentParser) extractOrganizationInfo(content string) CompanyOrganizationInfo {
	organizationInfo := CompanyOrganizationInfo{}

	// 提取组织架构
	if orgStructure := p.extractField(content, []string{"组织架构", "公司架构", "架构"}); orgStructure != "" {
		organizationInfo.OrganizationStructure = orgStructure
	}

	// 提取部门设置
	if departments := p.extractField(content, []string{"部门", "部门设置", "组织部门"}); departments != "" {
		organizationInfo.Departments = departments
	}

	// 提取人员规模
	if personnelScale := p.extractField(content, []string{"人员规模", "员工数量", "人员数量"}); personnelScale != "" {
		organizationInfo.PersonnelScale = personnelScale
	}

	// 提取管理层信息
	if managementInfo := p.extractField(content, []string{"管理层", "管理团队", "领导团队"}); managementInfo != "" {
		organizationInfo.ManagementInfo = managementInfo
	}

	return organizationInfo
}

// extractFinancialInfo 提取财务信息
func (p *CompanyDocumentParser) extractFinancialInfo(content string) CompanyFinancialInfo {
	financialInfo := CompanyFinancialInfo{}

	// 提取注册资本
	if registeredCapital := p.extractField(content, []string{"注册资本", "注册资金", "资本"}); registeredCapital != "" {
		financialInfo.RegisteredCapital = registeredCapital
	}

	// 提取年营业额
	if annualRevenue := p.extractField(content, []string{"年营业额", "营业收入", "年收入"}); annualRevenue != "" {
		financialInfo.AnnualRevenue = annualRevenue
	}

	// 提取融资情况
	if financingStatus := p.extractField(content, []string{"融资", "融资情况", "投资"}); financingStatus != "" {
		financialInfo.FinancingStatus = financingStatus
	}

	// 提取上市状态
	if listingStatus := p.extractField(content, []string{"上市", "上市状态", "挂牌"}); listingStatus != "" {
		financialInfo.ListingStatus = listingStatus
	}

	return financialInfo
}

// extractField 提取字段值
func (p *CompanyDocumentParser) extractField(content string, keywords []string) string {
	for _, keyword := range keywords {
		// 使用正则表达式匹配字段
		pattern := fmt.Sprintf(`%s[：:]\s*([^\n\r]+)`, keyword)
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

// extractYear 提取年份
func (p *CompanyDocumentParser) extractYear(content string) int {
	// 匹配成立年份
	pattern := `成立[于]?[：:]?\s*(\d{4})`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		year := 0
		fmt.Sscanf(matches[1], "%d", &year)
		if year > 1900 && year <= time.Now().Year() {
			return year
		}
	}
	return 0
}

// extractWebsite 提取网站
func (p *CompanyDocumentParser) extractWebsite(content string) string {
	// 匹配网站URL
	pattern := `(https?://[^\s]+)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// calculateConfidence 计算置信度
func (p *CompanyDocumentParser) calculateConfidence(basicInfo CompanyBasicInfo, businessInfo CompanyBusinessInfo, organizationInfo CompanyOrganizationInfo, financialInfo CompanyFinancialInfo) float64 {
	score := 0.0
	total := 0.0

	// 基本信息权重
	if basicInfo.Name != "" {
		score += 0.3
	}
	if basicInfo.Industry != "" {
		score += 0.2
	}
	if basicInfo.Location != "" {
		score += 0.1
	}
	if basicInfo.FoundedYear > 0 {
		score += 0.1
	}
	total += 0.7

	// 业务信息权重
	if businessInfo.MainBusiness != "" {
		score += 0.2
	}
	if businessInfo.Products != "" {
		score += 0.1
	}
	total += 0.3

	// 组织信息权重
	if organizationInfo.PersonnelScale != "" {
		score += 0.1
	}
	if organizationInfo.Departments != "" {
		score += 0.1
	}
	total += 0.2

	// 财务信息权重
	if financialInfo.RegisteredCapital != "" {
		score += 0.1
	}
	if financialInfo.AnnualRevenue != "" {
		score += 0.1
	}
	total += 0.2

	if total == 0 {
		return 0.0
	}

	return score / total
}
