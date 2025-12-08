package service

import (
	"context"
	"log"
	"spider-go/internal/cache"
	"time"
)

// TaskService 定时任务服务
type TaskService interface {
	// PrewarmUserData 预热活跃用户数据
	PrewarmUserData(ctx context.Context) error
}

// taskServiceImpl 定时任务服务实现
type taskServiceImpl struct {
	dauCache      cache.DAUCache
	sessionSvc    SessionService
	userDataCache cache.UserDataCache
	configCache   cache.ConfigCache
	// TODO: 重构这些服务为接口
	// courseSvc, gradeSvc, examSvc 现在在模块中，需要重新设计
}

// NewTaskService 创建定时任务服务
func NewTaskService(
	dauCache cache.DAUCache,
	sessionSvc SessionService,
	userDataCache cache.UserDataCache,
	configCache cache.ConfigCache,
) TaskService {
	return &taskServiceImpl{
		dauCache:      dauCache,
		sessionSvc:    sessionSvc,
		userDataCache: userDataCache,
		configCache:   configCache,
	}
}

// PrewarmUserData 预热活跃用户数据
// TODO: 需要重构，因为course/grade/exam服务现在在模块中
func (s *taskServiceImpl) PrewarmUserData(ctx context.Context) error {
	log.Println("数据预热任务暂时禁用，需要重构以适应新的模块架构")
	return nil
}

// prewarmSingleUser 预热单个用户数据
// TODO: 需要重构
func (s *taskServiceImpl) prewarmSingleUser(ctx context.Context, uidStr string, currentTerm string) {
	// 暂时禁用
}

// getRecentActiveUsers 获取近 N 天活跃用户
func (s *taskServiceImpl) getRecentActiveUsers(ctx context.Context, days int) []string {
	userSet := make(map[string]struct{})
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		users, err := s.dauCache.GetActiveUsers(ctx, date)
		if err != nil {
			continue
		}
		for _, uid := range users {
			userSet[uid] = struct{}{}
		}
	}

	result := make([]string, 0, len(userSet))
	for uid := range userSet {
		result = append(result, uid)
	}
	return result
}
