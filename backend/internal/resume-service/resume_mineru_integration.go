package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==============================================
// Resume服务MinerU集成处理 - 简历文档解析
// ==============================================

// ResumeMinerUIntegration Resume服务MinerU集成服务
type ResumeMinerUIntegration struct {
	baseURL string
	client  *http.Client
}

// NewResumeMinerUIntegration 创建Resume服务MinerU集成服务
func NewResumeMinerUIntegration() *ResumeMinerUIntegration {
	return &ResumeMinerUIntegration{
		baseURL: "http://localhost:8621", // 更新为正确的MinerU端口
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ResumeMinerUParseRequest Resume服务MinerU解析请求
type ResumeMinerUParseRequest struct {
	FilePath     string `json:"file_path"`
	UserID       uint   `json:"user_id"`
	TaskID       uint   `json:"task_id"`
	BusinessType string `json:"business_type"`
}

// ResumeMinerUParseResponse Resume服务MinerU解析响应
type ResumeMinerUParseResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
}

// ResumeParsingTask 简历解析任务
type ResumeParsingTask struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	ResumeID     uint       `json:"resume_id" gorm:"not null"`
	FileID       uint       `json:"file_id" gorm:"not null"`
	TaskType     string     `json:"task_type" gorm:"not null"`
	Status       string     `json:"status" gorm:"default:'pending'"`
	Progress     int        `json:"progress" gorm:"default:0"`
	ErrorMessage string     `json:"error_message" gorm:"type:text"`
	ResultData   string     `json:"result_data" gorm:"type:json"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ResumeStructuredDataRecord 简历结构化数据记录
type ResumeStructuredDataRecord struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ResumeID       uint      `json:"resume_id" gorm:"not null"`
	TaskID         uint      `json:"task_id" gorm:"not null"`
	BasicInfo      string    `json:"basic_info" gorm:"type:text"`
	EducationInfo  string    `json:"education_info" gorm:"type:text"`
	WorkExperience string    `json:"work_experience" gorm:"type:text"`
	SkillsInfo     string    `json:"skills_info" gorm:"type:text"`
	ProjectInfo    string    `json:"project_info" gorm:"type:text"`
	Confidence     float64   `json:"confidence" gorm:"default:0.0"`
	ParsingVersion string    `json:"parsing_version" gorm:"default:'mineru-v1.0'"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// handleResumeDocumentUploadWithMinerU 使用MinerU处理简历文档上传
func handleResumeDocumentUploadWithMinerU(c *gin.Context, db *gorm.DB, userID uint) {
	// 1. 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 2. 验证文件类型
	fileType := getResumeFileType(header.Filename)
	if !isValidResumeFileType(fileType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型，仅支持PDF、DOC、DOCX格式"})
		return
	}

	// 3. 保存文件到磁盘
	filePath, err := saveResumeUploadedFile(file, header, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败: " + err.Error()})
		return
	}

	// 4. 在MySQL中创建元数据记录
	resume, err := createResumeMetadata(db, userID, header, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建简历元数据失败: " + err.Error()})
		return
	}

	// 5. 创建MinerU解析任务
	taskID, err := createResumeMinerUParsingTask(db, resume.ID, userID, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建解析任务失败: " + err.Error()})
		return
	}

	// 6. 异步启动MinerU解析
	go func() {
		if err := processResumeWithMinerU(db, taskID, userID, filePath); err != nil {
			log.Printf("Resume MinerU解析失败: %v", err)
		}
	}()

	// 7. 返回响应
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "简历上传成功，解析任务已启动",
		"resume_id":  resume.ID,
		"task_id":    taskID,
		"file_path":  filePath,
		"status":     "processing",
		"created_at": time.Now().Format(time.RFC3339),
	})
}

