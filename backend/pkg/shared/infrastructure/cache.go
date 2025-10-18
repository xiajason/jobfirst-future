package infrastructure

import (
	"context"
	"encoding/json"
	"time"
)

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Clear(ctx context.Context) error
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client RedisClient
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client RedisClient) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

// Get 获取缓存
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}

// Set 设置缓存
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists 检查缓存是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Int64()
	return result > 0, err
}

// Expire 设置过期时间
func (c *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// Clear 清空缓存
func (c *RedisCache) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// RedisClient Redis客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) RedisResult
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) RedisResult
	Del(ctx context.Context, keys ...string) RedisResult
	Exists(ctx context.Context, keys ...string) RedisResult
	Expire(ctx context.Context, key string, expiration time.Duration) RedisResult
	FlushDB(ctx context.Context) RedisResult
}

// RedisResult Redis结果接口
type RedisResult interface {
	Result() (interface{}, error)
	Err() error
	Bytes() ([]byte, error)
	Int64() (int64, error)
	Bool() (bool, error)
}

// CacheService 缓存服务
type CacheService struct {
	cache Cache
}

// NewCacheService 创建缓存服务
func NewCacheService(cache Cache) *CacheService {
	return &CacheService{
		cache: cache,
	}
}

// GetObject 获取对象
func (s *CacheService) GetObject(ctx context.Context, key string, dest interface{}) error {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetObject 设置对象
func (s *CacheService) SetObject(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, data, expiration)
}

// GetString 获取字符串
func (s *CacheService) GetString(ctx context.Context, key string) (string, error) {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetString 设置字符串
func (s *CacheService) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	return s.cache.Set(ctx, key, []byte(value), expiration)
}
