package registry

import (
	"math/rand"
	"sync"
	"time"

	"github.com/jobfirst/jobfirst-core/errors"
)

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	strategy LoadBalanceStrategy
	mutex    sync.RWMutex
}

// LoadBalanceStrategy 负载均衡策略接口
type LoadBalanceStrategy interface {
	Select(services []*ServiceInfo) *ServiceInfo
	Name() string
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{}
	// 默认使用轮询策略
	lb.strategy = NewRoundRobinStrategy()
	return lb
}

// Select 选择服务
func (lb *LoadBalancer) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	lb.mutex.RLock()
	strategy := lb.strategy
	lb.mutex.RUnlock()

	return strategy.Select(services)
}

// SetStrategy 设置负载均衡策略
func (lb *LoadBalancer) SetStrategy(strategy LoadBalanceStrategy) {
	if strategy == nil {
		return
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.strategy = strategy
}

// GetStrategy 获取当前策略名称
func (lb *LoadBalancer) GetStrategy() string {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if lb.strategy == nil {
		return "none"
	}

	return lb.strategy.Name()
}

// RoundRobinStrategy 轮询策略
type RoundRobinStrategy struct {
	index int
	mutex sync.Mutex
}

// NewRoundRobinStrategy 创建轮询策略
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

// Select 轮询选择服务
func (rr *RoundRobinStrategy) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	service := services[rr.index%len(services)]
	rr.index++
	return service
}

// Name 返回策略名称
func (rr *RoundRobinStrategy) Name() string {
	return "round_robin"
}

// RandomStrategy 随机策略
type RandomStrategy struct {
	rand *rand.Rand
}

// NewRandomStrategy 创建随机策略
func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 随机选择服务
func (rs *RandomStrategy) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	index := rs.rand.Intn(len(services))
	return services[index]
}

// Name 返回策略名称
func (rs *RandomStrategy) Name() string {
	return "random"
}

// WeightedRoundRobinStrategy 加权轮询策略
type WeightedRoundRobinStrategy struct {
	weights map[string]int
	current map[string]int
	mutex   sync.Mutex
}

// NewWeightedRoundRobinStrategy 创建加权轮询策略
func NewWeightedRoundRobinStrategy(weights map[string]int) *WeightedRoundRobinStrategy {
	if weights == nil {
		weights = make(map[string]int)
	}

	return &WeightedRoundRobinStrategy{
		weights: weights,
		current: make(map[string]int),
	}
}

// Select 加权轮询选择服务
func (wrr *WeightedRoundRobinStrategy) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	// 如果没有设置权重，使用轮询
	if len(wrr.weights) == 0 {
		// 为所有服务设置默认权重
		for _, service := range services {
			if _, exists := wrr.weights[service.ID]; !exists {
				wrr.weights[service.ID] = 1
			}
		}
	}

	// 找到权重最大的服务
	var selected *ServiceInfo
	maxWeight := -1

	for _, service := range services {
		weight := wrr.weights[service.ID]
		if weight <= 0 {
			weight = 1 // 默认权重
		}

		current := wrr.current[service.ID]
		if current < weight {
			if current > maxWeight {
				maxWeight = current
				selected = service
			}
		}
	}

	if selected != nil {
		wrr.current[selected.ID]++
	} else {
		// 重置所有当前权重
		for id := range wrr.current {
			wrr.current[id] = 0
		}
		// 选择第一个服务
		selected = services[0]
		wrr.current[selected.ID] = 1
	}

	return selected
}

// Name 返回策略名称
func (wrr *WeightedRoundRobinStrategy) Name() string {
	return "weighted_round_robin"
}

// SetWeight 设置服务权重
func (wrr *WeightedRoundRobinStrategy) SetWeight(serviceID string, weight int) error {
	if serviceID == "" {
		return errors.NewError(errors.ErrCodeValidation, "service ID cannot be empty")
	}
	if weight <= 0 {
		return errors.NewError(errors.ErrCodeValidation, "weight must be positive")
	}

	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	wrr.weights[serviceID] = weight
	return nil
}

