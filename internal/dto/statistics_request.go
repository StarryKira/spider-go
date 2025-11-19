package dto

// DAURequest 日活查询请求
type DAURequest struct {
	Date string `form:"date"` // 日期，格式：YYYY-MM-DD，不传则查询今天
}

// DAURangeRequest 日活范围查询请求
type DAURangeRequest struct {
	StartDate string `form:"start_date" binding:"required"` // 起始日期，格式：YYYY-MM-DD
	EndDate   string `form:"end_date" binding:"required"`   // 结束日期，格式：YYYY-MM-DD
}
