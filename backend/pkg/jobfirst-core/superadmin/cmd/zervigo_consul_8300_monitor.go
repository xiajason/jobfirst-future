package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

// Consul8300Monitor Consul 8300端口监控器原型
type Consul8300Monitor struct {
	config     *Consul8300Config
	client     *api.Client
	metrics    *Consul8300Metrics
	logger     *Logger
}

// Consul8300Config 配置结构
type Consul8300Config struct {
	Host          string        `json:"host"`
	Port          int           `json:"port"`
	CheckInterval time.Duration `json:"check_interval"`
	Timeout       time.Duration `json:"timeout"`
	HTTPPort      int           `json:"http_port"` // 用于API调用
}

// Consul8300Metrics 监控指标
type Consul8300Metrics struct {
	PortAvailability *PortMetrics `json:"port_availability"`
	RPCCommunication *RPCMetrics  `json:"rpc_communication"`
	RaftProtocol     *RaftMetrics `json:"raft_protocol"`
	LastUpdated      time.Time    `json:"last_updated"`
}

// PortMetrics 端口指标
type PortMetrics struct {
	IsListening  bool          `json:"is_listening"`
	ResponseTime time.Duration `json:"response_time"`
	LastCheck    time.Time     `json:"last_check"`
	ErrorCount   int           `json:"error_count"`
	SuccessRate  float64       `json:"success_rate"`
}

// RPCMetrics RPC通信指标
type RPCMetrics struct {
	ActiveConnections int    `json:"active_connections"`
	RequestRate       float64 `json:"request_rate"`
	ErrorRate         float64 `json:"error_rate"`
	AverageLatency    time.Duration `json:"average_latency"`
	RaftState         string  `json:"raft_state"`
	LeaderAddress     string  `json:"leader_address"`
	IsHealthy         bool    `json:"is_healthy"`
	LastHeartbeat     time.Time `json:"last_heartbeat"`
}

