package apigateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// HealthStatus 健康状态
type HealthStatus struct {
	ServiceID    string    `json:"service_id"`
	ServiceName  string    `json:"service_name"`
	Status       string    `json:"status"`
	LastCheck    time.Time `json:"last_check"`
	ResponseTime int64     `json:"response_time_ms"`
	Error        string    `json:"error,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	serviceRegistry *ServiceRegistry
	interval        time.Duration
	timeout         time.Duration
	statuses        map[string]*HealthStatus
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(interval, timeout time.Duration) (*HealthChecker, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("健康检查间隔必须大于0")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("健康检查超时时间必须大于0")
	}

	ctx, cancel := context.WithCancel(context.Background())

	checker := &HealthChecker{
		interval: interval,
		timeout:  timeout,
		statuses: make(map[string]*HealthStatus),
		ctx:      ctx,
		cancel:   cancel,
	}

	return checker, nil
}

// SetServiceRegistry 设置服务注册管理器
func (hc *HealthChecker) SetServiceRegistry(registry *ServiceRegistry) {
	hc.serviceRegistry = registry
}

// Start 启动健康检查器
func (hc *HealthChecker) Start() error {
	if hc.serviceRegistry == nil {
		return fmt.Errorf("服务注册管理器未设置")
	}

	log.Println("🔍 启动健康检查器...")

	hc.wg.Add(1)
	go hc.run()

	return nil
}

// Stop 停止健康检查器
func (hc *HealthChecker) Stop() {
	log.Println("🛑 停止健康检查器...")
	hc.cancel()
	hc.wg.Wait()
	log.Println("✅ 健康检查器已停止")
}

// run 运行健康检查循环
func (hc *HealthChecker) run() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.checkAllServices()
		}
	}
}

// checkAllServices 检查所有服务
func (hc *HealthChecker) checkAllServices() {
	services, err := hc.serviceRegistry.ListServices()
	if err != nil {
		log.Printf("❌ 获取服务列表失败: %v", err)
		return
	}

	for _, service := range services {
		hc.checkService(service)
	}
}

// checkService 检查单个服务
func (hc *HealthChecker) checkService(service *service_registry.ServiceInfo) {
	start := time.Now()

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: hc.timeout,
	}

	// 构建健康检查URL
	healthURL := fmt.Sprintf("http://%s:%d%s", service.Address, service.Port, service.HealthCheck)

	// 发送健康检查请求
	resp, err := client.Get(healthURL)
	responseTime := time.Since(start).Milliseconds()

	// 更新健康状态
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	status := &HealthStatus{
		ServiceID:    service.ID,
		ServiceName:  service.Name,
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
	}

	if err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
		log.Printf("❌ 服务 %s 健康检查失败: %v", service.Name, err)
	} else {
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status.Status = "healthy"
			log.Printf("✅ 服务 %s 健康检查通过 (响应时间: %dms)", service.Name, responseTime)
		} else {
			status.Status = "unhealthy"
			status.Error = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
			log.Printf("❌ 服务 %s 健康检查失败: HTTP状态码 %d", service.Name, resp.StatusCode)
		}
	}

	hc.statuses[service.ID] = status
}

// GetHealthStatus 获取服务健康状态
func (hc *HealthChecker) GetHealthStatus(serviceID string) (*HealthStatus, error) {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	status, exists := hc.statuses[serviceID]
	if !exists {
		return nil, fmt.Errorf("服务健康状态不存在: %s", serviceID)
	}

	return status, nil
}

// GetAllHealthStatuses 获取所有服务健康状态
func (hc *HealthChecker) GetAllHealthStatuses() map[string]*HealthStatus {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	// 创建副本
	statuses := make(map[string]*HealthStatus)
	for id, status := range hc.statuses {
		statuses[id] = status
	}

	return statuses
}

// GetHealthyServices 获取健康的服务
func (hc *HealthChecker) GetHealthyServices() []string {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	var healthyServices []string
	for id, status := range hc.statuses {
		if status.Status == "healthy" {
			healthyServices = append(healthyServices, id)
		}
	}

	return healthyServices
}

// GetUnhealthyServices 获取不健康的服务
func (hc *HealthChecker) GetUnhealthyServices() []string {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	var unhealthyServices []string
	for id, status := range hc.statuses {
		if status.Status == "unhealthy" {
			unhealthyServices = append(unhealthyServices, id)
		}
	}

	return unhealthyServices
}

// GetServiceHealthSummary 获取服务健康摘要
func (hc *HealthChecker) GetServiceHealthSummary() map[string]interface{} {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	total := len(hc.statuses)
	healthy := 0
	unhealthy := 0

	for _, status := range hc.statuses {
		if status.Status == "healthy" {
			healthy++
		} else {
			unhealthy++
		}
	}

	return map[string]interface{}{
		"total":     total,
		"healthy":   healthy,
		"unhealthy": unhealthy,
		"timestamp": time.Now(),
	}
}
