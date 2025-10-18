package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientWrapper Redis客户端包装器
type RedisClientWrapper struct {
	client *redis.Client
}

// Get 获取值
func (w *RedisClientWrapper) Get(ctx context.Context, key string) RedisResult {
	return &RedisResultWrapper{result: w.client.Get(ctx, key)}
}

// Set 设置值
func (w *RedisClientWrapper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) RedisResult {
	return &RedisResultWrapper{result: w.client.Set(ctx, key, value, expiration)}
}

// Del 删除键
func (w *RedisClientWrapper) Del(ctx context.Context, keys ...string) RedisResult {
	return &RedisResultWrapper{result: w.client.Del(ctx, keys...)}
}

// Exists 检查键是否存在
func (w *RedisClientWrapper) Exists(ctx context.Context, keys ...string) RedisResult {
	return &RedisResultWrapper{result: w.client.Exists(ctx, keys...)}
}

// Expire 设置过期时间
func (w *RedisClientWrapper) Expire(ctx context.Context, key string, expiration time.Duration) RedisResult {
	return &RedisResultWrapper{result: w.client.Expire(ctx, key, expiration)}
}

// FlushDB 清空数据库
func (w *RedisClientWrapper) FlushDB(ctx context.Context) RedisResult {
	return &RedisResultWrapper{result: w.client.FlushDB(ctx)}
}

// RedisResultWrapper Redis结果包装器
type RedisResultWrapper struct {
	result interface{}
}

// Result 获取结果
func (w *RedisResultWrapper) Result() (interface{}, error) {
	switch cmd := w.result.(type) {
	case *redis.StringCmd:
		return cmd.Result()
	case *redis.StatusCmd:
		return cmd.Result()
	case *redis.IntCmd:
		return cmd.Result()
	case *redis.BoolCmd:
		return cmd.Result()
	default:
		return nil, nil
	}
}

// Err 获取错误
func (w *RedisResultWrapper) Err() error {
	switch cmd := w.result.(type) {
	case *redis.StringCmd:
		return cmd.Err()
	case *redis.StatusCmd:
		return cmd.Err()
	case *redis.IntCmd:
		return cmd.Err()
	case *redis.BoolCmd:
		return cmd.Err()
	default:
		return nil
	}
}

// Bytes 获取字节数组
func (w *RedisResultWrapper) Bytes() ([]byte, error) {
	if cmd, ok := w.result.(*redis.StringCmd); ok {
		return cmd.Bytes()
	}
	return nil, nil
}

// Int64 获取64位整数
func (w *RedisResultWrapper) Int64() (int64, error) {
	if cmd, ok := w.result.(*redis.IntCmd); ok {
		return cmd.Result()
	}
	return 0, nil
}

// Bool 获取布尔值
func (w *RedisResultWrapper) Bool() (bool, error) {
	if cmd, ok := w.result.(*redis.BoolCmd); ok {
		return cmd.Result()
	}
	return false, nil
}
