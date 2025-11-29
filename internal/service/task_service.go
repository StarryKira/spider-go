package service

import (
	"context"
	"log"
	"spider-go/internal/cache"
	"spider-go/internal/repository"
	"strconv"
	"time"
)

// TaskService 定时任务服务
type TaskService interface {
	// PrewarmUserData 预热活跃用户数据
	PrewarmUserData(ctx context.Context) error
}

// taskServiceImpl 定时任务服务实现
type taskServiceImpl struct {
	userRepo      repository.UserRepository
	dauCache      cache.DAUCache
	sessionSvc    SessionService
	courseSvc     CourseService
	gradeSvc      GradeService
	examSvc       ExamService
	userDataCache cache.UserDataCache
	configCache   cache.ConfigCache
}

// NewTaskService 创建定时任务服务
func NewTaskService(
	userRepo repository.UserRepository,
	dauCache cache.DAUCache,
	sessionSvc SessionService,
	courseSvc CourseService,
	gradeSvc GradeService,
	examSvc ExamService,
	userDataCache cache.UserDataCache,
	configCache cache.ConfigCache,
) TaskService {
	return &taskServiceImpl{
		userRepo:      userRepo,
		dauCache:      dauCache,
		sessionSvc:    sessionSvc,
		courseSvc:     courseSvc,
		gradeSvc:      gradeSvc,
		examSvc:       examSvc,
		userDataCache: userDataCache,
		configCache:   configCache,
	}
}

// PrewarmUserData 预热活跃用户数据
// 凌晨执行，获取近3天活跃用户，并缓存其数据
func (s *taskServiceImpl) PrewarmUserData(ctx context.Context) error {
	log.Println("开始执行数据预热任务...")

	// 1. 获取近3天活跃用户ID
	activeUserIDs := s.getRecentActiveUsers(ctx, 3)
	log.Printf("发现 %d 个活跃用户需要预热", len(activeUserIDs))

	// 2. 获取当前学期
	currentTerm, err := s.configCache.GetCurrentTerm(ctx)
	if err != nil {
		return err
	}

	// 3. 并发预热数据
	// 限制并发数，避免对教务系统造成过大压力
	sem := make(chan struct{}, 5) // 并发数 5

	for _, uidStr := range activeUserIDs {
		sem <- struct{}{} // 获取信号量
		go func(uidStr string) {
			defer func() { <-sem }() // 释放信号量
			s.prewarmSingleUser(ctx, uidStr, currentTerm)
		}(uidStr)
	}

	log.Println("数据预热任务已提交")
	return nil
}

// prewarmSingleUser 预热单个用户数据
func (s *taskServiceImpl) prewarmSingleUser(ctx context.Context, uidStr string, currentTerm string) {
	// 转换 UID
	uid, _ := strconv.Atoi(uidStr)
	// 获取用户信息
	user, err := s.userRepo.GetUserByUid(uid)
	if err != nil || user.Sid == "" || user.Spwd == "" {
		return
	}

	// 登录并获取会话
	if err := s.sessionSvc.LoginAndCache(ctx, uid, user.Sid, user.Spwd); err != nil {
		log.Printf("用户 %d 登录失败: %v", uid, err)
		return
	}

	// 1. 预热课表（当前周）
	// 这里假设预热第1-20周的数据可能太多，只预热当前周
	// 实际项目中可以配合 ConfigCache 获取当前周次
	currentWeek := 1 // 假设当前是第1周
	if _, err := s.courseSvc.GetCourseTableByWeek(ctx, currentWeek, currentTerm, uid); err != nil {
		log.Printf("用户 %d 课表预热失败: %v", uid, err)
	}

	// 2. 预热成绩（当前学期）
	if _, _, err := s.gradeSvc.GetGradeByTerm(ctx, uid, currentTerm); err != nil {
		log.Printf("用户 %d 成绩预热失败: %v", uid, err)
	}

	// 3. 预热考试安排
	if _, err := s.examSvc.GetAllExams(ctx, uid, currentTerm); err != nil {
		log.Printf("用户 %d 考试安排预热失败: %v", uid, err)
	}
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
