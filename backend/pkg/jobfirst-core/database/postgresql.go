package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgreSQLManager PostgreSQL数据库管理器
type PostgreSQLManager struct {
	db     *gorm.DB
	config PostgreSQLConfig
}

// PostgreSQLConfig PostgreSQL配置
type PostgreSQLConfig struct {
	Host        string          `json:"host"`
	Port        int             `json:"port"`
	Username    string          `json:"username"`
	Password    string          `json:"password"`
	Database    string          `json:"database"`
	SSLMode     string          `json:"ssl_mode"`
	MaxIdle     int             `json:"max_idle"`
	MaxOpen     int             `json:"max_open"`
	MaxLifetime time.Duration   `json:"max_lifetime"`
	LogLevel    logger.LogLevel `json:"log_level"`
}

// NewPostgreSQLManager 创建PostgreSQL管理器
func NewPostgreSQLManager(config PostgreSQLConfig) (*PostgreSQLManager, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL连接失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取PostgreSQL实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	manager := &PostgreSQLManager{
		db:     db,
		config: config,
	}

	return manager, nil
}

// GetDB 获取数据库实例
func (pm *PostgreSQLManager) GetDB() *gorm.DB {
	return pm.db
}

// Close 关闭数据库连接
func (pm *PostgreSQLManager) Close() error {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping 测试数据库连接
func (pm *PostgreSQLManager) Ping() error {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Migrate 执行数据库迁移
func (pm *PostgreSQLManager) Migrate(models ...interface{}) error {
	return pm.db.AutoMigrate(models...)
}

// Transaction 执行事务
func (pm *PostgreSQLManager) Transaction(fn func(*gorm.DB) error) error {
	return pm.db.Transaction(fn)
}

// Create 创建记录
func (pm *PostgreSQLManager) Create(value interface{}) error {
	return pm.db.Create(value).Error
}

// First 查找第一条记录
func (pm *PostgreSQLManager) First(dest interface{}, conds ...interface{}) error {
	return pm.db.First(dest, conds...).Error
}

// Find 查找记录
func (pm *PostgreSQLManager) Find(dest interface{}, conds ...interface{}) error {
	return pm.db.Find(dest, conds...).Error
}

// Update 更新记录
func (pm *PostgreSQLManager) Update(column string, value interface{}) error {
	return pm.db.Update(column, value).Error
}

// Delete 删除记录
func (pm *PostgreSQLManager) Delete(value interface{}, conds ...interface{}) error {
	return pm.db.Delete(value, conds...).Error
}

// Raw 执行原生SQL
func (pm *PostgreSQLManager) Raw(sql string, values ...interface{}) *gorm.DB {
	return pm.db.Raw(sql, values...)
}

// Exec 执行SQL
func (pm *PostgreSQLManager) Exec(sql string, values ...interface{}) error {
	return pm.db.Exec(sql, values...).Error
}

// Health 健康检查
func (pm *PostgreSQLManager) Health() map[string]interface{} {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"status":               "healthy",
		"host":                 pm.config.Host,
		"port":                 pm.config.Port,
		"database":             pm.config.Database,
		"max_open_conns":       stats.MaxOpenConnections,
		"open_conns":           stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// CreateVectorExtension 创建向量扩展
func (pm *PostgreSQLManager) CreateVectorExtension() error {
	return pm.db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
}

// CreateVectorIndex 创建向量索引
func (pm *PostgreSQLManager) CreateVectorIndex(tableName, columnName, indexName string) error {
	sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING ivfflat (%s vector_cosine_ops)",
		indexName, tableName, columnName)
	return pm.db.Exec(sql).Error
}

// VectorSearch 向量搜索
func (pm *PostgreSQLManager) VectorSearch(tableName, columnName string, queryVector []float64, limit int) (*gorm.DB, error) {
	// 构建向量搜索SQL
	sql := fmt.Sprintf("SELECT *, %s <=> ? as distance FROM %s ORDER BY %s <=> ? LIMIT ?",
		columnName, tableName, columnName)

	// 将查询向量转换为PostgreSQL向量格式
	vectorStr := fmt.Sprintf("[%f", queryVector[0])
	for i := 1; i < len(queryVector); i++ {
		vectorStr += fmt.Sprintf(",%f", queryVector[i])
	}
	vectorStr += "]"

	return pm.db.Raw(sql, vectorStr, vectorStr, limit), nil
}
