package multidatabase

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ConsistencyChecker 数据一致性检查器
type ConsistencyChecker struct {
	manager *MultiDatabaseManager
	config  *ConsistencyConfig
	results map[string]*ConsistencyResult
	mu      sync.RWMutex
}

// ConsistencyConfig 一致性检查配置
type ConsistencyConfig struct {
	// 检查间隔
	CheckInterval time.Duration `yaml:"check_interval"`

	// 超时时间
	Timeout time.Duration `yaml:"timeout"`

	// 重试次数
	MaxRetries int `yaml:"max_retries"`

	// 检查规则
	Rules []ConsistencyRule `yaml:"rules"`

	// 自动修复
	AutoRepair bool `yaml:"auto_repair"`
}

// ConsistencyRule 一致性检查规则
type ConsistencyRule struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Source      DatabaseType           `yaml:"source"`
	Target      DatabaseType           `yaml:"target"`
	Query       string                 `yaml:"query"`
	Conditions  map[string]interface{} `yaml:"conditions"`
	Enabled     bool                   `yaml:"enabled"`
}

// ConsistencyResult 一致性检查结果
type ConsistencyResult struct {
	RuleID          string            `json:"rule_id"`
	RuleName        string            `json:"rule_name"`
	Status          ConsistencyStatus `json:"status"`
	Message         string            `json:"message"`
	Inconsistencies []Inconsistency   `json:"inconsistencies"`
	CheckedAt       time.Time         `json:"checked_at"`
	Duration        time.Duration     `json:"duration"`
	Repaired        bool              `json:"repaired"`
}

// ConsistencyStatus 一致性状态
type ConsistencyStatus string

const (
	ConsistencyStatusConsistent   ConsistencyStatus = "consistent"
	ConsistencyStatusInconsistent ConsistencyStatus = "inconsistent"
	ConsistencyStatusError        ConsistencyStatus = "error"
	ConsistencyStatusSkipped      ConsistencyStatus = "skipped"
)

