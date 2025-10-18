package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
	"gorm.io/gorm"
)

// ==============================================
// Company服务MinerU集成处理 - 企业文档解析
// ==============================================

// CompanyMinerUIntegration Company服务MinerU集成服务
type CompanyMinerUIntegration struct {
	baseURL string
	client  *http.Client
}

// NewCompanyMinerUIntegration 创建Company服务MinerU集成服务
func NewCompanyMinerUIntegration() *CompanyMinerUIntegration {
	return &CompanyMinerUIntegration{
		baseURL: "http://localhost:8001",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CompanyMinerUParseRequest Company服务MinerU解析请求
type CompanyMinerUParseRequest struct {
	FilePath  string `json:"file_path"`
	CompanyID uint   `json:"company_id"`
	TaskID    uint   `json:"task_id"`
}

// CompanyMinerUParseResponse Company服务MinerU解析响应
type CompanyMinerUParseResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
}

// handleCompanyDocumentUploadWithMinerU 使用MinerU处理企业文档上传
func handleCompanyDocumentUploadWithMinerU(c *gin.Context, core *jobfirst.Core, companyID uint) {
	// 获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)
	// 1. 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 2. 验证文件类型
	fileType := getCompanyFileType(header.Filename)
	if !isValidCompanyFileType(fileType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型"})
		return
	}

	// 3. 保存文件到磁盘
	filePath, err := saveCompanyUploadedFile(file, header, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败: " + err.Error()})
		return
	}

	// 4. 在MySQL中创建元数据记录
	document, err := createCompanyDocumentMetadata(core, companyID, userID, header, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文档元数据失败: " + err.Error()})
		return
	}

	// 5. 创建MinerU解析任务
	taskID, err := createCompanyMinerUParsingTask(core, document.ID, companyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建解析任务失败: " + err.Error()})
		return
	}

	// 6. 异步调用MinerU解析
	go func() {
		err := callCompanyMinerUForParsing(filePath, taskID, companyID, userID, core)
		if err != nil {
			log.Printf("Company MinerU解析失败: %v", err)
			updateCompanyMinerUParsingTaskStatus(core, taskID, "failed", err.Error())
		}
	}()

	// 7. 立即返回任务ID
	c.JSON(http.StatusOK, gin.H{
		"task_id":      taskID,
		"document_id":  document.ID,
		"company_id":   companyID,
		"status":       "processing",
		"message":      "企业文档解析中，请稍后查询结果",
		"parsing_mode": "mineru",
	})
}

// createCompanyMinerUParsingTask 创建Company服务MinerU解析任务
func createCompanyMinerUParsingTask(core *jobfirst.Core, documentID, companyID, userID uint) (uint, error) {
	// 创建解析任务记录
	task := CompanyParsingTask{
		DocumentID:   documentID,
		CompanyID:    companyID,
		UserID:       userID, // 使用从JWT token获取的用户ID
		Status:       "pending",
		Progress:     0,
		ErrorMessage: "",
		ResultData:   "{}", // 初始化为空的JSON对象
		MineruTaskID: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	db := core.GetDB()
	if err := db.Create(&task).Error; err != nil {
		return 0, fmt.Errorf("创建解析任务失败: %v", err)
	}

	log.Printf("创建Company MinerU解析任务成功: taskID=%d, documentID=%d, companyID=%d", task.ID, documentID, companyID)
	return task.ID, nil
}

// callCompanyMinerUForParsing 调用MinerU解析企业文档服务
func callCompanyMinerUForParsing(filePath string, taskID uint, companyID uint, userID uint, core *jobfirst.Core) error {
	mineru := NewCompanyMinerUIntegration()

	// 调用MinerU服务 - 使用文件上传API
	// 首先读取文件内容
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 创建multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("创建文件字段失败: %v", err)
	}
	fileWriter.Write(fileContent)

	// 添加公司ID
	writer.WriteField("company_id", strconv.FormatUint(uint64(companyID), 10))

	// 添加用户ID
	writer.WriteField("user_id", strconv.FormatUint(uint64(userID), 10))

	writer.Close()

	// 调用MinerU上传解析API
	resp, err := mineru.client.Post(
		mineru.baseURL+"/api/v1/parse/upload",
		writer.FormDataContentType(),
		&buf,
	)
	if err != nil {
		return fmt.Errorf("调用MinerU服务失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取MinerU响应失败: %v", err)
	}

	// 解析响应 - MinerU返回的是直接的解析结果
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析MinerU响应失败: %v", err)
	}

	// 检查响应状态
	if status, ok := result["status"].(string); !ok || status != "success" {
		return fmt.Errorf("MinerU解析失败: %v", result)
	}

	// 获取解析结果
	parseResult, ok := result["result"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("MinerU响应格式错误: 缺少result字段")
	}

	// 保存解析结果到企业画像数据库
	err = saveCompanyMinerUResultToDatabase(core, taskID, companyID, parseResult)
	if err != nil {
		return fmt.Errorf("保存解析结果失败: %v", err)
	}

	// 更新任务状态为完成
	updateCompanyMinerUParsingTaskStatus(core, taskID, "completed", "")
	log.Printf("Company MinerU解析完成: taskID=%d, companyID=%d", taskID, companyID)
	return nil
}

