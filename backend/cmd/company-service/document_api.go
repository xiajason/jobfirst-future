package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// DocumentAPI 文档API处理器
type DocumentAPI struct {
	core            *jobfirst.Core
	mineruClient    *MinerUClient
	documentParser  *CompanyDocumentParser
	uploadDir       string
	quotaMiddleware *QuotaMiddleware
}

// NewDocumentAPI 创建文档API处理器
func NewDocumentAPI(core *jobfirst.Core) *DocumentAPI {
	mineruClient := NewMinerUClient("http://localhost:8001")
	documentParser := NewCompanyDocumentParser(mineruClient)
	quotaMiddleware := NewQuotaMiddleware(core.GetDB())

	return &DocumentAPI{
		core:            core,
		mineruClient:    mineruClient,
		documentParser:  documentParser,
		uploadDir:       "./uploads/company-documents",
		quotaMiddleware: quotaMiddleware,
	}
}

// UploadDocumentRequest 上传文档请求
type UploadDocumentRequest struct {
	CompanyID int    `json:"company_id" binding:"required"`
	Title     string `json:"title" binding:"required"`
}

// UploadDocumentResponse 上传文档响应
type UploadDocumentResponse struct {
	Status     string `json:"status"`
	DocumentID int    `json:"document_id"`
	Message    string `json:"message"`
	UploadTime string `json:"upload_time"`
}

// CompanyParseDocumentResponse 解析文档响应
type CompanyParseDocumentResponse struct {
	Status         string                 `json:"status"`
	TaskID         int                    `json:"task_id"`
	Message        string                 `json:"message"`
	StructuredData *CompanyStructuredData `json:"structured_data,omitempty"`
	QuotaInfo      interface{}            `json:"quota_info,omitempty"`
}

// DocumentStatusResponse 文档状态响应
type DocumentStatusResponse struct {
	Status         string                 `json:"status"`
	TaskID         int                    `json:"task_id"`
	Progress       int                    `json:"progress"`
	Message        string                 `json:"message"`
	StructuredData *CompanyStructuredData `json:"structured_data,omitempty"`
	Error          string                 `json:"error,omitempty"`
}

// SetupDocumentRoutes 设置文档相关路由
func (api *DocumentAPI) SetupDocumentRoutes(r *gin.Engine) {
	// 创建上传目录
	os.MkdirAll(api.uploadDir, 0755)

	// 需要认证的文档API
	authMiddleware := api.core.AuthMiddleware.RequireAuth()
	documents := r.Group("/api/v1/company/documents")
	documents.Use(authMiddleware)
	{
		// 上传文档
		documents.POST("/upload", api.uploadDocument)

		// 解析文档 - 添加AI配额检查
		documents.POST("/:id/parse",
			api.quotaMiddleware.QuotaCheckMiddleware("document_parsing"),
			api.quotaMiddleware.RecordUsageMiddleware("document_parsing", "company_document_parsing"),
			api.parseDocument)

		// 获取解析状态
		documents.GET("/:id/parse/status", api.getParseStatus)

		// 获取文档列表
		documents.GET("/", api.getDocumentList)

		// 获取文档详情
		documents.GET("/:id", api.getDocumentDetail)

		// 删除文档
		documents.DELETE("/:id", api.deleteDocument)

		// MinerU集成文档上传
		documents.POST("/upload-mineru", func(c *gin.Context) {
			// 获取用户信息
			_, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
				return
			}

			// 获取公司ID
			companyIDStr := c.PostForm("company_id")
			if companyIDStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "company_id不能为空"})
				return
			}
			companyID, err := strconv.ParseUint(companyIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的company_id"})
				return
			}

			handleCompanyDocumentUploadWithMinerU(c, api.core, uint(companyID))
		})

		// 检查MinerU解析状态
		documents.GET("/parsing-status/:task_id", func(c *gin.Context) {
			CheckCompanyMinerUParsingStatusHandler(c, api.core)
		})

		// 获取MinerU解析结果
		documents.GET("/parsed-data/:task_id", func(c *gin.Context) {
			GetCompanyMinerUParsedDataHandler(c, api.core)
		})
	}
}

