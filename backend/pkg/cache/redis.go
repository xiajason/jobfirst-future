package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/config"
)

func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Host + ":" + cfg.Port,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  parseDuration(cfg.DialTimeout),
		ReadTimeout:  parseDuration(cfg.ReadTimeout),
		WriteTimeout: parseDuration(cfg.WriteTimeout),
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Printf("Connected to Redis: %s:%s", cfg.Host, cfg.Port)
	return client, nil
}

func CloseRedis(client *redis.Client) {
	if client != nil {
		if err := client.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Println("Redis connection closed")
		}
	}
}

func parseDuration(duration string) time.Duration {
	if parsed, err := time.ParseDuration(duration); err == nil {
		return parsed
	}
	return 5 * time.Second // 默认5秒
}
