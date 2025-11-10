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

func (c *GradeController) GetAllGrade(*gin.Context) {

}
