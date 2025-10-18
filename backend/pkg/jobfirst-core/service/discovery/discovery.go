package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/jobfirst/jobfirst-core/errors"
	"github.com/jobfirst/jobfirst-core/service/registry"
)

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	registry *registry.SimpleServiceRegistry
	cache    *DiscoveryCache
	watchers map[string]*ServiceWatcher
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// DiscoveryCache 发现缓存
type DiscoveryCache struct {
	services map[string][]*registry.ServiceInfo
	ttl      time.Duration
	mutex    sync.RWMutex
}

// ServiceWatcher 服务监听器
type ServiceWatcher struct {
	serviceName string
	callback    func([]*registry.ServiceInfo)
	lastUpdate  time.Time
	ctx         context.Context
	cancel      context.CancelFunc
}

// DiscoveryConfig 服务发现配置
type DiscoveryConfig struct {
	CacheTTL      time.Duration `json:"cache_ttl"`
	WatchInterval time.Duration `json:"watch_interval"`
	MaxRetries    int           `json:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval"`
}

// NewServiceDiscovery 创建服务发现
func NewServiceDiscovery(registry *registry.SimpleServiceRegistry, config *DiscoveryConfig) (*ServiceDiscovery, error) {
	if registry == nil {
		return nil, errors.NewError(errors.ErrCodeValidation, "registry cannot be nil")
	}

	if config == nil {
		config = &DiscoveryConfig{
			CacheTTL:      5 * time.Minute,
			WatchInterval: 30 * time.Second,
			MaxRetries:    3,
			RetryInterval: 5 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	discovery := &ServiceDiscovery{
		registry: registry,
		cache:    NewDiscoveryCache(config.CacheTTL),
		watchers: make(map[string]*ServiceWatcher),
		ctx:      ctx,
		cancel:   cancel,
	}

	return discovery, nil
}

// Discover 发现服务
func (sd *ServiceDiscovery) Discover(serviceName string) ([]*registry.ServiceInfo, error) {
	if serviceName == "" {
		return nil, errors.NewError(errors.ErrCodeValidation, "service name cannot be empty")
	}

	// 先从缓存获取
	if services := sd.cache.Get(serviceName); services != nil {
		return services, nil
	}

	// 从注册中心获取
	services, err := sd.registry.GetServicesByName(serviceName)
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "failed to discover services", err)
	}

	// 更新缓存
	sd.cache.Set(serviceName, services)

	return services, nil
}

// DiscoverHealthy 发现健康的服务
func (sd *ServiceDiscovery) DiscoverHealthy(serviceName string) ([]*registry.ServiceInfo, error) {
	services, err := sd.Discover(serviceName)
	if err != nil {
		return nil, err
	}

	// 过滤健康的服务
	healthyServices := make([]*registry.ServiceInfo, 0)
	for _, service := range services {
		if service.Health != nil && service.Health.Status == "healthy" {
			healthyServices = append(healthyServices, service)
		}
	}

	return healthyServices, nil
}

// SelectService 选择服务（负载均衡）
func (sd *ServiceDiscovery) SelectService(serviceName string) (*registry.ServiceInfo, error) {
	return sd.registry.SelectService(serviceName)
}

// Watch 监听服务变化
func (sd *ServiceDiscovery) Watch(serviceName string, callback func([]*registry.ServiceInfo)) error {
	if serviceName == "" {
		return errors.NewError(errors.ErrCodeValidation, "service name cannot be empty")
	}
	if callback == nil {
		return errors.NewError(errors.ErrCodeValidation, "callback cannot be nil")
	}

	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	// 检查是否已经在监听
	if _, exists := sd.watchers[serviceName]; exists {
		return errors.NewError(errors.ErrCodeValidation, "service is already being watched")
	}

	// 创建监听器
	watcherCtx, watcherCancel := context.WithCancel(sd.ctx)
	watcher := &ServiceWatcher{
		serviceName: serviceName,
		callback:    callback,
		lastUpdate:  time.Now(),
		ctx:         watcherCtx,
		cancel:      watcherCancel,
	}

	sd.watchers[serviceName] = watcher

	// 启动监听
	go sd.startWatching(watcher)

	return nil
}

// Unwatch 停止监听服务
func (sd *ServiceDiscovery) Unwatch(serviceName string) error {
	if serviceName == "" {
		return errors.NewError(errors.ErrCodeValidation, "service name cannot be empty")
	}

	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	watcher, exists := sd.watchers[serviceName]
	if !exists {
		return errors.NewError(errors.ErrCodeNotFound, "service is not being watched")
	}

	// 停止监听
	watcher.cancel()
	delete(sd.watchers, serviceName)

	return nil
}

// startWatching 开始监听服务
func (sd *ServiceDiscovery) startWatching(watcher *ServiceWatcher) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-watcher.ctx.Done():
			return
		case <-ticker.C:
			services, err := sd.Discover(watcher.serviceName)
			if err == nil {
				// 检查是否有变化
				if sd.hasServiceChanged(watcher, services) {
					watcher.callback(services)
					watcher.lastUpdate = time.Now()
				}
			}
		}
	}
}

// hasServiceChanged 检查服务是否有变化
func (sd *ServiceDiscovery) hasServiceChanged(_ *ServiceWatcher, _ []*registry.ServiceInfo) bool {
	// 简化实现，实际应该比较服务列表的差异
	// 这里假设每次都有变化
	return true
}

// GetCacheStatus 获取缓存状态
func (sd *ServiceDiscovery) GetCacheStatus() map[string]interface{} {
	return sd.cache.GetStatus()
}

// ClearCache 清空缓存
func (sd *ServiceDiscovery) ClearCache() {
	sd.cache.Clear()
}

// Close 关闭服务发现
func (sd *ServiceDiscovery) Close() error {
	sd.cancel()
	return nil
}

// NewDiscoveryCache 创建发现缓存
func NewDiscoveryCache(ttl time.Duration) *DiscoveryCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	return &DiscoveryCache{
		services: make(map[string][]*registry.ServiceInfo),
		ttl:      ttl,
	}
}

// Get 从缓存获取服务
func (dc *DiscoveryCache) Get(serviceName string) []*registry.ServiceInfo {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	services, exists := dc.services[serviceName]
	if !exists {
		return nil
	}

	// 检查TTL（简化实现）
	return services
}

// Set 设置缓存
func (dc *DiscoveryCache) Set(serviceName string, services []*registry.ServiceInfo) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.services[serviceName] = services
}

// Clear 清空缓存
func (dc *DiscoveryCache) Clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.services = make(map[string][]*registry.ServiceInfo)
}

// GetStatus 获取缓存状态
func (dc *DiscoveryCache) GetStatus() map[string]interface{} {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	status := map[string]interface{}{
		"ttl":      dc.ttl,
		"services": len(dc.services),
	}

	// 统计缓存的服务
	serviceStats := make(map[string]int)
	for name, services := range dc.services {
		serviceStats[name] = len(services)
	}
	status["service_stats"] = serviceStats

	return status
}

// GetWatchers 获取监听器列表
func (sd *ServiceDiscovery) GetWatchers() []string {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	watchers := make([]string, 0, len(sd.watchers))
	for serviceName := range sd.watchers {
		watchers = append(watchers, serviceName)
	}

	return watchers
}

// GetDiscoveryStatus 获取服务发现状态
func (sd *ServiceDiscovery) GetDiscoveryStatus() map[string]interface{} {
	status := map[string]interface{}{
		"cache":    sd.GetCacheStatus(),
		"watchers": sd.GetWatchers(),
	}

	return status
}
