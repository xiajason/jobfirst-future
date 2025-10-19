package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jobfirst/jobfirst-core"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewStatisticsEnhancedService 创建统计增强服务实例
func NewStatisticsEnhancedService(core *jobfirst.Core) (*StatisticsEnhancedService, error) {
	service := &StatisticsEnhancedService{
		core:    core,
		mysqlDB: core.GetDB(), // 使用核心包的MySQL DB
	}

	// 初始化PostgreSQL
	// 直接连接PostgreSQL，不依赖core.Database.PostgreSQL配置
	postgresDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=szjason72 dbname=jobfirst_vector port=5432 sslmode=disable",
	}), &gorm.Config{})
	if err != nil {
		log.Printf("PostgreSQL连接失败: %v", err)
	} else {
		service.postgresDB = postgresDB
		log.Println("PostgreSQL连接成功")
		// 自动迁移表
		err = service.createAnalyticsTables()
		if err != nil {
			log.Printf("创建PostgreSQL分析表失败: %v", err)
		}
	}

	// 初始化Neo4j
	if core.Database.Neo4j != nil {
		neo4jDriver := core.Database.Neo4j.GetDriver()
		if neo4jDriver == nil {
			log.Printf("Neo4j驱动获取失败")
		} else {
			// 类型断言为正确的驱动类型
			if driver, ok := neo4jDriver.(neo4j.DriverWithContext); ok {
				service.neo4jDriver = driver
				// 创建分析关系索引
				service.createAnalyticsIndexes()
			} else {
				log.Printf("Neo4j驱动类型不匹配")
			}
		}
	}

	// 获取Redis连接
	redisManager := core.Database.GetRedis()
	if redisManager != nil {
		service.redisClient = redisManager.GetClient()
	}

	return service, nil
}

// createAnalyticsTables 创建分析表
func (s *StatisticsEnhancedService) createAnalyticsTables() error {
	if s.postgresDB == nil {
		return fmt.Errorf("PostgreSQL未连接，无法创建分析表")
	}

	// 创建实时分析表
	err := s.postgresDB.AutoMigrate(&RealTimeAnalytics{})
	if err != nil {
		return fmt.Errorf("创建实时分析表失败: %w", err)
	}

	// 创建历史分析表
	err = s.postgresDB.AutoMigrate(&HistoricalAnalysis{})
	if err != nil {
		return fmt.Errorf("创建历史分析表失败: %w", err)
	}

	// 创建预测模型表
	err = s.postgresDB.AutoMigrate(&PredictiveModel{})
	if err != nil {
		return fmt.Errorf("创建预测模型表失败: %w", err)
	}

	// 创建预测结果表
	err = s.postgresDB.AutoMigrate(&PredictionResult{})
	if err != nil {
		return fmt.Errorf("创建预测结果表失败: %w", err)
	}

	// 创建用户行为分析表
	err = s.postgresDB.AutoMigrate(&UserBehaviorAnalysis{})
	if err != nil {
		return fmt.Errorf("创建用户行为分析表失败: %w", err)
	}

	// 创建业务洞察表
	err = s.postgresDB.AutoMigrate(&BusinessInsight{})
	if err != nil {
		return fmt.Errorf("创建业务洞察表失败: %w", err)
	}

	// 创建异常检测表
	err = s.postgresDB.AutoMigrate(&AnomalyDetection{})
	if err != nil {
		return fmt.Errorf("创建异常检测表失败: %w", err)
	}

	// 创建可视化配置表
	err = s.postgresDB.AutoMigrate(&VisualizationConfig{})
	if err != nil {
		return fmt.Errorf("创建可视化配置表失败: %w", err)
	}

	// 创建统计报告表
	err = s.postgresDB.AutoMigrate(&StatisticsReport{})
	if err != nil {
		return fmt.Errorf("创建统计报告表失败: %w", err)
	}

	// 创建数据同步状态表
	err = s.postgresDB.AutoMigrate(&StatisticsSyncStatus{})
	if err != nil {
		return fmt.Errorf("创建数据同步状态表失败: %w", err)
	}

	log.Println("PostgreSQL分析表创建成功")
	return nil
}

