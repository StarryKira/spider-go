package admin

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrAdminNotFound = errors.New("admin not found")
)

// Repository 管理员数据访问接口
type Repository interface {
	Create(ctx context.Context, admin *Admin) error
	FindByID(ctx context.Context, uid int) (*Admin, error)
	FindByEmail(ctx context.Context, email string) (*Admin, error)
	UpdatePassword(ctx context.Context, uid int, password string) error
	CheckExists(ctx context.Context) (bool, error)
}

// adminRepository 管理员数据访问实现
type adminRepository struct {
	db *gorm.DB
}

// NewRepository 创建管理员数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &adminRepository{db: db}
}

// Create 创建管理员
func (r *adminRepository) Create(ctx context.Context, admin *Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

// FindByID 根据ID查找管理员
func (r *adminRepository) FindByID(ctx context.Context, uid int) (*Admin, error) {
	var admin Admin
	if err := r.db.WithContext(ctx).First(&admin, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}
	return &admin, nil
}

// FindByEmail 根据邮箱查找管理员
func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*Admin, error) {
	var admin Admin
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}
	return &admin, nil
}

// UpdatePassword 更新密码
func (r *adminRepository) UpdatePassword(ctx context.Context, uid int, password string) error {
	return r.db.WithContext(ctx).Model(&Admin{}).Where("uid = ?", uid).Update("password", password).Error
}

// CheckExists 检查是否存在管理员
func (r *adminRepository) CheckExists(ctx context.Context) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Admin{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