// saveCompanyMinerUResultToDatabase 将MinerU解析结果保存到企业画像数据库
func saveCompanyMinerUResultToDatabase(core *jobfirst.Core, taskID uint, companyID uint, data map[string]interface{}) error {
	// 获取MySQL数据库连接
	db := core.GetDB()

	// 获取任务信息
	var task CompanyParsingTask
	if err := db.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("获取任务信息失败: %v", err)
	}

	// 将MinerU解析结果转换为JSON字符串
	resultJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化解析结果失败: %v", err)
	}

	// 更新MySQL中的任务状态和结果
	task.Status = "completed"
	task.Progress = 100
	task.ResultData = string(resultJSON)
	task.UpdatedAt = time.Now()

	if err := db.Save(&task).Error; err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 从MinerU解析结果中提取结构化数据
	basicInfo, businessInfo, organizationInfo, financialInfo := extractCompanyStructuredData(data)

	// 将结构体转换为JSON字符串
	basicInfoJSON, _ := json.Marshal(basicInfo)
	businessInfoJSON, _ := json.Marshal(businessInfo)
	organizationInfoJSON, _ := json.Marshal(organizationInfo)
	financialInfoJSON, _ := json.Marshal(financialInfo)

	// 创建结构化数据记录
	structuredData := CompanyStructuredDataRecord{
		CompanyID:        companyID,
		TaskID:           taskID,
		BasicInfo:        string(basicInfoJSON),
		BusinessInfo:     string(businessInfoJSON),
		OrganizationInfo: string(organizationInfoJSON),
		FinancialInfo:    string(financialInfoJSON),
		Confidence:       extractConfidenceFromCompanyData(data),
		ParsingVersion:   "mineru-v1.0",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(&structuredData).Error; err != nil {
		log.Printf("警告: 保存结构化数据失败: %v", err)
		// 即使结构化数据保存失败，任务状态已经更新，所以不返回错误
	}

	// 解析并保存到企业画像表
	err = parseAndSaveCompanyProfileData(db, companyID, data)
	if err != nil {
		log.Printf("警告: 保存企业画像数据失败: %v", err)
		// 即使企业画像数据保存失败，也不返回错误，因为基础数据已经保存
	}

	log.Printf("✅ Company MinerU解析结果已成功保存: taskID=%d, companyID=%d", taskID, companyID)
	return nil
}

