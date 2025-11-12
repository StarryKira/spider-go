package controller

import (
	"spider-go/internal/dto"
	"spider-go/internal/handler/service"

	"github.com/gin-gonic/gin"
)

type GradeController struct {
	gradeSvc *service.GradeService
}

func NewGradeController(gradeSvc *service.GradeService) *GradeController {
	return &GradeController{gradeSvc: gradeSvc}
}

func (h *GradeController) GetAllGrade(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		dto.BadRequest(c, 114514, "invalid token")
	}
	grade, err := h.gradeSvc.GetAllGrade(uid.(int))
	if err != nil {
		dto.Error(c, 200, 111, err.Error())
		return
	}
	dto.Success(c, grade)
}
