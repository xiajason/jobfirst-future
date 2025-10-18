package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jobfirst/jobfirst-core"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TemplateEnhancedService 模板增强服务
type TemplateEnhancedService struct {
	core        *jobfirst.Core
	mysqlDB     *gorm.DB
	postgresDB  *gorm.DB
	neo4jDriver neo4j.DriverWithContext
	redisClient *redis.Client
}

// NewTemplateEnhancedService 创建模板增强服务
func NewTemplateEnhancedService(core *jobfirst.Core) (*TemplateEnhancedService, error) {
	service := &TemplateEnhancedService{
		core: core,
	}

	// 获取MySQL连接
	service.mysqlDB = core.GetDB()

	// 初始化PostgreSQL连接
	postgresDSN := "host=localhost user=postgres password= dbname=jobfirst port=5432 sslmode=disable"
	postgresDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Printf("PostgreSQL连接失败: %v", err)
		// 继续运行，但不提供向量化功能
	} else {
		service.postgresDB = postgresDB
		// 创建向量表
		service.createVectorTables()
	}

	// 初始化Neo4j连接
	neo4jDriver, err := neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		log.Printf("Neo4j连接失败: %v", err)
		// 继续运行，但不提供关系网络功能
	} else {
		service.neo4jDriver = neo4jDriver
		// 创建关系网络索引
		service.createRelationshipIndexes()
	}

	// 获取Redis连接
	redisManager := core.Database.GetRedis()
	if redisManager != nil {
		service.redisClient = redisManager.GetClient()
	}

	return service, nil
}