// createAnalyticsIndexes 创建Neo4j索引
func (s *StatisticsEnhancedService) createAnalyticsIndexes() {
	if s.neo4jDriver == nil {
		log.Println("Neo4j未连接，无法创建索引")
		return
	}
	ctx := context.Background()
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// 创建用户行为关系索引
	_, err := session.Run(ctx, "CREATE CONSTRAINT IF NOT EXISTS FOR (ub:UserBehavior) REQUIRE ub.id IS UNIQUE", map[string]interface{}{})
	if err != nil {
		log.Printf("创建Neo4j UserBehavior ID唯一性约束失败: %v", err)
	} else {
		log.Println("Neo4j UserBehavior ID唯一性约束已创建或已存在")
	}

	// 创建分析关系索引
	_, err = session.Run(ctx, "CREATE CONSTRAINT IF NOT EXISTS FOR (a:Analysis) REQUIRE a.id IS UNIQUE", map[string]interface{}{})
	if err != nil {
		log.Printf("创建Neo4j Analysis ID唯一性约束失败: %v", err)
	} else {
		log.Println("Neo4j Analysis ID唯一性约束已创建或已存在")
	}
}

// RecordRealTimeAnalytics 记录实时分析数据
func (s *StatisticsEnhancedService) RecordRealTimeAnalytics(metricType, metricName string, value float64, dimensions map[string]interface{}) error {
	if s.postgresDB == nil {
		return fmt.Errorf("PostgreSQL未连接，无法记录实时分析数据")
	}

	dimensionsJSON, err := json.Marshal(dimensions)
	if err != nil {
		return fmt.Errorf("序列化维度数据失败: %w", err)
	}

	analytics := RealTimeAnalytics{
		MetricType:  metricType,
		MetricName:  metricName,
		MetricValue: value,
		Dimensions:  string(dimensionsJSON),
		Timestamp:   time.Now(),
	}

	if err := s.postgresDB.Create(&analytics).Error; err != nil {
		return fmt.Errorf("保存实时分析数据失败: %w", err)
	}

	// 同步到Redis缓存
	if s.redisClient != nil {
		key := fmt.Sprintf("analytics:%s:%s", metricType, metricName)
		s.redisClient.Set(context.Background(), key, value, time.Minute*5)
	}

	return nil
}

// GetRealTimeAnalytics 获取实时分析数据
func (s *StatisticsEnhancedService) GetRealTimeAnalytics(metricType, metricName string, limit int) ([]RealTimeAnalytics, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法获取实时分析数据")
	}

	var analytics []RealTimeAnalytics
	query := s.postgresDB.Where("metric_type = ? AND metric_name = ?", metricType, metricName)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("timestamp DESC").Find(&analytics).Error; err != nil {
		return nil, fmt.Errorf("获取实时分析数据失败: %w", err)
	}

	return analytics, nil
}

// PerformHistoricalAnalysis 执行历史数据分析
func (s *StatisticsEnhancedService) PerformHistoricalAnalysis(analysisType, entityType string, entityID uint, startDate, endDate time.Time) (*AnalysisResult, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法执行历史分析")
	}

	// 根据分析类型执行不同的分析
	var result *AnalysisResult
	var err error

	switch analysisType {
	case "trend":
		result, err = s.analyzeTrend(entityType, entityID, startDate, endDate)
	case "pattern":
		result, err = s.analyzePattern(entityType, entityID, startDate, endDate)
	case "correlation":
		result, err = s.analyzeCorrelation(entityType, entityID, startDate, endDate)
	default:
		return nil, fmt.Errorf("不支持的分析类型: %s", analysisType)
	}

	if err != nil {
		return nil, err
	}

	// 保存分析结果到数据库
	analysis := HistoricalAnalysis{
		AnalysisType:   analysisType,
		EntityType:     entityType,
		EntityID:       entityID,
		AnalysisPeriod: "custom",
		StartDate:      startDate,
		EndDate:        endDate,
		AnalysisResult: fmt.Sprintf(`{"result": %s}`, result.Result),
		Insights:       fmt.Sprintf("%v", result.Insights),
		Confidence:     result.Confidence,
	}

	if err := s.postgresDB.Create(&analysis).Error; err != nil {
		log.Printf("保存历史分析结果失败: %v", err)
	}

	return result, nil
}

// analyzeTrend 分析趋势
func (s *StatisticsEnhancedService) analyzeTrend(entityType string, entityID uint, startDate, endDate time.Time) (*AnalysisResult, error) {
	// 模拟趋势分析
	result := &AnalysisResult{
		AnalysisType: "trend",
		EntityType:   entityType,
		EntityID:     entityID,
		Result: map[string]interface{}{
			"trend_direction": "increasing",
			"growth_rate":     0.15,
			"volatility":      0.05,
			"seasonality":     "moderate",
		},
		Insights: []string{
			"数据呈现稳定增长趋势",
			"增长率保持在15%左右",
			"存在适度的季节性波动",
		},
		Confidence: 0.85,
		Timestamp:  time.Now(),
	}

	return result, nil
}

