package cluster

import (
	"time"
)

// ClusterConfig 集群配置
type ClusterConfig struct {
	// 集群基础配置
	ClusterID         string        `json:"cluster_id" yaml:"cluster_id"`
	NodeID           string        `json:"node_id" yaml:"node_id"`
	MaxNodes         int           `json:"max_nodes" yaml:"max_nodes"`
	
	// 负载均衡配置
	LoadBalanceStrategy string     `json:"load_balance_strategy" yaml:"load_balance_strategy"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	NodeTimeout        time.Duration `json:"node_timeout" yaml:"node_timeout"`
	
	// 数据同步配置
	SyncWorkers       int           `json:"sync_workers" yaml:"sync_workers"`
	SyncQueueSize     int           `json:"sync_queue_size" yaml:"sync_queue_size"`
	SyncRetryInterval time.Duration `json:"sync_retry_interval" yaml:"sync_retry_interval"`
	MaxSyncRetries    int           `json:"max_sync_retries" yaml:"max_sync_retries"`
	SyncTimeout       time.Duration `json:"sync_timeout" yaml:"sync_timeout"`
	
	// 监控配置
	MetricsEnabled bool   `json:"metrics_enabled" yaml:"metrics_enabled"`
	LogLevel      string `json:"log_level" yaml:"log_level"`
}

// NodeConfig 节点配置
type NodeConfig struct {
	NodeID    string            `json:"node_id" yaml:"node_id"`
	Host      string            `json:"host" yaml:"host"`
	Port      int               `json:"port" yaml:"port"`
	Weight    int               `json:"weight" yaml:"weight"`
	Status    NodeStatus        `json:"status" yaml:"status"`
	Metadata  map[string]string `json:"metadata" yaml:"metadata"`
	CreatedAt time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"`
}

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusActive   NodeStatus = "active"
	NodeStatusInactive NodeStatus = "inactive"
	NodeStatusFailed   NodeStatus = "failed"
	NodeStatusMaintenance NodeStatus = "maintenance"
)

// ClusterNode 集群节点
type ClusterNode struct {
	Config     *NodeConfig `json:"config" yaml:"config"`
	LastSeen   time.Time   `json:"last_seen" yaml:"last_seen"`
	Connections int        `json:"connections" yaml:"connections"`
	ResponseTime time.Duration `json:"response_time" yaml:"response_time"`
	SuccessRate float64    `json:"success_rate" yaml:"success_rate"`
}

// ClusterStatus 集群状态
type ClusterStatus struct {
	ClusterID     string                 `json:"cluster_id" yaml:"cluster_id"`
	TotalNodes    int                    `json:"total_nodes" yaml:"total_nodes"`
	ActiveNodes   int                    `json:"active_nodes" yaml:"active_nodes"`
	FailedNodes   int                    `json:"failed_nodes" yaml:"failed_nodes"`
	TotalRequests int64                  `json:"total_requests" yaml:"total_requests"`
	SuccessfulRequests int64             `json:"successful_requests" yaml:"successful_requests"`
	FailedRequests int64                 `json:"failed_requests" yaml:"failed_requests"`
	AverageResponseTime time.Duration    `json:"average_response_time" yaml:"average_response_time"`
	LastUpdated   time.Time              `json:"last_updated" yaml:"last_updated"`
	Nodes         map[string]*ClusterNode `json:"nodes" yaml:"nodes"`
}

// NodeStats 节点统计信息
type NodeStats struct {
	NodeID       string        `json:"node_id" yaml:"node_id"`
	Status       NodeStatus    `json:"status" yaml:"status"`
	Connections  int           `json:"connections" yaml:"connections"`
	ResponseTime time.Duration `json:"response_time" yaml:"response_time"`
	SuccessRate  float64       `json:"success_rate" yaml:"success_rate"`
	LastSeen     time.Time     `json:"last_seen" yaml:"last_seen"`
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	StrategyRoundRobin LoadBalanceStrategy = "round-robin"
	StrategyLeastConn  LoadBalanceStrategy = "least-connections"
	StrategyWeighted   LoadBalanceStrategy = "weighted"
	StrategyIPHash     LoadBalanceStrategy = "ip-hash"
)

