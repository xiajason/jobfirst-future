package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SimpleManager 简化的配置管理器
type SimpleManager struct {
	configs     map[string]interface{}
	versions    []*ConfigVersion
	watchers    []*ConfigWatcher
	changes     []*ConfigChange
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	storagePath string
}

// NewSimpleManager 创建简化配置管理器
func NewSimpleManager(storagePath string) (*SimpleManager, error) {
	if storagePath == "" {
		storagePath = "./configs"
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &SimpleManager{
		configs:     make(map[string]interface{}),
		versions:    make([]*ConfigVersion, 0),
		watchers:    make([]*ConfigWatcher, 0),
		changes:     make([]*ConfigChange, 0),
		ctx:         ctx,
		cancel:      cancel,
		storagePath: storagePath,
	}

	// 创建存储目录
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// 加载配置
	if err := manager.loadConfigs(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to load configs: %v", err)
	}

	return manager, nil
}

// loadConfigs 加载配置
func (m *SimpleManager) loadConfigs() error {
	configFile := filepath.Join(m.storagePath, "current.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return m.createDefaultConfig()
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var configs map[string]interface{}
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	m.mutex.Lock()
	m.configs = configs
	m.mutex.Unlock()

	return nil
}

// createDefaultConfig 创建默认配置
func (m *SimpleManager) createDefaultConfig() error {
	defaultConfigs := map[string]interface{}{
		"database.mysql.host":         "localhost",
		"database.mysql.port":         3306,
		"database.mysql.username":     "root",
		"database.mysql.password":     "",
		"database.mysql.database":     "jobfirst",
		"database.redis.host":         "localhost",
		"database.redis.port":         6379,
		"database.redis.password":     "",
		"service.api_gateway.port":    8080,
		"service.user_service.port":   8081,
		"service.resume_service.port": 8082,
		"system.log_level":            "info",
		"system.debug":                false,
		"security.jwt_secret":         "default-secret",
		"security.token_expiry":       3600,
	}

	m.mutex.Lock()
	m.configs = defaultConfigs
	m.mutex.Unlock()

	return m.saveConfigs()
}

// saveConfigs 保存配置
func (m *SimpleManager) saveConfigs() error {
	m.mutex.RLock()
	configs := make(map[string]interface{})
	for k, v := range m.configs {
		configs[k] = v
	}
	m.mutex.RUnlock()

	return m.saveConfigsUnlocked(configs)
}

// saveConfigsUnlocked 保存配置（不需要锁）
func (m *SimpleManager) saveConfigsUnlocked(configs map[string]interface{}) error {
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configFile := filepath.Join(m.storagePath, "current.json")
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Get 获取配置值
func (m *SimpleManager) Get(key string) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	value, exists := m.configs[key]
	if !exists {
		return nil, fmt.Errorf("config key '%s' not found", key)
	}

	return value, nil
}

// Set 设置配置值
func (m *SimpleManager) Set(key string, value interface{}) error {
	m.mutex.Lock()
	oldValue, exists := m.configs[key]
	m.configs[key] = value

	// 记录变更
	change := &ConfigChange{
		ID:          uuid.New().String(),
		Version:     "1.0.0",
		Type:        m.getConfigType(key),
		Changes:     map[string]interface{}{key: value},
		OldValues:   map[string]interface{}{key: oldValue},
		NewValues:   map[string]interface{}{key: value},
		Timestamp:   time.Now(),
		Author:      "system",
		Description: fmt.Sprintf("Updated config key '%s'", key),
		Status:      ChangeStatusApplied,
	}

	m.changes = append(m.changes, change)

	// 复制配置用于保存（避免死锁）
	configs := make(map[string]interface{})
	for k, v := range m.configs {
		configs[k] = v
	}
	m.mutex.Unlock()

	// 保存配置（不需要锁）
	if err := m.saveConfigsUnlocked(configs); err != nil {
		// 回滚变更
		m.mutex.Lock()
		if exists {
			m.configs[key] = oldValue
		} else {
			delete(m.configs, key)
		}
		change.Status = ChangeStatusFailed
		m.mutex.Unlock()
		return fmt.Errorf("failed to save config: %v", err)
	}

	// 通知监听器
	m.mutex.RLock()
	m.notifyWatchers(change)
	m.mutex.RUnlock()

	return nil
}

// GetAll 获取所有配置
func (m *SimpleManager) GetAll() (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	configs := make(map[string]interface{})
	for k, v := range m.configs {
		configs[k] = v
	}

	return configs, nil
}

// GetByType 根据类型获取配置
func (m *SimpleManager) GetByType(configType ConfigType) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	configs := make(map[string]interface{})
	prefix := string(configType) + "."

	for k, v := range m.configs {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			configs[k] = v
		}
	}

	return configs, nil
}

// CreateVersion 创建新版本
func (m *SimpleManager) CreateVersion(description string, author string) (*ConfigVersion, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 获取当前配置
	configs := make(map[string]interface{})
	for k, v := range m.configs {
		configs[k] = v
	}

	version := &ConfigVersion{
		Version:     fmt.Sprintf("1.%d.0", len(m.versions)+1),
		Timestamp:   time.Now(),
		Author:      author,
		Description: description,
		Config:      configs,
		Checksum:    fmt.Sprintf("%d", len(configs)),
		Tags:        []string{},
	}

	m.versions = append(m.versions, version)
	return version, nil
}

// WatchChanges 监听配置变更
func (m *SimpleManager) WatchChanges(callback func(*ConfigChange)) error {
	watcher := &ConfigWatcher{
		ID:       uuid.New().String(),
		Callback: callback,
		Filter:   "",
		Active:   true,
	}

	m.mutex.Lock()
	m.watchers = append(m.watchers, watcher)
	m.mutex.Unlock()

	return nil
}

// 辅助方法

func (m *SimpleManager) getConfigType(key string) ConfigType {
	if len(key) > 8 && key[:8] == "database" {
		return ConfigTypeDatabase
	} else if len(key) > 7 && key[:7] == "service" {
		return ConfigTypeService
	} else if len(key) > 6 && key[:6] == "system" {
		return ConfigTypeSystem
	} else if len(key) > 8 && key[:8] == "security" {
		return ConfigTypeSecurity
	} else if len(key) > 7 && key[:7] == "logging" {
		return ConfigTypeLogging
	}
	return ConfigTypeSystem
}

func (m *SimpleManager) notifyWatchers(change *ConfigChange) {
	m.mutex.RLock()
	watchers := make([]*ConfigWatcher, len(m.watchers))
	copy(watchers, m.watchers)
	m.mutex.RUnlock()

	for _, watcher := range watchers {
		if watcher.Active && (watcher.Filter == "" || watcher.Filter == change.Type) {
			go watcher.Callback(change)
		}
	}
}

// Close 关闭配置管理器
func (m *SimpleManager) Close() error {
	m.cancel()
	return nil
}
