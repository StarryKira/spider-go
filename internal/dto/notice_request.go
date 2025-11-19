package dto

// NoticeCreateRequest 创建通知请求
type NoticeCreateRequest struct {
	Content    string `json:"content" binding:"required"` // 通知内容
	NoticeType string `json:"notice_type"`                // 通知类型
	IsShow     bool   `json:"is_show"`                    // 是否显示
	IsTop      bool   `json:"is_top"`                     // 是否置顶
	IsHtml     bool   `json:"is_html"`                    // 是否HTML格式
}

// NoticeUpdateRequest 更新通知请求
type NoticeUpdateRequest struct {
	Content    string `json:"content" binding:"required"` // 通知内容
	NoticeType string `json:"notice_type"`                // 通知类型
	IsShow     bool   `json:"is_show"`                    // 是否显示
	IsTop      bool   `json:"is_top"`                     // 是否置顶
	IsHtml     bool   `json:"is_html"`                    // 是否HTML格式
}
