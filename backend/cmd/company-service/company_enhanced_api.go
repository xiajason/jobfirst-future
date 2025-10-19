package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// JobData 职位数据模型（PostgreSQL）
type JobData struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	CompanyID    uint      `json:"company_id" gorm:"not null"`
	Title        string    `json:"title" gorm:"size:200;not null"`
	Description  string    `json:"description" gorm:"type:text"`
	Requirements string    `json:"requirements" gorm:"type:text"`
	Location     string    `json:"location" gorm:"size:200"`
	Salary       string    `json:"salary" gorm:"size:100"`
	Status       string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// setupCompanyEnhancedRoutes 设置Company服务增强API路由
func setupCompanyEnhancedRoutes(r *gin.Engine, core *jobfirst.Core, dataSyncService *CompanyDataSyncService) {
	// 需要认证的增强API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	enhanced := r.Group("/api/v1/company/enhanced")
	enhanced.Use(authMiddleware)
	{
		// 企业数据同步API
		sync := enhanced.Group("/sync")
		{
			// 同步企业数据到所有数据库
			sync.POST("/:id", func(c *gin.Context) {
				companyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
					return
				}

				// 获取企业
				var company EnhancedCompany
				if err := core.GetDB().First(&company, companyID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
					return
				}

				// 同步到所有数据库
				if err := dataSyncService.SyncCompanyData(uint(companyID)); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "同步失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":     "success",
					"message":    "企业数据同步成功",
					"company_id": companyID,
					"timestamp":  time.Now(),
				})
			})

			// 获取同步状态
			sync.GET("/:id/status", func(c *gin.Context) {
				companyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
					return
				}

				status := map[string]interface{}{
					"company_id":  companyID,
					"sync_status": make(map[string]interface{}),
				}

				// 检查MySQL状态
				var company EnhancedCompany
				if err := core.GetDB().First(&company, companyID).Error; err == nil {
					status["sync_status"].(map[string]interface{})["mysql"] = map[string]interface{}{
						"status":       "synced",
						"last_updated": company.UpdatedAt,
					}
				} else {
					status["sync_status"].(map[string]interface{})["mysql"] = map[string]interface{}{
						"status": "not_found",
					}
				}

				// 检查PostgreSQL状态
				if dataSyncService.postgresDB != nil {
					var jobData JobData
					if err := dataSyncService.postgresDB.Where("company_id = ?", companyID).First(&jobData).Error; err == nil {
						status["sync_status"].(map[string]interface{})["postgresql"] = map[string]interface{}{
							"status":       "synced",
							"has_job_data": true,
							"last_updated": jobData.UpdatedAt,
						}
					} else {
						status["sync_status"].(map[string]interface{})["postgresql"] = map[string]interface{}{
							"status":       "not_synced",
							"has_job_data": false,
						}
					}
				} else {
					status["sync_status"].(map[string]interface{})["postgresql"] = map[string]interface{}{
						"status": "unavailable",
					}
				}

				// 检查Neo4j状态
				if dataSyncService.neo4jDriver != nil {
					relationships, err := dataSyncService.GetCompanyRelationships(uint(companyID))
					if err == nil {
						status["sync_status"].(map[string]interface{})["neo4j"] = map[string]interface{}{
							"status":             "synced",
							"relationship_count": len(relationships),
						}
					} else {
						status["sync_status"].(map[string]interface{})["neo4j"] = map[string]interface{}{
							"status": "not_synced",
						}
					}
				} else {
					status["sync_status"].(map[string]interface{})["neo4j"] = map[string]interface{}{
						"status": "unavailable",
					}
				}

				// 检查Redis状态
				if dataSyncService.redisClient != nil {
					cachedCompany, err := dataSyncService.GetCachedCompany(uint(companyID))
					if err == nil {
						status["sync_status"].(map[string]interface{})["redis"] = map[string]interface{}{
							"status":      "cached",
							"cached_name": cachedCompany.Name,
						}
					} else {
						status["sync_status"].(map[string]interface{})["redis"] = map[string]interface{}{
							"status": "not_cached",
						}
					}
				} else {
					status["sync_status"].(map[string]interface{})["redis"] = map[string]interface{}{
						"status": "unavailable",
					}
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   status,
				})
			})
		}

		// 企业地理位置API
		location := enhanced.Group("/location")
		{
			// 获取企业地理位置信息
			location.GET("/:id", func(c *gin.Context) {
				companyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
					return
				}

				var company EnhancedCompany
				if err := core.GetDB().First(&company, companyID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
					return
				}

				locationData := map[string]interface{}{
					"company_id":   company.ID,
					"company_name": company.Name,
					"location":     company.Location,
					"address":      company.Address,
					"city":         company.City,
					"district":     company.District,
					"area":         company.Area,
					"postal_code":  company.PostalCode,
					"bd_coordinates": map[string]interface{}{
						"latitude":  company.BDLatitude,
						"longitude": company.BDLongitude,
						"altitude":  company.BDAltitude,
						"accuracy":  company.BDAccuracy,
						"timestamp": company.BDTimestamp,
					},
					"location_codes": map[string]interface{}{
						"city_code":     company.CityCode,
						"district_code": company.DistrictCode,
						"area_code":     company.AreaCode,
					},
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   locationData,
				})
			})

			// 更新企业地理位置信息
			location.PUT("/:id", func(c *gin.Context) {
				companyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
					return
				}

				var updateData struct {
					BDLatitude   *float64 `json:"bd_latitude"`
					BDLongitude  *float64 `json:"bd_longitude"`
					BDAltitude   *float64 `json:"bd_altitude"`
					BDAccuracy   *float64 `json:"bd_accuracy"`
					BDTimestamp  *int64   `json:"bd_timestamp"`
					Address      string   `json:"address"`
					City         string   `json:"city"`
					District     string   `json:"district"`
					Area         string   `json:"area"`
					PostalCode   string   `json:"postal_code"`
					CityCode     string   `json:"city_code"`
					DistrictCode string   `json:"district_code"`
					AreaCode     string   `json:"area_code"`
				}

				if err := c.ShouldBindJSON(&updateData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				var company EnhancedCompany
				if err := core.GetDB().First(&company, companyID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "企业不存在"})
					return
				}

				// 更新地理位置信息
				company.BDLatitude = updateData.BDLatitude
				company.BDLongitude = updateData.BDLongitude
				company.BDAltitude = updateData.BDAltitude
				company.BDAccuracy = updateData.BDAccuracy
				company.BDTimestamp = updateData.BDTimestamp
				company.Address = updateData.Address
				company.City = updateData.City
				company.District = updateData.District
				company.Area = updateData.Area
				company.PostalCode = updateData.PostalCode
				company.CityCode = updateData.CityCode
				company.DistrictCode = updateData.DistrictCode
				company.AreaCode = updateData.AreaCode
				company.UpdatedAt = time.Now()

				if err := core.GetDB().Save(&company).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "更新地理位置信息失败"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "地理位置信息更新成功",
					"data":    company,
				})
			})
		}

		// 企业关系网络API
		relationships := enhanced.Group("/relationships")
		{
			// 获取企业关系
			relationships.GET("/:id", func(c *gin.Context) {
				companyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
					return
				}

				relationships, err := dataSyncService.GetCompanyRelationships(uint(companyID))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取企业关系失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":     "success",
					"data":       relationships,
					"count":      len(relationships),
					"company_id": companyID,
				})
			})

			// 创建企业关系
			relationships.POST("/", func(c *gin.Context) {
				var req struct {
					SourceID     uint    `json:"source_id" binding:"required"`
					TargetID     uint    `json:"target_id" binding:"required"`
					Relationship string  `json:"relationship" binding:"required"`
					Weight       float64 `json:"weight" binding:"required,min=0,max=1"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := dataSyncService.CreateCompanyRelationship(
					req.SourceID, req.TargetID, req.Relationship, req.Weight,
				); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "创建企业关系失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"status":  "success",
					"message": "企业关系创建成功",
					"data":    req,
				})
			})
		}

		// 企业分析API
		analysis := enhanced.Group("/analysis")
		{
			// 获取企业分析数据
			analysis.GET("/:id", func(c *gin.Context) {
				companyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的企业ID"})
					return
				}

				analysis, err := dataSyncService.AnalyzeCompanyData(uint(companyID))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "分析失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   analysis,
				})
			})
		}

		// 企业智能推荐API
		recommendations := enhanced.Group("/recommendations")
		{
			// 基于地理位置的推荐
			recommendations.POST("/location-based", func(c *gin.Context) {
				var req struct {
					CompanyID uint    `json:"company_id" binding:"required"`
					Radius    float64 `json:"radius,omitempty"` // 搜索半径(公里)
					Limit     int     `json:"limit,omitempty"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				radius := req.Radius
				if radius <= 0 {
					radius = 10.0 // 默认10公里
				}

				limit := req.Limit
				if limit <= 0 || limit > 20 {
					limit = 10
				}

				recommendations, err := dataSyncService.GetLocationBasedRecommendations(req.CompanyID, radius, limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "推荐失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"recommendations": recommendations,
						"count":           len(recommendations),
						"type":            "location-based",
						"radius":          radius,
					},
				})
			})

			// 基于行业关系的推荐
			recommendations.POST("/industry-based", func(c *gin.Context) {
				var req struct {
					CompanyID uint `json:"company_id" binding:"required"`
					Limit     int  `json:"limit,omitempty"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				limit := req.Limit
				if limit <= 0 || limit > 20 {
					limit = 10
				}

				recommendations, err := dataSyncService.GetIndustryBasedRecommendations(req.CompanyID, limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "推荐失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"recommendations": recommendations,
						"count":           len(recommendations),
						"type":            "industry-based",
					},
				})
			})
		}
	}
}
