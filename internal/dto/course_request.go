package dto

type CourseRequest struct {
	Term string `json:"term" binding:"required"`
}
