package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// UserDataCache 用户数据缓存接口
type UserDataCache interface {
	// CacheGrades 缓存成绩数据
	CacheGrades(ctx context.Context, uid int, term string, data interface{}, expiration time.Duration) error
	// GetGrades 获取成绩缓存
	GetGrades(ctx context.Context, uid int, term string, target interface{}) error

	// CacheCourseTable 缓存课表数据
	CacheCourseTable(ctx context.Context, uid int, term string, week int, data interface{}, expiration time.Duration) error
	// GetCourseTable 获取课表缓存
	GetCourseTable(ctx context.Context, uid int, term string, week int, target interface{}) error

	// CacheExams 缓存考试安排
	CacheExams(ctx context.Context, uid int, term string, data interface{}, expiration time.Duration) error
	// GetExams 获取考试安排缓存
	GetExams(ctx context.Context, uid int, term string, target interface{}) error
}

// RedisUserDataCache Redis 实现的用户数据缓存
type RedisUserDataCache struct {
	client *redis.Client
}

// NewRedisUserDataCache 创建 Redis 用户数据缓存
func NewRedisUserDataCache(client *redis.Client) UserDataCache {
	return &RedisUserDataCache{
		client: client,
	}
}

// CacheGrades 缓存成绩数据
func (c *RedisUserDataCache) CacheGrades(ctx context.Context, uid int, term string, data interface{}, expiration time.Duration) error {
	key := c.getGradesKey(uid, term)
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, bytes, expiration).Err()
}

// GetGrades 获取成绩缓存
func (c *RedisUserDataCache) GetGrades(ctx context.Context, uid int, term string, target interface{}) error {
	key := c.getGradesKey(uid, term)
	bytes, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}

// CacheCourseTable 缓存课表数据
func (c *RedisUserDataCache) CacheCourseTable(ctx context.Context, uid int, term string, week int, data interface{}, expiration time.Duration) error {
	key := c.getCourseKey(uid, term, week)
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, bytes, expiration).Err()
}

// GetCourseTable 获取课表缓存
func (c *RedisUserDataCache) GetCourseTable(ctx context.Context, uid int, term string, week int, target interface{}) error {
	key := c.getCourseKey(uid, term, week)
	bytes, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}

// CacheExams 缓存考试安排
func (c *RedisUserDataCache) CacheExams(ctx context.Context, uid int, term string, data interface{}, expiration time.Duration) error {
	key := c.getExamKey(uid, term)
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, bytes, expiration).Err()
}

// GetExams 获取考试安排缓存
func (c *RedisUserDataCache) GetExams(ctx context.Context, uid int, term string, target interface{}) error {
	key := c.getExamKey(uid, term)
	bytes, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}

// 键生成辅助方法
func (c *RedisUserDataCache) getGradesKey(uid int, term string) string {
	if term == "" {
		return fmt.Sprintf("data:grades:%d:all", uid)
	}
	return fmt.Sprintf("data:grades:%d:%s", uid, term)
}

func (c *RedisUserDataCache) getCourseKey(uid int, term string, week int) string {
	return fmt.Sprintf("data:course:%d:%s:%d", uid, term, week)
}

func (c *RedisUserDataCache) getExamKey(uid int, term string) string {
	return fmt.Sprintf("data:exam:%d:%s", uid, term)
}