// analyzePattern 分析模式
func (s *StatisticsEnhancedService) analyzePattern(entityType string, entityID uint, startDate, endDate time.Time) (*AnalysisResult, error) {
	// 模拟模式分析
	result := &AnalysisResult{
		AnalysisType: "pattern",
		EntityType:   entityType,
		EntityID:     entityID,
		Result: map[string]interface{}{
			"pattern_type":     "cyclical",
			"cycle_length":     7, // 天
			"pattern_strength": 0.75,
			"anomalies":        []string{"2025-09-10", "2025-09-15"},
		},
		Insights: []string{
			"数据呈现7天周期性模式",
			"模式强度为75%",
			"发现2个异常数据点",
		},
		Confidence: 0.78,
		Timestamp:  time.Now(),
	}

	return result, nil
}

// analyzeCorrelation 分析相关性
func (s *StatisticsEnhancedService) analyzeCorrelation(entityType string, entityID uint, startDate, endDate time.Time) (*AnalysisResult, error) {
	// 模拟相关性分析
	result := &AnalysisResult{
		AnalysisType: "correlation",
		EntityType:   entityType,
		EntityID:     entityID,
		Result: map[string]interface{}{
			"correlation_matrix": map[string]float64{
				"user_activity":  0.65,
				"template_usage": 0.42,
				"company_growth": 0.38,
			},
			"strongest_correlation": "user_activity",
			"correlation_strength":  0.65,
		},
		Insights: []string{
			"用户活跃度与整体指标相关性最强",
			"模板使用率与用户活跃度正相关",
			"企业增长与用户活跃度存在中等相关性",
		},
		Confidence: 0.72,
		Timestamp:  time.Now(),
	}

	return result, nil
}

// CreatePredictiveModel 创建预测模型
func (s *StatisticsEnhancedService) CreatePredictiveModel(modelName, modelType, targetEntity string, parameters map[string]interface{}) (*PredictiveModel, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法创建预测模型")
	}

	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		return nil, fmt.Errorf("序列化模型参数失败: %w", err)
	}

	model := PredictiveModel{
		ModelName:       modelName,
		ModelType:       modelType,
		TargetEntity:    targetEntity,
		ModelVersion:    "1.0.0",
		ModelParameters: string(parametersJSON),
		ModelAccuracy:   0.0, // 初始准确度
		Status:          "training",
		LastTrained:     time.Now(),
	}

	if err := s.postgresDB.Create(&model).Error; err != nil {
		return nil, fmt.Errorf("创建预测模型失败: %w", err)
	}

	return &model, nil
}

// TrainPredictiveModel 训练预测模型
func (s *StatisticsEnhancedService) TrainPredictiveModel(modelID uint, trainingData map[string]interface{}) error {
	if s.postgresDB == nil {
		return fmt.Errorf("PostgreSQL未连接，无法训练预测模型")
	}

	// 模拟模型训练
	trainingDataJSON, err := json.Marshal(trainingData)
	if err != nil {
		return fmt.Errorf("序列化训练数据失败: %w", err)
	}

	// 更新模型状态和准确度
	model := PredictiveModel{
		TrainingData:  string(trainingDataJSON),
		ModelAccuracy: 0.85, // 模拟训练后的准确度
		Status:        "active",
		LastTrained:   time.Now(),
	}

	if err := s.postgresDB.Model(&PredictiveModel{}).Where("id = ?", modelID).Updates(model).Error; err != nil {
		return fmt.Errorf("更新预测模型失败: %w", err)
	}

	return nil
}

// GeneratePrediction 生成预测
func (s *StatisticsEnhancedService) GeneratePrediction(modelID uint, entityType string, entityID uint, predictionType string) (*PredictionResultInterface, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法生成预测")
	}

	// 获取模型信息
	var model PredictiveModel
	if err := s.postgresDB.First(&model, modelID).Error; err != nil {
		return nil, fmt.Errorf("获取预测模型失败: %w", err)
	}

	if model.Status != "active" {
		return nil, fmt.Errorf("预测模型未激活")
	}

	// 模拟预测生成
	predictedValue := 100.0 + math.Sin(float64(time.Now().Unix())/86400)*50                   // 模拟预测值
	confidence := model.ModelAccuracy * (0.8 + 0.2*math.Sin(float64(time.Now().Unix())/3600)) // 模拟置信度

	result := &PredictionResultInterface{
		ModelID:        modelID,
		EntityType:     entityType,
		EntityID:       entityID,
		PredictionType: predictionType,
		PredictedValue: predictedValue,
		Confidence:     confidence,
		Details: map[string]interface{}{
			"model_version": model.ModelVersion,
			"training_date": model.LastTrained,
			"parameters":    model.ModelParameters,
		},
		Timestamp: time.Now(),
	}

	// 保存预测结果
	predictionRecord := PredictionResult{
		ModelID:        modelID,
		EntityType:     entityType,
		EntityID:       entityID,
		PredictionType: predictionType,
		PredictedValue: predictedValue,
		Confidence:     confidence,
		PredictionDate: time.Now(),
	}

	if err := s.postgresDB.Create(&predictionRecord).Error; err != nil {
		log.Printf("保存预测结果失败: %v", err)
	}

	return result, nil
}

