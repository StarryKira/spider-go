package grade

// Grade 成绩信息
type Grade struct {
	SerialNo string  `json:"serialNo"` // 序号
	Term     string  `json:"Year"`     // 学期
	Code     string  `json:"Code"`     // 课程代码
	Subject  string  `json:"subject"`  // 课程名称
	Score    string  `json:"score"`    // 分数
	Credit   float64 `json:"credit"`   // 学分
	Gpa      float64 `json:"gpa"`      // 绩点
	Status   int     `json:"Status"`   // 状态：0=正常考试，1=补考/重修
	Property string  `json:"property"` // 课程性质：必修/选修
}

// GPA 绩点信息
type GPA struct {
	AverageGPA   float64 `json:"averageGPA"`   // 平均绩点
	AverageScore float64 `json:"averageScore"` // 平均分
	BasicScore   float64 `json:"basicScore"`   // 基本分
}

// LevelGrade 等级考试成绩
type LevelGrade struct {
	No         string `json:"no"`         // 序号
	CourseName string `json:"CourseName"` // 考试名称
	LevGrade   string `json:"LevelGrade"` // 成绩/等级
	Time       string `json:"Time"`       // 考试时间
}

// GetGradesRequest 获取成绩请求
type GetGradesRequest struct {
	Term string `form:"term"` // 学期（可选），格式：2024-2025-1
}

// GradesResponse 成绩响应
type GradesResponse struct {
	Grades []Grade `json:"grades"`
	GPA    *GPA    `json:"gpa"`
}
