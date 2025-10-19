package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
	"gorm.io/gorm"
)

// 标准响应格式
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// 标准成功响应
func standardSuccessResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// 标准错误响应
func standardErrorResponse(c *gin.Context, statusCode int, message, error string) {
	c.JSON(statusCode, StandardResponse{
		Success: false,
		Message: message,
		Error:   error,
	})
}

// 设置完整的Job服务API路由
func setupJobRoutes(r *gin.Engine, core *jobfirst.Core) {
	// 公开API路由组（无需认证）
	public := r.Group("/api/v1/job/public")
	{
		// 获取职位列表
		public.GET("/jobs", func(c *gin.Context) {
			getPublicJobs(c, core)
		})

		// 获取职位详情
		public.GET("/jobs/:id", func(c *gin.Context) {
			getPublicJobDetail(c, core)
		})

		// 获取公司职位列表
		public.GET("/companies/:company_id/jobs", func(c *gin.Context) {
			getCompanyJobs(c, core)
		})

		// 获取行业列表
		public.GET("/industries", func(c *gin.Context) {
			getIndustries(c, core)
		})

		// 获取工作类型列表
		public.GET("/job-types", func(c *gin.Context) {
			getJobTypes(c, core)
		})
	}

	// 职位管理API路由组（需要认证）
	jobs := r.Group("/api/v1/job/jobs")
	authMiddleware := core.AuthMiddleware.RequireAuth()
	jobs.Use(authMiddleware)
	{
		// 创建职位
		jobs.POST("/", func(c *gin.Context) {
			createJob(c, core)
		})

		// 更新职位
		jobs.PUT("/:id", func(c *gin.Context) {
			updateJob(c, core)
		})

		// 删除职位
		jobs.DELETE("/:id", func(c *gin.Context) {
			deleteJob(c, core)
		})

		// 获取我的职位列表
		jobs.GET("/my-jobs", func(c *gin.Context) {
			getMyJobs(c, core)
		})

		// 获取职位详情（需要认证）
		jobs.GET("/:id", func(c *gin.Context) {
			getJobDetail(c, core)
		})
	}

	// 职位申请API路由组（需要认证）
	applications := r.Group("/api/v1/job/jobs")
	applications.Use(authMiddleware)
	{
		// 申请职位
		applications.POST("/:id/apply", func(c *gin.Context) {
			applyJob(c, core)
		})

		// 获取我的申请历史
		applications.GET("/my-applications", func(c *gin.Context) {
			getMyApplications(c, core)
		})

		// 取消申请
		applications.DELETE("/:id/apply", func(c *gin.Context) {
			cancelApplication(c, core)
		})
	}

	// 职位匹配API路由组（需要认证）
	matching := r.Group("/api/v1/job/matching")
	matching.Use(authMiddleware)
	{
		// 智能职位匹配
		matching.POST("/match", func(c *gin.Context) {
			smartJobMatching(c, core)
		})

		// 获取匹配历史
		matching.GET("/history", func(c *gin.Context) {
			getMatchingHistory(c, core)
		})
	}

	// 管理员API路由组（需要管理员权限）
	admin := r.Group("/api/v1/job/admin")
	admin.Use(authMiddleware)
	{
		// 获取所有职位（管理员）
		admin.GET("/jobs", func(c *gin.Context) {
			getAllJobs(c, core)
		})

		// 更新职位状态（管理员）
		admin.PUT("/jobs/:id/status", func(c *gin.Context) {
			updateJobStatus(c, core)
		})

		// 获取职位申请列表（管理员）
		admin.GET("/jobs/:id/applications", func(c *gin.Context) {
			getJobApplications(c, core)
		})

		// 审核申请（管理员）
		admin.PUT("/applications/:id/review", func(c *gin.Context) {
			reviewApplication(c, core)
		})
	}
}

// 公开API实现

// 获取公开职位列表
func getPublicJobs(c *gin.Context, core *jobfirst.Core) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	industry := c.Query("industry")
	location := c.Query("location")
	jobType := c.Query("job_type")
	keyword := c.Query("keyword")

	db := core.GetDB()
	var jobs []Job
	offset := (page - 1) * pageSize

	query := db.Model(&Job{}).Preload("Company").Where("status = ?", JobStatusActive)

	if industry != "" {
		query = query.Where("industry = ?", industry)
	}
	if location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}
	if jobType != "" {
		query = query.Where("job_type = ?", jobType)
	}
	if keyword != "" {
		query = query.Where("(title LIKE ? OR description LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&jobs).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get jobs", err.Error())
		return
	}

	var total int64
	query.Count(&total)

	standardSuccessResponse(c, gin.H{
		"jobs":  jobs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	}, "Jobs retrieved successfully")
}

// 获取公开职位详情
func getPublicJobDetail(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))

	db := core.GetDB()
	var job Job
	if err := db.Preload("Company").Where("id = ? AND status = ?", jobID, JobStatusActive).First(&job).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	// 增加浏览次数
	db.Model(&job).Update("view_count", job.ViewCount+1)

	standardSuccessResponse(c, job, "Job detail retrieved successfully")
}