// uploadDocument 上传文档
func (api *DocumentAPI) uploadDocument(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败: " + err.Error()})
		return
	}

	// 获取请求参数
	companyIDStr := c.PostForm("company_id")
	title := c.PostForm("title")

	if companyIDStr == "" || title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id和title不能为空"})
		return
	}

	companyID, err := strconv.Atoi(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id格式错误"})
		return
	}

	// 验证文件类型
	allowedTypes := map[string]bool{
		".pdf":  true,
		".docx": true,
		".doc":  true,
		".txt":  true,
	}

	ext := filepath.Ext(file.Filename)
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型，仅支持PDF、DOCX、DOC、TXT"})
		return
	}

	// 生成唯一文件名
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%d_%d_%s_%s", userID, companyID, timestamp, file.Filename)
	filePath := filepath.Join(api.uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	// 读取文件内容并转换为Base64
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件内容失败: " + err.Error()})
		return
	}

	fileContentBase64 := base64.StdEncoding.EncodeToString(fileContent)

	// 保存到数据库
	db := api.core.GetDB()
	document := CompanyDocument{
		CompanyID:    uint(companyID),
		UserID:       userID,
		Title:        title,
		OriginalFile: filePath,
		FileContent:  fileContentBase64,
		FileType:     ext[1:], // 去掉点号
		FileSize:     file.Size,
		UploadTime:   time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.Create(&document).Error; err != nil {
		// 删除已保存的文件
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文档记录失败: " + err.Error()})
		return
	}

	response := UploadDocumentResponse{
		Status:     "success",
		DocumentID: int(document.ID),
		Message:    "文档上传成功",
		UploadTime: document.UploadTime.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// parseDocument 解析文档
func (api *DocumentAPI) parseDocument(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取文档ID
	documentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档ID格式错误"})
		return
	}

	// 查询文档
	db := api.core.GetDB()
	var document CompanyDocument
	if err := db.First(&document, documentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 检查权限
	if document.UserID != userID {
		// 检查是否为管理员
		role := c.GetString("role")
		if role != "admin" && role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
	}

	// 检查是否已有解析任务
	var existingTask CompanyParsingTask
	if err := db.Where("document_id = ?", documentID).First(&existingTask).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "该文档已有解析任务"})
		return
	}

	// 创建解析任务
	task := CompanyParsingTask{
		CompanyID:  document.CompanyID,
		DocumentID: document.ID,
		UserID:     userID,
		Status:     "pending",
		Progress:   0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建解析任务失败: " + err.Error()})
		return
	}

	// 异步解析文档
	go api.asyncParseDocument(&task, &document)

	// 获取配额信息
	quotaInfo, _ := c.Get("quota_info")

	response := CompanyParseDocumentResponse{
		Status:  "success",
		TaskID:  int(task.ID),
		Message: "解析任务已创建，正在处理中",
	}

	// 添加配额信息到响应
	if quotaInfo != nil {
		response.QuotaInfo = quotaInfo
	}

	c.JSON(http.StatusOK, response)
}

// asyncParseDocument 异步解析文档
func (api *DocumentAPI) asyncParseDocument(task *CompanyParsingTask, document *CompanyDocument) {
	db := api.core.GetDB()

	// 更新任务状态为处理中
	task.Status = "processing"
	task.Progress = 10
	db.Save(task)

	// 解析文档
	structuredData, err := api.documentParser.ParseCompanyDocument(document.OriginalFile, int(document.UserID))
	if err != nil {
		// 更新任务状态为失败
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		task.Progress = 100
		db.Save(task)
		return
	}

	// 更新任务状态为完成
	task.Status = "completed"
	task.Progress = 100

	// 保存解析结果
	resultData, _ := json.Marshal(structuredData)
	task.ResultData = string(resultData)
	db.Save(task)

	// 将结构体转换为JSON字符串
	basicInfoJSON, _ := json.Marshal(structuredData.BasicInfo)
	businessInfoJSON, _ := json.Marshal(structuredData.BusinessInfo)
	organizationInfoJSON, _ := json.Marshal(structuredData.OrganizationInfo)
	financialInfoJSON, _ := json.Marshal(structuredData.FinancialInfo)

	// 保存结构化数据
	structuredDataRecord := CompanyStructuredDataRecord{
		CompanyID:        document.CompanyID,
		TaskID:           task.ID,
		BasicInfo:        string(basicInfoJSON),
		BusinessInfo:     string(businessInfoJSON),
		OrganizationInfo: string(organizationInfoJSON),
		FinancialInfo:    string(financialInfoJSON),
		Confidence:       structuredData.Confidence,
		ParsingVersion:   structuredData.ParsingVersion,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	db.Create(&structuredDataRecord)
}

// getParseStatus 获取解析状态
func (api *DocumentAPI) getParseStatus(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取文档ID
	documentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档ID格式错误"})
		return
	}

	// 查询解析任务
	db := api.core.GetDB()
	var task CompanyParsingTask
	if err := db.Where("document_id = ?", documentID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "解析任务不存在"})
		return
	}

	// 检查权限
	if task.UserID != userID {
		// 检查是否为管理员
		role := c.GetString("role")
		if role != "admin" && role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
	}

	response := DocumentStatusResponse{
		Status:   task.Status,
		TaskID:   int(task.ID),
		Progress: task.Progress,
		Message:  getStatusMessage(task.Status),
		Error:    task.ErrorMessage,
	}

	// 如果解析完成，返回结构化数据
	if task.Status == "completed" && task.ResultData != "" {
		var structuredData CompanyStructuredData
		if err := json.Unmarshal([]byte(task.ResultData), &structuredData); err == nil {
			response.StructuredData = &structuredData
		}
	}

	c.JSON(http.StatusOK, response)
}

