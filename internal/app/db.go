package app

import (
	"fmt"
	"spider-go/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDBWithConfig 使用配置初始化数据库
func InitDBWithConfig(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.User, config.Database.Pass, config.Database.Host, config.Database.Port, config.Database.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.User{}, &model.Notice{}, &model.Administrator{}); err != nil {
		return nil, err
	}

	return db, nil
}
