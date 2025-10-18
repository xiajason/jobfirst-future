package config

import (
	"time"
)

// ConfigType 配置类型
type ConfigType string

const (
	ConfigTypeDatabase ConfigType = "database"
	ConfigTypeService  ConfigType = "service"
	ConfigTypeSystem   ConfigType = "system"
	ConfigTypeSecurity ConfigType = "security"
	ConfigTypeLogging  ConfigType = "logging"
	ConfigTypeCache    ConfigType = "cache"
)

// ConfigItem 配置项
type ConfigItem struct {
	Key         string            `json:"key"`
	Value       interface{}       `json:"value"`
	Type        ConfigType        `json:"type"`
	Description string            `json:"description"`
	Default     interface{}       `json:"default"`
	Required    bool              `json:"required"`
	Validators  []Validator       `json:"validators,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ConfigVersion 配置版本
type ConfigVersion struct {
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Author      string                 `json:"author"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Checksum    string                 `json:"checksum"`
	Tags        []string               `json:"tags,omitempty"`
}

// ConfigChange 配置变更
type ConfigChange struct {
	ID          string                 `json:"id"`
	Version     string                 `json:"version"`
	Type        ConfigType             `json:"type"`
	Changes     map[string]interface{} `json:"changes"`
	OldValues   map[string]interface{} `json:"old_values"`
	NewValues   map[string]interface{} `json:"new_values"`
	Timestamp   time.Time              `json:"timestamp"`
	Author      string                 `json:"author"`
	Description string                 `json:"description"`
	Status      ChangeStatus           `json:"status"`
}

// ChangeStatus 变更状态
type ChangeStatus string

const (
	ChangeStatusPending    ChangeStatus = "pending"
	ChangeStatusApplied    ChangeStatus = "applied"
	ChangeStatusFailed     ChangeStatus = "failed"
	ChangeStatusRolledBack ChangeStatus = "rolled_back"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// 基础配置管理
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Delete(key string) error
	GetAll() (map[string]interface{}, error)

	// 配置类型管理
	GetByType(configType ConfigType) (map[string]interface{}, error)
	SetByType(configType ConfigType, config map[string]interface{}) error

	// 版本管理
	GetVersion(version string) (*ConfigVersion, error)
	GetVersions() ([]*ConfigVersion, error)
	CreateVersion(description string, author string) (*ConfigVersion, error)
	RollbackToVersion(version string) error

	// 热更新
	WatchChanges(callback func(*ConfigChange)) error
	StopWatching() error

	// 验证
	Validate(config map[string]interface{}) error
	ValidateItem(item *ConfigItem) error
}

// Validator 配置验证器接口
type Validator interface {
	Validate(value interface{}) error
	GetName() string
}

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	ID       string
	Callback func(*ConfigChange)
	Filter   ConfigType
	Active   bool
}

// ConfigManagerConfig 配置管理器配置
type ConfigManagerConfig struct {
	StoragePath   string        `json:"storage_path"`
	BackupPath    string        `json:"backup_path"`
	MaxVersions   int           `json:"max_versions"`
	WatchInterval time.Duration `json:"watch_interval"`
	AutoBackup    bool          `json:"auto_backup"`
	EncryptionKey string        `json:"encryption_key,omitempty"`
	Compression   bool          `json:"compression"`
}
