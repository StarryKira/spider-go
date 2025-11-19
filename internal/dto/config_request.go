package dto

// SetCurrentTermRequest 设置当前学期请求
type SetCurrentTermRequest struct {
	Term string `json:"term" binding:"required"` // 学期，格式：2024-2025-1
}
