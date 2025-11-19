package model

import "time"

// Administrator 管理员
type Administrator struct {
	Uid       int       `gorm:"primary_key;AUTO_INCREMENT" json:"uid"`
	Email     string    `gorm:"unique" json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"` // 密码不返回给前端
	CreatedAt time.Time `json:"created_at"`
	Avatar    string    `json:"avatar"`
}
