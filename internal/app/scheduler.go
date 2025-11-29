package app

import (
	"context"
	"log"
	"spider-go/internal/service"

	"github.com/robfig/cron/v3"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron        *cron.Cron
	taskService service.TaskService
}

// NewScheduler 创建调度器
func NewScheduler(taskService service.TaskService) *Scheduler {
	return &Scheduler{
		cron:        cron.New(),
		taskService: taskService,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	// 每天凌晨 2 点执行数据预热
	_, err := s.cron.AddFunc("0 2 * * *", func() {
		log.Println("触发定时任务：数据预热")
		if err := s.taskService.PrewarmUserData(context.Background()); err != nil {
			log.Printf("数据预热任务失败: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("添加定时任务失败: %v", err)
	}

	s.cron.Start()
	log.Println("定时任务调度器已启动")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		log.Println("定时任务调度器已停止")
	}
}
