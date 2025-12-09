package statistics

import (
	"context"
	"spider-go/internal/common"
	"spider-go/internal/service"
	"time"
)

// Service 统计服务接口
type Service interface {
	// GetTodayDAU 获取今日DAU
	GetTodayDAU(ctx context.Context) (int64, error)
	// GetDAUByDate 获取指定日期的DAU
	GetDAUByDate(ctx context.Context, date time.Time) (int64, error)
	// GetDAURange 获取指定日期范围的DAU
	GetDAURange(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error)
}

type statisticsService struct {
	dauService service.DAUService
}

// NewService 创建统计服务
func NewService(dauService service.DAUService) Service {
	return &statisticsService{
		dauService: dauService,
	}
}

// GetTodayDAU 获取今日DAU
func (s *statisticsService) GetTodayDAU(ctx context.Context) (int64, error) {
	count, err := s.dauService.GetTodayDAU(ctx)
	if err != nil {
		return 0, common.NewAppError(common.CodeInternalError, "获取今日DAU失败")
	}
	return count, nil
}

// GetDAUByDate 获取指定日期的DAU
func (s *statisticsService) GetDAUByDate(ctx context.Context, date time.Time) (int64, error) {
	count, err := s.dauService.GetDAUByDate(ctx, date)
	if err != nil {
		return 0, common.NewAppError(common.CodeInternalError, "获取DAU失败")
	}
	return count, nil
}

// GetDAURange 获取指定日期范围的DAU
func (s *statisticsService) GetDAURange(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error) {
	// 验证日期范围（最多查询31天）
	if endDate.Sub(startDate) > 31*24*time.Hour {
		return nil, common.NewAppError(common.CodeInvalidParams, "日期范围不能超过31天")
	}

	// 验证开始日期不能晚于结束日期
	if startDate.After(endDate) {
		return nil, common.NewAppError(common.CodeInvalidParams, "开始日期不能晚于结束日期")
	}

	data, err := s.dauService.GetDAURange(ctx, startDate, endDate)
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, "获取DAU范围失败")
	}
	return data, nil
}
