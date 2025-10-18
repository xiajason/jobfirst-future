package cache

import (
	"time"
)

// CacheType 缓存类型
type CacheType string

const (
	CacheTypeRedis  CacheType = "redis"  // Redis缓存
	CacheTypeMemory CacheType = "memory" // 内存缓存
	CacheTypeLocal  CacheType = "local"  // 本地缓存
)

// CacheConfig 缓存配置
type CacheConfig struct {
	Type     CacheType     `json:"type"`      // 缓存类型
	Host     string        `json:"host"`      // 主机地址
	Port     int           `json:"port"`      // 端口
	Password string        `json:"password"`  // 密码
	DB       int           `json:"db"`        // 数据库编号
	PoolSize int           `json:"pool_size"` // 连接池大小
	TTL      time.Duration `json:"ttl"`       // 默认过期时间
}

// CacheItem 缓存项
type CacheItem struct {
	Key        string      `json:"key"`        // 缓存键
	Value      interface{} `json:"value"`      // 缓存值
	Expiration time.Time   `json:"expiration"` // 过期时间
	CreatedAt  time.Time   `json:"created_at"` // 创建时间
	UpdatedAt  time.Time   `json:"updated_at"` // 更新时间
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits    int64 `json:"hits"`    // 命中次数
	Misses  int64 `json:"misses"`  // 未命中次数
	Keys    int64 `json:"keys"`    // 键数量
	Memory  int64 `json:"memory"`  // 内存使用量
	Expired int64 `json:"expired"` // 过期键数量
	Evicted int64 `json:"evicted"` // 驱逐键数量
}

// CacheOptions 缓存选项
type CacheOptions struct {
	TTL       time.Duration `json:"ttl"`       // 过期时间
	NX        bool          `json:"nx"`        // 仅当键不存在时设置
	XX        bool          `json:"xx"`        // 仅当键存在时设置
	KeepTTL   bool          `json:"keep_ttl"`  // 保持原有TTL
	Condition string        `json:"condition"` // 条件
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Type:     CacheTypeRedis,
		Host:     "localhost",
		Port:     8201,
		Password: "",
		DB:       0,
		PoolSize: 10,
		TTL:      24 * time.Hour,
	}
}
