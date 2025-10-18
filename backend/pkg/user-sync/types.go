package usersync

import (
	"time"
)

// UserEvent 用户事件
type UserEvent struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	UserID    uint        `json:"user_id"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
}

// EventType 事件类型
type EventType string

const (
	EventTypeUserCreated       EventType = "user.created"
	EventTypeUserUpdated       EventType = "user.updated"
	EventTypeUserDeleted       EventType = "user.deleted"
	EventTypeUserStatusChanged EventType = "user.status_changed"
)

// UserSyncTask 用户同步任务
type UserSyncTask struct {
	ID         string                 `json:"id"`
	UserID     uint                   `json:"user_id"`
	Username   string                 `json:"username"`
	EventType  EventType              `json:"event_type"`
	Targets    []SyncTarget           `json:"targets"`
	Data       map[string]interface{} `json:"data"`
	Priority   int                    `json:"priority"`
	Status     SyncTaskStatus         `json:"status"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Error      string                 `json:"error,omitempty"`
}

// SyncTarget 同步目标
type SyncTarget struct {
	Service string `json:"service"`
	URL     string `json:"url"`
	Method  string `json:"method"`
	Enabled bool   `json:"enabled"`
}

// SyncTaskStatus 同步任务状态
type SyncTaskStatus string

const (
	SyncTaskStatusPending    SyncTaskStatus = "pending"
	SyncTaskStatusProcessing SyncTaskStatus = "processing"
	SyncTaskStatusCompleted  SyncTaskStatus = "completed"
	SyncTaskStatusFailed     SyncTaskStatus = "failed"
	SyncTaskStatusRetrying   SyncTaskStatus = "retrying"
)

// SyncResult 同步结果
type SyncResult struct {
	TaskID    string        `json:"task_id"`
	Target    string        `json:"target"`
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// User 用户数据结构
type User struct {
	ID           uint       `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // 不同步密码哈希
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	Phone        string     `json:"phone"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// SyncConfig 同步配置
type SyncConfig struct {
	Enabled          bool          `json:"enabled"`
	Workers          int           `json:"workers"`
	QueueSize        int           `json:"queue_size"`
	RetryInterval    time.Duration `json:"retry_interval"`
	MaxRetries       int           `json:"max_retries"`
	Timeout          time.Duration `json:"timeout"`
	ConsistencyCheck bool          `json:"consistency_check"`
	CheckInterval    time.Duration `json:"check_interval"`
	AutoRepair       bool          `json:"auto_repair"`
}

// DefaultSyncConfig 默认同步配置
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		Enabled:          true,
		Workers:          3,
		QueueSize:        1000,
		RetryInterval:    5 * time.Second,
		MaxRetries:       3,
		Timeout:          30 * time.Second,
		ConsistencyCheck: true,
		CheckInterval:    5 * time.Minute,
		AutoRepair:       true,
	}
}
