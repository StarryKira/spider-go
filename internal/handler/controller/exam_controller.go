package controller

import (
	"spider-go/internal/dto"
	"spider-go/internal/handler/service"

	"github.com/gin-gonic/gin"
)

type ExamController struct {
	examSvc *service.ExamService
}

func NewExamController(examSvc *service.ExamService) *ExamController {
	return &ExamController{examSvc: examSvc}
}

func (h *ExamController) GetExams(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		dto.BadRequest(c, 114514, "invalid token")
		return
	}
	req := dto.ExamRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, 114514, "请求体错误")
		return
	}
	exams, err := h.examSvc.GetAllExams(uid.(int), req.Term)
	if err != nil {
		dto.BadRequest(c, 113514, err.Error())
		return
	}
	if exams == nil {
		dto.Success(c, "暂时没有数据")
		return
	}

	dto.Success(c, exams)
	return
}
