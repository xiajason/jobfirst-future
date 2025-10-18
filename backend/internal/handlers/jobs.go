package handlers

import (
	"github.com/xiajason/zervi-basic/basic/backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// JobHandler 职位处理器
type JobHandler struct {
	db *gorm.DB
}

// NewJobHandler 创建职位处理器
func NewJobHandler() *JobHandler {
	// 使用全局数据库连接
	return &JobHandler{
		db: getGlobalDB(),
	}
}

// getGlobalDB 获取全局数据库连接
func getGlobalDB() *gorm.DB {
	// 从main包获取数据库连接
	// 这里需要导入main包，但由于循环依赖问题，我们暂时使用一个全局变量
	// 在实际项目中，应该使用依赖注入容器
	return globalDB
}

// 全局数据库连接变量
var globalDB *gorm.DB

// SetGlobalDB 设置全局数据库连接
func SetGlobalDB(db *gorm.DB) {
	globalDB = db
}

// GetJobsV2 获取职位列表（新版本）
func (h *JobHandler) GetJobsV2(c *gin.Context) {
	// 获取查询参数
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 从数据库查询职位数据
	var jobs []models.Job
	query := h.db.Preload("Company").Preload("Category").Where("status = ?", "published")

	if err := query.Limit(limit).Order("priority DESC, created_at DESC").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to fetch jobs",
			"error":   err.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	var result []map[string]interface{}
	for _, job := range jobs {
		jobData := map[string]interface{}{
			"id":                  job.ID,
			"title":               job.Title,
			"company_id":          job.CompanyID,
			"company_name":        job.Company.Name,
			"company_short_name":  job.Company.ShortName,
			"company_logo":        job.Company.LogoURL,
			"location":            job.Location,
			"salary_min":          job.SalaryMin,
			"salary_max":          job.SalaryMax,
			"salary_type":         job.SalaryType,
			"experience_required": job.ExperienceRequired,
			"education_required":  job.EducationRequired,
			"job_type":            job.JobType,
			"status":              job.Status,
			"view_count":          job.ViewCount,
			"application_count":   job.ApplicationCount,
			"favorite_count":      job.FavoriteCount,
			"created_at":          job.CreatedAt,
		}
		result = append(result, jobData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// GetJobDetailV2 获取职位详情（新版本）
func (h *JobHandler) GetJobDetailV2(c *gin.Context) {
	jobID := c.Param("id")

	// 从数据库查询职位详情
	var job models.Job
	if err := h.db.Preload("Company").Preload("Category").Where("id = ? AND status = ?", jobID, "published").First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Job not found",
			"error":   err.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	jobData := map[string]interface{}{
		"id":                  job.ID,
		"title":               job.Title,
		"company_id":          job.CompanyID,
		"company_name":        job.Company.Name,
		"company_short_name":  job.Company.ShortName,
		"company_logo":        job.Company.LogoURL,
		"location":            job.Location,
		"salary_min":          job.SalaryMin,
		"salary_max":          job.SalaryMax,
		"salary_type":         job.SalaryType,
		"experience_required": job.ExperienceRequired,
		"education_required":  job.EducationRequired,
		"job_type":            job.JobType,
		"status":              job.Status,
		"description":         job.Description,
		"requirements":        job.Requirements,
		"benefits":            job.Benefits,
		"skills":              job.Skills,
		"tags":                job.Tags,
		"view_count":          job.ViewCount,
		"application_count":   job.ApplicationCount,
		"favorite_count":      job.FavoriteCount,
		"created_at":          job.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    jobData,
	})
}

// SearchJobsV2 搜索职位（新版本）
func (h *JobHandler) SearchJobsV2(c *gin.Context) {
	// 获取搜索参数
	keyword := c.Query("keyword")
	location := c.Query("location")
	experience := c.Query("experience")

	// 构建查询条件
	query := h.db.Preload("Company").Preload("Category").Where("status = ?", "published")

	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}

	if experience != "" {
		query = query.Where("experience_required = ?", experience)
	}

	// 执行查询
	var jobs []models.Job
	if err := query.Order("priority DESC, created_at DESC").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to search jobs",
			"error":   err.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	var result []map[string]interface{}
	for _, job := range jobs {
		jobData := map[string]interface{}{
			"id":                  job.ID,
			"title":               job.Title,
			"company_id":          job.CompanyID,
			"company_name":        job.Company.Name,
			"company_short_name":  job.Company.ShortName,
			"company_logo":        job.Company.LogoURL,
			"location":            job.Location,
			"salary_min":          job.SalaryMin,
			"salary_max":          job.SalaryMax,
			"salary_type":         job.SalaryType,
			"experience_required": job.ExperienceRequired,
			"education_required":  job.EducationRequired,
			"job_type":            job.JobType,
			"status":              job.Status,
			"view_count":          job.ViewCount,
			"application_count":   job.ApplicationCount,
			"favorite_count":      job.FavoriteCount,
			"created_at":          job.CreatedAt,
		}
		result = append(result, jobData)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}
