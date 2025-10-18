package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Manager 配置管理器
type Manager struct {
	viper *viper.Viper
}

// NewManager 创建配置管理器
func NewManager(configPath string) (*Manager, error) {
	v := viper.New()

	// 设置配置文件路径
	if configPath != "" {
		dir := filepath.Dir(configPath)
		filename := filepath.Base(configPath)
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)

		v.AddConfigPath(dir)
		v.SetConfigName(name)
		v.SetConfigType(strings.TrimPrefix(ext, "."))
	} else {
		// 默认配置路径
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath("../../configs")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// 设置环境变量前缀
	v.SetEnvPrefix("JOBFIRST")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在，使用默认配置
	}

	manager := &Manager{
		viper: v,
	}

	return manager, nil
}

// Get 获取配置值
func (cm *Manager) Get(key string) interface{} {
	return cm.viper.Get(key)
}

// GetString 获取字符串配置
func (cm *Manager) GetString(key string) string {
	return cm.viper.GetString(key)
}

// GetInt 获取整数配置
func (cm *Manager) GetInt(key string) int {
	return cm.viper.GetInt(key)
}

// GetBool 获取布尔配置
func (cm *Manager) GetBool(key string) bool {
	return cm.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置
func (cm *Manager) GetFloat64(key string) float64 {
	return cm.viper.GetFloat64(key)
}

// GetStringSlice 获取字符串切片配置
func (cm *Manager) GetStringSlice(key string) []string {
	return cm.viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置
func (cm *Manager) GetStringMap(key string) map[string]interface{} {
	return cm.viper.GetStringMap(key)
}

// Set 设置配置值
func (cm *Manager) Set(key string, value interface{}) {
	cm.viper.Set(key, value)
}

// IsSet 检查配置是否已设置
func (cm *Manager) IsSet(key string) bool {
	return cm.viper.IsSet(key)
}

// AllSettings 获取所有配置
func (cm *Manager) AllSettings() map[string]interface{} {
	return cm.viper.AllSettings()
}

// WatchConfig 监听配置文件变化
func (cm *Manager) WatchConfig() {
	cm.viper.WatchConfig()
}

// OnConfigChange 配置文件变化回调
func (cm *Manager) OnConfigChange(fn func()) {
	cm.viper.OnConfigChange(func(e fsnotify.Event) {
		fn()
	})
}

// SaveConfig 保存配置到文件
func (cm *Manager) SaveConfig() error {
	return cm.viper.WriteConfig()
}

// SaveConfigAs 保存配置到指定文件
func (cm *Manager) SaveConfigAs(filename string) error {
	return cm.viper.WriteConfigAs(filename)
}

// 预定义的配置结构

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	Charset     string `mapstructure:"charset"`
	MaxIdle     int    `mapstructure:"max_idle"`
	MaxOpen     int    `mapstructure:"max_open"`
	MaxLifetime string `mapstructure:"max_lifetime"`
	LogLevel    string `mapstructure:"log_level"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
	PoolSize int    `mapstructure:"pool_size"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret        string `mapstructure:"jwt_secret"`
	TokenExpiry      string `mapstructure:"token_expiry"`
	RefreshExpiry    string `mapstructure:"refresh_expiry"`
	PasswordMin      int    `mapstructure:"password_min_length"`
	MaxLoginAttempts int    `mapstructure:"max_login_attempts"`
	LockoutDuration  string `mapstructure:"lockout_duration"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
	File   string `mapstructure:"file"`
}

// AppConfig 应用配置
type AppConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Log      LogConfig      `mapstructure:"log"`
}

// LoadAppConfig 加载应用配置
func (cm *Manager) LoadAppConfig() (*AppConfig, error) {
	var config AppConfig
	if err := cm.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return &config, nil
}

// GetDatabaseConfig 获取数据库配置
func (cm *Manager) GetDatabaseConfig() (*DatabaseConfig, error) {
	var config DatabaseConfig
	if err := cm.viper.UnmarshalKey("database", &config); err != nil {
		return nil, fmt.Errorf("解析数据库配置失败: %w", err)
	}
	return &config, nil
}

// GetRedisConfig 获取Redis配置
func (cm *Manager) GetRedisConfig() (*RedisConfig, error) {
	var config RedisConfig
	if err := cm.viper.UnmarshalKey("redis", &config); err != nil {
		return nil, fmt.Errorf("解析Redis配置失败: %w", err)
	}
	return &config, nil
}

// GetServerConfig 获取服务器配置
func (cm *Manager) GetServerConfig() (*ServerConfig, error) {
	var config ServerConfig
	if err := cm.viper.UnmarshalKey("server", &config); err != nil {
		return nil, fmt.Errorf("解析服务器配置失败: %w", err)
	}
	return &config, nil
}

// GetAuthConfig 获取认证配置
func (cm *Manager) GetAuthConfig() (*AuthConfig, error) {
	var config AuthConfig
	if err := cm.viper.UnmarshalKey("auth", &config); err != nil {
		return nil, fmt.Errorf("解析认证配置失败: %w", err)
	}
	return &config, nil
}

// GetLogConfig 获取日志配置
func (cm *Manager) GetLogConfig() (*LogConfig, error) {
	var config LogConfig
	if err := cm.viper.UnmarshalKey("log", &config); err != nil {
		return nil, fmt.Errorf("解析日志配置失败: %w", err)
	}
	return &config, nil
}

// 环境变量辅助函数

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt 获取整数环境变量
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool 获取布尔环境变量
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
