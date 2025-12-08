package exam

// ExamArrangement 考试安排
type ExamArrangement struct {
	SerialNo  string `json:"serial_no"`  // 序号
	ClassNo   string `json:"class_no"`   // 课程号
	ClassName string `json:"class_name"` // 课程名称
	Time      string `json:"time"`       // 考试时间
	Place     string `json:"place"`      // 考试地点
	Execution string `json:"execution"`  // 执行情况
}

// GetExamsRequest 获取考试安排请求
type GetExamsRequest struct {
	Term string `form:"term" binding:"required"` // 学期：2024-2025-1
}