// ClusterSyncData 集群同步数据
type ClusterSyncData struct {
	TaskID     string                 `json:"task_id" yaml:"task_id"`
	SourceNode string                 `json:"source_node" yaml:"source_node"`
	TargetNodes []string              `json:"target_nodes" yaml:"target_nodes"`
	DataType   string                 `json:"data_type" yaml:"data_type"`
	Data       map[string]interface{} `json:"data" yaml:"data"`
	Priority   int                    `json:"priority" yaml:"priority"`
	Timestamp  time.Time              `json:"timestamp" yaml:"timestamp"`
}

// SyncTarget 同步目标
type SyncTarget string

const (
	SyncTargetBasicServer   SyncTarget = "basic-server"
	SyncTargetUnifiedAuth   SyncTarget = "unified-auth"
	SyncTargetUserService   SyncTarget = "user-service"
	SyncTargetAIService     SyncTarget = "ai-service"
	SyncTargetResumeService SyncTarget = "resume-service"
	SyncTargetCompanyService SyncTarget = "company-service"
	SyncTargetJobService    SyncTarget = "job-service"
)

// SyncMode 同步模式
type SyncMode string

const (
	SyncModeRealtime   SyncMode = "realtime"   // 实时同步
	SyncModeNearRealtime SyncMode = "near-realtime" // 准实时同步
	SyncModeBatch      SyncMode = "batch"      // 批量同步
)

// SyncStrategy 同步策略
type SyncStrategy struct {
	Mode            SyncMode        `json:"mode" yaml:"mode"`
	RetryConfig     RetryConfig     `json:"retry_config" yaml:"retry_config"`
	Timeout         time.Duration   `json:"timeout" yaml:"timeout"`
	BatchSize       int             `json:"batch_size" yaml:"batch_size"`
	Compression     bool            `json:"compression" yaml:"compression"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	RetryInterval   time.Duration `json:"retry_interval" yaml:"retry_interval"`
	BackoffFactor   float64       `json:"backoff_factor" yaml:"backoff_factor"`
	MaxRetryInterval time.Duration `json:"max_retry_interval" yaml:"max_retry_interval"`
}

// SyncStats 同步统计信息
type SyncStats struct {
	TotalTasks      int64             `json:"total_tasks" yaml:"total_tasks"`
	CompletedTasks  int64             `json:"completed_tasks" yaml:"completed_tasks"`
	FailedTasks     int64             `json:"failed_tasks" yaml:"failed_tasks"`
	RetryTasks      int64             `json:"retry_tasks" yaml:"retry_tasks"`
	AverageLatency  time.Duration     `json:"average_latency" yaml:"average_latency"`
	LastUpdated     time.Time         `json:"last_updated" yaml:"last_updated"`
	TargetStats     map[string]int64  `json:"target_stats" yaml:"target_stats"`
}

// UserData 用户数据
type UserData struct {
	ID       uint   `json:"id" yaml:"id"`
	Username string `json:"username" yaml:"username"`
	Email    string `json:"email" yaml:"email"`
	Role     string `json:"role" yaml:"role"`
	Status   string `json:"status" yaml:"status"`
	Phone    string `json:"phone" yaml:"phone"`
}

// ConfigData 配置数据
type ConfigData struct {
	ConfigID string                 `json:"config_id" yaml:"config_id"`
	Data     map[string]interface{} `json:"data" yaml:"data"`
	Version  int                    `json:"version" yaml:"version"`
}

// DefaultClusterConfig 返回默认集群配置
func DefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		ClusterID:         "jobfirst-cluster-1",
		NodeID:           "basic-server-node-1",
		MaxNodes:         10,
		LoadBalanceStrategy: "round-robin",
		HealthCheckInterval: 30 * time.Second,
		NodeTimeout:       60 * time.Second,
		SyncWorkers:       5,
		SyncQueueSize:     2000,
		SyncRetryInterval: 5 * time.Second,
		MaxSyncRetries:    3,
		SyncTimeout:       30 * time.Second,
		MetricsEnabled:    true,
		LogLevel:         "info",
	}
}
