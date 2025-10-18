package database

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jManager Neo4j图数据库管理器
type Neo4jManager struct {
	driver neo4j.DriverWithContext
	config Neo4jConfig
}

// Neo4jConfig Neo4j配置
type Neo4jConfig struct {
	URI      string `json:"uri"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// NewNeo4jManager 创建Neo4j管理器
func NewNeo4jManager(config Neo4jConfig) (*Neo4jManager, error) {
	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		neo4j.BasicAuth(config.Username, config.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("Neo4j连接失败: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("Neo4j连接验证失败: %w", err)
	}

	return &Neo4jManager{
		driver: driver,
		config: config,
	}, nil
}

// GetDriver 获取Neo4j驱动
func (nm *Neo4jManager) GetDriver() neo4j.DriverWithContext {
	return nm.driver
}

// Close 关闭连接
func (nm *Neo4jManager) Close(ctx context.Context) error {
	return nm.driver.Close(ctx)
}

// Ping 测试连接
func (nm *Neo4jManager) Ping(ctx context.Context) error {
	return nm.driver.VerifyConnectivity(ctx)
}

// ExecuteQuery 执行查询
func (nm *Neo4jManager) ExecuteQuery(ctx context.Context, query string, parameters map[string]interface{}) ([]map[string]interface{}, error) {
	session := nm.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: nm.config.Database,
	})
	defer session.Close(ctx)

	result, err := session.Run(ctx, query, parameters)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()
		recordMap := make(map[string]interface{})

		for _, key := range record.Keys {
			value, _ := record.Get(key)
			recordMap[key] = value
		}

		records = append(records, recordMap)
	}

	return records, result.Err()
}

// ExecuteWrite 执行写入操作
func (nm *Neo4jManager) ExecuteWrite(ctx context.Context, query string, parameters map[string]interface{}) error {
	session := nm.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: nm.config.Database,
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.Run(ctx, query, parameters)
	return err
}

// CreateNode 创建节点
func (nm *Neo4jManager) CreateNode(ctx context.Context, labels []string, properties map[string]interface{}) (string, error) {
	query := "CREATE (n"

	// 添加标签
	for _, label := range labels {
		query += ":" + label
	}

	// 添加属性
	if len(properties) > 0 {
		query += " $props"
	}

	query += ") RETURN id(n) as node_id"

	parameters := map[string]interface{}{
		"props": properties,
	}

	records, err := nm.ExecuteQuery(ctx, query, parameters)
	if err != nil {
		return "", err
	}

	if len(records) > 0 {
		if nodeID, ok := records[0]["node_id"].(int64); ok {
			return fmt.Sprintf("%d", nodeID), nil
		}
	}

	return "", fmt.Errorf("创建节点失败")
}

// CreateRelationship 创建关系
func (nm *Neo4jManager) CreateRelationship(ctx context.Context, fromNodeID, toNodeID string, relType string, properties map[string]interface{}) error {
	query := fmt.Sprintf("MATCH (a), (b) WHERE id(a) = %s AND id(b) = %s CREATE (a)-[r:%s $props]->(b) RETURN r", fromNodeID, toNodeID, relType)

	parameters := map[string]interface{}{
		"props": properties,
	}

	_, err := nm.ExecuteQuery(ctx, query, parameters)
	return err
}

// FindNodes 查找节点
func (nm *Neo4jManager) FindNodes(ctx context.Context, labels []string, conditions map[string]interface{}) ([]map[string]interface{}, error) {
	query := "MATCH (n"

	// 添加标签
	for _, label := range labels {
		query += ":" + label
	}

	// 添加条件
	if len(conditions) > 0 {
		query += " WHERE "
		first := true
		for key := range conditions {
			if !first {
				query += " AND "
			}
			query += fmt.Sprintf("n.%s = $%s", key, key)
			first = false
		}
	}

	query += ") RETURN n"

	return nm.ExecuteQuery(ctx, query, conditions)
}

// FindRelationships 查找关系
func (nm *Neo4jManager) FindRelationships(ctx context.Context, fromLabels, toLabels []string, relType string) ([]map[string]interface{}, error) {
	query := "MATCH (a"

	// 添加起始节点标签
	for _, label := range fromLabels {
		query += ":" + label
	}

	query += ")-[r"
	if relType != "" {
		query += ":" + relType
	}

	query += "]->(b"

	// 添加目标节点标签
	for _, label := range toLabels {
		query += ":" + label
	}

	query += ") RETURN a, r, b"

	return nm.ExecuteQuery(ctx, query, nil)
}

// DeleteNode 删除节点
func (nm *Neo4jManager) DeleteNode(ctx context.Context, nodeID string) error {
	query := fmt.Sprintf("MATCH (n) WHERE id(n) = %s DETACH DELETE n", nodeID)
	return nm.ExecuteWrite(ctx, query, nil)
}

// DeleteRelationship 删除关系
func (nm *Neo4jManager) DeleteRelationship(ctx context.Context, relID string) error {
	query := fmt.Sprintf("MATCH ()-[r]->() WHERE id(r) = %s DELETE r", relID)
	return nm.ExecuteWrite(ctx, query, nil)
}

// UpdateNode 更新节点
func (nm *Neo4jManager) UpdateNode(ctx context.Context, nodeID string, properties map[string]interface{}) error {
	query := fmt.Sprintf("MATCH (n) WHERE id(n) = %s SET n += $props", nodeID)
	parameters := map[string]interface{}{
		"props": properties,
	}
	return nm.ExecuteWrite(ctx, query, parameters)
}

// GraphSearch 图搜索
func (nm *Neo4jManager) GraphSearch(ctx context.Context, startNodeID string, maxDepth int, relationshipTypes []string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("MATCH (start) WHERE id(start) = %s", startNodeID)

	if len(relationshipTypes) > 0 {
		query += " CALL apoc.path.subgraphAll(start, {"
		query += "maxLevel: " + fmt.Sprintf("%d", maxDepth)
		query += ", relationshipFilter: '"
		for i, relType := range relationshipTypes {
			if i > 0 {
				query += "|"
			}
			query += relType
		}
		query += "'}) YIELD nodes, relationships RETURN nodes, relationships"
	} else {
		query += " MATCH path = (start)-[*1.." + fmt.Sprintf("%d", maxDepth) + "]->(end) RETURN path"
	}

	return nm.ExecuteQuery(ctx, query, nil)
}

// Health 健康检查
func (nm *Neo4jManager) Health() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	pingErr := nm.Ping(ctx)
	status := "healthy"
	if pingErr != nil {
		status = "unhealthy"
	}

	// 获取数据库信息
	var dbInfo map[string]interface{}
	if pingErr == nil {
		records, err := nm.ExecuteQuery(ctx, "CALL db.info()", nil)
		if err == nil && len(records) > 0 {
			dbInfo = records[0]
		}
	}

	return map[string]interface{}{
		"status":   status,
		"uri":      nm.config.URI,
		"database": nm.config.Database,
		"error":    pingErr,
		"info":     dbInfo,
	}
}
