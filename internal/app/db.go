package app

import (
	"spider-go/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	if err := LoadConfig(); err != nil {
		panic(err.Error())
	}
	db, err := gorm.Open(mysql.Open(GetDSN()), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	_ = db.AutoMigrate(
		&model.User{},
	)
	return db
}
