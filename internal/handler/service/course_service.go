package service

import "spider-go/internal/repository"

type CourseService struct {
	uRepo repository.UserRepository
}

func NewCourseService(uRepo repository.UserRepository) *CourseService {
	return &CourseService{uRepo: uRepo}
}

type DaySchedule struct {
	Weekday int      `json:"weekday"` //值为1-7，表示周一到周日
	Courses []Course `json:"courses"` //当天课程
}

type WeekSchedule struct {
	WeekNo    int           `json:"weekno"`
	Starttime string        `json:"starttime"`
	Endtime   string        `json:"endtime"`
	Days      []DaySchedule `json:"days"`
}

type Course struct {
	ID          int    `json:"id"`           // 课程唯一ID（可选）
	Name        string `json:"name"`         // 课程名称：高等数学
	Teacher     string `json:"teacher"`      // 任课老师
	Classroom   string `json:"classroom"`    // 教室：A1-203
	Weekday     int    `json:"weekday"`      // 周几：1~7 表示周一~周日
	StartPeriod int    `json:"start_period"` // 第几节开始：1 表示第一节
	EndPeriod   int    `json:"end_period"`   // 第几节结束：2 表示上到第二节
	StartTime   string `json:"start_time"`   // "08:00"
	EndTime     string `json:"end_time"`     // "09:40"
}

func (s *CourseService) getCourseTable(week int) (*WeekSchedule, error) {

}