// parseAndSaveCompanyProfileData 解析并保存企业画像数据
func parseAndSaveCompanyProfileData(db *gorm.DB, companyID uint, data map[string]interface{}) error {
	// 生成报告ID
	reportID := fmt.Sprintf("FSCR%s%06d", time.Now().Format("20060102150405"), companyID)

	// 解析基本信息
	if basicInfo, ok := data["basic_info"].(map[string]interface{}); ok {
		profileBasicInfo := CompanyProfileBasicInfo{
			CompanyID:               companyID,
			ReportID:                reportID,
			CompanyName:             getStringValue(basicInfo, "name"),
			UsedName:                getStringValue(basicInfo, "used_name"),
			UnifiedSocialCreditCode: getStringValue(basicInfo, "unified_social_credit_code"),
			LegalRepresentative:     getStringValue(basicInfo, "legal_representative"),
			BusinessStatus:          getStringValue(basicInfo, "business_status"),
			RegisteredCapital:       getFloatValue(basicInfo, "registered_capital"),
			Currency:                getStringValue(basicInfo, "currency"),
			InsuredCount:            getIntValue(basicInfo, "insured_count"),
			IndustryCategory:        getStringValue(basicInfo, "industry_category"),
			RegistrationAuthority:   getStringValue(basicInfo, "registration_authority"),
			BusinessScope:           getStringValue(basicInfo, "business_scope"),
			Tags:                    getJSONString(basicInfo, "tags"),
			DataSource:              "mineru_parsing",
			DataUpdateTime:          time.Now(),
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		}

		// 保存基本信息
		if err := db.Create(&profileBasicInfo).Error; err != nil {
			log.Printf("保存企业基本信息失败: %v", err)
		}
	}

	// 解析财务信息
	if financialInfo, ok := data["financial_info"].(map[string]interface{}); ok {
		profileFinancialInfo := CompanyProfileFinancialInfo{
			CompanyID:        companyID,
			ReportID:         reportID,
			AnnualRevenue:    getFloatValue(financialInfo, "annual_revenue"),
			NetProfit:        getFloatValue(financialInfo, "net_profit"),
			TotalAssets:      getFloatValue(financialInfo, "total_assets"),
			TotalLiabilities: getFloatValue(financialInfo, "total_liabilities"),
			Equity:           getFloatValue(financialInfo, "equity"),
			CashFlow:         getFloatValue(financialInfo, "cash_flow"),
			ROE:              getFloatValue(financialInfo, "roe"),
			ROA:              getFloatValue(financialInfo, "roa"),
			DebtRatio:        getFloatValue(financialInfo, "debt_ratio"),
			CurrentRatio:     getFloatValue(financialInfo, "current_ratio"),
			QuickRatio:       getFloatValue(financialInfo, "quick_ratio"),
			DataSource:       "mineru_parsing",
			DataUpdateTime:   time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		// 保存财务信息
		if err := db.Create(&profileFinancialInfo).Error; err != nil {
			log.Printf("保存企业财务信息失败: %v", err)
		}
	}

	// 解析风险信息
	if riskInfo, ok := data["risk_info"].(map[string]interface{}); ok {
		profileRiskInfo := CompanyProfileRiskInfo{
			CompanyID:        companyID,
			ReportID:         reportID,
			RiskLevel:        getStringValue(riskInfo, "risk_level"),
			RiskFactors:      getJSONString(riskInfo, "risk_factors"),
			CreditRating:     getStringValue(riskInfo, "credit_rating"),
			LegalDisputes:    getJSONString(riskInfo, "legal_disputes"),
			FinancialHealth:  getStringValue(riskInfo, "financial_health"),
			OperationalRisk:  getStringValue(riskInfo, "operational_risk"),
			MarketRisk:       getStringValue(riskInfo, "market_risk"),
			ComplianceStatus: getStringValue(riskInfo, "compliance_status"),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		// 保存风险信息
		if err := db.Create(&profileRiskInfo).Error; err != nil {
			log.Printf("保存企业风险信息失败: %v", err)
		}
	}

	return nil
}

// 辅助函数
func getStringValue(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatValue(data map[string]interface{}, key string) float64 {
	if value, exists := data[key]; exists {
		if f, ok := value.(float64); ok {
			return f
		}
	}
	return 0.0
}

func getIntValue(data map[string]interface{}, key string) int {
	if value, exists := data[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
		if f, ok := value.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func getJSONString(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists {
		if jsonStr, err := json.Marshal(value); err == nil {
			return string(jsonStr)
		}
	}
	return "{}"
}

func extractConfidenceFromCompanyData(data map[string]interface{}) float64 {
	if confidence, exists := data["confidence"]; exists {
		if confFloat, ok := confidence.(float64); ok {
			return confFloat
		}
	}
	return 0.95 // 默认置信度
}

// updateCompanyMinerUParsingTaskStatus 更新Company服务MinerU解析任务状态
func updateCompanyMinerUParsingTaskStatus(core *jobfirst.Core, taskID uint, status, errorMessage string) error {
	db := core.GetDB()

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	switch status {
	case "processing":
		updates["progress"] = 50
	case "completed":
		updates["progress"] = 100
	case "failed":
		updates["error_message"] = errorMessage
		updates["progress"] = 0
	}

	if err := db.Model(&CompanyParsingTask{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新解析任务状态失败: %v", err)
	}

	return nil
}

// CheckCompanyMinerUParsingStatusHandler 查询Company服务MinerU解析状态
func CheckCompanyMinerUParsingStatusHandler(c *gin.Context, core *jobfirst.Core) {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	// 查询解析任务状态
	var task CompanyParsingTask
	db := core.GetDB()
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "解析任务不存在"})
		return
	}

	// 构建响应
	response := gin.H{
		"task_id":       task.ID,
		"document_id":   task.DocumentID,
		"company_id":    task.CompanyID,
		"status":        task.Status,
		"progress":      task.Progress,
		"task_type":     "mineru_parsing",
		"error_message": task.ErrorMessage,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}

	// 如果任务完成，包含结果数据
	if task.Status == "completed" && task.ResultData != "" {
		var resultData map[string]interface{}
		if err := json.Unmarshal([]byte(task.ResultData), &resultData); err == nil {
			response["result_data"] = resultData
		}
	}

	// 如果任务失败，包含错误信息
	if task.Status == "failed" {
		response["error"] = task.ErrorMessage
	}

	c.JSON(http.StatusOK, response)
}

// GetCompanyMinerUParsedDataHandler 获取Company服务MinerU解析结果
func GetCompanyMinerUParsedDataHandler(c *gin.Context, core *jobfirst.Core) {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	// 验证任务是否存在
	var task CompanyParsingTask
	db := core.GetDB()
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "解析任务不存在"})
		return
	}

	// 检查任务状态
	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析任务尚未完成"})
		return
	}

	// 获取结构化数据
	var structuredData CompanyStructuredDataRecord
	if err := db.Where("task_id = ?", taskID).First(&structuredData).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "解析结果不存在"})
		return
	}

	// 解析结构化数据 - 从BasicInfo中获取
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(structuredData.BasicInfo), &parsedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析结果格式错误"})
		return
	}

	// 构建响应
	response := gin.H{
		"task_id":         task.ID,
		"document_id":     task.DocumentID,
		"company_id":      task.CompanyID,
		"status":          task.Status,
		"parsed_data":     parsedData,
		"confidence":      structuredData.Confidence,
		"parsing_version": structuredData.ParsingVersion,
		"created_at":      task.CreatedAt,
		"updated_at":      task.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// 文件处理辅助函数
func getCompanyFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "word"
	case ".txt":
		return "text"
	default:
		return "unknown"
	}
}

