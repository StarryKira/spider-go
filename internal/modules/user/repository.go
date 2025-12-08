package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// Repository 用户数据访问接口
type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, uid int) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, uid int, password string) error
	UpdateJwc(ctx context.Context, uid int, sid, spwd string) error
	Delete(ctx context.Context, uid int) error
}

// repository 用户数据访问实现
type repository struct {
	db *gorm.DB
}

// NewRepository 创建用户数据访问层
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create 创建用户
func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID 根据ID查找用户
func (r *repository) FindByID(ctx context.Context, uid int) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdatePassword 更新密码
func (r *repository) UpdatePassword(ctx context.Context, uid int, password string) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("uid = ?", uid).Update("password", password).Error
}

// UpdateJwc 更新教务系统绑定
func (r *repository) UpdateJwc(ctx context.Context, uid int, sid, spwd string) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"sid":  sid,
		"spwd": spwd,
	}).Error
}

// Delete 删除用户
func (r *repository) Delete(ctx context.Context, uid int) error {
	return r.db.WithContext(ctx).Delete(&User{}, uid).Error
}
