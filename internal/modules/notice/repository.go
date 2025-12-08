package notice

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNoticeNotFound = errors.New("notice not found")
)

// Repository 通知数据访问接口
type Repository interface {
	Create(ctx context.Context, notice *Notice) error
	Update(ctx context.Context, notice *Notice) error
	Delete(ctx context.Context, nid int) error
	FindByID(ctx context.Context, nid int) (*Notice, error)
	FindAll(ctx context.Context) ([]*Notice, error)
	FindVisible(ctx context.Context) ([]*Notice, error)
}

// repository 通知数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建通知数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建通知
func (r *repository) Create(ctx context.Context, notice *Notice) error {
	return r.db.WithContext(ctx).Create(notice).Error
}

// Update 更新通知
func (r *repository) Update(ctx context.Context, notice *Notice) error {
	return r.db.WithContext(ctx).Save(notice).Error
}

// Delete 删除通知
func (r *repository) Delete(ctx context.Context, nid int) error {
	return r.db.WithContext(ctx).Delete(&Notice{}, nid).Error
}

// FindByID 根据ID查找通知
func (r *repository) FindByID(ctx context.Context, nid int) (*Notice, error) {
	var notice Notice
	if err := r.db.WithContext(ctx).First(&notice, nid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoticeNotFound
		}
		return nil, err
	}
	return &notice, nil
}

// FindAll 获取所有通知（管理员）
func (r *repository) FindAll(ctx context.Context) ([]*Notice, error) {
	var notices []*Notice
	err := r.db.WithContext(ctx).
		Order("is_top DESC, create_time DESC").
		Find(&notices).Error
	return notices, err
}

// FindVisible 获取可见通知（普通用户）
func (r *repository) FindVisible(ctx context.Context) ([]*Notice, error) {
	var notices []*Notice
	err := r.db.WithContext(ctx).
		Where("is_show = ?", true).
		Order("is_top DESC, create_time DESC").
		Find(&notices).Error
	return notices, err
}
