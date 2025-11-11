package controller

import (
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
	uid := c.GetInt("uid")
	grade, err := h.gradeSvc.GetAllGrade(uid)
	if err != nil {
		return
	}

	c.JSON(200, grade)

}