// createVectorTables 创建向量表
func (s *TemplateEnhancedService) createVectorTables() {
	if s.postgresDB == nil {
		return
	}

	// 创建模板向量表
	s.postgresDB.Exec(`
		CREATE TABLE IF NOT EXISTS template_vectors (
			id SERIAL PRIMARY KEY,
			template_id INTEGER NOT NULL,
			content_vector VECTOR(384),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	// 创建向量索引
	s.postgresDB.Exec(`
		CREATE INDEX IF NOT EXISTS template_vectors_template_id_idx 
		ON template_vectors(template_id)
	`)

	s.postgresDB.Exec(`
		CREATE INDEX IF NOT EXISTS template_vectors_content_vector_idx 
		ON template_vectors USING ivfflat (content_vector vector_cosine_ops)
	`)
}

// createRelationshipIndexes 创建关系网络索引
func (s *TemplateEnhancedService) createRelationshipIndexes() {
	if s.neo4jDriver == nil {
		return
	}

	ctx := context.Background()
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// 创建模板节点索引
	session.Run(ctx, "CREATE INDEX template_id_index IF NOT EXISTS FOR (t:Template) ON (t.id)", nil)
	session.Run(ctx, "CREATE INDEX template_category_index IF NOT EXISTS FOR (t:Template) ON (t.category)", nil)

	// 创建关系索引
	session.Run(ctx, "CREATE INDEX relationship_type_index IF NOT EXISTS FOR ()-[r:RELATED_TO]-() ON (r.type)", nil)
}

// TemplateVector 模板向量模型
type TemplateVector struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TemplateID    uint      `json:"template_id" gorm:"not null"`
	ContentVector []float64 `json:"content_vector" gorm:"type:vector(384)"`
	Metadata      string    `json:"metadata" gorm:"type:jsonb"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TemplateRelationship 模板关系模型
type TemplateRelationship struct {
	ID           uint    `json:"id"`
	SourceID     uint    `json:"source_id"`
	TargetID     uint    `json:"target_id"`
	Relationship string  `json:"relationship"`
	Weight       float64 `json:"weight"`
	Metadata     string  `json:"metadata"`
}

// GenerateTemplateVector 生成模板向量
func (s *TemplateEnhancedService) GenerateTemplateVector(templateID uint, content string) error {
	if s.postgresDB == nil {
		return fmt.Errorf("PostgreSQL未连接，无法生成向量")
	}

	// 调用AI服务生成向量
	vector, err := s.callAIServiceForVector(content)
	if err != nil {
		return fmt.Errorf("生成向量失败: %v", err)
	}

	// 保存向量到PostgreSQL
	templateVector := TemplateVector{
		TemplateID:    templateID,
		ContentVector: vector,
		Metadata:      `{"content_length": ` + fmt.Sprintf("%d", len(content)) + `}`,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 检查是否已存在
	var existing TemplateVector
	if err := s.postgresDB.Where("template_id = ?", templateID).First(&existing).Error; err == nil {
		// 更新现有向量
		existing.ContentVector = vector
		existing.UpdatedAt = time.Now()
		return s.postgresDB.Save(&existing).Error
	} else {
		// 创建新向量
		return s.postgresDB.Create(&templateVector).Error
	}
}

// callAIServiceForVector 调用AI服务生成向量
func (s *TemplateEnhancedService) callAIServiceForVector(content string) ([]float64, error) {
	// 这里应该调用容器化AI服务的embedding API
	// 为了演示，我们返回一个模拟向量
	vector := make([]float64, 384)
	for i := range vector {
		vector[i] = float64(i%100) / 100.0
	}
	return vector, nil
}

// CreateTemplateRelationship 创建模板关系
func (s *TemplateEnhancedService) CreateTemplateRelationship(sourceID, targetID uint, relationship string, weight float64) error {
	if s.neo4jDriver == nil {
		return fmt.Errorf("Neo4j未连接，无法创建关系")
	}

	ctx := context.Background()
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// 创建关系
	_, err := session.Run(ctx, `
		MATCH (source:Template {id: $sourceID})
		MATCH (target:Template {id: $targetID})
		MERGE (source)-[r:RELATED_TO {type: $relationship}]->(target)
		SET r.weight = $weight, r.created_at = datetime()
	`, map[string]interface{}{
		"sourceID":     sourceID,
		"targetID":     targetID,
		"relationship": relationship,
		"weight":       weight,
	})

	return err
}

// GetSimilarTemplates 获取相似模板
func (s *TemplateEnhancedService) GetSimilarTemplates(templateID uint, limit int) ([]Template, error) {
	if s.postgresDB == nil {
		return nil, fmt.Errorf("PostgreSQL未连接，无法获取相似模板")
	}

	// 获取目标模板的向量
	var targetVector TemplateVector
	if err := s.postgresDB.Where("template_id = ?", templateID).First(&targetVector).Error; err != nil {
		return nil, fmt.Errorf("模板向量不存在: %v", err)
	}

	// 使用向量相似度搜索
	var similarVectors []TemplateVector
	err := s.postgresDB.Raw(`
		SELECT tv.*, 1 - (tv.content_vector <=> $1) as similarity
		FROM template_vectors tv
		WHERE tv.template_id != $2
		ORDER BY similarity DESC
		LIMIT $3
	`, targetVector.ContentVector, templateID, limit).Scan(&similarVectors).Error

	if err != nil {
		return nil, fmt.Errorf("相似度搜索失败: %v", err)
	}

	// 获取模板详细信息
	var templates []Template
	for _, vector := range similarVectors {
		var template Template
		if err := s.mysqlDB.First(&template, vector.TemplateID).Error; err == nil {
			templates = append(templates, template)
		}
	}

	return templates, nil
}

// GetTemplateRelationships 获取模板关系
func (s *TemplateEnhancedService) GetTemplateRelationships(templateID uint) ([]TemplateRelationship, error) {
	if s.neo4jDriver == nil {
		return nil, fmt.Errorf("Neo4j未连接，无法获取关系")
	}

	ctx := context.Background()
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (source:Template {id: $templateID})-[r:RELATED_TO]->(target:Template)
		RETURN target.id as target_id, r.type as relationship, r.weight as weight
		ORDER BY r.weight DESC
	`, map[string]interface{}{
		"templateID": templateID,
	})

	if err != nil {
		return nil, err
	}

	var relationships []TemplateRelationship
	for result.Next(ctx) {
		record := result.Record()
		rel := TemplateRelationship{
			SourceID:     templateID,
			TargetID:     uint(record.Values[0].(int64)),
			Relationship: record.Values[1].(string),
			Weight:       record.Values[2].(float64),
		}
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

// SyncTemplateToAllDatabases 同步模板到所有数据库
func (s *TemplateEnhancedService) SyncTemplateToAllDatabases(template *Template) error {
	// 1. 同步到MySQL (主数据库)
	if err := s.mysqlDB.Save(template).Error; err != nil {
		return fmt.Errorf("MySQL同步失败: %v", err)
	}

	// 2. 生成并同步向量到PostgreSQL
	if s.postgresDB != nil {
		if err := s.GenerateTemplateVector(template.ID, template.Content); err != nil {
			log.Printf("PostgreSQL向量同步失败: %v", err)
		}
	}

	// 3. 同步到Neo4j
	if s.neo4jDriver != nil {
		if err := s.syncTemplateToNeo4j(template); err != nil {
			log.Printf("Neo4j同步失败: %v", err)
		}
	}

	// 4. 同步到Redis缓存
	if s.redisClient != nil {
		if err := s.syncTemplateToRedis(template); err != nil {
			log.Printf("Redis同步失败: %v", err)
		}
	}

	return nil
}

// syncTemplateToNeo4j 同步模板到Neo4j
func (s *TemplateEnhancedService) syncTemplateToNeo4j(template *Template) error {
	ctx := context.Background()
	session := s.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.Run(ctx, `
		MERGE (t:Template {id: $id})
		SET t.name = $name,
		    t.category = $category,
		    t.description = $description,
		    t.usage = $usage,
		    t.rating = $rating,
		    t.created_by = $created_by,
		    t.updated_at = datetime()
	`, map[string]interface{}{
		"id":          template.ID,
		"name":        template.Name,
		"category":    template.Category,
		"description": template.Description,
		"usage":       template.Usage,
		"rating":      template.Rating,
		"created_by":  template.CreatedBy,
	})

	return err
}

// syncTemplateToRedis 同步模板到Redis
func (s *TemplateEnhancedService) syncTemplateToRedis(template *Template) error {
	ctx := context.Background()

	// 缓存模板基本信息
	templateData, err := json.Marshal(template)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("template:%d", template.ID)
	return s.redisClient.Set(ctx, key, templateData, 24*time.Hour).Err()
}

// GetCachedTemplate 从Redis获取缓存的模板
func (s *TemplateEnhancedService) GetCachedTemplate(templateID uint) (*Template, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis未连接")
	}

	ctx := context.Background()
	key := fmt.Sprintf("template:%d", templateID)

	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var template Template
	if err := json.Unmarshal([]byte(data), &template); err != nil {
		return nil, err
	}

	return &template, nil
}

// AnalyzeTemplateUsage 分析模板使用情况
func (s *TemplateEnhancedService) AnalyzeTemplateUsage(templateID uint) (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// 从MySQL获取基础统计
	var template Template
	if err := s.mysqlDB.First(&template, templateID).Error; err != nil {
		return nil, err
	}

	analysis["basic_stats"] = map[string]interface{}{
		"usage":  template.Usage,
		"rating": template.Rating,
		"name":   template.Name,
	}

	// 从Neo4j获取关系分析
	if s.neo4jDriver != nil {
		relationships, err := s.GetTemplateRelationships(templateID)
		if err == nil {
			analysis["relationships"] = relationships
			analysis["relationship_count"] = len(relationships)
		}
	}

	// 从PostgreSQL获取向量分析
	if s.postgresDB != nil {
		var vector TemplateVector
		if err := s.postgresDB.Where("template_id = ?", templateID).First(&vector).Error; err == nil {
			analysis["has_vector"] = true
			analysis["vector_dimension"] = len(vector.ContentVector)
		} else {
			analysis["has_vector"] = false
		}
	}

	return analysis, nil
}

// Close 关闭服务连接
func (s *TemplateEnhancedService) Close() {
	if s.neo4jDriver != nil {
		s.neo4jDriver.Close(context.Background())
	}
}
