package model

import "time"

type Administrator struct {
	Uid       int    `gorm:"primary_key;AUTO_INCREMENT"`
	Email     string `gorm:"unique"`
	Name      string
	Password  string
	CreatedAt time.Time
	Avatar    string
}
