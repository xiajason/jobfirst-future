package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	Load() error
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetFloat(key string) float64
	GetDuration(key string) time.Duration
	GetStringSlice(key string) []string
	Set(key string, value interface{}) error
	Watch(callback func()) error
	Reload() error
}

// Config 配置结构
type Config struct {
	mu    sync.RWMutex
	data  map[string]interface{}
	env   map[string]string
	file  string
	watch bool
}

// NewConfig 创建配置管理器
func NewConfig() *Config {
	return &Config{
		data: make(map[string]interface{}),
		env:  make(map[string]string),
	}
}

// LoadFromFile 从文件加载配置
func (c *Config) LoadFromFile(filepath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.file = filepath

	// 读取文件
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// 根据文件扩展名解析
	switch {
	case strings.HasSuffix(filepath, ".json"):
		err = json.Unmarshal(data, &c.data)
	case strings.HasSuffix(filepath, ".yaml") || strings.HasSuffix(filepath, ".yml"):
		err = yaml.Unmarshal(data, &c.data)
	default:
		return fmt.Errorf("unsupported config file format: %s", filepath)
	}

	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	return nil
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key, value := pair[0], pair[1]
		if strings.HasPrefix(key, prefix) {
			// 移除前缀并转换为小写
			configKey := strings.ToLower(strings.TrimPrefix(key, prefix))
			configKey = strings.ReplaceAll(configKey, "_", ".")
			c.env[configKey] = value
		}
	}
}

// Load 加载配置
func (c *Config) Load() error {
	// 加载环境变量
	c.LoadFromEnv("JOBFIRST_")

	// 如果有配置文件，则加载
	if c.file != "" {
		return c.LoadFromFile(c.file)
	}

	return nil
}

// Get 获取配置值
func (c *Config) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 优先从环境变量获取
	if value, exists := c.env[key]; exists {
		return value
	}

	// 从配置文件获取
	return c.getNestedValue(c.data, key)
}

// GetString 获取字符串配置
func (c *Config) GetString(key string) string {
	value := c.Get(key)
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		return ""
	}
}

// GetInt 获取整数配置
func (c *Config) GetInt(key string) int {
	value := c.Get(key)
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return 0
}

// GetBool 获取布尔配置
func (c *Config) GetBool(key string) bool {
	value := c.Get(key)
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true" || v == "1"
	case int:
		return v != 0
	}

	return false
}

// GetFloat 获取浮点数配置
func (c *Config) GetFloat(key string) float64 {
	value := c.Get(key)
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}

	return 0.0
}

// GetDuration 获取时间间隔配置
func (c *Config) GetDuration(key string) time.Duration {
	value := c.GetString(key)
	if value == "" {
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}

	return duration
}

// GetStringSlice 获取字符串切片配置
func (c *Config) GetStringSlice(key string) []string {
	value := c.Get(key)
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	case string:
		return strings.Split(v, ",")
	}

	return nil
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置到环境变量
	if str, ok := value.(string); ok {
		c.env[key] = str
	} else {
		c.env[key] = fmt.Sprintf("%v", value)
	}

	// 设置到配置文件数据
	c.setNestedValue(c.data, key, value)

	return nil
}

// Watch 监听配置变化
func (c *Config) Watch(callback func()) error {
	if c.file == "" {
		return fmt.Errorf("no config file specified for watching")
	}

	c.watch = true
	// 这里可以实现文件监听逻辑
	// 暂时返回nil
	return nil
}

// Reload 重新加载配置
func (c *Config) Reload() error {
	return c.Load()
}

// getNestedValue 获取嵌套值
func (c *Config) getNestedValue(data map[string]interface{}, key string) interface{} {
	keys := strings.Split(key, ".")
	current := data

	for i, k := range keys {
		if i == len(keys)-1 {
			return current[k]
		}

		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// setNestedValue 设置嵌套值
func (c *Config) setNestedValue(data map[string]interface{}, key string, value interface{}) {
	keys := strings.Split(key, ".")
	current := data

	for i, k := range keys {
		if i == len(keys)-1 {
			current[k] = value
			return
		}

		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			current[k] = make(map[string]interface{})
			current = current[k].(map[string]interface{})
		}
	}
}

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: NewConfig(),
	}
}

// WithFile 设置配置文件
func (b *ConfigBuilder) WithFile(filepath string) *ConfigBuilder {
	b.config.file = filepath
	return b
}

// WithEnvPrefix 设置环境变量前缀
func (b *ConfigBuilder) WithEnvPrefix(prefix string) *ConfigBuilder {
	b.config.LoadFromEnv(prefix)
	return b
}

// WithDefaults 设置默认值
func (b *ConfigBuilder) WithDefaults(defaults map[string]interface{}) *ConfigBuilder {
	for k, v := range defaults {
		b.config.Set(k, v)
	}
	return b
}

// Build 构建配置
func (b *ConfigBuilder) Build() (ConfigManager, error) {
	err := b.config.Load()
	if err != nil {
		return nil, err
	}
	return b.config, nil
}

// 全局配置实例
var globalConfig ConfigManager

// InitGlobalConfig 初始化全局配置
func InitGlobalConfig(config ConfigManager) {
	globalConfig = config
}

// GetConfig 获取全局配置
func GetConfig() ConfigManager {
	return globalConfig
}

// 便捷函数
func GetConfigString(key string) string {
	return GetConfig().GetString(key)
}

func GetConfigInt(key string) int {
	return GetConfig().GetInt(key)
}

func GetConfigBool(key string) bool {
	return GetConfig().GetBool(key)
}

func GetConfigFloat(key string) float64 {
	return GetConfig().GetFloat(key)
}

func GetConfigDuration(key string) time.Duration {
	return GetConfig().GetDuration(key)
}
