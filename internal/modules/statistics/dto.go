package statistics

// DAUResponse DAU响应
type DAUResponse struct {
	Date  string `json:"date"`  // 日期：2024-01-01
	Count int64  `json:"count"` // 日活数量
}

// DAURangeResponse DAU范围响应
type DAURangeResponse struct {
	StartDate string           `json:"start_date"` // 开始日期
	EndDate   string           `json:"end_date"`   // 结束日期
	Data      []DAUDayResponse `json:"data"`       // 每日数据
}

// DAUDayResponse 每日DAU数据
type DAUDayResponse struct {
	Date  string `json:"date"`  // 日期：2024-01-01
	Count int64  `json:"count"` // 日活数量
}
