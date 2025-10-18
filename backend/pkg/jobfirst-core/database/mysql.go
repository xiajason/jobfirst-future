package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLManager MySQL数据库管理器
type MySQLManager struct {
	db     *gorm.DB
	config MySQLConfig
}

// NewMySQLManager 创建MySQL管理器
func NewMySQLManager(config MySQLConfig) (*MySQLManager, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(config.LogLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("MySQL连接失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取MySQL实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	return &MySQLManager{
		db:     db,
		config: config,
	}, nil
}

// GetDB 获取数据库实例
func (mm *MySQLManager) GetDB() *gorm.DB {
	return mm.db
}

// Close 关闭数据库连接
func (mm *MySQLManager) Close() error {
	sqlDB, err := mm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping 测试数据库连接
func (mm *MySQLManager) Ping() error {
	sqlDB, err := mm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Migrate 执行数据库迁移（安全模式）
func (mm *MySQLManager) Migrate(models ...interface{}) error {
	// 检查表是否存在，如果存在则跳过迁移
	for _, model := range models {
		stmt := &gorm.Statement{DB: mm.db}
		if err := stmt.Parse(model); err != nil {
			return fmt.Errorf("解析模型失败: %w", err)
		}

		// 检查表是否存在
		if mm.db.Migrator().HasTable(stmt.Schema.Table) {
			// 表已存在，只添加缺失的列，不修改现有约束
			if err := mm.db.Migrator().AutoMigrate(model); err != nil {
				// 如果迁移失败，记录错误但不中断
				fmt.Printf("警告: 表 %s 迁移失败: %v\n", stmt.Schema.Table, err)
			}
		} else {
			// 表不存在，正常创建
			if err := mm.db.Migrator().CreateTable(model); err != nil {
				return fmt.Errorf("创建表失败: %w", err)
			}
		}
	}
	return nil
}

// Transaction 执行事务
func (mm *MySQLManager) Transaction(fn func(*gorm.DB) error) error {
	return mm.db.Transaction(fn)
}

// Create 创建记录
func (mm *MySQLManager) Create(value interface{}) error {
	return mm.db.Create(value).Error
}

// First 查找第一条记录
func (mm *MySQLManager) First(dest interface{}, conds ...interface{}) error {
	return mm.db.First(dest, conds...).Error
}

// Find 查找记录
func (mm *MySQLManager) Find(dest interface{}, conds ...interface{}) error {
	return mm.db.Find(dest, conds...).Error
}

// Update 更新记录
func (mm *MySQLManager) Update(column string, value interface{}) error {
	return mm.db.Update(column, value).Error
}

// Delete 删除记录
func (mm *MySQLManager) Delete(value interface{}, conds ...interface{}) error {
	return mm.db.Delete(value, conds...).Error
}

// Raw 执行原生SQL
func (mm *MySQLManager) Raw(sql string, values ...interface{}) *gorm.DB {
	return mm.db.Raw(sql, values...)
}

// Exec 执行SQL
func (mm *MySQLManager) Exec(sql string, values ...interface{}) error {
	return mm.db.Exec(sql, values...).Error
}

// Health 健康检查
func (mm *MySQLManager) Health() map[string]interface{} {
	sqlDB, err := mm.db.DB()
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"status":               "healthy",
		"host":                 mm.config.Host,
		"port":                 mm.config.Port,
		"database":             mm.config.Database,
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
