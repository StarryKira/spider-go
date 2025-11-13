package controller

import (
	"spider-go/internal/dto"
	"spider-go/internal/handler/service"

	"github.com/gin-gonic/gin"
)

type GradeController struct {
	gradeSvc *service.GradeService
}

type GradeResponse struct {
	Total     int             `json:"total"`
	GPA       service.GPA     `json:"gpa"`
	GradeList []service.Grade `json:"grades"`
}

func NewGradeController(gradeSvc *service.GradeService) *GradeController {
	return &GradeController{gradeSvc: gradeSvc}
}

func (h *GradeController) GetAllGrade(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		dto.BadRequest(c, 114514, "invalid token")
	}
	grade, gpa, err := h.gradeSvc.GetAllGrade(uid.(int))
	if err != nil {
		dto.Error(c, 200, 111, err.Error())
		return
	}
	dto.Success(c, GradeResponse{
		GradeList: grade,
		GPA:       *gpa,
		Total:     len(grade),
	})
}

func (h *GradeController) GetGradeByTerm(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		dto.BadRequest(c, 114514, "invalid token")
	}
	req := dto.GradeRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, 114514, "invalid request")
		return
	}
	grades, gpa, err := h.gradeSvc.GetGradeByTerm(uid.(int), req.Term)
	if err != nil {
		dto.Error(c, 200, 114514, err.Error())
		return
	}
	dto.Success(c, GradeResponse{
		GradeList: grades,
		GPA:       *gpa,
		Total:     len(grades),
	})
}
