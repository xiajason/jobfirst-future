package multidatabase

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MultiDatabaseManager 统一的多数据库管理器
type MultiDatabaseManager struct {
	// 数据库连接
	MySQL      *gorm.DB
	PostgreSQL *gorm.DB
	Neo4j      neo4j.Driver
	Redis      *redis.Client

	// 连接池配置
	config *ConnectionPoolConfig

	// 监控指标
	metrics *DatabaseMetrics

	// 同步锁
	mu sync.RWMutex

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc
}

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	// MySQL配置
	MySQL MySQLConfig `yaml:"mysql"`

	// PostgreSQL配置
	PostgreSQL PostgreSQLConfig `yaml:"postgresql"`

	// Neo4j配置
	Neo4j Neo4jConfig `yaml:"neo4j"`

	// Redis配置
	Redis RedisConfig `yaml:"redis"`

	// 连接池通用配置
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// PostgreSQLConfig PostgreSQL数据库配置
type PostgreSQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

// Neo4jConfig Neo4j图数据库配置
type Neo4jConfig struct {
	URI      string `yaml:"uri"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// RedisConfig Redis缓存配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// DatabaseMetrics 数据库监控指标
type DatabaseMetrics struct {
	// 连接池指标
	MySQLActiveConns      int `json:"mysql_active_conns"`
	MySQLIdleConns        int `json:"mysql_idle_conns"`
	PostgreSQLActiveConns int `json:"postgresql_active_conns"`
	PostgreSQLIdleConns   int `json:"postgresql_idle_conns"`
	RedisActiveConns      int `json:"redis_active_conns"`

	// 健康状态
	MySQLHealthy      bool `json:"mysql_healthy"`
	PostgreSQLHealthy bool `json:"postgresql_healthy"`
	Neo4jHealthy      bool `json:"neo4j_healthy"`
	RedisHealthy      bool `json:"redis_healthy"`

	// 最后检查时间
	LastCheckTime time.Time `json:"last_check_time"`

	// 错误计数
	MySQLErrors      int `json:"mysql_errors"`
	PostgreSQLErrors int `json:"postgresql_errors"`
	Neo4jErrors      int `json:"neo4j_errors"`
	RedisErrors      int `json:"redis_errors"`
}

// NewMultiDatabaseManager 创建新的多数据库管理器
func NewMultiDatabaseManager(config *ConnectionPoolConfig) (*MultiDatabaseManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &MultiDatabaseManager{
		config: config,
		metrics: &DatabaseMetrics{
			LastCheckTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化数据库连接
	if err := manager.initializeConnections(); err != nil {
		cancel()
		return nil, fmt.Errorf("初始化数据库连接失败: %w", err)
	}

	// 启动健康检查
	go manager.startHealthCheck()

	// 启动指标收集
	go manager.startMetricsCollection()

	return manager, nil
}

// initializeConnections 初始化所有数据库连接
func (m *MultiDatabaseManager) initializeConnections() error {
	// 初始化MySQL连接
	if err := m.initMySQL(); err != nil {
		return fmt.Errorf("MySQL连接初始化失败: %w", err)
	}

	// 初始化PostgreSQL连接
	if err := m.initPostgreSQL(); err != nil {
		return fmt.Errorf("PostgreSQL连接初始化失败: %w", err)
	}

	// 初始化Neo4j连接
	if err := m.initNeo4j(); err != nil {
		return fmt.Errorf("Neo4j连接初始化失败: %w", err)
	}

	// 初始化Redis连接
	if err := m.initRedis(); err != nil {
		return fmt.Errorf("Redis连接初始化失败: %w", err)
	}

	return nil
}

// initMySQL 初始化MySQL连接
func (m *MultiDatabaseManager) initMySQL() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		m.config.MySQL.User,
		m.config.MySQL.Password,
		m.config.MySQL.Host,
		m.config.MySQL.Port,
		m.config.MySQL.Database,
		m.config.MySQL.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(m.config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(m.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(m.config.ConnMaxIdleTime)

	m.MySQL = db
	return nil
}

// initPostgreSQL 初始化PostgreSQL连接
func (m *MultiDatabaseManager) initPostgreSQL() error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		m.config.PostgreSQL.User,
		m.config.PostgreSQL.Password,
		m.config.PostgreSQL.Host,
		m.config.PostgreSQL.Port,
		m.config.PostgreSQL.Database,
		m.config.PostgreSQL.SSLMode,
	)

	log.Printf("PostgreSQL DSN: %s", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(m.config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(m.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(m.config.ConnMaxIdleTime)

	m.PostgreSQL = db
	return nil
}

// initNeo4j 初始化Neo4j连接
func (m *MultiDatabaseManager) initNeo4j() error {
	driver, err := neo4j.NewDriver(
		m.config.Neo4j.URI,
		neo4j.BasicAuth(m.config.Neo4j.Username, m.config.Neo4j.Password, ""),
	)
	if err != nil {
		return err
	}

	m.Neo4j = driver
	return nil
}

// initRedis 初始化Redis连接
func (m *MultiDatabaseManager) initRedis() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", m.config.Redis.Host, m.config.Redis.Port),
		Password: m.config.Redis.Password,
		DB:       m.config.Redis.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	m.Redis = rdb
	return nil
}

// startHealthCheck 启动健康检查
func (m *MultiDatabaseManager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (m *MultiDatabaseManager) performHealthCheck() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查MySQL
	if m.MySQL != nil {
		sqlDB, err := m.MySQL.DB()
		if err == nil {
			if err := sqlDB.Ping(); err != nil {
				m.metrics.MySQLHealthy = false
				m.metrics.MySQLErrors++
				log.Printf("MySQL健康检查失败: %v", err)
			} else {
				m.metrics.MySQLHealthy = true
				stats := sqlDB.Stats()
				m.metrics.MySQLActiveConns = stats.OpenConnections
				m.metrics.MySQLIdleConns = stats.Idle
			}
		}
	}

	// 检查PostgreSQL
	if m.PostgreSQL != nil {
		sqlDB, err := m.PostgreSQL.DB()
		if err == nil {
			if err := sqlDB.Ping(); err != nil {
				m.metrics.PostgreSQLHealthy = false
				m.metrics.PostgreSQLErrors++
				log.Printf("PostgreSQL健康检查失败: %v", err)
			} else {
				m.metrics.PostgreSQLHealthy = true
				stats := sqlDB.Stats()
				m.metrics.PostgreSQLActiveConns = stats.OpenConnections
				m.metrics.PostgreSQLIdleConns = stats.Idle
			}
		}
	}

	// 检查Neo4j
	if m.Neo4j != nil {
		if err := m.Neo4j.VerifyConnectivity(); err != nil {
			m.metrics.Neo4jHealthy = false
			m.metrics.Neo4jErrors++
			log.Printf("Neo4j健康检查失败: %v", err)
		} else {
			m.metrics.Neo4jHealthy = true
		}
	}

	// 检查Redis
	if m.Redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.Redis.Ping(ctx).Err(); err != nil {
			m.metrics.RedisHealthy = false
			m.metrics.RedisErrors++
			log.Printf("Redis健康检查失败: %v", err)
		} else {
			m.metrics.RedisHealthy = true
			poolStats := m.Redis.PoolStats()
			m.metrics.RedisActiveConns = int(poolStats.TotalConns)
		}
	}

	m.metrics.LastCheckTime = time.Now()
}

// startMetricsCollection 启动指标收集
func (m *MultiDatabaseManager) startMetricsCollection() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// collectMetrics 收集指标
func (m *MultiDatabaseManager) collectMetrics() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 这里可以添加更多的指标收集逻辑
	// 比如查询性能、连接池使用率等
	log.Printf("数据库指标收集完成: %+v", m.metrics)
}

// GetMetrics 获取当前指标
func (m *MultiDatabaseManager) GetMetrics() *DatabaseMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回指标的副本
	metrics := *m.metrics
	return &metrics
}

// IsHealthy 检查整体健康状态
func (m *MultiDatabaseManager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.metrics.MySQLHealthy &&
		m.metrics.PostgreSQLHealthy &&
		m.metrics.Neo4jHealthy &&
		m.metrics.RedisHealthy
}

// Close 关闭所有数据库连接
func (m *MultiDatabaseManager) Close() error {
	m.cancel()

	var errs []error

	// 关闭MySQL
	if m.MySQL != nil {
		if sqlDB, err := m.MySQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("关闭MySQL连接失败: %w", err))
			}
		}
	}

	// 关闭PostgreSQL
	if m.PostgreSQL != nil {
		if sqlDB, err := m.PostgreSQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("关闭PostgreSQL连接失败: %w", err))
			}
		}
	}

	// 关闭Neo4j
	if m.Neo4j != nil {
		if err := m.Neo4j.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭Neo4j连接失败: %w", err))
		}
	}

	// 关闭Redis
	if m.Redis != nil {
		if err := m.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭Redis连接失败: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭数据库连接时发生错误: %v", errs)
	}

	return nil
}