// DetectAnomalies 检测异常
func (s *StatisticsEnhancedService) DetectAnomalies(metricName string, threshold float64) ([]AnomalyDetection, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法检测异常")
	}

	// 模拟异常检测
	anomalies := []AnomalyDetection{
		{
			AnomalyType:   "spike",
			EntityType:    "user_activity",
			MetricName:    metricName,
			ExpectedValue: 100.0,
			ActualValue:   150.0,
			Deviation:     0.5,
			Severity:      "medium",
			Description:   "用户活跃度出现异常峰值",
			Status:        "detected",
			DetectedAt:    time.Now(),
		},
	}

	// 保存异常检测结果
	for _, anomaly := range anomalies {
		if err := s.postgresDB.Create(&anomaly).Error; err != nil {
			log.Printf("保存异常检测结果失败: %v", err)
		}
	}

	return anomalies, nil
}

// GenerateBusinessInsights 生成业务洞察
func (s *StatisticsEnhancedService) GenerateBusinessInsights() ([]BusinessInsightInterface, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法生成业务洞察")
	}

	// 模拟业务洞察生成
	insights := []BusinessInsightInterface{
		{
			InsightType:     "trend",
			InsightCategory: "user_growth",
			Title:           "用户增长趋势分析",
			Description:     "过去30天用户增长率达到15%，主要增长来源为新用户注册",
			Impact:          "high",
			Confidence:      0.88,
			DataPoints: map[string]interface{}{
				"growth_rate":    0.15,
				"new_users":      150,
				"retention_rate": 0.75,
			},
			Recommendations: []string{
				"继续优化用户注册流程",
				"加强新用户引导",
				"提升用户留存率",
			},
			Timestamp: time.Now(),
		},
		{
			InsightType:     "opportunity",
			InsightCategory: "template_usage",
			Title:           "模板使用优化机会",
			Description:     "发现某些模板类型使用率较低，存在优化空间",
			Impact:          "medium",
			Confidence:      0.72,
			DataPoints: map[string]interface{}{
				"low_usage_templates":    5,
				"optimization_potential": 0.3,
			},
			Recommendations: []string{
				"分析低使用率模板的原因",
				"优化模板分类和推荐算法",
				"增加模板使用引导",
			},
			Timestamp: time.Now(),
		},
	}

	// 保存业务洞察
	for _, insight := range insights {
		dataPointsJSON, _ := json.Marshal(insight.DataPoints)
		recommendationsJSON, _ := json.Marshal(insight.Recommendations)

		insightRecord := BusinessInsight{
			InsightType:     insight.InsightType,
			InsightCategory: insight.InsightCategory,
			Title:           insight.Title,
			Description:     insight.Description,
			Impact:          insight.Impact,
			Confidence:      insight.Confidence,
			DataPoints:      string(dataPointsJSON),
			Recommendations: string(recommendationsJSON),
			Status:          "active",
		}

		if err := s.postgresDB.Create(&insightRecord).Error; err != nil {
			log.Printf("保存业务洞察失败: %v", err)
		}
	}

	return insights, nil
}

// GetSyncStatus 获取同步状态
func (s *StatisticsEnhancedService) GetSyncStatus(entityType string, entityID uint) (*StatisticsSyncStatus, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法获取同步状态")
	}

	var syncStatus StatisticsSyncStatus
	if err := s.postgresDB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).First(&syncStatus).Error; err != nil {
		return nil, fmt.Errorf("获取同步状态失败: %w", err)
	}

	return &syncStatus, nil
}

// Close 关闭数据库连接
func (s *StatisticsEnhancedService) Close() {
	if s.neo4jDriver != nil {
		ctx := context.Background()
		s.neo4jDriver.Close(ctx)
	}
	// GORM会自动关闭DB连接池，Redis客户端也通常不需要手动关闭
}
