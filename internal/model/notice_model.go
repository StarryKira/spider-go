package model

import "time"

// Notice 通知公告
type Notice struct {
	Nid        int       `gorm:"primary_key;AUTO_INCREMENT" json:"nid"`
	Content    string    `gorm:"type:text" json:"content"`          // 通知内容
	NoticeType string    `json:"notice_type"`                       // 通知类型
	IsShow     bool      `json:"is_show"`                           // 是否显示
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"` // 创建时间
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"` // 更新时间
	IsTop      bool      `json:"is_top"`                            // 是否置顶
	IsHtml     bool      `json:"is_html"`                           // 是否HTML格式
}