// GetWeight 获取服务权重
func (wrr *WeightedRoundRobinStrategy) GetWeight(serviceID string) int {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	weight, exists := wrr.weights[serviceID]
	if !exists {
		return 1 // 默认权重
	}

	return weight
}

// LeastConnectionsStrategy 最少连接策略
type LeastConnectionsStrategy struct {
	connections map[string]int
	mutex       sync.Mutex
}

// NewLeastConnectionsStrategy 创建最少连接策略
func NewLeastConnectionsStrategy() *LeastConnectionsStrategy {
	return &LeastConnectionsStrategy{
		connections: make(map[string]int),
	}
}

// Select 选择连接数最少的服务
func (lcs *LeastConnectionsStrategy) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	lcs.mutex.Lock()
	defer lcs.mutex.Unlock()

	var selected *ServiceInfo
	minConnections := int(^uint(0) >> 1) // 最大int值

	for _, service := range services {
		connections := lcs.connections[service.ID]
		if connections < minConnections {
			minConnections = connections
			selected = service
		}
	}

	if selected != nil {
		lcs.connections[selected.ID]++
	}

	return selected
}

// Name 返回策略名称
func (lcs *LeastConnectionsStrategy) Name() string {
	return "least_connections"
}

// ReleaseConnection 释放连接
func (lcs *LeastConnectionsStrategy) ReleaseConnection(serviceID string) {
	lcs.mutex.Lock()
	defer lcs.mutex.Unlock()

	if connections := lcs.connections[serviceID]; connections > 0 {
		lcs.connections[serviceID]--
	}
}

// GetConnections 获取服务连接数
func (lcs *LeastConnectionsStrategy) GetConnections(serviceID string) int {
	lcs.mutex.Lock()
	defer lcs.mutex.Unlock()

	return lcs.connections[serviceID]
}

// IPHashStrategy IP哈希策略
type IPHashStrategy struct {
	hashFunc func(string) uint32
}

// NewIPHashStrategy 创建IP哈希策略
func NewIPHashStrategy() *IPHashStrategy {
	return &IPHashStrategy{
		hashFunc: simpleHash,
	}
}

// Select 根据IP哈希选择服务
func (ihs *IPHashStrategy) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	// 这里简化处理，实际应该从请求上下文中获取客户端IP
	// 为了演示，我们使用服务ID的哈希
	hash := ihs.hashFunc(services[0].ID)
	index := int(hash) % len(services)

	return services[index]
}

// Name 返回策略名称
func (ihs *IPHashStrategy) Name() string {
	return "ip_hash"
}

// simpleHash 简单哈希函数
func simpleHash(s string) uint32 {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// GetAvailableStrategies 获取可用的负载均衡策略
func GetAvailableStrategies() []string {
	return []string{
		"round_robin",
		"random",
		"weighted_round_robin",
		"least_connections",
		"ip_hash",
	}
}

// NewStrategy 根据名称创建策略
func NewStrategy(name string, config map[string]interface{}) (LoadBalanceStrategy, error) {
	switch name {
	case "round_robin":
		return NewRoundRobinStrategy(), nil
	case "random":
		return NewRandomStrategy(), nil
	case "weighted_round_robin":
		weights := make(map[string]int)
		if config != nil {
			if weightsConfig, ok := config["weights"].(map[string]int); ok {
				weights = weightsConfig
			}
		}
		return NewWeightedRoundRobinStrategy(weights), nil
	case "least_connections":
		return NewLeastConnectionsStrategy(), nil
	case "ip_hash":
		return NewIPHashStrategy(), nil
	default:
		return nil, errors.NewError(errors.ErrCodeValidation, "unknown load balance strategy: "+name)
	}
}
