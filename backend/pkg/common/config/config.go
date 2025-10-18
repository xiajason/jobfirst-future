package config

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config 配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Consul   ConsulConfig   `mapstructure:"consul"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address    string `mapstructure:"address"`
	Datacenter string `mapstructure:"datacenter"`
	Token      string `mapstructure:"token"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return &Config{}
}

// Load 加载配置
func (c *Config) Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	c.setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 配置文件不存在时使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// 绑定环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 解析配置
	return viper.Unmarshal(c)
}

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	viper.SetDefault("server.port", "8000")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	viper.SetDefault("consul.address", "localhost:8202")
	viper.SetDefault("consul.datacenter", "dc1")

	viper.SetDefault("redis.address", "localhost:8201")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "8200")
	viper.SetDefault("database.user", "jobfirst")
	viper.SetDefault("database.password", "jobfirst123")
	viper.SetDefault("database.name", "jobfirst")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	viper.SetDefault("jwt.secret", "jobfirst-secret-key")
	viper.SetDefault("jwt.expiration", "24h")
}

// GetServerPort 获取服务器端口
func (c *Config) GetServerPort() string {
	return c.Server.Port
}

// GetConsulAddress 获取Consul地址
func (c *Config) GetConsulAddress() string {
	return c.Consul.Address
}

// GetConsulDatacenter 获取Consul数据中心
func (c *Config) GetConsulDatacenter() string {
	return c.Consul.Datacenter
}

// GetRedisAddress 获取Redis地址
func (c *Config) GetRedisAddress() string {
	return c.Redis.Address
}

// GetDatabaseDSN 获取数据库DSN
func (c *Config) GetDatabaseDSN() string {
	return c.Database.User + ":" + c.Database.Password + "@tcp(" + c.Database.Host + ":" + c.Database.Port + ")/" + c.Database.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// GetLogLevel 获取日志级别
func (c *Config) GetLogLevel() logrus.Level {
	level, err := logrus.ParseLevel(c.Logging.Level)
	if err != nil {
		return logrus.InfoLevel
	}
	return level
}

// GetJWTSecret 获取JWT密钥
func (c *Config) GetJWTSecret() string {
	return c.JWT.Secret
}

// GetJWTExpiration 获取JWT过期时间
func (c *Config) GetJWTExpiration() time.Duration {
	return c.JWT.Expiration
}