// getDocumentList 获取文档列表
func (api *DocumentAPI) getDocumentList(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取查询参数
	companyID := c.Query("company_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 构建查询
	db := api.core.GetDB()
	query := db.Model(&CompanyDocument{}).Where("user_id = ?", userID)

	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var documents []CompanyDocument
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文档列表失败"})
		return
	}

	var total int64
	query.Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"documents": documents,
			"total":     total,
			"page":      page,
			"size":      pageSize,
		},
	})
}

// getDocumentDetail 获取文档详情
func (api *DocumentAPI) getDocumentDetail(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取文档ID
	documentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档ID格式错误"})
		return
	}

	// 查询文档
	db := api.core.GetDB()
	var document CompanyDocument
	if err := db.First(&document, documentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 检查权限
	if document.UserID != userID {
		// 检查是否为管理员
		role := c.GetString("role")
		if role != "admin" && role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   document,
	})
}

// deleteDocument 删除文档
func (api *DocumentAPI) deleteDocument(c *gin.Context) {
	// 获取用户信息
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
		return
	}
	userID := userIDInterface.(uint)

	// 获取文档ID
	documentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档ID格式错误"})
		return
	}

	// 查询文档
	db := api.core.GetDB()
	var document CompanyDocument
	if err := db.First(&document, documentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 检查权限
	if document.UserID != userID {
		// 检查是否为管理员
		role := c.GetString("role")
		if role != "admin" && role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
	}

	// 删除文件
	if err := os.Remove(document.OriginalFile); err != nil {
		// 记录错误但不阻止删除数据库记录
		fmt.Printf("删除文件失败: %v\n", err)
	}

	// 删除数据库记录
	if err := db.Delete(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文档失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "文档删除成功"})
}

// getStatusMessage 获取状态消息
func getStatusMessage(status string) string {
	switch status {
	case "pending":
		return "等待解析"
	case "processing":
		return "正在解析"
	case "completed":
		return "解析完成"
	case "failed":
		return "解析失败"
	default:
		return "未知状态"
	}
}

// 数据模型定义
type CompanyDocument struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	CompanyID    uint      `json:"company_id" gorm:"not null"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	Title        string    `json:"title" gorm:"size:255;not null"`
	OriginalFile string    `json:"original_file" gorm:"type:text;not null"`
	FileContent  string    `json:"file_content" gorm:"type:longtext;not null"`
	FileType     string    `json:"file_type" gorm:"size:50;not null"`
	FileSize     int64     `json:"file_size" gorm:"not null"`
	UploadTime   time.Time `json:"upload_time" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CompanyParsingTask struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	CompanyID    uint      `json:"company_id" gorm:"not null"`
	DocumentID   uint      `json:"document_id" gorm:"not null"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	Status       string    `json:"status" gorm:"size:20;default:pending"`
	Progress     int       `json:"progress" gorm:"default:0"`
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	ResultData   string    `json:"result_data" gorm:"type:text"`
	MineruTaskID string    `json:"mineru_task_id" gorm:"size:100"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CompanyStructuredDataRecord struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	CompanyID        uint      `json:"company_id" gorm:"not null"`
	TaskID           uint      `json:"task_id" gorm:"not null"`
	BasicInfo        string    `json:"basic_info" gorm:"type:json"`
	BusinessInfo     string    `json:"business_info" gorm:"type:json"`
	OrganizationInfo string    `json:"organization_info" gorm:"type:json"`
	FinancialInfo    string    `json:"financial_info" gorm:"type:json"`
	Confidence       float64   `json:"confidence"`
	ParsingVersion   string    `json:"parsing_version" gorm:"size:50;default:mineru-v1.0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CompanyStructuredDataRecord) TableName() string {
	return "company_structured_data"
}
