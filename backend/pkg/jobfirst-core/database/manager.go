package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	// MySQL配置
	MySQL MySQLConfig `json:"mysql"`
	// Redis配置
	Redis RedisConfig `json:"redis"`
	// PostgreSQL配置
	PostgreSQL PostgreSQLConfig `json:"postgresql"`
	// Neo4j配置
	Neo4j Neo4jConfig `json:"neo4j"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host        string          `json:"host"`
	Port        int             `json:"port"`
	Username    string          `json:"username"`
	Password    string          `json:"password"`
	Database    string          `json:"database"`
	Charset     string          `json:"charset"`
	MaxIdle     int             `json:"max_idle"`
	MaxOpen     int             `json:"max_open"`
	MaxLifetime time.Duration   `json:"max_lifetime"`
	LogLevel    logger.LogLevel `json:"log_level"`
}

// Manager 统一数据库管理器
type Manager struct {
	MySQL      *MySQLManager
	Redis      *RedisManager
	PostgreSQL *PostgreSQLManager
	Neo4j      *Neo4jManager
	config     Config
}

// NewManager 创建统一数据库管理器
func NewManager(config Config) (*Manager, error) {
	manager := &Manager{
		config: config,
	}

	// 初始化MySQL
	if config.MySQL.Host != "" {
		mysqlManager, err := NewMySQLManager(config.MySQL)
		if err != nil {
			return nil, fmt.Errorf("初始化MySQL失败: %w", err)
		}
		manager.MySQL = mysqlManager
	}

	// 初始化Redis
	if config.Redis.Host != "" {
		redisManager, err := NewRedisManager(config.Redis)
		if err != nil {
			return nil, fmt.Errorf("初始化Redis失败: %w", err)
		}
		manager.Redis = redisManager
	}

	// 初始化PostgreSQL
	if config.PostgreSQL.Host != "" {
		postgresManager, err := NewPostgreSQLManager(config.PostgreSQL)
		if err != nil {
			return nil, fmt.Errorf("初始化PostgreSQL失败: %w", err)
		}
		manager.PostgreSQL = postgresManager
	}

	// 初始化Neo4j
	if config.Neo4j.URI != "" {
		neo4jManager, err := NewNeo4jManager(config.Neo4j)
		if err != nil {
			return nil, fmt.Errorf("初始化Neo4j失败: %w", err)
		}
		manager.Neo4j = neo4jManager
	}

	return manager, nil
}

// GetDB 获取MySQL数据库实例（向后兼容）
func (dm *Manager) GetDB() *gorm.DB {
	if dm.MySQL != nil {
		return dm.MySQL.GetDB()
	}
	return nil
}

// GetMySQL 获取MySQL管理器
func (dm *Manager) GetMySQL() *MySQLManager {
	return dm.MySQL
}

// GetRedis 获取Redis管理器
func (dm *Manager) GetRedis() *RedisManager {
	return dm.Redis
}

// GetPostgreSQL 获取PostgreSQL管理器
func (dm *Manager) GetPostgreSQL() *PostgreSQLManager {
	return dm.PostgreSQL
}

// GetNeo4j 获取Neo4j管理器
func (dm *Manager) GetNeo4j() *Neo4jManager {
	return dm.Neo4j
}

// Close 关闭所有数据库连接
func (dm *Manager) Close() error {
	var errors []error

	if dm.MySQL != nil {
		if err := dm.MySQL.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭MySQL失败: %w", err))
		}
	}

	if dm.Redis != nil {
		if err := dm.Redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭Redis失败: %w", err))
		}
	}

	if dm.PostgreSQL != nil {
		if err := dm.PostgreSQL.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭PostgreSQL失败: %w", err))
		}
	}

	if dm.Neo4j != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := dm.Neo4j.Close(ctx); err != nil {
			errors = append(errors, fmt.Errorf("关闭Neo4j失败: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("关闭数据库连接时发生错误: %v", errors)
	}

	return nil
}

// Ping 测试所有数据库连接
func (dm *Manager) Ping() error {
	var errors []error

	if dm.MySQL != nil {
		if err := dm.MySQL.Ping(); err != nil {
			errors = append(errors, fmt.Errorf("MySQL连接失败: %w", err))
		}
	}

	if dm.Redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := dm.Redis.Ping(ctx); err != nil {
			errors = append(errors, fmt.Errorf("Redis连接失败: %w", err))
		}
	}

	if dm.PostgreSQL != nil {
		if err := dm.PostgreSQL.Ping(); err != nil {
			errors = append(errors, fmt.Errorf("PostgreSQL连接失败: %w", err))
		}
	}

	if dm.Neo4j != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := dm.Neo4j.Ping(ctx); err != nil {
			errors = append(errors, fmt.Errorf("Neo4j连接失败: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("数据库连接测试失败: %v", errors)
	}

	return nil
}

// Migrate 执行数据库迁移（MySQL）
func (dm *Manager) Migrate(models ...interface{}) error {
	if dm.MySQL == nil {
		return fmt.Errorf("MySQL未初始化")
	}
	return dm.MySQL.Migrate(models...)
}

// Transaction 执行事务（MySQL）
func (dm *Manager) Transaction(fn func(*gorm.DB) error) error {
	if dm.MySQL == nil {
		return fmt.Errorf("MySQL未初始化")
	}
	return dm.MySQL.Transaction(fn)
}

// MultiDBTransaction 多数据库事务
type MultiDBTransaction struct {
	MySQL      *gorm.DB
	Redis      *RedisManager
	PostgreSQL *gorm.DB
	Neo4j      *Neo4jManager
}

// Transaction 执行多数据库事务
func (dm *Manager) MultiDBTransaction(fn func(*MultiDBTransaction) error) error {
	tx := &MultiDBTransaction{}

	// 准备MySQL事务
	if dm.MySQL != nil {
		mysqlTx := dm.MySQL.GetDB().Begin()
		if mysqlTx.Error != nil {
			return fmt.Errorf("开始MySQL事务失败: %w", mysqlTx.Error)
		}
		tx.MySQL = mysqlTx
		defer func() {
			if r := recover(); r != nil {
				if tx.MySQL != nil {
					tx.MySQL.Rollback()
				}
				panic(r)
			}
		}()
	}

	// 准备其他数据库连接
	if dm.Redis != nil {
		tx.Redis = dm.Redis
	}
	if dm.PostgreSQL != nil {
		tx.PostgreSQL = dm.PostgreSQL.GetDB()
	}
	if dm.Neo4j != nil {
		tx.Neo4j = dm.Neo4j
	}

	// 执行事务函数
	if err := fn(tx); err != nil {
		// 回滚MySQL事务
		if tx.MySQL != nil {
			tx.MySQL.Rollback()
		}
		return err
	}

	// 提交MySQL事务
	if tx.MySQL != nil {
		if err := tx.MySQL.Commit().Error; err != nil {
			return fmt.Errorf("提交MySQL事务失败: %w", err)
		}
	}

	return nil
}

// Health 健康检查
func (dm *Manager) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// 检查MySQL
	if dm.MySQL != nil {
		health["mysql"] = dm.MySQL.Health()
	}

	// 检查Redis
	if dm.Redis != nil {
		health["redis"] = dm.Redis.Health()
	}

	// 检查PostgreSQL
	if dm.PostgreSQL != nil {
		health["postgresql"] = dm.PostgreSQL.Health()
	}

	// 检查Neo4j
	if dm.Neo4j != nil {
		health["neo4j"] = dm.Neo4j.Health()
	}

	// 计算总体状态
	allHealthy := true
	for dbType, dbHealth := range health {
		if dbType == "status" || dbType == "timestamp" {
			continue
		}
		if healthMap, ok := dbHealth.(map[string]interface{}); ok {
			if status, exists := healthMap["status"]; exists && status != "healthy" {
				allHealthy = false
				break
			}
		}
	}

	if !allHealthy {
		health["status"] = "unhealthy"
	}

	return health
}
