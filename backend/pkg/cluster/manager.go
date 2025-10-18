package cluster

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ClusterManager 集群管理器接口
type ClusterManager interface {
	// 集群节点管理
	RegisterNode(nodeID string, config NodeConfig) error
	UnregisterNode(nodeID string) error
	GetNodeStatus(nodeID string) (*NodeStatus, error)
	
	// 集群数据同步
	SyncUserData(userID uint, data UserData) error
	SyncConfigData(configID string, data ConfigData) error
	
	// 集群状态管理
	GetClusterStatus() (*ClusterStatus, error)
	HandleNodeFailure(nodeID string) error
	
	// 启动和停止
	Start() error
	Stop()
}

// Manager 集群管理器实现
type Manager struct {
	config      *ClusterConfig
	nodes       map[string]*ClusterNode
	nodesMutex  sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	started     bool
	startedMutex sync.RWMutex
}

// NewManager 创建新的集群管理器
func NewManager(config *ClusterConfig) *Manager {
	if config == nil {
		config = DefaultClusterConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Manager{
		config: config,
		nodes:  make(map[string]*ClusterNode),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动集群管理器
func (m *Manager) Start() error {
	m.startedMutex.Lock()
	defer m.startedMutex.Unlock()
	
	if m.started {
		return fmt.Errorf("cluster manager is already started")
	}
	
	log.Printf("启动集群管理器: %s", m.config.ClusterID)
	
	// 启动健康检查
	go m.startHealthCheck()
	
	// 启动统计更新
	go m.startStatsUpdate()
	
	m.started = true
	log.Printf("集群管理器启动成功: %s", m.config.ClusterID)
	
	return nil
}

// Stop 停止集群管理器
func (m *Manager) Stop() {
	m.startedMutex.Lock()
	defer m.startedMutex.Unlock()
	
	if !m.started {
		return
	}
	
	log.Printf("停止集群管理器: %s", m.config.ClusterID)
	
	m.cancel()
	m.started = false
	
	log.Printf("集群管理器已停止: %s", m.config.ClusterID)
}

// RegisterNode 注册节点到集群
func (m *Manager) RegisterNode(nodeID string, config NodeConfig) error {
	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()
	
	if len(m.nodes) >= m.config.MaxNodes {
		return fmt.Errorf("cluster is full, max nodes: %d", m.config.MaxNodes)
	}
	
	// 检查节点是否已存在
	if _, exists := m.nodes[nodeID]; exists {
		return fmt.Errorf("node %s already exists", nodeID)
	}
	
	// 创建集群节点
	clusterNode := &ClusterNode{
		Config:      &config,
		LastSeen:    time.Now(),
		Connections: 0,
		ResponseTime: 0,
		SuccessRate: 100.0,
	}
	
	m.nodes[nodeID] = clusterNode
	
	log.Printf("节点注册成功: %s (%s:%d)", nodeID, config.Host, config.Port)
	
	return nil
}

// UnregisterNode 从集群中注销节点
func (m *Manager) UnregisterNode(nodeID string) error {
	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()
	
	if _, exists := m.nodes[nodeID]; !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}
	
	delete(m.nodes, nodeID)
	
	log.Printf("节点注销成功: %s", nodeID)
	
	return nil
}

// GetNodeStatus 获取节点状态
func (m *Manager) GetNodeStatus(nodeID string) (*NodeStatus, error) {
	m.nodesMutex.RLock()
	defer m.nodesMutex.RUnlock()
	
	node, exists := m.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}
	
	return &node.Config.Status, nil
}

// SyncUserData 同步用户数据到集群
func (m *Manager) SyncUserData(userID uint, data UserData) error {
	// 这里实现用户数据同步逻辑
	// 暂时返回成功，后续会实现具体的同步逻辑
	log.Printf("同步用户数据: userID=%d, username=%s", userID, data.Username)
	return nil
}

// SyncConfigData 同步配置数据到集群
func (m *Manager) SyncConfigData(configID string, data ConfigData) error {
	// 这里实现配置数据同步逻辑
	// 暂时返回成功，后续会实现具体的同步逻辑
	log.Printf("同步配置数据: configID=%s, version=%d", configID, data.Version)
	return nil
}

// GetClusterStatus 获取集群状态
func (m *Manager) GetClusterStatus() (*ClusterStatus, error) {
	m.nodesMutex.RLock()
	defer m.nodesMutex.RUnlock()
	
	status := &ClusterStatus{
		ClusterID:     m.config.ClusterID,
		TotalNodes:    len(m.nodes),
		ActiveNodes:   0,
		FailedNodes:   0,
		TotalRequests: 0,
		SuccessfulRequests: 0,
		FailedRequests: 0,
		AverageResponseTime: 0,
		LastUpdated:   time.Now(),
		Nodes:         make(map[string]*ClusterNode),
	}
	
	// 统计节点状态
	for nodeID, node := range m.nodes {
		status.Nodes[nodeID] = node
		status.TotalRequests += int64(node.Connections)
		
		switch node.Config.Status {
		case NodeStatusActive:
			status.ActiveNodes++
		case NodeStatusFailed:
			status.FailedNodes++
		}
	}
	
	return status, nil
}

// HandleNodeFailure 处理节点故障
func (m *Manager) HandleNodeFailure(nodeID string) error {
	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()
	
	node, exists := m.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}
	
	// 标记节点为故障状态
	node.Config.Status = NodeStatusFailed
	node.LastSeen = time.Now()
	
	log.Printf("节点故障处理: %s", nodeID)
	
	// 这里可以添加故障转移逻辑
	// 例如：将流量转移到其他节点
	
	return nil
}

// startHealthCheck 启动健康检查
func (m *Manager) startHealthCheck() {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (m *Manager) performHealthCheck() {
	m.nodesMutex.RLock()
	nodes := make(map[string]*ClusterNode)
	for nodeID, node := range m.nodes {
		nodes[nodeID] = node
	}
	m.nodesMutex.RUnlock()
	
	for nodeID, node := range nodes {
		// 检查节点是否超时
		if time.Since(node.LastSeen) > m.config.NodeTimeout {
			log.Printf("节点超时: %s", nodeID)
			m.HandleNodeFailure(nodeID)
		}
	}
}

// startStatsUpdate 启动统计更新
func (m *Manager) startStatsUpdate() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateStats()
		}
	}
}

// updateStats 更新统计信息
func (m *Manager) updateStats() {
	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()
	
	// 更新节点统计信息
	for _, node := range m.nodes {
		node.LastSeen = time.Now()
		// 这里可以添加更复杂的统计逻辑
	}
}
