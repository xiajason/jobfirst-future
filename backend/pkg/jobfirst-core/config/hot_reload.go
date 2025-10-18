package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// HotReloader 配置热更新器
type HotReloader struct {
	configManager *SimpleManager
	watcher       *fsnotify.Watcher
	configFile    string
	ctx           context.Context
	cancel        context.CancelFunc
	mutex         sync.RWMutex
	callbacks     []func(*ConfigChange)
	enabled       bool
}

// NewHotReloader 创建配置热更新器
func NewHotReloader(configManager *SimpleManager, configFile string) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	reloader := &HotReloader{
		configManager: configManager,
		watcher:       watcher,
		configFile:    configFile,
		ctx:           ctx,
		cancel:        cancel,
		callbacks:     make([]func(*ConfigChange), 0),
		enabled:       false,
	}

	return reloader, nil
}

// Start 启动热更新
func (hr *HotReloader) Start() error {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	if hr.enabled {
		return fmt.Errorf("hot reloader is already running")
	}

	// 添加文件监听
	if err := hr.watcher.Add(hr.configFile); err != nil {
		return fmt.Errorf("failed to add file to watcher: %v", err)
	}

	// 添加目录监听（用于监听新文件）
	configDir := filepath.Dir(hr.configFile)
	if err := hr.watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to add directory to watcher: %v", err)
	}

	hr.enabled = true

	// 启动监听协程
	go hr.watch()

	return nil
}

// Stop 停止热更新
func (hr *HotReloader) Stop() error {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	if !hr.enabled {
		return nil
	}

	hr.cancel()
	hr.enabled = false

	if err := hr.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %v", err)
	}

	return nil
}

// AddCallback 添加变更回调
func (hr *HotReloader) AddCallback(callback func(*ConfigChange)) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()

	hr.callbacks = append(hr.callbacks, callback)
}

// watch 监听文件变更
func (hr *HotReloader) watch() {
	for {
		select {
		case event, ok := <-hr.watcher.Events:
			if !ok {
				return
			}

			// 只处理写入事件
			if event.Op&fsnotify.Write == fsnotify.Write {
				// 检查是否是目标配置文件
				if event.Name == hr.configFile {
					hr.handleConfigChange()
				}
			}

		case err, ok := <-hr.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)

		case <-hr.ctx.Done():
			return
		}
	}
}

// handleConfigChange 处理配置变更
func (hr *HotReloader) handleConfigChange() {
	// 等待文件写入完成
	time.Sleep(100 * time.Millisecond)

	// 读取新配置
	newConfig, err := hr.loadConfigFromFile()
	if err != nil {
		fmt.Printf("Failed to load config from file: %v\n", err)
		return
	}

	// 获取当前配置
	currentConfig, err := hr.configManager.GetAll()
	if err != nil {
		fmt.Printf("Failed to get current config: %v\n", err)
		return
	}

	// 比较配置差异
	changes := hr.compareConfigs(currentConfig, newConfig)
	if len(changes) == 0 {
		return // 没有变更
	}

	// 应用新配置
	if err := hr.applyConfigChanges(newConfig); err != nil {
		fmt.Printf("Failed to apply config changes: %v\n", err)
		return
	}

	// 通知回调
	hr.notifyCallbacks(changes)
}

// loadConfigFromFile 从文件加载配置
func (hr *HotReloader) loadConfigFromFile() (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(hr.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return config, nil
}

// compareConfigs 比较配置差异
func (hr *HotReloader) compareConfigs(oldConfig, newConfig map[string]interface{}) []*ConfigChange {
	var changes []*ConfigChange

	// 检查新增和修改的配置
	for key, newValue := range newConfig {
		oldValue, exists := oldConfig[key]
		if !exists {
			// 新增配置
			change := &ConfigChange{
				ID:          fmt.Sprintf("add_%s_%d", key, time.Now().Unix()),
				Version:     "hot_reload",
				Type:        hr.getConfigType(key),
				Changes:     map[string]interface{}{key: newValue},
				OldValues:   map[string]interface{}{key: nil},
				NewValues:   map[string]interface{}{key: newValue},
				Timestamp:   time.Now(),
				Author:      "hot_reload",
				Description: fmt.Sprintf("Added config key '%s' via hot reload", key),
				Status:      ChangeStatusApplied,
			}
			changes = append(changes, change)
		} else if oldValue != newValue {
			// 修改配置
			change := &ConfigChange{
				ID:          fmt.Sprintf("update_%s_%d", key, time.Now().Unix()),
				Version:     "hot_reload",
				Type:        hr.getConfigType(key),
				Changes:     map[string]interface{}{key: newValue},
				OldValues:   map[string]interface{}{key: oldValue},
				NewValues:   map[string]interface{}{key: newValue},
				Timestamp:   time.Now(),
				Author:      "hot_reload",
				Description: fmt.Sprintf("Updated config key '%s' via hot reload", key),
				Status:      ChangeStatusApplied,
			}
			changes = append(changes, change)
		}
	}

	// 检查删除的配置
	for key, oldValue := range oldConfig {
		if _, exists := newConfig[key]; !exists {
			// 删除配置
			change := &ConfigChange{
				ID:          fmt.Sprintf("delete_%s_%d", key, time.Now().Unix()),
				Version:     "hot_reload",
				Type:        hr.getConfigType(key),
				Changes:     map[string]interface{}{key: nil},
				OldValues:   map[string]interface{}{key: oldValue},
				NewValues:   map[string]interface{}{key: nil},
				Timestamp:   time.Now(),
				Author:      "hot_reload",
				Description: fmt.Sprintf("Deleted config key '%s' via hot reload", key),
				Status:      ChangeStatusApplied,
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// applyConfigChanges 应用配置变更
func (hr *HotReloader) applyConfigChanges(newConfig map[string]interface{}) error {
	// 这里应该调用配置管理器的内部方法来更新配置
	// 由于我们使用的是简化版本，这里直接更新内存中的配置
	hr.configManager.mutex.Lock()
	hr.configManager.configs = newConfig
	hr.configManager.mutex.Unlock()

	// 保存到文件
	return hr.configManager.saveConfigs()
}

// notifyCallbacks 通知回调
func (hr *HotReloader) notifyCallbacks(changes []*ConfigChange) {
	hr.mutex.RLock()
	callbacks := make([]func(*ConfigChange), len(hr.callbacks))
	copy(callbacks, hr.callbacks)
	hr.mutex.RUnlock()

	for _, change := range changes {
		for _, callback := range callbacks {
			go callback(change)
		}
	}
}

// getConfigType 获取配置类型
func (hr *HotReloader) getConfigType(key string) ConfigType {
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

// IsEnabled 检查是否启用
func (hr *HotReloader) IsEnabled() bool {
	hr.mutex.RLock()
	defer hr.mutex.RUnlock()
	return hr.enabled
}

// GetConfigFile 获取配置文件路径
func (hr *HotReloader) GetConfigFile() string {
	return hr.configFile
}