// 获取公司职位列表
func getCompanyJobs(c *gin.Context, core *jobfirst.Core) {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	db := core.GetDB()
	var jobs []Job
	offset := (page - 1) * pageSize

	if err := db.Preload("Company").Where("company_id = ? AND status = ?", companyID, JobStatusActive).Offset(offset).Limit(pageSize).Find(&jobs).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get company jobs", err.Error())
		return
	}

	var total int64
	db.Model(&Job{}).Where("company_id = ? AND status = ?", companyID, JobStatusActive).Count(&total)

	standardSuccessResponse(c, gin.H{
		"jobs":  jobs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	}, "Company jobs retrieved successfully")
}

// 获取行业列表
func getIndustries(c *gin.Context, core *jobfirst.Core) {
	industries := []string{
		"technology", "finance", "healthcare", "education",
		"marketing", "sales", "hr", "design", "media", "other",
	}
	standardSuccessResponse(c, industries, "Industries retrieved successfully")
}

// 获取工作类型列表
func getJobTypes(c *gin.Context, core *jobfirst.Core) {
	jobTypes := []string{
		"full-time", "part-time", "contract", "internship",
	}
	standardSuccessResponse(c, jobTypes, "Job types retrieved successfully")
}

// 职位管理API实现

// 创建职位
func createJob(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	db := core.GetDB()
	job := Job{
		Title:        req.Title,
		Description:  req.Description,
		Requirements: req.Requirements,
		CompanyID:    req.CompanyID,
		Industry:     req.Industry,
		Location:     req.Location,
		SalaryMin:    *req.SalaryMin,
		SalaryMax:    *req.SalaryMax,
		Experience:   req.Experience,
		Education:    req.Education,
		JobType:      req.JobType,
		Status:       JobStatusActive,
		CreatedBy:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.Create(&job).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to create job", err.Error())
		return
	}

	standardSuccessResponse(c, job, "Job created successfully")
}

// 更新职位
func updateJob(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var req UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	db := core.GetDB()
	var job Job
	if err := db.First(&job, jobID).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	// 检查权限
	if job.CreatedBy != userID {
		standardErrorResponse(c, http.StatusForbidden, "No permission to update this job", "")
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Requirements != nil {
		updates["requirements"] = *req.Requirements
	}
	if req.Industry != nil {
		updates["industry"] = *req.Industry
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.SalaryMin != nil {
		updates["salary_min"] = *req.SalaryMin
	}
	if req.SalaryMax != nil {
		updates["salary_max"] = *req.SalaryMax
	}
	if req.Experience != nil {
		updates["experience"] = *req.Experience
	}
	if req.Education != nil {
		updates["education"] = *req.Education
	}
	if req.JobType != nil {
		updates["job_type"] = *req.JobType
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	updates["updated_at"] = time.Now()

	if err := db.Model(&job).Updates(updates).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to update job", err.Error())
		return
	}

	standardSuccessResponse(c, job, "Job updated successfully")
}

// 删除职位
func deleteJob(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	db := core.GetDB()
	var job Job
	if err := db.First(&job, jobID).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	// 检查权限
	if job.CreatedBy != userID {
		standardErrorResponse(c, http.StatusForbidden, "No permission to delete this job", "")
		return
	}

	if err := db.Delete(&job).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to delete job", err.Error())
		return
	}

	standardSuccessResponse(c, gin.H{}, "Job deleted successfully")
}

// 获取我的职位列表
func getMyJobs(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	db := core.GetDB()
	var jobs []Job
	offset := (page - 1) * pageSize

	if err := db.Preload("Company").Where("created_by = ?", userID).Offset(offset).Limit(pageSize).Find(&jobs).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get my jobs", err.Error())
		return
	}

	var total int64
	db.Model(&Job{}).Where("created_by = ?", userID).Count(&total)

	standardSuccessResponse(c, gin.H{
		"jobs":  jobs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	}, "My jobs retrieved successfully")
}

// 获取职位详情（需要认证）
func getJobDetail(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))

	db := core.GetDB()
	var job Job
	if err := db.Preload("Company").First(&job, jobID).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "Job not found", err.Error())
		return
	}

	standardSuccessResponse(c, job, "Job detail retrieved successfully")
}

// 职位申请API实现

// 申请职位
func applyJob(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var req ApplyJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	db := core.GetDB()

	// 检查是否已经申请过
	var existingApplication JobApplication
	if err := db.Where("job_id = ? AND user_id = ?", jobID, userID).First(&existingApplication).Error; err == nil {
		standardErrorResponse(c, http.StatusConflict, "Already applied for this job", "")
		return
	}

	// 创建申请
	application := JobApplication{
		JobID:       uint(jobID),
		UserID:      userID,
		ResumeID:    req.ResumeID,
		Status:      ApplicationStatusPending,
		CoverLetter: req.CoverLetter,
		AppliedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(&application).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to apply for job", err.Error())
		return
	}

	// 增加申请次数
	db.Model(&Job{}).Where("id = ?", jobID).Update("apply_count", gorm.Expr("apply_count + 1"))

	standardSuccessResponse(c, application, "Job application submitted successfully")
}

