package multidatabase

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(configPath string) (*ConnectionPoolConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config ConnectionPoolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	config.setDefaults()

	return &config, nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *ConnectionPoolConfig {
	config := &ConnectionPoolConfig{
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnvInt("MYSQL_PORT", 3306),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
			Database: getEnv("MYSQL_DATABASE", "jobfirst"),
			Charset:  getEnv("MYSQL_CHARSET", "utf8mb4"),
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "szjason72"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			Database: getEnv("POSTGRES_DATABASE", "jobfirst_vector"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Neo4j: Neo4jConfig{
			URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
			Username: getEnv("NEO4J_USERNAME", "neo4j"),
			Password: getEnv("NEO4J_PASSWORD", "jobfirst_password_2024"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 100),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", "1h"),
		ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", "30m"),
	}

	return config
}

// setDefaults 设置默认值
func (c *ConnectionPoolConfig) setDefaults() {
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 100
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = time.Hour
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 30 * time.Minute
	}
	if c.MySQL.Charset == "" {
		c.MySQL.Charset = "utf8mb4"
	}
	if c.PostgreSQL.SSLMode == "" {
		c.PostgreSQL.SSLMode = "disable"
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

// getEnvDuration 获取持续时间环境变量
func getEnvDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Hour // 默认值
}
