package cache

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionCache 会话缓存接口
type SessionCache interface {
	// GetCookies 获取用户的 cookies
	GetCookies(ctx context.Context, uid int) ([]*http.Cookie, error)
	// SetCookies 设置用户的 cookies
	SetCookies(ctx context.Context, uid int, cookies []*http.Cookie, expiration time.Duration) error
	// DeleteCookies 删除用户的 cookies
	DeleteCookies(ctx context.Context, uid int) error
	// HasCookies 检查用户是否有缓存的 cookies
	HasCookies(ctx context.Context, uid int) (bool, error)
}

// RedisSessionCache Redis 实现的会话缓存
type RedisSessionCache struct {
	client *redis.Client
}

// NewRedisSessionCache 创建 Redis 会话缓存
func NewRedisSessionCache(client *redis.Client) SessionCache {
	return &RedisSessionCache{
		client: client,
	}
}

// GetCookies 获取用户的 cookies
func (c *RedisSessionCache) GetCookies(ctx context.Context, uid int) ([]*http.Cookie, error) {
	key := c.getUserKey(uid)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var cookies []*http.Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, err
	}

	return cookies, nil
}

// SetCookies 设置用户的 cookies
func (c *RedisSessionCache) SetCookies(ctx context.Context, uid int, cookies []*http.Cookie, expiration time.Duration) error {
	key := c.getUserKey(uid)
	data, err := json.Marshal(cookies)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

// DeleteCookies 删除用户的 cookies
func (c *RedisSessionCache) DeleteCookies(ctx context.Context, uid int) error {
	key := c.getUserKey(uid)
	return c.client.Del(ctx, key).Err()
}

// HasCookies 检查用户是否有缓存的 cookies
func (c *RedisSessionCache) HasCookies(ctx context.Context, uid int) (bool, error) {
	key := c.getUserKey(uid)
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// getUserKey 获取用户的 Redis key
func (c *RedisSessionCache) getUserKey(uid int) string {
	return "session:" + strconv.Itoa(uid)
}
