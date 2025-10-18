package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jobfirst/jobfirst-core"
)

// setupStatisticsEnhancedRoutes 设置Statistics服务增强API路由
func setupStatisticsEnhancedRoutes(r *gin.Engine, core *jobfirst.Core, enhancedService *StatisticsEnhancedService) {
	// 需要认证的增强API路由
	authMiddleware := core.AuthMiddleware.RequireAuth()
	enhanced := r.Group("/api/v1/statistics/enhanced")
	enhanced.Use(authMiddleware)
	{
		// 实时分析API
		realtime := enhanced.Group("/realtime")
		{
			// 记录实时分析数据
			realtime.POST("/record", func(c *gin.Context) {
				var req struct {
					MetricType  string                 `json:"metric_type" binding:"required"`
					MetricName  string                 `json:"metric_name" binding:"required"`
					MetricValue float64                `json:"metric_value" binding:"required"`
					Dimensions  map[string]interface{} `json:"dimensions"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := enhancedService.RecordRealTimeAnalytics(req.MetricType, req.MetricName, req.MetricValue, req.Dimensions); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "记录实时分析数据失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "实时分析数据记录成功",
					"data":    req,
				})
			})

			// 获取实时分析数据
			realtime.GET("/metrics/:type/:name", func(c *gin.Context) {
				metricType := c.Param("type")
				metricName := c.Param("name")
				limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

				analytics, err := enhancedService.GetRealTimeAnalytics(metricType, metricName, limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取实时分析数据失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   analytics,
					"count":  len(analytics),
				})
			})
		}

		// 历史分析API
		historical := enhanced.Group("/historical")
		{
			// 执行历史数据分析
			historical.POST("/analyze", func(c *gin.Context) {
				var req struct {
					AnalysisType string    `json:"analysis_type" binding:"required"` // trend, pattern, correlation
					EntityType   string    `json:"entity_type" binding:"required"`
					EntityID     uint      `json:"entity_id" binding:"required"`
					StartDate    time.Time `json:"start_date" binding:"required"`
					EndDate      time.Time `json:"end_date" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				result, err := enhancedService.PerformHistoricalAnalysis(
					req.AnalysisType, req.EntityType, req.EntityID, req.StartDate, req.EndDate,
				)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "执行历史分析失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   result,
				})
			})
		}

		// 预测模型API
		predictive := enhanced.Group("/predictive")
		{
			// 创建预测模型
			predictive.POST("/models", func(c *gin.Context) {
				var req struct {
					ModelName    string                 `json:"model_name" binding:"required"`
					ModelType    string                 `json:"model_type" binding:"required"` // regression, classification, clustering
					TargetEntity string                 `json:"target_entity" binding:"required"`
					Parameters   map[string]interface{} `json:"parameters"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				model, err := enhancedService.CreatePredictiveModel(
					req.ModelName, req.ModelType, req.TargetEntity, req.Parameters,
				)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "创建预测模型失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"status":  "success",
					"message": "预测模型创建成功",
					"data":    model,
				})
			})

			// 训练预测模型
			predictive.POST("/models/:id/train", func(c *gin.Context) {
				modelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模型ID"})
					return
				}

				var req struct {
					TrainingData map[string]interface{} `json:"training_data" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := enhancedService.TrainPredictiveModel(uint(modelID), req.TrainingData); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "训练预测模型失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":   "success",
					"message":  "预测模型训练成功",
					"model_id": modelID,
				})
			})

			// 生成预测
			predictive.POST("/predict", func(c *gin.Context) {
				var req struct {
					ModelID        uint   `json:"model_id" binding:"required"`
					EntityType     string `json:"entity_type" binding:"required"`
					EntityID       uint   `json:"entity_id" binding:"required"`
					PredictionType string `json:"prediction_type" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				result, err := enhancedService.GeneratePrediction(
					req.ModelID, req.EntityType, req.EntityID, req.PredictionType,
				)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "生成预测失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   result,
				})
			})
		}

		// 异常检测API
		anomaly := enhanced.Group("/anomaly")
		{
			// 检测异常
			anomaly.POST("/detect", func(c *gin.Context) {
				var req struct {
					MetricName string  `json:"metric_name" binding:"required"`
					Threshold  float64 `json:"threshold" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				anomalies, err := enhancedService.DetectAnomalies(req.MetricName, req.Threshold)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "检测异常失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   anomalies,
					"count":  len(anomalies),
				})
			})
		}

		// 业务洞察API
		insights := enhanced.Group("/insights")
		{
			// 生成业务洞察
			insights.POST("/generate", func(c *gin.Context) {
				insights, err := enhancedService.GenerateBusinessInsights()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "生成业务洞察失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   insights,
					"count":  len(insights),
				})
			})
		}

		// 数据同步API
		sync := enhanced.Group("/sync")
		{
			// 获取同步状态
			sync.GET("/status/:entity_type/:entity_id", func(c *gin.Context) {
				entityType := c.Param("entity_type")
				entityID, err := strconv.ParseUint(c.Param("entity_id"), 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的实体ID"})
					return
				}

				status, err := enhancedService.GetSyncStatus(entityType, uint(entityID))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取同步状态失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   status,
				})
			})
		}

		// 可视化API
		visualization := enhanced.Group("/visualization")
		{
			// 创建可视化配置
			visualization.POST("/configs", func(c *gin.Context) {
				var req struct {
					ChartType       string                 `json:"chart_type" binding:"required"`
					ChartName       string                 `json:"chart_name" binding:"required"`
					DataSource      string                 `json:"data_source" binding:"required"`
					QueryConfig     map[string]interface{} `json:"query_config"`
					DisplayConfig   map[string]interface{} `json:"display_config"`
					RefreshInterval int                    `json:"refresh_interval"`
					IsPublic        bool                   `json:"is_public"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				// 获取当前用户ID
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
					return
				}
				userID := userIDInterface.(uint)

				queryConfigJSON, _ := json.Marshal(req.QueryConfig)
				displayConfigJSON, _ := json.Marshal(req.DisplayConfig)

				config := VisualizationConfig{
					ChartType:       req.ChartType,
					ChartName:       req.ChartName,
					DataSource:      req.DataSource,
					QueryConfig:     string(queryConfigJSON),
					DisplayConfig:   string(displayConfigJSON),
					RefreshInterval: req.RefreshInterval,
					IsPublic:        req.IsPublic,
					CreatedBy:       userID,
				}

				if err := enhancedService.postgresDB.Create(&config).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "创建可视化配置失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"status":  "success",
					"message": "可视化配置创建成功",
					"data":    config,
				})
			})

			// 获取可视化配置列表
			visualization.GET("/configs", func(c *gin.Context) {
				var configs []VisualizationConfig
				if err := enhancedService.postgresDB.Find(&configs).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取可视化配置失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   configs,
					"count":  len(configs),
				})
			})
		}

		// 报告生成API
		reports := enhanced.Group("/reports")
		{
			// 生成统计报告
			reports.POST("/generate", func(c *gin.Context) {
				var req struct {
					ReportType   string    `json:"report_type" binding:"required"` // daily, weekly, monthly, custom
					ReportName   string    `json:"report_name" binding:"required"`
					ReportPeriod string    `json:"report_period"`
					StartDate    time.Time `json:"start_date" binding:"required"`
					EndDate      time.Time `json:"end_date" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				// 获取当前用户ID
				userIDInterface, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "用户ID不存在"})
					return
				}
				userID := userIDInterface.(uint)

				// 模拟报告生成
				reportData := map[string]interface{}{
					"total_users":    1500,
					"new_users":      150,
					"active_users":   1200,
					"template_usage": 8500,
					"company_growth": 25,
				}

				reportDataJSON, _ := json.Marshal(reportData)

				report := StatisticsReport{
					ReportType:      req.ReportType,
					ReportName:      req.ReportName,
					ReportPeriod:    req.ReportPeriod,
					StartDate:       req.StartDate,
					EndDate:         req.EndDate,
					ReportData:      string(reportDataJSON),
					Summary:         "系统整体运行良好，用户增长稳定",
					Insights:        "用户活跃度提升，模板使用率增长",
					Recommendations: "继续优化用户体验，加强功能推广",
					Status:          "completed",
					GeneratedBy:     userID,
					GeneratedAt:     time.Now(),
				}

				if err := enhancedService.postgresDB.Create(&report).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "生成统计报告失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "统计报告生成成功",
					"data":    report,
				})
			})

			// 获取报告列表
			reports.GET("/", func(c *gin.Context) {
				var reports []StatisticsReport
				if err := enhancedService.postgresDB.Find(&reports).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "获取报告列表失败: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"data":   reports,
					"count":  len(reports),
				})
			})
		}
	}
}
