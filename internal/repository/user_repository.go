package repository

import (
	"spider-go/internal/model"

	"gorm.io/gorm"
)

type UserWriter interface {
	CreateUser(user *model.User) error
	DeleteUser(uid int) error
	UpdatePassword(uid int, Password string) error
	UpdateJwc(uid int, sid string, spwd string) error
	UpdateAvatar(uid int, AvatarLink string) error
	UpdateName(uid int, name string) error
}

type UserReader interface {
	GetUserByUid(uid int) (*model.User, error)
	GetUserByEmail(Email string) (*model.User, error)
	GetUserByName(name string) (*model.User, error)
}

type UserRepository interface {
	UserReader
	UserWriter
}

type gormUserRepository struct {
	gormDB *gorm.DB
}

func (repo *gormUserRepository) GetUserByEmail(Email string) (*model.User, error) {
	var user = &model.User{}
	err := repo.gormDB.Where("email=?", Email).First(user).Error
	return user, err
}

func (repo *gormUserRepository) GetUserByUid(uid int) (*model.User, error) {
	var user = &model.User{}
	err := repo.gormDB.Find(user, uid).Error
	return user, err
}

func (repo *gormUserRepository) GetUserByName(name string) (*model.User, error) {
	var user = &model.User{}
	err := repo.gormDB.Where("name=?", name).First(user).Error
	return user, err
}

// 更新教务处信息
func (repo *gormUserRepository) UpdateJwc(uid int, sid string, spwd string) error {
	err := repo.gormDB.Model(&model.User{}).Where(uid).Update("sid", sid).Update("spwd", spwd).Error
	return err
}

func NewGormUserRepository(gormDB *gorm.DB) UserRepository {
	return &gormUserRepository{gormDB: gormDB}
}

func (repo *gormUserRepository) CreateUser(user *model.User) error {
	err := repo.gormDB.Create(user).Error
	return err
}

func (repo *gormUserRepository) DeleteUser(uid int) error {
	err := repo.gormDB.Delete(&model.User{}, uid).Error
	return err
}

func (repo *gormUserRepository) UpdatePassword(uid int, Password string) error {
	err := repo.gormDB.Model(&model.User{}).Where(uid).Update("password", Password).Error
	return err
}

func (repo *gormUserRepository) UpdateAvatar(uid int, AvatarLink string) error {
	err := repo.gormDB.Model(&model.User{}).Where(uid).Update("avatar", AvatarLink).Error
	return err
}

func (repo *gormUserRepository) UpdateName(uid int, name string) error {
	err := repo.gormDB.Model(&model.User{}).Where(uid).Update("name", name).Error
	return err
}