func isValidCompanyFileType(fileType string) bool {
	validTypes := []string{"pdf", "word", "text"}
	for _, validType := range validTypes {
		if fileType == validType {
			return true
		}
	}
	return false
}

func saveCompanyUploadedFile(file multipart.File, header *multipart.FileHeader, companyID uint) (string, error) {
	// 创建上传目录
	uploadDir := fmt.Sprintf("./uploads/companies/%d", companyID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %v", err)
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// 创建文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	return filePath, nil
}

func createCompanyDocumentMetadata(core *jobfirst.Core, companyID uint, userID uint, header *multipart.FileHeader, filePath string) (*CompanyDocument, error) {
	// 读取文件内容并转换为Base64
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %v", err)
	}

	// 将文件内容转换为Base64编码
	base64Content := base64.StdEncoding.EncodeToString(fileContent)

	document := CompanyDocument{
		CompanyID:    companyID,
		UserID:       userID, // 使用从JWT token获取的用户ID
		Title:        header.Filename,
		OriginalFile: header.Filename,
		FileContent:  base64Content, // 存储Base64编码的内容
		FileType:     getCompanyFileType(header.Filename),
		FileSize:     header.Size,
		UploadTime:   time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	db := core.GetDB()
	if err := db.Create(&document).Error; err != nil {
		return nil, fmt.Errorf("创建文档元数据失败: %v", err)
	}

	return &document, nil
}

// extractCompanyStructuredData 从MinerU解析结果中提取结构化数据
func extractCompanyStructuredData(data map[string]interface{}) (CompanyBasicInfo, CompanyBusinessInfo, CompanyOrganizationInfo, CompanyFinancialInfo) {
	// 提取基本信息
	basicInfo := CompanyBasicInfo{}
	if basicData, ok := data["basic_info"].(map[string]interface{}); ok {
		basicInfo.Name = getStringValue(basicData, "name")
		basicInfo.ShortName = getStringValue(basicData, "short_name")
		basicInfo.FoundedYear = getIntValue(basicData, "founded_year")
		basicInfo.CompanySize = getStringValue(basicData, "company_size")
		basicInfo.Industry = getStringValue(basicData, "industry")
		basicInfo.Location = getStringValue(basicData, "location")
		basicInfo.Website = getStringValue(basicData, "website")
	}

	// 提取业务信息
	businessInfo := CompanyBusinessInfo{}
	if businessData, ok := data["business_info"].(map[string]interface{}); ok {
		businessInfo.MainBusiness = getStringValue(businessData, "main_business")
		businessInfo.Products = getStringValue(businessData, "products")
		businessInfo.TargetCustomers = getStringValue(businessData, "target_customers")
		businessInfo.CompetitiveAdvantage = getStringValue(businessData, "competitive_advantage")
	}

	// 提取组织信息
	organizationInfo := CompanyOrganizationInfo{}
	if orgData, ok := data["organization_info"].(map[string]interface{}); ok {
		organizationInfo.OrganizationStructure = getStringValue(orgData, "organization_structure")
		organizationInfo.Departments = getStringValue(orgData, "departments")
		organizationInfo.PersonnelScale = getStringValue(orgData, "personnel_scale")
		organizationInfo.ManagementInfo = getStringValue(orgData, "management_info")
	}

	// 提取财务信息
	financialInfo := CompanyFinancialInfo{}
	if financialData, ok := data["financial_info"].(map[string]interface{}); ok {
		financialInfo.RegisteredCapital = getStringValue(financialData, "registered_capital")
		financialInfo.AnnualRevenue = getStringValue(financialData, "annual_revenue")
		financialInfo.FinancingStatus = getStringValue(financialData, "financing_status")
		financialInfo.ListingStatus = getStringValue(financialData, "listing_status")
	}

	return basicInfo, businessInfo, organizationInfo, financialInfo
}
