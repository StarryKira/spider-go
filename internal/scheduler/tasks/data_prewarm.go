package tasks

import (
	"context"
	"log"
)

// DataPrewarmTask 数据预热任务
type DataPrewarmTask struct {
	// TODO: 需要在模块层提供数据预热接口后再实现
}

// NewDataPrewarmTask 创建数据预热任务
func NewDataPrewarmTask() *DataPrewarmTask {
	return &DataPrewarmTask{}
}

// Name 任务名称
func (t *DataPrewarmTask) Name() string {
	return "数据预热"
}

// Cron Cron 表达式（每天凌晨2点执行）
func (t *DataPrewarmTask) Cron() string {
	return "0 2 * * *"
}

// Run 执行任务
func (t *DataPrewarmTask) Run(ctx context.Context) error {
	// TODO: 实现数据预热逻辑
	// 需要各模块提供预热接口后才能实现
	log.Println("数据预热任务暂时禁用，等待模块层提供预热接口")
	return nil
}
