package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisManager Redis数据库管理器
type RedisManager struct {
	client *redis.Client
	config RedisConfig
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
	PoolSize int    `json:"pool_size"`
	MinIdle  int    `json:"min_idle"`
}

// NewRedisManager 创建Redis管理器
func NewRedisManager(config RedisConfig) (*RedisManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdle,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	return &RedisManager{
		client: rdb,
		config: config,
	}, nil
}

// GetClient 获取Redis客户端
func (rm *RedisManager) GetClient() *redis.Client {
	return rm.client
}

// Set 设置键值
func (rm *RedisManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rm.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (rm *RedisManager) Get(ctx context.Context, key string) (string, error) {
	return rm.client.Get(ctx, key).Result()
}

// Del 删除键
func (rm *RedisManager) Del(ctx context.Context, keys ...string) error {
	return rm.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (rm *RedisManager) Exists(ctx context.Context, keys ...string) (int64, error) {
	return rm.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (rm *RedisManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rm.client.Expire(ctx, key, expiration).Err()
}

// HSet 设置哈希字段
func (rm *RedisManager) HSet(ctx context.Context, key string, values ...interface{}) error {
	return rm.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (rm *RedisManager) HGet(ctx context.Context, key, field string) (string, error) {
	return rm.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (rm *RedisManager) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return rm.client.HGetAll(ctx, key).Result()
}

// LPush 列表左推入
func (rm *RedisManager) LPush(ctx context.Context, key string, values ...interface{}) error {
	return rm.client.LPush(ctx, key, values...).Err()
}

// RPush 列表右推入
func (rm *RedisManager) RPush(ctx context.Context, key string, values ...interface{}) error {
	return rm.client.RPush(ctx, key, values...).Err()
}

// LPop 列表左弹出
func (rm *RedisManager) LPop(ctx context.Context, key string) (string, error) {
	return rm.client.LPop(ctx, key).Result()
}

// RPop 列表右弹出
func (rm *RedisManager) RPop(ctx context.Context, key string) (string, error) {
	return rm.client.RPop(ctx, key).Result()
}

// SAdd 集合添加
func (rm *RedisManager) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return rm.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (rm *RedisManager) SMembers(ctx context.Context, key string) ([]string, error) {
	return rm.client.SMembers(ctx, key).Result()
}

// ZAdd 有序集合添加
func (rm *RedisManager) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return rm.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围
func (rm *RedisManager) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rm.client.ZRange(ctx, key, start, stop).Result()
}

// Pipeline 创建管道
func (rm *RedisManager) Pipeline() redis.Pipeliner {
	return rm.client.Pipeline()
}

// Close 关闭连接
func (rm *RedisManager) Close() error {
	return rm.client.Close()
}

// Ping 测试连接
func (rm *RedisManager) Ping(ctx context.Context) error {
	return rm.client.Ping(ctx).Err()
}

// Health 健康检查
func (rm *RedisManager) Health() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stats := rm.client.PoolStats()

	// 测试连接
	pingErr := rm.Ping(ctx)
	status := "healthy"
	if pingErr != nil {
		status = "unhealthy"
	}

	return map[string]interface{}{
		"status":      status,
		"host":        rm.config.Host,
		"port":        rm.config.Port,
		"database":    rm.config.Database,
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
		"error":       pingErr,
	}
}
