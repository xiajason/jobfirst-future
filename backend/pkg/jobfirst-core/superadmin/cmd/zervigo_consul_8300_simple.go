package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

// Consul8300SimpleMonitor 简化的Consul 8300端口监控器
type Consul8300SimpleMonitor struct {
	config *Consul8300Config
	client *api.Client
}

// Consul8300Config 配置结构
type Consul8300Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	HTTPPort int    `json:"http_port"`
}

// Consul8300Status 监控状态
type Consul8300Status struct {
	PortAvailable    bool          `json:"port_available"`
	HTTPAPIHealthy   bool          `json:"http_api_healthy"`
	RaftLeaderExists bool          `json:"raft_leader_exists"`
	LastCheck        time.Time     `json:"last_check"`
	ResponseTime     time.Duration `json:"response_time"`
}

// NewConsul8300SimpleMonitor 创建简化的监控器
func NewConsul8300SimpleMonitor(config *Consul8300Config) (*Consul8300SimpleMonitor, error) {
	// 创建Consul客户端
	clientConfig := api.DefaultConfig()
	clientConfig.Address = fmt.Sprintf("%s:%d", config.Host, config.HTTPPort)

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Consul客户端失败: %v", err)
	}

	return &Consul8300SimpleMonitor{
		config: config,
		client: client,
	}, nil
}

// CheckPort 检查8300端口是否可用
func (m *Consul8300SimpleMonitor) CheckPort() bool {
	conn, err := net.DialTimeout("tcp",
		fmt.Sprintf("%s:%d", m.config.Host, m.config.Port),
		5*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// CheckHTTPAPI 检查HTTP API是否健康
func (m *Consul8300SimpleMonitor) CheckHTTPAPI() bool {
	_, err := m.client.Status().Leader()
	return err == nil
}

// CheckRaftLeader 检查Raft领导者是否存在
func (m *Consul8300SimpleMonitor) CheckRaftLeader() bool {
	leader, err := m.client.Status().Leader()
	return err == nil && leader != ""
}

// GetStatus 获取完整状态
func (m *Consul8300SimpleMonitor) GetStatus() *Consul8300Status {
	start := time.Now()

	status := &Consul8300Status{
		LastCheck: time.Now(),
	}

	// 检查8300端口
	status.PortAvailable = m.CheckPort()

	// 检查HTTP API
	status.HTTPAPIHealthy = m.CheckHTTPAPI()

	// 检查Raft领导者
	status.RaftLeaderExists = m.CheckRaftLeader()

	status.ResponseTime = time.Since(start)

	return status
}

// GenerateReport 生成监控报告
func (m *Consul8300SimpleMonitor) GenerateReport() {
	fmt.Println("🔍 Consul 8300端口监控报告")
	fmt.Println("==============================")
	fmt.Printf("📅 报告时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	status := m.GetStatus()

	// 端口状态
	fmt.Println("🔌 端口状态:")
	portStatus := "❌ 不可用"
	if status.PortAvailable {
		portStatus = "✅ 可用"
	}
	fmt.Printf("  8300端口: %s\n", portStatus)
	fmt.Printf("  响应时间: %v\n", status.ResponseTime)
	fmt.Println()

	// HTTP API状态
	fmt.Println("🌐 HTTP API状态:")
	apiStatus := "❌ 不健康"
	if status.HTTPAPIHealthy {
		apiStatus = "✅ 健康"
	}
	fmt.Printf("  8500端口API: %s\n", apiStatus)
	fmt.Println()

	// Raft状态
	fmt.Println("🏛️  Raft协议状态:")
	raftStatus := "❌ 无领导者"
	if status.RaftLeaderExists {
		raftStatus = "✅ 有领导者"
	}
	fmt.Printf("  集群状态: %s\n", raftStatus)
	fmt.Println()

	// 综合评估
	fmt.Println("📊 综合评估:")
	overallHealth := "❌ 不健康"
	if status.PortAvailable && status.HTTPAPIHealthy && status.RaftLeaderExists {
		overallHealth = "✅ 健康"
	} else if status.PortAvailable && (status.HTTPAPIHealthy || status.RaftLeaderExists) {
		overallHealth = "⚠️  部分健康"
	}
	fmt.Printf("  整体状态: %s\n", overallHealth)

	// 建议
	fmt.Println("\n💡 建议:")
	if !status.PortAvailable {
		fmt.Println("  - 检查Consul服务是否正常运行")
		fmt.Println("  - 检查8300端口是否被防火墙阻止")
	}
	if !status.HTTPAPIHealthy {
		fmt.Println("  - 检查Consul HTTP API配置")
		fmt.Println("  - 检查8500端口是否正常监听")
	}
	if !status.RaftLeaderExists {
		fmt.Println("  - 检查Consul集群配置")
		fmt.Println("  - 检查Raft协议状态")
	}
	if status.PortAvailable && status.HTTPAPIHealthy && status.RaftLeaderExists {
		fmt.Println("  - 系统运行正常，建议定期监控")
	}
}

// SaveReport 保存报告到文件
func (m *Consul8300SimpleMonitor) SaveReport() error {
	status := m.GetStatus()

	report := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    status,
		"summary": map[string]interface{}{
			"port_available":     status.PortAvailable,
			"http_api_healthy":   status.HTTPAPIHealthy,
			"raft_leader_exists": status.RaftLeaderExists,
			"overall_health":     status.PortAvailable && status.HTTPAPIHealthy && status.RaftLeaderExists,
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("consul_8300_simple_report.json", data, 0644)
}

// 主函数 - 演示Consul 8300监控
func main() {
	fmt.Println("🚀 Consul 8300端口监控器 (简化版)")
	fmt.Println("==================================")
	fmt.Println("监控Consul RPC通信和Raft协议状态")
	fmt.Println()

	// 配置
	config := &Consul8300Config{
		Host:     "localhost",
		Port:     8300,
		HTTPPort: 8500,
	}

	// 创建监控器
	monitor, err := NewConsul8300SimpleMonitor(config)
	if err != nil {
		fmt.Printf("❌ 创建监控器失败: %v\n", err)
		os.Exit(1)
	}

	// 生成报告
	monitor.GenerateReport()

	// 保存报告
	if err := monitor.SaveReport(); err != nil {
		fmt.Printf("❌ 保存报告失败: %v\n", err)
	} else {
		fmt.Println("\n📄 详细报告已保存到: consul_8300_simple_report.json")
	}

	fmt.Println("\n🎉 Consul 8300端口监控演示完成！")
}
