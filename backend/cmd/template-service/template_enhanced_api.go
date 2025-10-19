package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// setupEnhancedRoutes 设置增强API路由
func setupEnhancedRoutes(r *gin.Engine, core *jobfirst.Core, enhancedService *TemplateEnhancedService) {
	// 需要认证的增强API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	enhanced := r.Group("/api/v1/template/enhanced")
	enhanced.Use(authMiddleware)
	{
		// 模板向量化API
		vectors := enhanced.Group("/vectors")
		{
			// 生成模板向量
			vectors.POST("/:id/generate", func(c *gin.Context) {
				templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
					return
				}

				// 获取模板内容
				var template Template
				if err := enhancedService.mysqlDB.First(&template, templateID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
					return
				}

				// 生成向量
				if err := enhancedService.GenerateTemplateVector(uint(templateID), template.Content); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "向量生成失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"message":     "模板向量生成成功",
					"template_id": templateID,
				})
			})

			// 获取相似模板
			vectors.GET("/:id/similar", func(c *gin.Context) {
				templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
					return
				}

				limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
				if limit > 50 {
					limit = 50
				}

				similarTemplates, err := enhancedService.GetSimilarTemplates(uint(templateID), limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取相似模板失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"data":        similarTemplates,
					"count":       len(similarTemplates),
					"template_id": templateID,
				})
			})
		}

		// 模板关系网络API
		relationships := enhanced.Group("/relationships")
		{
			// 创建模板关系
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

				if err := enhancedService.CreateTemplateRelationship(
					req.SourceID, req.TargetID, req.Relationship, req.Weight,
				); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "创建关系失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"status":  "success",
					"message": "模板关系创建成功",
					"data":    req,
				})
			})

			// 获取模板关系
			relationships.GET("/:id", func(c *gin.Context) {
				templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
					return
				}

				relationships, err := enhancedService.GetTemplateRelationships(uint(templateID))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关系失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"data":        relationships,
					"count":       len(relationships),
					"template_id": templateID,
				})
			})
		}

		// 模板同步API
		sync := enhanced.Group("/sync")
		{
			// 同步模板到所有数据库
			sync.POST("/:id", func(c *gin.Context) {
				templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
					return
				}

				// 获取模板
				var template Template
				if err := enhancedService.mysqlDB.First(&template, templateID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
					return
				}

				// 同步到所有数据库
				if err := enhancedService.SyncTemplateToAllDatabases(&template); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "同步失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"message":     "模板同步成功",
					"template_id": templateID,
					"timestamp":   time.Now(),
				})
			})

			// 获取同步状态
			sync.GET("/:id/status", func(c *gin.Context) {
				templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
					return
				}

				status := map[string]interface{}{
					"template_id": templateID,
					"sync_status": make(map[string]interface{}),
				}

				// 检查MySQL状态
				var template Template
				if err := enhancedService.mysqlDB.First(&template, templateID).Error; err == nil {
					status["sync_status"].(map[string]interface{})["mysql"] = map[string]interface{}{
						"status":       "synced",
						"last_updated": template.UpdatedAt,
					}
				} else {
					status["sync_status"].(map[string]interface{})["mysql"] = map[string]interface{}{
						"status": "not_found",
					}
				}

				// 检查PostgreSQL状态
				if enhancedService.postgresDB != nil {
					var vector TemplateVector
					if err := enhancedService.postgresDB.Where("template_id = ?", templateID).First(&vector).Error; err == nil {
						status["sync_status"].(map[string]interface{})["postgresql"] = map[string]interface{}{
							"status":       "synced",
							"has_vector":   true,
							"last_updated": vector.UpdatedAt,
						}
					} else {
						status["sync_status"].(map[string]interface{})["postgresql"] = map[string]interface{}{
							"status":     "not_synced",
							"has_vector": false,
						}
					}
				} else {
					status["sync_status"].(map[string]interface{})["postgresql"] = map[string]interface{}{
						"status": "unavailable",
					}
				}

				// 检查Neo4j状态
				if enhancedService.neo4jDriver != nil {
					relationships, err := enhancedService.GetTemplateRelationships(uint(templateID))
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
				if enhancedService.redisClient != nil {
					cachedTemplate, err := enhancedService.GetCachedTemplate(uint(templateID))
					if err == nil {
						status["sync_status"].(map[string]interface{})["redis"] = map[string]interface{}{
							"status":      "cached",
							"cached_name": cachedTemplate.Name,
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

		// 模板分析API
		analysis := enhanced.Group("/analysis")
		{
			// 获取模板使用分析
			analysis.GET("/:id/usage", func(c *gin.Context) {
				templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
					return
				}

				analysis, err := enhancedService.AnalyzeTemplateUsage(uint(templateID))
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

		// 智能推荐API
		recommendations := enhanced.Group("/recommendations")
		{
			// 基于内容的推荐
			recommendations.POST("/content-based", func(c *gin.Context) {
				var req struct {
					TemplateID uint `json:"template_id" binding:"required"`
					Limit      int  `json:"limit,omitempty"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				limit := req.Limit
				if limit <= 0 || limit > 20 {
					limit = 10
				}

				similarTemplates, err := enhancedService.GetSimilarTemplates(req.TemplateID, limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "推荐失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"recommendations": similarTemplates,
						"count":           len(similarTemplates),
						"type":            "content-based",
					},
				})
			})

			// 基于关系的推荐
			recommendations.POST("/relationship-based", func(c *gin.Context) {
				var req struct {
					TemplateID uint `json:"template_id" binding:"required"`
					Limit      int  `json:"limit,omitempty"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				limit := req.Limit
				if limit <= 0 || limit > 20 {
					limit = 10
				}

				relationships, err := enhancedService.GetTemplateRelationships(req.TemplateID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "推荐失败: " + err.Error()})
					return
				}

				// 获取相关模板
				var recommendations []Template
				for i, rel := range relationships {
					if i >= limit {
						break
					}
					var template Template
					if err := enhancedService.mysqlDB.First(&template, rel.TargetID).Error; err == nil {
						recommendations = append(recommendations, template)
					}
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data": gin.H{
						"recommendations": recommendations,
						"count":           len(recommendations),
						"type":            "relationship-based",
					},
				})
			})
		}
	}
}