// createResumeMetadata 创建简历元数据记录
func createResumeMetadata(db *gorm.DB, userID uint, header *multipart.FileHeader, filePath string) (*ResumeMetadata, error) {

	resume := &ResumeMetadata{
		UserID:        userID,
		Title:         header.Filename,
		CreationMode:  "upload",
		Status:        "draft",
		ParsingStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.Create(resume).Error; err != nil {
		return nil, fmt.Errorf("创建简历元数据失败: %v", err)
	}

	log.Printf("✅ 简历元数据创建成功: ID=%d, UserID=%d, Title=%s", resume.ID, userID, header.Filename)
	return resume, nil
}

// createResumeMinerUParsingTask 创建Resume MinerU解析任务
func createResumeMinerUParsingTask(db *gorm.DB, resumeID uint, userID uint, filePath string) (uint, error) {

	task := &ResumeParsingTask{
		ResumeID:   resumeID,
		FileID:     1, // 使用虚拟的file_id，因为当前没有文件表
		TaskType:   "mineru_parse",
		Status:     "pending",
		Progress:   0,
		ResultData: "{}", // 初始化为空的JSON对象
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := db.Create(task).Error; err != nil {
		return 0, fmt.Errorf("创建解析任务失败: %v", err)
	}

	log.Printf("✅ Resume MinerU解析任务创建成功: TaskID=%d, ResumeID=%d, UserID=%d", task.ID, resumeID, userID)
	return task.ID, nil
}

// processResumeWithMinerU 使用MinerU处理简历解析
func processResumeWithMinerU(db *gorm.DB, taskID uint, userID uint, filePath string) error {
	log.Printf("开始Resume MinerU解析: taskID=%d, userID=%d, filePath=%s", taskID, userID, filePath)

	// 更新任务状态为处理中
	updateResumeMinerUParsingTaskStatus(db, taskID, "processing", "")

	// 创建MinerU集成服务
	mineruIntegration := NewResumeMinerUIntegration()

	// 调用MinerU服务进行解析
	result, err := mineruIntegration.parseResumeDocument(filePath, userID)
	if err != nil {
		updateResumeMinerUParsingTaskStatus(db, taskID, "failed", err.Error())
		return fmt.Errorf("MinerU解析失败: %v", err)
	}

	// 检查响应状态
	if status, ok := result["status"].(string); !ok || status != "success" {
		updateResumeMinerUParsingTaskStatus(db, taskID, "failed", "MinerU解析失败")
		return fmt.Errorf("MinerU解析失败: %v", result)
	}

	// 获取解析结果
	parseResult, ok := result["result"].(map[string]interface{})
	if !ok {
		updateResumeMinerUParsingTaskStatus(db, taskID, "failed", "MinerU响应格式错误")
		return fmt.Errorf("MinerU响应格式错误: 缺少result字段")
	}

	// 保存解析结果到简历数据库
	err = saveResumeMinerUResultToDatabase(db, taskID, userID, parseResult)
	if err != nil {
		updateResumeMinerUParsingTaskStatus(db, taskID, "failed", err.Error())
		return fmt.Errorf("保存解析结果失败: %v", err)
	}

	// 更新任务状态为完成
	updateResumeMinerUParsingTaskStatus(db, taskID, "completed", "")
	log.Printf("✅ Resume MinerU解析完成: taskID=%d, userID=%d", taskID, userID)
	return nil
}

// parseResumeDocument 调用MinerU服务解析简历文档
func (r *ResumeMinerUIntegration) parseResumeDocument(filePath string, userID uint) (map[string]interface{}, error) {
	// 构建请求数据
	requestData := ResumeMinerUParseRequest{
		FilePath:     filePath,
		UserID:       userID,
		BusinessType: "resume", // 简历业务类型
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送请求到MinerU服务
	url := fmt.Sprintf("%s/api/v1/parse/document", r.baseURL)
	resp, err := r.client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("发送请求到MinerU服务失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MinerU服务返回错误: %s (状态码: %d)", response["error"], resp.StatusCode)
	}

	return response, nil
}

// saveResumeMinerUResultToDatabase 将MinerU解析结果保存到简历数据库
func saveResumeMinerUResultToDatabase(db *gorm.DB, taskID uint, userID uint, data map[string]interface{}) error {

	// 获取任务信息
	var task ResumeParsingTask
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
	now := time.Now()
	task.CompletedAt = &now
	task.UpdatedAt = time.Now()

	if err := db.Save(&task).Error; err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 更新resume_metadata表的解析状态和结果
	var resume ResumeMetadata
	if err := db.First(&resume, task.ResumeID).Error; err != nil {
		return fmt.Errorf("获取简历信息失败: %v", err)
	}

	resume.ParsingStatus = "completed"
	resume.ParsedData = string(resultJSON)
	resume.UpdatedAt = time.Now()

	if err := db.Save(&resume).Error; err != nil {
		return fmt.Errorf("更新简历解析状态失败: %v", err)
	}

	// 从MinerU解析结果中提取结构化数据
	basicInfo, educationInfo, workExperience, skillsInfo, projectInfo := extractResumeStructuredData(data)

	// 将结构体转换为JSON字符串
	basicInfoJSON, _ := json.Marshal(basicInfo)
	educationInfoJSON, _ := json.Marshal(educationInfo)
	workExperienceJSON, _ := json.Marshal(workExperience)
	skillsInfoJSON, _ := json.Marshal(skillsInfo)
	projectInfoJSON, _ := json.Marshal(projectInfo)

	// 创建结构化数据记录
	structuredData := ResumeStructuredDataRecord{
		ResumeID:       task.ResumeID,
		TaskID:         taskID,
		BasicInfo:      string(basicInfoJSON),
		EducationInfo:  string(educationInfoJSON),
		WorkExperience: string(workExperienceJSON),
		SkillsInfo:     string(skillsInfoJSON),
		ProjectInfo:    string(projectInfoJSON),
		Confidence:     extractConfidenceFromResumeData(data),
		ParsingVersion: "mineru-v1.0",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.Create(&structuredData).Error; err != nil {
		log.Printf("警告: 保存结构化数据失败: %v", err)
		// 即使结构化数据保存失败，任务状态已经更新，所以不返回错误
	}

	log.Printf("✅ Resume MinerU解析结果已成功保存: taskID=%d, userID=%d", taskID, userID)
	return nil
}

// extractResumeStructuredData 从MinerU解析结果中提取简历结构化数据
func extractResumeStructuredData(data map[string]interface{}) (map[string]interface{}, map[string]interface{}, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	// 基本信息
	basicInfo := map[string]interface{}{
		"name":        extractStringFromData(data, "name", "姓名"),
		"phone":       extractStringFromData(data, "phone", "电话"),
		"email":       extractStringFromData(data, "email", "邮箱"),
		"location":    extractStringFromData(data, "location", "地址"),
		"birth_date":  extractStringFromData(data, "birth_date", "出生日期"),
		"gender":      extractStringFromData(data, "gender", "性别"),
		"nationality": extractStringFromData(data, "nationality", "国籍"),
	}

	// 教育信息
	educationInfo := map[string]interface{}{
		"degree":     extractStringFromData(data, "degree", "学历"),
		"school":     extractStringFromData(data, "school", "学校"),
		"major":      extractStringFromData(data, "major", "专业"),
		"graduation": extractStringFromData(data, "graduation", "毕业时间"),
		"gpa":        extractStringFromData(data, "gpa", "GPA"),
	}

	// 工作经历
	workExperience := map[string]interface{}{
		"companies":    extractArrayFromData(data, "companies", "工作经历"),
		"positions":    extractArrayFromData(data, "positions", "职位"),
		"durations":    extractArrayFromData(data, "durations", "工作时间"),
		"descriptions": extractArrayFromData(data, "descriptions", "工作描述"),
	}

	// 技能信息
	skillsInfo := map[string]interface{}{
		"technical_skills": extractArrayFromData(data, "technical_skills", "技术技能"),
		"languages":        extractArrayFromData(data, "languages", "语言技能"),
		"certifications":   extractArrayFromData(data, "certifications", "证书"),
		"soft_skills":      extractArrayFromData(data, "soft_skills", "软技能"),
	}

	// 项目信息
	projectInfo := map[string]interface{}{
		"projects":     extractArrayFromData(data, "projects", "项目经历"),
		"technologies": extractArrayFromData(data, "technologies", "使用技术"),
		"roles":        extractArrayFromData(data, "roles", "项目角色"),
		"achievements": extractArrayFromData(data, "achievements", "项目成果"),
	}

	return basicInfo, educationInfo, workExperience, skillsInfo, projectInfo
}

// extractStringFromData 从数据中提取字符串值
func extractStringFromData(data map[string]interface{}, key string, fallback string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return fallback
}

// extractArrayFromData 从数据中提取数组值
func extractArrayFromData(data map[string]interface{}, key string, fallback string) []interface{} {
	if value, ok := data[key]; ok {
		if arr, ok := value.([]interface{}); ok {
			return arr
		}
	}
	return []interface{}{fallback}
}

// extractConfidenceFromResumeData 从简历数据中提取置信度
func extractConfidenceFromResumeData(data map[string]interface{}) float64 {
	if confidence, ok := data["confidence"]; ok {
		if conf, ok := confidence.(float64); ok {
			return conf
		}
	}
	return 0.85 // 默认置信度
}

// updateResumeMinerUParsingTaskStatus 更新Resume MinerU解析任务状态
func updateResumeMinerUParsingTaskStatus(db *gorm.DB, taskID uint, status string, errorMsg string) {

	var task ResumeParsingTask
	if err := db.First(&task, taskID).Error; err != nil {
		log.Printf("更新任务状态失败: 找不到任务 %d", taskID)
		return
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	if status == "processing" {
		task.Progress = 50
		now := time.Now()
		task.StartedAt = &now
	} else if status == "completed" {
		task.Progress = 100
		now := time.Now()
		task.CompletedAt = &now
	} else if status == "failed" {
		task.ErrorMessage = errorMsg
		task.Progress = 0
	}

	if err := db.Save(&task).Error; err != nil {
		log.Printf("更新任务状态失败: %v", err)
	} else {
		log.Printf("✅ 任务状态已更新: TaskID=%d, Status=%s", taskID, status)
	}
}

// getResumeFileType 获取简历文件类型
func getResumeFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc":
		return "doc"
	case ".docx":
		return "docx"
	default:
		return "unknown"
	}
}

// isValidResumeFileType 验证简历文件类型是否有效
func isValidResumeFileType(fileType string) bool {
	validTypes := []string{"pdf", "doc", "docx"}
	for _, validType := range validTypes {
		if fileType == validType {
			return true
		}
	}
	return false
}

// saveResumeUploadedFile 保存上传的简历文件
func saveResumeUploadedFile(file multipart.File, header *multipart.FileHeader, userID uint) (string, error) {
	// 创建上传目录 - 保存到MinerU容器可以访问的位置（通过Docker卷映射）
	uploadDir := fmt.Sprintf("/Users/szjason72/zervi-basic/basic/backend/internal/resume/uploads/resumes/%d", userID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %v", err)
	}

	// 生成唯一文件名
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s", timestamp, header.Filename)
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

	// 返回容器内的路径（通过卷映射）
	containerPath := fmt.Sprintf("/app/resume_uploads/resumes/%d/%s", userID, filename)
	log.Printf("✅ 简历文件保存成功: %s (容器路径: %s)", filePath, containerPath)
	return containerPath, nil
}

// ResumeMetadata 简历元数据模型（与数据库表对应）
type ResumeMetadata struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"not null"`
	FileID        *int      `json:"file_id"`
	Title         string    `json:"title" gorm:"not null"`
	CreationMode  string    `json:"creation_mode" gorm:"default:'upload'"`
	Status        string    `json:"status" gorm:"default:'draft'"`
	ParsingStatus string    `json:"parsing_status" gorm:"default:'pending'"`
	ParsingError  string    `json:"parsing_error" gorm:"type:text"`
	SQLiteDBPath  string    `json:"sqlite_db_path" gorm:"column:sqlite_db_path"`
	ParsedData    string    `json:"parsed_data" gorm:"type:text"`
	TemplateID    *int      `json:"template_id"`
	IsPublic      bool      `json:"is_public" gorm:"default:false"`
	ViewCount     int       `json:"view_count" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ResumeMetadata) TableName() string {
	return "resume_metadata"
}

// ==============================================
// 缺失的关键业务逻辑函数
// ==============================================

// createResumeContent 在用户SQLite数据库中创建简历内容记录
func createResumeContent(userID uint, resumeMetadataID int, title, content string) (*ResumeContent, error) {
	// 获取用户SQLite数据库连接
	userDB, err := GetSecureUserDatabase(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户SQLite数据库失败: %v", err)
	}

	// 创建简历内容记录
	resumeContent := &ResumeContent{
		ResumeMetadataID: uint(resumeMetadataID),
		Title:            title,
		Content:          content,
		FileType:         "pdf", // 默认类型，后续可以从文件扩展名推断
		FileSize:         0,     // 后续可以从文件信息获取
		FilePath:         "",    // 后续可以从文件路径获取
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := userDB.Create(resumeContent).Error; err != nil {
		return nil, fmt.Errorf("创建简历内容记录失败: %v", err)
	}

	log.Printf("✅ 简历内容记录创建成功: ID=%d, UserID=%d, ResumeMetadataID=%d, Title=%s",
		resumeContent.ID, userID, resumeMetadataID, title)

	return resumeContent, nil
}

// saveOriginalResumeData 保存原始简历数据到用户SQLite数据库
func saveOriginalResumeData(userID uint, resumeMetadataID uint, filePath, filename string) error {
	// 获取用户SQLite数据库连接
	userDB, err := GetSecureUserDatabase(userID)
	if err != nil {
		return fmt.Errorf("获取用户SQLite数据库失败: %v", err)
	}

	// 读取文件内容
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件内容失败: %v", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 查找对应的简历内容记录
	var resumeContent ResumeContent
	if err := userDB.Where("resume_metadata_id = ?", resumeMetadataID).First(&resumeContent).Error; err != nil {
		// 如果找不到记录，创建一个新的
		resumeContent = ResumeContent{
			ResumeMetadataID: resumeMetadataID,
			Title:            filename,
			Content:          string(fileContent),
			FileType:         strings.ToLower(filepath.Ext(filename)),
			FileSize:         fileInfo.Size(),
			FilePath:         filePath,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := userDB.Create(&resumeContent).Error; err != nil {
			return fmt.Errorf("创建简历内容记录失败: %v", err)
		}
	} else {
		// 更新现有记录
		resumeContent.Content = string(fileContent)
		resumeContent.FileType = strings.ToLower(filepath.Ext(filename))
		resumeContent.FileSize = fileInfo.Size()
		resumeContent.FilePath = filePath
		resumeContent.UpdatedAt = time.Now()

		if err := userDB.Save(&resumeContent).Error; err != nil {
			return fmt.Errorf("更新简历内容记录失败: %v", err)
		}
	}

	log.Printf("✅ 原始简历数据保存成功: UserID=%d, ResumeMetadataID=%d, FileSize=%d, FileType=%s",
		userID, resumeMetadataID, fileInfo.Size(), resumeContent.FileType)

	return nil
}

// ==============================================
// SQLite数据库模型定义
// ==============================================

// ResumeContent SQLite中的简历内容模型
type ResumeContent struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ResumeMetadataID uint      `json:"resume_metadata_id" gorm:"not null"`
	Title            string    `json:"title" gorm:"not null"`
	Content          string    `json:"content" gorm:"type:text"`
	FileType         string    `json:"file_type" gorm:"size:20"`
	FileSize         int64     `json:"file_size"`
	FilePath         string    `json:"file_path" gorm:"size:500"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ResumeContent) TableName() string {
	return "resume_content"
}

// ParsedResumeDataDB SQLite中的解析结果模型
type ParsedResumeDataDB struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	ResumeContentID uint      `json:"resume_content_id" gorm:"not null"`
	PersonalInfo    []byte    `json:"personal_info" gorm:"type:blob"`
	WorkExperience  []byte    `json:"work_experience" gorm:"type:blob"`
	Education       []byte    `json:"education" gorm:"type:blob"`
	Skills          []byte    `json:"skills" gorm:"type:blob"`
	Projects        []byte    `json:"projects" gorm:"type:blob"`
	Certifications  []byte    `json:"certifications" gorm:"type:blob"`
	Keywords        []byte    `json:"keywords" gorm:"type:blob"`
	Confidence      float64   `json:"confidence" gorm:"default:0.0"`
	ParsingVersion  string    `json:"parsing_version" gorm:"size:50"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ParsedResumeDataDB) TableName() string {
	return "parsed_resume_data"
}
