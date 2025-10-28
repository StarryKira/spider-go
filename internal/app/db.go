package app

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() *gorm.DB {
	if err := LoadConfig(); err != nil {
		panic(err.Error())
	}
	db, err := gorm.Open(mysql.Open(GetDSN()), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	_ = db.AutoMigrate()
	return db
}
