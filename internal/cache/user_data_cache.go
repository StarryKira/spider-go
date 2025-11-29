package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type UserDataCache interface {
	// GetCourseTable 从缓存读取课程表
	GetCourseTable(ctx context.Context, uid int)
	// SetCourseTable 缓存课程表到缓存
	SetCourseTable(ctx context.Context, uid int)
	// HasCourseTable 检查是否有用户缓存的课程表
	HasCourseTable(ctx context.Context, uid int) bool
	// GetExamArrangement 从缓存读取考试安排
	GetExamArrangement(ctx context.Context, uid int)
	// SetExamArrangement 缓存考试安排到缓存
	SetExamArrangement(ctx context.Context, uid int)
	// HasExamArrangement 检查是否有用户缓存的考试安排
	HasExamArrangement(ctx context.Context, uid int) bool
	// GetGradeTable 从缓存中读取成绩单
	GetGradeTable(ctx context.Context, uid int)
	// SetGradeTable 缓存成绩单到缓存
	SetGradeTable(ctx context.Context, uid int)
	// HasGradeTable 检查是否有用户缓存的成绩单
	HasGradeTable(ctx context.Context, uid int) bool
}

type RedisUserdataCache struct {
	client *redis.Client
}

func NewRedisUserDataCache(client *redis.Client) UserDataCache {
	return &RedisUserdataCache{
		client: client,
	}
}
