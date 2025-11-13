package dto

type GradeRequest struct {
	Term string `json:"term" binding:"required"`
}
