package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheManager 缓存管理器
type CacheManager struct {
	config *CacheConfig
	client redis.Client
	memory map[string]*CacheItem
	mutex  sync.RWMutex
	stats  *CacheStats
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(config *CacheConfig) (*CacheManager, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cm := &CacheManager{
		config: config,
		memory: make(map[string]*CacheItem),
		stats:  &CacheStats{},
	}

	// 根据缓存类型初始化
	switch config.Type {
	case CacheTypeRedis:
		return cm.initRedis()
	case CacheTypeMemory:
		return cm.initMemory()
	case CacheTypeLocal:
		return cm.initLocal()
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", config.Type)
	}
}

// initRedis 初始化Redis缓存
func (c *CacheManager) initRedis() (*CacheManager, error) {
	c.client = *redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Password: c.config.Password,
		DB:       c.config.DB,
		PoolSize: c.config.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return c, nil
}

// initMemory 初始化内存缓存
func (c *CacheManager) initMemory() (*CacheManager, error) {
	// 启动清理过期键的goroutine
	go c.cleanupExpiredKeys()
	return c, nil
}

// initLocal 初始化本地缓存
func (c *CacheManager) initLocal() (*CacheManager, error) {
	// 启动清理过期键的goroutine
	go c.cleanupExpiredKeys()
	return c, nil
}

// Set 设置缓存
func (c *CacheManager) Set(ctx context.Context, key string, value interface{}, options *CacheOptions) error {
	if options == nil {
		options = &CacheOptions{TTL: c.config.TTL}
	}

	switch c.config.Type {
	case CacheTypeRedis:
		return c.setRedis(ctx, key, value, options)
	case CacheTypeMemory, CacheTypeLocal:
		return c.setMemory(key, value, options)
	default:
		return fmt.Errorf("unsupported cache type: %s", c.config.Type)
	}
}

// setRedis Redis设置缓存
func (c *CacheManager) setRedis(ctx context.Context, key string, value interface{}, options *CacheOptions) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	if options.NX {
		cmd := c.client.SetNX(ctx, key, data, options.TTL)
		return cmd.Err()
	} else if options.XX {
		cmd := c.client.SetXX(ctx, key, data, options.TTL)
		return cmd.Err()
	} else {
		cmd := c.client.Set(ctx, key, data, options.TTL)
		return cmd.Err()
	}
}

// setMemory 内存设置缓存
func (c *CacheManager) setMemory(key string, value interface{}, options *CacheOptions) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item := &CacheItem{
		Key:        key,
		Value:      value,
		Expiration: time.Now().Add(options.TTL),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	c.memory[key] = item
	return nil
}

// Get 获取缓存
func (c *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	switch c.config.Type {
	case CacheTypeRedis:
		return c.getRedis(ctx, key, dest)
	case CacheTypeMemory, CacheTypeLocal:
		return c.getMemory(key, dest)
	default:
		return fmt.Errorf("unsupported cache type: %s", c.config.Type)
	}
}

// getRedis Redis获取缓存
func (c *CacheManager) getRedis(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			c.stats.Misses++
			return fmt.Errorf("key not found: %s", key)
		}
		return err
	}

	c.stats.Hits++
	return json.Unmarshal(data, dest)
}

// getMemory 内存获取缓存
func (c *CacheManager) getMemory(key string, dest interface{}) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.memory[key]
	if !exists {
		c.stats.Misses++
		return fmt.Errorf("key not found: %s", key)
	}

	// 检查是否过期
	if time.Now().After(item.Expiration) {
		c.stats.Misses++
		c.stats.Expired++
		return fmt.Errorf("key expired: %s", key)
	}

	c.stats.Hits++

	// 复制值到目标
	switch v := item.Value.(type) {
	case []byte:
		return json.Unmarshal(v, dest)
	default:
		// 通过JSON序列化/反序列化来复制
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, dest)
	}
}

// Delete 删除缓存
func (c *CacheManager) Delete(ctx context.Context, key string) error {
	switch c.config.Type {
	case CacheTypeRedis:
		return c.deleteRedis(ctx, key)
	case CacheTypeMemory, CacheTypeLocal:
		return c.deleteMemory(key)
	default:
		return fmt.Errorf("unsupported cache type: %s", c.config.Type)
	}
}

// deleteRedis Redis删除缓存
func (c *CacheManager) deleteRedis(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// deleteMemory 内存删除缓存
func (c *CacheManager) deleteMemory(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.memory, key)
	return nil
}

// Exists 检查键是否存在
func (c *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	switch c.config.Type {
	case CacheTypeRedis:
		return c.existsRedis(ctx, key)
	case CacheTypeMemory, CacheTypeLocal:
		return c.existsMemory(key), nil
	default:
		return false, fmt.Errorf("unsupported cache type: %s", c.config.Type)
	}
}

// existsRedis Redis检查键是否存在
func (c *CacheManager) existsRedis(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}

// existsMemory 内存检查键是否存在
func (c *CacheManager) existsMemory(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.memory[key]
	if !exists {
		return false
	}

	// 检查是否过期
	if time.Now().After(item.Expiration) {
		return false
	}

	return true
}

// TTL 获取键的剩余生存时间
func (c *CacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	switch c.config.Type {
	case CacheTypeRedis:
		return c.ttlRedis(ctx, key)
	case CacheTypeMemory, CacheTypeLocal:
		return c.ttlMemory(key), nil
	default:
		return 0, fmt.Errorf("unsupported cache type: %s", c.config.Type)
	}
}

// ttlRedis Redis获取TTL
func (c *CacheManager) ttlRedis(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// ttlMemory 内存获取TTL
func (c *CacheManager) ttlMemory(key string) time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.memory[key]
	if !exists {
		return -1
	}

	ttl := time.Until(item.Expiration)
	if ttl <= 0 {
		return -2 // 已过期
	}

	return ttl
}

// GetStats 获取缓存统计
func (c *CacheManager) GetStats() *CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := *c.stats
	stats.Keys = int64(len(c.memory))

	return &stats
}

// cleanupExpiredKeys 清理过期键
func (c *CacheManager) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()

		for key, item := range c.memory {
			if now.After(item.Expiration) {
				delete(c.memory, key)
				c.stats.Expired++
			}
		}
		c.mutex.Unlock()
	}
}

// Close 关闭缓存连接
func (c *CacheManager) Close() error {
	if c.config.Type == CacheTypeRedis {
		return c.client.Close()
	}
	return nil
}
