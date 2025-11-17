package model

import "time"

type Notice struct {
	Nid        int `gorm:"primary_key;AUTO_INCREMENT"`
	NoticeType string
	IsShow     bool
	CreateTime time.Time
	updateTime time.Time
	isTop      bool
	isHtml     bool
}
