package dto

// SetCurrentTermRequest 设置当前学期请求
type SetCurrentTermRequest struct {
	Term string `json:"term" binding:"required"` // 学期，格式：2024-2025-1
}

// SetSemesterDatesRequest 设置学期开学和放假时间请求
type SetSemesterDatesRequest struct {
	Term      string `json:"term" binding:"required"`       // 学期，格式：2024-2025-1
	StartDate string `json:"start_date" binding:"required"` // 开学时间，格式：2024-09-01
	EndDate   string `json:"end_date" binding:"required"`   // 放假时间，格式：2025-01-15
}

// GetSemesterDatesRequest 获取学期开学和放假时间请求
type GetSemesterDatesRequest struct {
	Term string `form:"term" binding:"required"` // 学期，格式：2024-2025-1
}
