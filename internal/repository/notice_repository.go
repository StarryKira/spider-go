package repository

import (
	"spider-go/internal/model"

	"gorm.io/gorm"
)

// NoticeRepository 通知仓储接口
type NoticeRepository interface {
	// CreateNotice 创建通知
	CreateNotice(notice *model.Notice) error
	// UpdateNotice 更新通知
	UpdateNotice(notice *model.Notice) error
	// DeleteNotice 删除通知
	DeleteNotice(nid int) error
	// GetNoticeByNid 根据 NID 获取通知
	GetNoticeByNid(nid int) (*model.Notice, error)
	// GetAllNotices 获取所有通知（管理员）
	GetAllNotices() ([]model.Notice, error)
	// GetVisibleNotices 获取可见通知（普通用户）
	GetVisibleNotices() ([]model.Notice, error)
}

// gormNoticeRepository GORM 实现的通知仓储
type gormNoticeRepository struct {
	db *gorm.DB
}

// NewGormNoticeRepository 创建 GORM 通知仓储
func NewGormNoticeRepository(db *gorm.DB) NoticeRepository {
	return &gormNoticeRepository{db: db}
}

// CreateNotice 创建通知
func (r *gormNoticeRepository) CreateNotice(notice *model.Notice) error {
	return r.db.Create(notice).Error
}

// UpdateNotice 更新通知
func (r *gormNoticeRepository) UpdateNotice(notice *model.Notice) error {
	return r.db.Save(notice).Error
}

// DeleteNotice 删除通知
func (r *gormNoticeRepository) DeleteNotice(nid int) error {
	return r.db.Delete(&model.Notice{}, nid).Error
}

// GetNoticeByNid 根据 NID 获取通知
func (r *gormNoticeRepository) GetNoticeByNid(nid int) (*model.Notice, error) {
	var notice model.Notice
	err := r.db.First(&notice, nid).Error
	return &notice, err
}

// GetAllNotices 获取所有通知（管理员）
func (r *gormNoticeRepository) GetAllNotices() ([]model.Notice, error) {
	var notices []model.Notice
	err := r.db.Order("is_top DESC, create_time DESC").Find(&notices).Error
	return notices, err
}

// GetVisibleNotices 获取可见通知（普通用户）
func (r *gormNoticeRepository) GetVisibleNotices() ([]model.Notice, error) {
	var notices []model.Notice
	err := r.db.Where("is_show = ?", true).
		Order("is_top DESC, create_time DESC").
		Find(&notices).Error
	return notices, err
}
