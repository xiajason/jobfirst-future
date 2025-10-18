package database

import (
	"fmt"
	"log"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL      MySQLConfig      `yaml:"mysql"`
	PostgreSQL PostgreSQLConfig `yaml:"postgresql"`
	Neo4j      Neo4jConfig      `yaml:"neo4j"`
	Redis      RedisConfig      `yaml:"redis"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
}

// PostgreSQLConfig PostgreSQL配置
type PostgreSQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

// Neo4jConfig Neo4j配置
type Neo4jConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	Config *DatabaseConfig
}

// NewConfigManager 创建配置管理器
func NewConfigManager(config *DatabaseConfig) *ConfigManager {
	return &ConfigManager{
		Config: config,
	}
}

// GetMySQLDSN 获取MySQL连接字符串
func (cm *ConfigManager) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cm.Config.MySQL.Username,
		cm.Config.MySQL.Password,
		cm.Config.MySQL.Host,
		cm.Config.MySQL.Port,
		cm.Config.MySQL.Database,
		cm.Config.MySQL.Charset,
	)
}

// GetPostgreSQLDSN 获取PostgreSQL连接字符串
func (cm *ConfigManager) GetPostgreSQLDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cm.Config.PostgreSQL.Host,
		cm.Config.PostgreSQL.Port,
		cm.Config.PostgreSQL.Username,
		cm.Config.PostgreSQL.Password,
		cm.Config.PostgreSQL.Database,
		cm.Config.PostgreSQL.SSLMode,
	)
}

// GetNeo4jURI 获取Neo4j连接URI
func (cm *ConfigManager) GetNeo4jURI() string {
	return fmt.Sprintf("neo4j://%s:%d", cm.Config.Neo4j.Host, cm.Config.Neo4j.Port)
}

// GetRedisAddr 获取Redis连接地址
func (cm *ConfigManager) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", cm.Config.Redis.Host, cm.Config.Redis.Port)
}

// LogConnectionInfo 打印连接信息
func (cm *ConfigManager) LogConnectionInfo() {
	log.Println("=== Database Connection Information ===")

	log.Printf("MySQL: %s:%d/%s",
		cm.Config.MySQL.Host,
		cm.Config.MySQL.Port,
		cm.Config.MySQL.Database)

	log.Printf("PostgreSQL: %s:%d/%s",
		cm.Config.PostgreSQL.Host,
		cm.Config.PostgreSQL.Port,
		cm.Config.PostgreSQL.Database)

	log.Printf("Neo4j: %s:%d",
		cm.Config.Neo4j.Host,
		cm.Config.Neo4j.Port)

	log.Printf("Redis: %s:%d",
		cm.Config.Redis.Host,
		cm.Config.Redis.Port)

	log.Println("=======================================")
}

// ValidateConfig 验证配置
func (cm *ConfigManager) ValidateConfig() error {
	// 验证MySQL配置
	if cm.Config.MySQL.Host == "" || cm.Config.MySQL.Port == 0 {
		return fmt.Errorf("invalid MySQL configuration")
	}

	// 验证PostgreSQL配置
	if cm.Config.PostgreSQL.Host == "" || cm.Config.PostgreSQL.Port == 0 {
		return fmt.Errorf("invalid PostgreSQL configuration")
	}

	// 验证Neo4j配置
	if cm.Config.Neo4j.Host == "" || cm.Config.Neo4j.Port == 0 {
		return fmt.Errorf("invalid Neo4j configuration")
	}

	// 验证Redis配置
	if cm.Config.Redis.Host == "" || cm.Config.Redis.Port == 0 {
		return fmt.Errorf("invalid Redis configuration")
	}

	return nil
}

// CreateDefaultConfig 创建默认配置
func CreateDefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MySQL: MySQLConfig{
			Host:     "localhost",
			Port:     8200,
			Database: "jobfirst",
			Username: "jobfirst",
			Password: "jobfirst123",
			Charset:  "utf8mb4",
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     "localhost",
			Port:     8203,
			Database: "jobfirst_vector",
			Username: "jobfirst",
			Password: "jobfirst123",
			SSLMode:  "disable",
		},
		Neo4j: Neo4jConfig{
			Host:     "localhost",
			Port:     8205,
			Username: "neo4j",
			Password: "jobfirst123",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     8201,
			Password: "",
			DB:       0,
		},
	}
}