// RaftMetrics Raft协议指标
type RaftMetrics struct {
	State        string    `json:"state"`        // Leader/Follower/Candidate
	Term         uint64    `json:"term"`
	LastLogIndex uint64    `json:"last_log_index"`
	CommitIndex  uint64    `json:"commit_index"`
	AppliedIndex uint64    `json:"applied_index"`
	PeerCount    int       `json:"peer_count"`
	IsHealthy    bool      `json:"is_healthy"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// Logger 简单日志器
type Logger struct {
	verbose bool
}

func (l *Logger) Info(msg string) {
	if l.verbose {
		fmt.Printf("ℹ️  %s\n", msg)
	}
}

func (l *Logger) Error(msg string) {
	fmt.Printf("❌ %s\n", msg)
}

func (l *Logger) Success(msg string) {
	fmt.Printf("✅ %s\n", msg)
}

func (l *Logger) Warning(msg string) {
	fmt.Printf("⚠️  %s\n", msg)
}

// NewConsul8300Monitor 创建新的Consul 8300监控器
func NewConsul8300Monitor(config *Consul8300Config) (*Consul8300Monitor, error) {
	// 创建Consul客户端
	clientConfig := api.DefaultConfig()
	clientConfig.Address = fmt.Sprintf("%s:%d", config.Host, config.HTTPPort)
	
	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Consul客户端失败: %v", err)
	}

	return &Consul8300Monitor{
		config: config,
		client: client,
		metrics: &Consul8300Metrics{
			PortAvailability: &PortMetrics{},
			RPCCommunication: &RPCMetrics{},
			RaftProtocol:     &RaftMetrics{},
		},
		logger: &Logger{verbose: true},
	}, nil
}

// Start 启动监控
func (m *Consul8300Monitor) Start() error {
	m.logger.Info("启动Consul 8300端口监控...")
	
	// 启动端口可用性检查
	go m.monitorPortAvailability()
	
	// 启动RPC通信监控
	go m.monitorRPCCommunication()
	
	// 启动Raft协议监控
	go m.monitorRaftProtocol()
	
	m.logger.Success("Consul 8300端口监控已启动")
	return nil
}

// monitorPortAvailability 监控端口可用性
func (m *Consul8300Monitor) monitorPortAvailability() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		start := time.Now()
		
		// 检查端口是否监听
		isListening := m.checkPortListening()
		
		responseTime := time.Since(start)
		
		// 更新指标
		m.metrics.PortAvailability.IsListening = isListening
		m.metrics.PortAvailability.ResponseTime = responseTime
		m.metrics.PortAvailability.LastCheck = time.Now()
		
		if isListening {
			m.metrics.PortAvailability.ErrorCount = 0
			m.metrics.PortAvailability.SuccessRate = 100.0
		} else {
			m.metrics.PortAvailability.ErrorCount++
			m.metrics.PortAvailability.SuccessRate = 0.0
		}
		
		// 检查告警条件
		m.checkPortAlerts(isListening, responseTime)
	}
}

// checkPortListening 检查端口是否监听
func (m *Consul8300Monitor) checkPortListening() bool {
	conn, err := net.DialTimeout("tcp", 
		fmt.Sprintf("%s:%d", m.config.Host, m.config.Port), 
		m.config.Timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// checkPortAlerts 检查端口告警
func (m *Consul8300Monitor) checkPortAlerts(isListening bool, responseTime time.Duration) {
	if !isListening {
		m.logger.Error(fmt.Sprintf("Consul 8300端口不可用 - 响应时间: %v", responseTime))
	} else if responseTime > 1*time.Second {
		m.logger.Warning(fmt.Sprintf("Consul 8300端口响应时间过长: %v", responseTime))
	}
}

// monitorRPCCommunication 监控RPC通信
func (m *Consul8300Monitor) monitorRPCCommunication() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		// 获取RPC统计信息
		stats, err := m.getRPCStats()
		if err != nil {
			m.logger.Error(fmt.Sprintf("获取RPC统计信息失败: %v", err))
			continue
		}
		
		// 更新指标
		m.metrics.RPCCommunication = stats
		
		// 检查告警条件
		m.checkRPCAlerts(stats)
	}
}

// getRPCStats 获取RPC统计信息
func (m *Consul8300Monitor) getRPCStats() (*RPCMetrics, error) {
	// 通过Consul API获取统计信息
	agent := m.client.Agent()
	
	// 获取节点信息
	self, err := agent.Self()
	if err != nil {
		return nil, err
	}
	
	// 获取Raft配置
	raft, err := m.client.Operator().RaftGetConfiguration(nil)
	if err != nil {
		return nil, err
	}
	
	// 解析节点状态
	memberStatus := "unknown"
	if member, ok := self["Member"].(map[string]interface{}); ok {
		if status, ok := member["Status"].(string); ok {
			}
				}
			}
		}
	}
	
	stats := &RPCMetrics{
		ActiveConnections: len(raft.Servers),
		RaftState:         memberStatus,
		LeaderAddress:     m.getLeaderAddress(raft),
		IsHealthy:         memberStatus == "alive",
		LastHeartbeat:     time.Now(),
		RequestRate:       10.5, // 模拟数据
		ErrorRate:         0.1,  // 模拟数据
		AverageLatency:    50 * time.Millisecond, // 模拟数据
	}
	
	return stats, nil
}

// getLeaderAddress 获取领导者地址
func (m *Consul8300Monitor) getLeaderAddress(raft *api.RaftConfiguration) string {
	for _, server := range raft.Servers {
		if server.Leader {
			return server.Address
		}
	}
	return "unknown"
}

// checkRPCAlerts 检查RPC告警
func (m *Consul8300Monitor) checkRPCAlerts(stats *RPCMetrics) {
	if !stats.IsHealthy {
		m.logger.Error("Consul RPC通信不健康")
	} else if stats.ErrorRate > 5.0 {
		m.logger.Warning(fmt.Sprintf("Consul RPC错误率过高: %.2f%%", stats.ErrorRate))
	}
}

// monitorRaftProtocol 监控Raft协议
func (m *Consul8300Monitor) monitorRaftProtocol() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		// 获取Raft状态
		raft, err := m.getRaftStatus()
		if err != nil {
			m.logger.Error(fmt.Sprintf("获取Raft状态失败: %v", err))
			continue
		}
		
		// 更新指标
		m.metrics.RaftProtocol = raft
		
		// 检查告警条件
		m.checkRaftAlerts(raft)
	}
}

// getRaftStatus 获取Raft状态
func (m *Consul8300Monitor) getRaftStatus() (*RaftMetrics, error) {
	// 获取Raft配置
	config, err := m.client.Operator().RaftGetConfiguration(nil)
	if err != nil {
		return nil, err
	}
	
	// 获取Raft统计信息 (使用Status API)
	status, err := m.client.Status().Leader()
	if err != nil {
		return nil, err
	}
	
	// 模拟Raft统计信息
	stats := map[string]interface{}{
		"state":           "Leader",
		"term":            float64(1),
		"last_log_index":  float64(100),
		"commit_index":    float64(95),
		"applied_index":   float64(90),
	}
	
	// 解析Raft状态
	state := "unknown"
	term := uint64(0)
	lastLogIndex := uint64(0)
	commitIndex := uint64(0)
	appliedIndex := uint64(0)
	
	if stateVal, ok := stats["state"].(string); ok {
		state = stateVal
	}
	if termVal, ok := stats["term"].(float64); ok {
		term = uint64(termVal)
	}
	if lastLogIndexVal, ok := stats["last_log_index"].(float64); ok {
		lastLogIndex = uint64(lastLogIndexVal)
	}
	if commitIndexVal, ok := stats["commit_index"].(float64); ok {
		commitIndex = uint64(commitIndexVal)
	}
	if appliedIndexVal, ok := stats["applied_index"].(float64); ok {
		appliedIndex = uint64(appliedIndexVal)
	}
	
	// 使用leader状态判断健康状态
	isHealthy := status != ""
	
	raft := &RaftMetrics{
		State:         state,
		Term:          term,
		LastLogIndex:  lastLogIndex,
		CommitIndex:   commitIndex,
		AppliedIndex:  appliedIndex,
		PeerCount:     len(config.Servers),
		IsHealthy:     isHealthy,
		LastHeartbeat: time.Now(),
	}
	
	return raft, nil
}

// checkRaftAlerts 检查Raft告警
func (m *Consul8300Monitor) checkRaftAlerts(raft *RaftMetrics) {
	if !raft.IsHealthy {
		m.logger.Error(fmt.Sprintf("Consul Raft状态不健康: %s", raft.State))
	} else if raft.State == "Candidate" {
		m.logger.Warning("Consul Raft正在进行领导者选举")
	}
}

// GetMetrics 获取监控指标
func (m *Consul8300Monitor) GetMetrics() *Consul8300Metrics {
	m.metrics.LastUpdated = time.Now()
	return m.metrics
}

// GenerateReport 生成监控报告
func (m *Consul8300Monitor) GenerateReport() {
	fmt.Println("🔍 Consul 8300端口监控报告")
	fmt.Println("==============================")
	fmt.Printf("📅 报告时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	
	// 端口可用性报告
	fmt.Println("🔌 端口可用性:")
	port := m.metrics.PortAvailability
	status := "❌ 不可用"
	if port.IsListening {
		status = "✅ 可用"
	}
	fmt.Printf("  状态: %s\n", status)
	fmt.Printf("  响应时间: %v\n", port.ResponseTime)
	fmt.Printf("  最后检查: %s\n", port.LastCheck.Format("15:04:05"))
	fmt.Printf("  成功率: %.1f%%\n", port.SuccessRate)
	fmt.Println()
	
	// RPC通信报告
	fmt.Println("📡 RPC通信状态:")
	rpc := m.metrics.RPCCommunication
	healthStatus := "❌ 不健康"
	if rpc.IsHealthy {
		healthStatus = "✅ 健康"
	}
	fmt.Printf("  健康状态: %s\n", healthStatus)
	fmt.Printf("  活跃连接: %d\n", rpc.ActiveConnections)
	fmt.Printf("  请求速率: %.1f req/s\n", rpc.RequestRate)
	fmt.Printf("  错误率: %.2f%%\n", rpc.ErrorRate)
	fmt.Printf("  平均延迟: %v\n", rpc.AverageLatency)
	fmt.Printf("  Raft状态: %s\n", rpc.RaftState)
	fmt.Printf("  领导者地址: %s\n", rpc.LeaderAddress)
	fmt.Println()
	
	// Raft协议报告
	fmt.Println("🏛️  Raft协议状态:")
	raft := m.metrics.RaftProtocol
	raftHealth := "❌ 不健康"
	if raft.IsHealthy {
		raftHealth = "✅ 健康"
	}
	fmt.Printf("  健康状态: %s\n", raftHealth)
	fmt.Printf("  当前状态: %s\n", raft.State)
	fmt.Printf("  任期: %d\n", raft.Term)
	fmt.Printf("  最后日志索引: %d\n", raft.LastLogIndex)
	fmt.Printf("  提交索引: %d\n", raft.CommitIndex)
	fmt.Printf("  应用索引: %d\n", raft.AppliedIndex)
	fmt.Printf("  节点数量: %d\n", raft.PeerCount)
	fmt.Println()
	
	// 综合评估
	fmt.Println("📊 综合评估:")
	overallHealth := "❌ 不健康"
	if port.IsListening && rpc.IsHealthy && raft.IsHealthy {
		overallHealth = "✅ 健康"
	} else if port.IsListening && (rpc.IsHealthy || raft.IsHealthy) {
		overallHealth = "⚠️  部分健康"
	}
	fmt.Printf("  整体状态: %s\n", overallHealth)
	
	// 建议
	fmt.Println("\n💡 建议:")
	if !port.IsListening {
		fmt.Println("  - 检查Consul服务是否正常运行")
		fmt.Println("  - 检查8300端口是否被防火墙阻止")
	}
	if !rpc.IsHealthy {
		fmt.Println("  - 检查Consul集群配置")
		fmt.Println("  - 检查网络连接状态")
	}
	if !raft.IsHealthy {
		fmt.Println("  - 检查Raft协议配置")
		fmt.Println("  - 检查集群节点状态")
	}
	if port.IsListening && rpc.IsHealthy && raft.IsHealthy {
		fmt.Println("  - 系统运行正常，建议定期监控")
	}
}

// SaveReport 保存报告到文件
func (m *Consul8300Monitor) SaveReport() error {
	report := map[string]interface{}{
		"timestamp": time.Now(),
		"metrics":   m.metrics,
		"summary": map[string]interface{}{
			"port_available":    m.metrics.PortAvailability.IsListening,
			"rpc_healthy":       m.metrics.RPCCommunication.IsHealthy,
			"raft_healthy":      m.metrics.RaftProtocol.IsHealthy,
			"overall_health":    m.metrics.PortAvailability.IsListening && 
								m.metrics.RPCCommunication.IsHealthy && 
								m.metrics.RaftProtocol.IsHealthy,
		},
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile("consul_8300_monitor_report.json", data, 0644)
}

// 主函数 - 演示Consul 8300监控
func main() {
	fmt.Println("🚀 Consul 8300端口监控器")
	fmt.Println("=========================")
	fmt.Println("监控Consul RPC通信和Raft协议状态")
	fmt.Println()
	
	// 配置
	config := &Consul8300Config{
		Host:          "localhost",
		Port:          8300,
		HTTPPort:      8500,
		CheckInterval: 30 * time.Second,
		Timeout:       5 * time.Second,
	}
	
	// 创建监控器
	monitor, err := NewConsul8300Monitor(config)
	if err != nil {
		fmt.Printf("❌ 创建监控器失败: %v\n", err)
		os.Exit(1)
	}
	
	// 启动监控
	if err := monitor.Start(); err != nil {
		fmt.Printf("❌ 启动监控失败: %v\n", err)
		os.Exit(1)
	}
	
	// 等待一段时间收集数据
	fmt.Println("⏳ 收集监控数据中...")
	time.Sleep(35 * time.Second)
	
	// 生成报告
	monitor.GenerateReport()
	
	// 保存报告
	if err := monitor.SaveReport(); err != nil {
		fmt.Printf("❌ 保存报告失败: %v\n", err)
	} else {
		fmt.Println("\n📄 详细报告已保存到: consul_8300_monitor_report.json")
	}
	
	fmt.Println("\n🎉 Consul 8300端口监控演示完成！")
}
