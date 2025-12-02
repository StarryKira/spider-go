package app

import (
	"context"
	"log"
	"spider-go/internal/service"

	"github.com/robfig/cron/v3"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron          *cron.Cron
	taskService   service.TaskService
	rsaKeyService service.RSAKeyService
}

// NewScheduler 创建调度器
func NewScheduler(taskService service.TaskService, rsaKeyService service.RSAKeyService) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		taskService:   taskService,
		rsaKeyService: rsaKeyService,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	// 1. 每天凌晨 2 点执行数据预热
	_, err := s.cron.AddFunc("0 2 * * *", func() {
		log.Println("触发定时任务：数据预热")
		if err := s.taskService.PrewarmUserData(context.Background()); err != nil {
			log.Printf("数据预热任务失败: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("添加定时任务（数据预热）失败: %v", err)
	}

	// 2. 每小时更新一次 RSA 公钥
	_, err = s.cron.AddFunc("0 * * * *", func() {
		log.Println("触发定时任务：更新 RSA 公钥")
		if err := s.rsaKeyService.FetchAndUpdate(); err != nil {
			log.Printf("更新 RSA 公钥失败: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("添加定时任务（RSA 公钥更新）失败: %v", err)
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
