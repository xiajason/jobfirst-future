package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager 数据库管理器接口
type DatabaseManager interface {
	Connect() error
	GetMySQLConnection() *gorm.DB
	GetPostgreSQLConnection() *gorm.DB
	GetNeo4jConnection() neo4j.Driver
	GetRedisConnection() *redis.Client
	HealthCheck() map[string]bool
	Close() error
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL      MySQLConfig      `yaml:"mysql" json:"mysql"`
	PostgreSQL PostgreSQLConfig `yaml:"postgresql" json:"postgresql"`
	Neo4j      Neo4jConfig      `yaml:"neo4j" json:"neo4j"`
	Redis      RedisConfig      `yaml:"redis" json:"redis"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	Database        string        `yaml:"database" json:"database"`
	Username        string        `yaml:"username" json:"username"`
	Password        string        `yaml:"password" json:"password"`
	Charset         string        `yaml:"charset" json:"charset"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// PostgreSQLConfig PostgreSQL配置
type PostgreSQLConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	Database        string        `yaml:"database" json:"database"`
	Username        string        `yaml:"username" json:"username"`
	Password        string        `yaml:"password" json:"password"`
	SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// Neo4jConfig Neo4j配置
type Neo4jConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string        `yaml:"host" json:"host"`
	Port     int           `yaml:"port" json:"port"`
	Password string        `yaml:"password" json:"password"`
	DB       int           `yaml:"db" json:"db"`
	PoolSize int           `yaml:"pool_size" json:"pool_size"`
	MinIdle  int           `yaml:"min_idle" json:"min_idle"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
}

// Manager 数据库管理器实现
type Manager struct {
	mu           sync.RWMutex
	config       *DatabaseConfig
	mysqlDB      *gorm.DB
	postgresDB   *gorm.DB
	neo4jDriver  neo4j.Driver
	redisClient  *redis.Client
	healthStatus map[string]bool
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager(config *DatabaseConfig) DatabaseManager {
	return &Manager{
		config:       config,
		healthStatus: make(map[string]bool),
	}
}

// Connect 连接所有数据库
func (m *Manager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	// 连接MySQL
	if err := m.connectMySQL(); err != nil {
		errors = append(errors, fmt.Errorf("MySQL connection failed: %v", err))
	}

	// 连接PostgreSQL
	if err := m.connectPostgreSQL(); err != nil {
		errors = append(errors, fmt.Errorf("PostgreSQL connection failed: %v", err))
	}

	// 连接Neo4j
	if err := m.connectNeo4j(); err != nil {
		errors = append(errors, fmt.Errorf("Neo4j connection failed: %v", err))
	}

	// 连接Redis
	if err := m.connectRedis(); err != nil {
		errors = append(errors, fmt.Errorf("Redis connection failed: %v", err))
	}

	// 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// connectMySQL 连接MySQL
func (m *Manager) connectMySQL() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		m.config.MySQL.Username,
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

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if m.config.MySQL.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(m.config.MySQL.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}

	if m.config.MySQL.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(m.config.MySQL.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}

	if m.config.MySQL.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(m.config.MySQL.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	m.mysqlDB = db
	m.healthStatus["mysql"] = true

	Info("MySQL connected successfully", Field{Key: "host", Value: m.config.MySQL.Host})
	return nil
}

// connectPostgreSQL 连接PostgreSQL
func (m *Manager) connectPostgreSQL() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		m.config.PostgreSQL.Host,
		m.config.PostgreSQL.Port,
		m.config.PostgreSQL.Username,
		m.config.PostgreSQL.Password,
		m.config.PostgreSQL.Database,
		m.config.PostgreSQL.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if m.config.PostgreSQL.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(m.config.PostgreSQL.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}

	if m.config.PostgreSQL.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(m.config.PostgreSQL.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}

	if m.config.PostgreSQL.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(m.config.PostgreSQL.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	m.postgresDB = db
	m.healthStatus["postgresql"] = true

	Info("PostgreSQL connected successfully", Field{Key: "host", Value: m.config.PostgreSQL.Host})
	return nil
}

// connectNeo4j 连接Neo4j
func (m *Manager) connectNeo4j() error {
	uri := fmt.Sprintf("neo4j://%s:%d", m.config.Neo4j.Host, m.config.Neo4j.Port)

	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(m.config.Neo4j.Username, m.config.Neo4j.Password, ""))
	if err != nil {
		return err
	}

	// 测试连接
	err = driver.VerifyConnectivity()
	if err != nil {
		return err
	}

	m.neo4jDriver = driver
	m.healthStatus["neo4j"] = true

	Info("Neo4j connected successfully", Field{Key: "host", Value: m.config.Neo4j.Host})
	return nil
}

// connectRedis 连接Redis
func (m *Manager) connectRedis() error {
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", m.config.Redis.Host, m.config.Redis.Port),
		Password: m.config.Redis.Password,
		DB:       m.config.Redis.DB,
	}

	if m.config.Redis.PoolSize > 0 {
		options.PoolSize = m.config.Redis.PoolSize
	} else {
		options.PoolSize = 10
	}

	if m.config.Redis.MinIdle > 0 {
		options.MinIdleConns = m.config.Redis.MinIdle
	} else {
		options.MinIdleConns = 5
	}

	if m.config.Redis.Timeout > 0 {
		options.DialTimeout = m.config.Redis.Timeout
	} else {
		options.DialTimeout = 5 * time.Second
	}

	client := redis.NewClient(options)

	// 测试连接
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return err
	}

	m.redisClient = client
	m.healthStatus["redis"] = true

	Info("Redis connected successfully", Field{Key: "host", Value: m.config.Redis.Host})
	return nil
}

// GetMySQLConnection 获取MySQL连接
func (m *Manager) GetMySQLConnection() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mysqlDB
}

// GetPostgreSQLConnection 获取PostgreSQL连接
func (m *Manager) GetPostgreSQLConnection() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.postgresDB
}

// GetNeo4jConnection 获取Neo4j连接
func (m *Manager) GetNeo4jConnection() neo4j.Driver {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.neo4jDriver
}

// GetRedisConnection 获取Redis连接
func (m *Manager) GetRedisConnection() *redis.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.redisClient
}

// HealthCheck 健康检查
func (m *Manager) HealthCheck() map[string]bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// MySQL健康检查
	if m.mysqlDB != nil {
		sqlDB, err := m.mysqlDB.DB()
		if err == nil {
			m.healthStatus["mysql"] = sqlDB.Ping() == nil
		} else {
			m.healthStatus["mysql"] = false
		}
	} else {
		m.healthStatus["mysql"] = false
	}

	// PostgreSQL健康检查
	if m.postgresDB != nil {
		sqlDB, err := m.postgresDB.DB()
		if err == nil {
			m.healthStatus["postgresql"] = sqlDB.Ping() == nil
		} else {
			m.healthStatus["postgresql"] = false
		}
	} else {
		m.healthStatus["postgresql"] = false
	}

	// Neo4j健康检查
	if m.neo4jDriver != nil {
		m.healthStatus["neo4j"] = m.neo4jDriver.VerifyConnectivity() == nil
	} else {
		m.healthStatus["neo4j"] = false
	}

	// Redis健康检查
	if m.redisClient != nil {
		ctx := context.Background()
		_, err := m.redisClient.Ping(ctx).Result()
		m.healthStatus["redis"] = err == nil
	} else {
		m.healthStatus["redis"] = false
	}

	return m.healthStatus
}

// Close 关闭所有连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	// 关闭MySQL连接
	if m.mysqlDB != nil {
		sqlDB, err := m.mysqlDB.DB()
		if err == nil {
			errors = append(errors, sqlDB.Close())
		}
	}

	// 关闭PostgreSQL连接
	if m.postgresDB != nil {
		sqlDB, err := m.postgresDB.DB()
		if err == nil {
			errors = append(errors, sqlDB.Close())
		}
	}

	// 关闭Neo4j连接
	if m.neo4jDriver != nil {
		errors = append(errors, m.neo4jDriver.Close())
	}

	// 关闭Redis连接
	if m.redisClient != nil {
		errors = append(errors, m.redisClient.Close())
	}

	// 返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// CreateDefaultConfig 创建默认配置
func CreateDefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MySQL: MySQLConfig{
			Host:            "mysql",
			Port:            3306,
			Database:        "jobfirst",
			Username:        "jobfirst",
			Password:        "jobfirst123",
			Charset:         "utf8mb4",
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
		},
		PostgreSQL: PostgreSQLConfig{
			Host:            "postgresql",
			Port:            5432,
			Database:        "jobfirst_vector",
			Username:        "jobfirst",
			Password:        "jobfirst123",
			SSLMode:         "disable",
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
		},
		Neo4j: Neo4jConfig{
			Host:     "neo4j",
			Port:     7687,
			Username: "neo4j",
			Password: "jobfirst123",
		},
		Redis: RedisConfig{
			Host:     "redis",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
			MinIdle:  5,
			Timeout:  5 * time.Second,
		},
	}
}

// 全局数据库管理器实例
var globalDBManager DatabaseManager

// InitGlobalDatabaseManager 初始化全局数据库管理器
func InitGlobalDatabaseManager(config *DatabaseConfig) error {
	manager := NewDatabaseManager(config)
	err := manager.Connect()
	if err != nil {
		return err
	}
	globalDBManager = manager
	return nil
}

// GetDatabaseManager 获取全局数据库管理器
func GetDatabaseManager() DatabaseManager {
	return globalDBManager
}
