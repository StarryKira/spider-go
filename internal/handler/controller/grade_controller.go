package controller

import (
	"flag"
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
	dto.Success(c, gin.H{"total": len(grade), "gradeList": grade})
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

}
