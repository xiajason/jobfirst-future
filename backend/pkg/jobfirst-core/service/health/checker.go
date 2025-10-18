package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jobfirst/jobfirst-core/errors"
	"github.com/jobfirst/jobfirst-core/service/registry"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	checks   map[string]HealthCheckFunc
	interval time.Duration
	timeout  time.Duration
	mutex    sync.RWMutex
}

// HealthCheckFunc 健康检查函数类型
type HealthCheckFunc func(ctx context.Context, service *registry.ServiceInfo) error

// NewHealthChecker 创建健康检查器实现
func NewHealthChecker(interval, timeout time.Duration) (*HealthChecker, error) {
	if interval <= 0 {
		return nil, errors.NewError(errors.ErrCodeValidation, "health check interval must be positive")
	}
	if timeout <= 0 {
		return nil, errors.NewError(errors.ErrCodeValidation, "health check timeout must be positive")
	}

	checker := &HealthChecker{
		checks:   make(map[string]HealthCheckFunc),
		interval: interval,
		timeout:  timeout,
	}

	// 注册默认的健康检查
	checker.registerDefaultChecks()

	return checker, nil
}

// RegisterCheck 注册健康检查函数
func (hc *HealthChecker) RegisterCheck(name string, checkFunc HealthCheckFunc) error {
	if name == "" {
		return errors.NewError(errors.ErrCodeValidation, "check name cannot be empty")
	}
	if checkFunc == nil {
		return errors.NewError(errors.ErrCodeValidation, "check function cannot be nil")
	}

	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.checks[name] = checkFunc
	return nil
}

// CheckHealth 检查服务健康状态
func (hc *HealthChecker) CheckHealth(service *registry.ServiceInfo) (*registry.HealthStatus, error) {
	if service == nil {
		return nil, errors.NewError(errors.ErrCodeValidation, "service cannot be nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	// 执行所有注册的健康检查
	var lastError error
	details := make(map[string]string)

	hc.mutex.RLock()
	checks := make(map[string]HealthCheckFunc)
	for name, checkFunc := range hc.checks {
		checks[name] = checkFunc
	}
	hc.mutex.RUnlock()

	for name, checkFunc := range checks {
		if err := checkFunc(ctx, service); err != nil {
			details[name] = err.Error()
			lastError = err
		} else {
			details[name] = "ok"
		}
	}

	// 确定整体健康状态
	status := "healthy"
	message := "all checks passed"

	if lastError != nil {
		status = "unhealthy"
		message = lastError.Error()
	}

	return &registry.HealthStatus{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Details:   details,
	}, nil
}

// registerDefaultChecks 注册默认的健康检查
func (hc *HealthChecker) registerDefaultChecks() {
	// HTTP健康检查
	hc.RegisterCheck("http", hc.httpHealthCheck)

	// TCP健康检查
	hc.RegisterCheck("tcp", hc.tcpHealthCheck)

	// 服务响应时间检查
	hc.RegisterCheck("response_time", hc.responseTimeCheck)
}

// httpHealthCheck HTTP健康检查
func (hc *HealthChecker) httpHealthCheck(ctx context.Context, service *registry.ServiceInfo) error {
	if service.Endpoint == "" {
		return fmt.Errorf("service endpoint is empty")
	}

	// 构建健康检查URL
	healthURL := fmt.Sprintf("%s/health", service.Endpoint)

	// 执行HTTP请求
	client := &http.Client{
		Timeout: hc.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// tcpHealthCheck TCP健康检查
func (hc *HealthChecker) tcpHealthCheck(ctx context.Context, service *registry.ServiceInfo) error {
	if service.Endpoint == "" {
		return fmt.Errorf("service endpoint is empty")
	}

	// 解析端点地址
	conn, err := net.DialTimeout("tcp", service.Endpoint, hc.timeout)
	if err != nil {
		return fmt.Errorf("tcp connection failed: %w", err)
	}
	defer conn.Close()

	return nil
}

// responseTimeCheck 响应时间检查
func (hc *HealthChecker) responseTimeCheck(ctx context.Context, service *registry.ServiceInfo) error {
	if service.Endpoint == "" {
		return fmt.Errorf("service endpoint is empty")
	}

	start := time.Now()

	// 执行HTTP请求
	client := &http.Client{
		Timeout: hc.timeout,
	}

	healthURL := fmt.Sprintf("%s/health", service.Endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseTime := time.Since(start)

	// 检查响应时间（默认阈值：5秒）
	maxResponseTime := 5 * time.Second
	if responseTime > maxResponseTime {
		return fmt.Errorf("response time too slow: %v (max: %v)", responseTime, maxResponseTime)
	}

	return nil
}

// GetCheckNames 获取所有注册的检查名称
func (hc *HealthChecker) GetCheckNames() []string {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	names := make([]string, 0, len(hc.checks))
	for name := range hc.checks {
		names = append(names, name)
	}

	return names
}

// RemoveCheck 移除健康检查
func (hc *HealthChecker) RemoveCheck(name string) error {
	if name == "" {
		return errors.NewError(errors.ErrCodeValidation, "check name cannot be empty")
	}

	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	if _, exists := hc.checks[name]; !exists {
		return errors.NewError(errors.ErrCodeNotFound, "check not found")
	}

	delete(hc.checks, name)
	return nil
}

// SetInterval 设置检查间隔
func (hc *HealthChecker) SetInterval(interval time.Duration) error {
	if interval <= 0 {
		return errors.NewError(errors.ErrCodeValidation, "interval must be positive")
	}

	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.interval = interval
	return nil
}

// SetTimeout 设置检查超时
func (hc *HealthChecker) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return errors.NewError(errors.ErrCodeValidation, "timeout must be positive")
	}

	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.timeout = timeout
	return nil
}

// GetConfig 获取健康检查器配置
func (hc *HealthChecker) GetConfig() map[string]interface{} {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	return map[string]interface{}{
		"interval": hc.interval,
		"timeout":  hc.timeout,
		"checks":   hc.GetCheckNames(),
	}
}
