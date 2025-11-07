package model

import "time"

type User struct {
	Uid       int `gorm:"primary_key;AUTO_INCREMENT"`
	Email     string
	Name      string
	Password  string
	Sid       string
	Spwd      string
	CreatedAt time.Time
	Avatar    string
}