// Inconsistency 不一致项
type Inconsistency struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      map[string]interface{} `json:"source"`
	Target      map[string]interface{} `json:"target"`
	Difference  map[string]interface{} `json:"difference"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
}

// NewConsistencyChecker 创建新的一致性检查器
func NewConsistencyChecker(manager *MultiDatabaseManager, config *ConsistencyConfig) *ConsistencyChecker {
	return &ConsistencyChecker{
		manager: manager,
		config:  config,
		results: make(map[string]*ConsistencyResult),
	}
}

// Start 启动一致性检查
func (c *ConsistencyChecker) Start(ctx context.Context) {
	log.Printf("启动数据一致性检查，检查间隔: %v", c.config.CheckInterval)

	ticker := time.NewTicker(c.config.CheckInterval)
	defer ticker.Stop()

	// 立即执行一次检查
	c.performAllChecks()

	for {
		select {
		case <-ctx.Done():
			log.Println("停止数据一致性检查")
			return
		case <-ticker.C:
			c.performAllChecks()
		}
	}
}

// performAllChecks 执行所有一致性检查
func (c *ConsistencyChecker) performAllChecks() {
	log.Println("开始执行数据一致性检查")

	for _, rule := range c.config.Rules {
		if !rule.Enabled {
			continue
		}

		result := c.checkRule(rule)
		c.mu.Lock()
		c.results[rule.ID] = result
		c.mu.Unlock()

		// 如果启用自动修复且发现不一致
		if c.config.AutoRepair && result.Status == ConsistencyStatusInconsistent {
			c.repairInconsistencies(result)
		}
	}

	log.Println("数据一致性检查完成")
}

// checkRule 检查单个规则
func (c *ConsistencyChecker) checkRule(rule ConsistencyRule) *ConsistencyResult {
	startTime := time.Now()
	result := &ConsistencyResult{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		CheckedAt: startTime,
	}

	log.Printf("检查一致性规则: %s (%s)", rule.Name, rule.ID)

	// 根据源数据库和目标数据库类型执行检查
	switch {
	case rule.Source == DatabaseTypeMySQL && rule.Target == DatabaseTypePostgreSQL:
		result = c.checkMySQLToPostgreSQL(rule)
	case rule.Source == DatabaseTypeMySQL && rule.Target == DatabaseTypeNeo4j:
		result = c.checkMySQLToNeo4j(rule)
	case rule.Source == DatabaseTypeMySQL && rule.Target == DatabaseTypeRedis:
		result = c.checkMySQLToRedis(rule)
	case rule.Source == DatabaseTypePostgreSQL && rule.Target == DatabaseTypeNeo4j:
		result = c.checkPostgreSQLToNeo4j(rule)
	case rule.Source == DatabaseTypePostgreSQL && rule.Target == DatabaseTypeRedis:
		result = c.checkPostgreSQLToRedis(rule)
	case rule.Source == DatabaseTypeNeo4j && rule.Target == DatabaseTypeRedis:
		result = c.checkNeo4jToRedis(rule)
	default:
		result.Status = ConsistencyStatusSkipped
		result.Message = fmt.Sprintf("不支持的检查类型: %s -> %s", rule.Source, rule.Target)
	}

	result.Duration = time.Since(startTime)
	return result
}

// checkMySQLToPostgreSQL 检查MySQL到PostgreSQL的一致性
func (c *ConsistencyChecker) checkMySQLToPostgreSQL(rule ConsistencyRule) *ConsistencyResult {
	result := &ConsistencyResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ConsistencyStatusError,
	}

	if c.manager.MySQL == nil || c.manager.PostgreSQL == nil {
		result.Message = "MySQL或PostgreSQL连接未初始化"
		return result
	}

	// 执行MySQL查询
	var mysqlResults []map[string]interface{}
	if err := c.manager.MySQL.Raw(rule.Query).Scan(&mysqlResults).Error; err != nil {
		result.Message = fmt.Sprintf("MySQL查询失败: %v", err)
		return result
	}

	// 执行PostgreSQL查询
	var postgresResults []map[string]interface{}
	if err := c.manager.PostgreSQL.Raw(rule.Query).Scan(&postgresResults).Error; err != nil {
		result.Message = fmt.Sprintf("PostgreSQL查询失败: %v", err)
		return result
	}

	// 比较结果
	inconsistencies := c.compareResults(mysqlResults, postgresResults, rule)
	if len(inconsistencies) == 0 {
		result.Status = ConsistencyStatusConsistent
		result.Message = "数据一致"
	} else {
		result.Status = ConsistencyStatusInconsistent
		result.Message = fmt.Sprintf("发现 %d 个不一致项", len(inconsistencies))
		result.Inconsistencies = inconsistencies
	}

	return result
}

// checkMySQLToNeo4j 检查MySQL到Neo4j的一致性
func (c *ConsistencyChecker) checkMySQLToNeo4j(rule ConsistencyRule) *ConsistencyResult {
	result := &ConsistencyResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ConsistencyStatusError,
	}

	if c.manager.MySQL == nil || c.manager.Neo4j == nil {
		result.Message = "MySQL或Neo4j连接未初始化"
		return result
	}

	// 执行MySQL查询
	var mysqlResults []map[string]interface{}
	if err := c.manager.MySQL.Raw(rule.Query).Scan(&mysqlResults).Error; err != nil {
		result.Message = fmt.Sprintf("MySQL查询失败: %v", err)
		return result
	}

	// 执行Neo4j查询
	session := c.manager.Neo4j.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	neo4jResults, err := session.Run(rule.Query, nil)
	if err != nil {
		result.Message = fmt.Sprintf("Neo4j查询失败: %v", err)
		return result
	}

	// 转换Neo4j结果为map
	var neo4jMaps []map[string]interface{}
	for neo4jResults.Next() {
		record := neo4jResults.Record()
		recordMap := make(map[string]interface{})
		for _, key := range record.Keys {
			recordMap[key] = record.AsMap()[key]
		}
		neo4jMaps = append(neo4jMaps, recordMap)
	}

	// 比较结果
	inconsistencies := c.compareResults(mysqlResults, neo4jMaps, rule)
	if len(inconsistencies) == 0 {
		result.Status = ConsistencyStatusConsistent
		result.Message = "数据一致"
	} else {
		result.Status = ConsistencyStatusInconsistent
		result.Message = fmt.Sprintf("发现 %d 个不一致项", len(inconsistencies))
		result.Inconsistencies = inconsistencies
	}

	return result
}

// checkMySQLToRedis 检查MySQL到Redis的一致性
func (c *ConsistencyChecker) checkMySQLToRedis(rule ConsistencyRule) *ConsistencyResult {
	result := &ConsistencyResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ConsistencyStatusError,
	}

	if c.manager.MySQL == nil || c.manager.Redis == nil {
		result.Message = "MySQL或Redis连接未初始化"
		return result
	}

	// 执行MySQL查询
	var mysqlResults []map[string]interface{}
	if err := c.manager.MySQL.Raw(rule.Query).Scan(&mysqlResults).Error; err != nil {
		result.Message = fmt.Sprintf("MySQL查询失败: %v", err)
		return result
	}

	// 检查Redis中的对应数据
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	var inconsistencies []Inconsistency
	for _, mysqlRecord := range mysqlResults {
		// 根据规则生成Redis键
		redisKey := c.generateRedisKey(rule, mysqlRecord)

		// 检查Redis中是否存在该键
		exists, err := c.manager.Redis.Exists(ctx, redisKey).Result()
		if err != nil {
			inconsistencies = append(inconsistencies, Inconsistency{
				ID:          fmt.Sprintf("%v", mysqlRecord["id"]),
				Type:        "redis_missing",
				Source:      mysqlRecord,
				Target:      nil,
				Severity:    "high",
				Description: fmt.Sprintf("Redis中缺少键: %s", redisKey),
			})
			continue
		}

		if exists == 0 {
			inconsistencies = append(inconsistencies, Inconsistency{
				ID:          fmt.Sprintf("%v", mysqlRecord["id"]),
				Type:        "redis_missing",
				Source:      mysqlRecord,
				Target:      nil,
				Severity:    "high",
				Description: fmt.Sprintf("Redis中缺少键: %s", redisKey),
			})
		}
	}

	if len(inconsistencies) == 0 {
		result.Status = ConsistencyStatusConsistent
		result.Message = "数据一致"
	} else {
		result.Status = ConsistencyStatusInconsistent
		result.Message = fmt.Sprintf("发现 %d 个不一致项", len(inconsistencies))
		result.Inconsistencies = inconsistencies
	}

	return result
}

// checkPostgreSQLToNeo4j 检查PostgreSQL到Neo4j的一致性
func (c *ConsistencyChecker) checkPostgreSQLToNeo4j(rule ConsistencyRule) *ConsistencyResult {
	// 实现PostgreSQL到Neo4j的一致性检查
	return &ConsistencyResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ConsistencyStatusSkipped,
		Message:  "PostgreSQL到Neo4j检查待实现",
	}
}

// checkPostgreSQLToRedis 检查PostgreSQL到Redis的一致性
func (c *ConsistencyChecker) checkPostgreSQLToRedis(rule ConsistencyRule) *ConsistencyResult {
	// 实现PostgreSQL到Redis的一致性检查
	return &ConsistencyResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ConsistencyStatusSkipped,
		Message:  "PostgreSQL到Redis检查待实现",
	}
}

// checkNeo4jToRedis 检查Neo4j到Redis的一致性
func (c *ConsistencyChecker) checkNeo4jToRedis(rule ConsistencyRule) *ConsistencyResult {
	// 实现Neo4j到Redis的一致性检查
	return &ConsistencyResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Status:   ConsistencyStatusSkipped,
		Message:  "Neo4j到Redis检查待实现",
	}
}

// compareResults 比较两个结果集
func (c *ConsistencyChecker) compareResults(source, target []map[string]interface{}, rule ConsistencyRule) []Inconsistency {
	var inconsistencies []Inconsistency

	// 创建目标数据的索引
	targetIndex := make(map[string]map[string]interface{})
	for _, record := range target {
		if id, ok := record["id"]; ok {
			targetIndex[fmt.Sprintf("%v", id)] = record
		}
	}

	// 检查源数据中的每一项是否在目标中存在
	for _, sourceRecord := range source {
		if id, ok := sourceRecord["id"]; ok {
			idStr := fmt.Sprintf("%v", id)
			if targetRecord, exists := targetIndex[idStr]; exists {
				// 比较字段值
				if diff := c.compareRecords(sourceRecord, targetRecord); len(diff) > 0 {
					inconsistencies = append(inconsistencies, Inconsistency{
						ID:          idStr,
						Type:        "field_mismatch",
						Source:      sourceRecord,
						Target:      targetRecord,
						Difference:  diff,
						Severity:    "medium",
						Description: "字段值不匹配",
					})
				}
			} else {
				inconsistencies = append(inconsistencies, Inconsistency{
					ID:          idStr,
					Type:        "missing_in_target",
					Source:      sourceRecord,
					Target:      nil,
					Severity:    "high",
					Description: "目标数据库中缺少记录",
				})
			}
		}
	}

	return inconsistencies
}

// compareRecords 比较两个记录
func (c *ConsistencyChecker) compareRecords(source, target map[string]interface{}) map[string]interface{} {
	differences := make(map[string]interface{})

	for key, sourceValue := range source {
		if targetValue, exists := target[key]; exists {
			if sourceValue != targetValue {
				differences[key] = map[string]interface{}{
					"source": sourceValue,
					"target": targetValue,
				}
			}
		}
	}

	return differences
}

// generateRedisKey 生成Redis键
func (c *ConsistencyChecker) generateRedisKey(rule ConsistencyRule, record map[string]interface{}) string {
	// 根据规则和记录生成Redis键
	// 这里可以根据实际需求实现更复杂的键生成逻辑
	if id, ok := record["id"]; ok {
		return fmt.Sprintf("%s:%s:%v", rule.Source, rule.ID, id)
	}
	return fmt.Sprintf("%s:%s", rule.Source, rule.ID)
}

// repairInconsistencies 修复不一致项
func (c *ConsistencyChecker) repairInconsistencies(result *ConsistencyResult) {
	log.Printf("开始修复不一致项: %s", result.RuleName)

	for _, inconsistency := range result.Inconsistencies {
		if err := c.repairInconsistency(inconsistency); err != nil {
			log.Printf("修复不一致项失败: %s, 错误: %v", inconsistency.ID, err)
		} else {
			log.Printf("成功修复不一致项: %s", inconsistency.ID)
		}
	}

	result.Repaired = true
}

// repairInconsistency 修复单个不一致项
func (c *ConsistencyChecker) repairInconsistency(inconsistency Inconsistency) error {
	// 根据不一致类型执行修复
	switch inconsistency.Type {
	case "redis_missing":
		return c.repairRedisMissing(inconsistency)
	case "missing_in_target":
		return c.repairMissingInTarget(inconsistency)
	case "field_mismatch":
		return c.repairFieldMismatch(inconsistency)
	default:
		return fmt.Errorf("不支持的不一致类型: %s", inconsistency.Type)
	}
}

// repairRedisMissing 修复Redis中缺失的数据
func (c *ConsistencyChecker) repairRedisMissing(inconsistency Inconsistency) error {
	// 实现Redis数据修复逻辑
	return fmt.Errorf("Redis数据修复待实现")
}

// repairMissingInTarget 修复目标数据库中缺失的数据
func (c *ConsistencyChecker) repairMissingInTarget(inconsistency Inconsistency) error {
	// 实现目标数据库数据修复逻辑
	return fmt.Errorf("目标数据库数据修复待实现")
}

// repairFieldMismatch 修复字段不匹配
func (c *ConsistencyChecker) repairFieldMismatch(inconsistency Inconsistency) error {
	// 实现字段不匹配修复逻辑
	return fmt.Errorf("字段不匹配修复待实现")
}

// GetResults 获取所有检查结果
func (c *ConsistencyChecker) GetResults() map[string]*ConsistencyResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[string]*ConsistencyResult)
	for k, v := range c.results {
		results[k] = v
	}
	return results
}

// GetResult 获取特定规则的检查结果
func (c *ConsistencyChecker) GetResult(ruleID string) (*ConsistencyResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, exists := c.results[ruleID]
	return result, exists
}
