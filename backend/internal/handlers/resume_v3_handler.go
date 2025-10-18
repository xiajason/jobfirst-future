package handlers

import (
	"fmt"
	"github.com/xiajason/zervi-basic/basic/backend/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResumeV3Handler V3.0简历处理器
type ResumeV3Handler struct {
	db *gorm.DB
}

// NewResumeV3Handler 创建V3.0简历处理器
func NewResumeV3Handler(db *gorm.DB) *ResumeV3Handler {
	return &ResumeV3Handler{db: db}
}

// GetResumes 获取简历列表
func (h *ResumeV3Handler) GetResumes(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")
	search := c.Query("search")

	offset := (page - 1) * size
	query := h.db.Model(&models.ResumeV3{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("title LIKE ? OR summary LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var resumes []models.ResumeV3
	var total int64

	query.Count(&total)
	err := query.Preload("Template").
		Preload("Skills.Skill").
		Preload("WorkExperiences.Company").
		Preload("WorkExperiences.Position").
		Preload("Projects.Company").
		Preload("Educations").
		Preload("Certifications").
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&resumes).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取简历列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": models.ResumeListResponse{
			Resumes: resumes,
			Total:   total,
			Page:    page,
			Size:    size,
		},
	})
}

// GetResume 获取简历详情
func (h *ResumeV3Handler) GetResume(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	resumeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的简历ID"})
		return
	}

	var resume models.ResumeV3
	err = h.db.Preload("Template").
		Preload("Skills.Skill").
		Preload("WorkExperiences.Company").
		Preload("WorkExperiences.Position").
		Preload("Projects.Company").
		Preload("Educations").
		Preload("Certifications").
		Preload("Comments.User").
		Preload("Comments.Replies.User").
		Where("id = ? AND user_id = ?", resumeID, userID).
		First(&resume).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "简历不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取简历失败", "error": err.Error()})
		}
		return
	}

	// 增加浏览次数
	h.db.Model(&resume).Update("view_count", gorm.Expr("view_count + 1"))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    resume,
	})
}

// CreateResume 创建简历
func (h *ResumeV3Handler) CreateResume(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	var req models.CreateResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	// 生成UUID和Slug
	uuid := generateUUID()
	slug := generateSlug(req.Title)

	resume := models.ResumeV3{
		UUID:       uuid,
		UserID:     userID,
		Title:      req.Title,
		Slug:       slug,
		TemplateID: req.TemplateID,
		Content:    req.Content,
		Status:     "draft",
		Visibility: req.Visibility,
		CanComment: true,
	}

	if req.Visibility == "" {
		resume.Visibility = "private"
	}

	err := h.db.Create(&resume).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建简历失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "简历创建成功",
		"data":    resume,
	})
}

// GetSkills 获取技能列表
func (h *ResumeV3Handler) GetSkills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	category := c.Query("category")
	search := c.Query("search")
	popular := c.Query("popular")

	offset := (page - 1) * size
	query := h.db.Model(&models.Skill{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if popular == "true" {
		query = query.Where("is_popular = ?", true)
	}

	var skills []models.Skill
	var total int64

	query.Count(&total)
	err := query.Offset(offset).Limit(size).Order("search_count DESC, name ASC").Find(&skills).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取技能列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": models.SkillListResponse{
			Skills: skills,
			Total:  total,
			Page:   page,
			Size:   size,
		},
	})
}

// GetCompanies 获取公司列表
func (h *ResumeV3Handler) GetCompanies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	industry := c.Query("industry")
	search := c.Query("search")
	verified := c.Query("verified")

	offset := (page - 1) * size
	query := h.db.Model(&models.Company{})

	if industry != "" {
		query = query.Where("industry = ?", industry)
	}

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if verified == "true" {
		query = query.Where("is_verified = ?", true)
	}

	var companies []models.Company
	var total int64

	query.Count(&total)
	err := query.Offset(offset).Limit(size).Order("is_verified DESC, name ASC").Find(&companies).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取公司列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": models.CompanyListResponse{
			Companies: companies,
			Total:     total,
			Page:      page,
			Size:      size,
		},
	})
}

// GetPositions 获取职位列表
func (h *ResumeV3Handler) GetPositions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	category := c.Query("category")
	level := c.Query("level")
	search := c.Query("search")

	offset := (page - 1) * size
	query := h.db.Model(&models.Position{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if level != "" {
		query = query.Where("level = ?", level)
	}

	if search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}

	var positions []models.Position
	var total int64

	query.Count(&total)
	err := query.Offset(offset).Limit(size).Order("title ASC").Find(&positions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取职位列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": models.PositionListResponse{
			Positions: positions,
			Total:     total,
			Page:      page,
			Size:      size,
		},
	})
}

// ==================== 辅助函数 ====================

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(c *gin.Context) uint {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// generateUUID 生成UUID
func generateUUID() string {
	return uuid.New().String()
}

// generateSlug 生成URL友好的slug
func generateSlug(title string) string {
	// 简单的slug生成逻辑
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "，", "-")
	slug = strings.ReplaceAll(slug, "。", "-")
	slug = strings.ReplaceAll(slug, "、", "-")
	slug = strings.ReplaceAll(slug, "（", "-")
	slug = strings.ReplaceAll(slug, "）", "-")
	slug = strings.ReplaceAll(slug, "(", "-")
	slug = strings.ReplaceAll(slug, ")", "-")
	slug = strings.ReplaceAll(slug, "&", "and")
	slug = strings.ReplaceAll(slug, "@", "at")
	slug = strings.ReplaceAll(slug, "#", "hash")
	slug = strings.ReplaceAll(slug, "%", "percent")
	slug = strings.ReplaceAll(slug, "+", "plus")
	slug = strings.ReplaceAll(slug, "=", "equals")
	slug = strings.ReplaceAll(slug, "?", "question")
	slug = strings.ReplaceAll(slug, "!", "exclamation")
	slug = strings.ReplaceAll(slug, "*", "star")
	slug = strings.ReplaceAll(slug, "^", "caret")
	slug = strings.ReplaceAll(slug, "~", "tilde")
	slug = strings.ReplaceAll(slug, "`", "backtick")
	slug = strings.ReplaceAll(slug, "|", "pipe")
	slug = strings.ReplaceAll(slug, "\\", "backslash")
	slug = strings.ReplaceAll(slug, "/", "slash")
	slug = strings.ReplaceAll(slug, ":", "colon")
	slug = strings.ReplaceAll(slug, ";", "semicolon")
	slug = strings.ReplaceAll(slug, "\"", "quote")
	slug = strings.ReplaceAll(slug, "'", "apostrophe")
	slug = strings.ReplaceAll(slug, "<", "less")
	slug = strings.ReplaceAll(slug, ">", "greater")
	slug = strings.ReplaceAll(slug, ",", "comma")
	slug = strings.ReplaceAll(slug, ".", "dot")

	// 移除连续的分隔符
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// 移除开头和结尾的分隔符
	slug = strings.Trim(slug, "-")

	// 如果为空，使用时间戳
	if slug == "" {
		slug = fmt.Sprintf("resume-%d", time.Now().Unix())
	}

	return slug
}
