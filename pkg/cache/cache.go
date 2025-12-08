package cache

import (
	"context"
	"time"
)

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// Options 缓存配置选项
type Options struct {
	Prefix string
	TTL    time.Duration
}

// Option 配置函数
type Option func(*Options)

// WithPrefix 设置缓存键前缀
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

// WithTTL 设置默认过期时间
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}