// 获取我的申请历史
func getMyApplications(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	db := core.GetDB()
	var applications []JobApplication
	offset := (page - 1) * pageSize

	if err := db.Preload("Job").Preload("Job.Company").Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Find(&applications).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get my applications", err.Error())
		return
	}

	var total int64
	db.Model(&JobApplication{}).Where("user_id = ?", userID).Count(&total)

	standardSuccessResponse(c, gin.H{
		"applications": applications,
		"total":        total,
		"page":         page,
		"size":         pageSize,
	}, "My applications retrieved successfully")
}

// 取消申请
func cancelApplication(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	db := core.GetDB()
	var application JobApplication
	if err := db.Where("job_id = ? AND user_id = ?", jobID, userID).First(&application).Error; err != nil {
		standardErrorResponse(c, http.StatusNotFound, "Application not found", err.Error())
		return
	}

	if err := db.Delete(&application).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to cancel application", err.Error())
		return
	}

	standardSuccessResponse(c, gin.H{}, "Application cancelled successfully")
}

// 职位匹配API实现

// 智能职位匹配
func smartJobMatching(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	var req JobMatchingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// 这里应该调用AI服务进行智能匹配
	// 暂时返回模拟数据
	matches := []JobMatchResult{
		{
			JobID:      1,
			MatchScore: 85.5,
			Breakdown: map[string]float64{
				"skills":     90.0,
				"experience": 80.0,
				"location":   100.0,
			},
			Confidence: 0.85,
		},
	}

	// 记录匹配日志
	db := core.GetDB()
	filtersJSON, _ := json.Marshal(req.Filters)
	matchingLog := JobMatchingLog{
		UserID:         userID,
		ResumeID:       req.ResumeID,
		MatchesCount:   len(matches),
		FiltersApplied: string(filtersJSON),
		ProcessingTime: 150, // 毫秒
		CreatedAt:      time.Now(),
	}
	db.Create(&matchingLog)

	response := JobMatchingResponse{
		Matches:        matches,
		Total:          len(matches),
		ResumeID:       req.ResumeID,
		UserID:         userID,
		FiltersApplied: req.Filters,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	standardSuccessResponse(c, response, "Job matching completed successfully")
}

// 获取匹配历史
func getMatchingHistory(c *gin.Context, core *jobfirst.Core) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		standardErrorResponse(c, http.StatusUnauthorized, "User ID not found", "")
		return
	}
	userID := userIDInterface.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	db := core.GetDB()
	var logs []JobMatchingLog
	offset := (page - 1) * pageSize

	if err := db.Preload("User").Preload("Resume").Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get matching history", err.Error())
		return
	}

	var total int64
	db.Model(&JobMatchingLog{}).Where("user_id = ?", userID).Count(&total)

	standardSuccessResponse(c, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	}, "Matching history retrieved successfully")
}

// 管理员API实现

// 获取所有职位（管理员）
func getAllJobs(c *gin.Context, core *jobfirst.Core) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	db := core.GetDB()
	var jobs []Job
	offset := (page - 1) * pageSize

	query := db.Model(&Job{}).Preload("Company")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&jobs).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get all jobs", err.Error())
		return
	}

	var total int64
	query.Count(&total)

	standardSuccessResponse(c, gin.H{
		"jobs":  jobs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	}, "All jobs retrieved successfully")
}

// 更新职位状态（管理员）
func updateJobStatus(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	db := core.GetDB()
	if err := db.Model(&Job{}).Where("id = ?", jobID).Update("status", req.Status).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to update job status", err.Error())
		return
	}

	standardSuccessResponse(c, gin.H{}, "Job status updated successfully")
}

// 获取职位申请列表（管理员）
func getJobApplications(c *gin.Context, core *jobfirst.Core) {
	jobID, _ := strconv.Atoi(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	db := core.GetDB()
	var applications []JobApplication
	offset := (page - 1) * pageSize

	if err := db.Preload("User").Preload("Resume").Where("job_id = ?", jobID).Offset(offset).Limit(pageSize).Find(&applications).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to get job applications", err.Error())
		return
	}

	var total int64
	db.Model(&JobApplication{}).Where("job_id = ?", jobID).Count(&total)

	standardSuccessResponse(c, gin.H{
		"applications": applications,
		"total":        total,
		"page":         page,
		"size":         pageSize,
	}, "Job applications retrieved successfully")
}

// 审核申请（管理员）
func reviewApplication(c *gin.Context, core *jobfirst.Core) {
	applicationID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		standardErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	db := core.GetDB()
	now := time.Now()
	updates := map[string]interface{}{
		"status":      req.Status,
		"reviewed_at": &now,
		"updated_at":  now,
	}

	if req.Notes != "" {
		updates["cover_letter"] = req.Notes
	}

	if err := db.Model(&JobApplication{}).Where("id = ?", applicationID).Updates(updates).Error; err != nil {
		standardErrorResponse(c, http.StatusInternalServerError, "Failed to review application", err.Error())
		return
	}

	standardSuccessResponse(c, gin.H{}, "Application reviewed successfully")
}
