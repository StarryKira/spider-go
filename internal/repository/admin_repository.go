package repository

import (
	"spider-go/internal/model"

	"gorm.io/gorm"
)

// AdminRepository 管理员仓储接口
type AdminRepository interface {
	// GetAdminByEmail 根据邮箱获取管理员
	GetAdminByEmail(email string) (*model.Administrator, error)
	// GetAdminByUid 根据 UID 获取管理员
	GetAdminByUid(uid int) (*model.Administrator, error)
	// CreateAdmin 创建管理员
	CreateAdmin(admin *model.Administrator) error
	// CheckAdminExists 检查管理员是否存在
	CheckAdminExists() (bool, error)
	//
	UpdateAdminPassword(email string, password string) error
}

// gormAdminRepository GORM 实现的管理员仓储
type gormAdminRepository struct {
	db *gorm.DB
}

func (r *gormAdminRepository) UpdateAdminPassword(email string, password string) error {
	err := r.db.Model(&model.Administrator{}).Where("email=?", email).Update("password", password).Error
	return err
}

// NewGormAdminRepository 创建 GORM 管理员仓储
func NewGormAdminRepository(db *gorm.DB) AdminRepository {
	return &gormAdminRepository{db: db}
}

// GetAdminByEmail 根据邮箱获取管理员
func (r *gormAdminRepository) GetAdminByEmail(email string) (*model.Administrator, error) {
	var admin model.Administrator
	err := r.db.Where("email = ?", email).First(&admin).Error
	return &admin, err
}

// GetAdminByUid 根据 UID 获取管理员
func (r *gormAdminRepository) GetAdminByUid(uid int) (*model.Administrator, error) {
	var admin model.Administrator
	err := r.db.First(&admin, uid).Error
	return &admin, err
}

// CreateAdmin 创建管理员
func (r *gormAdminRepository) CreateAdmin(admin *model.Administrator) error {
	return r.db.Create(admin).Error
}

// CheckAdminExists 检查是否存在管理员
func (r *gormAdminRepository) CheckAdminExists() (bool, error) {
	var count int64
	err := r.db.Model(&model.Administrator{}).Count(&count).Error
	return count > 0, err
}
