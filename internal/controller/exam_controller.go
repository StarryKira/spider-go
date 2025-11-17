package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
	"spider-go/internal/service"

	"github.com/gin-gonic/gin"
)

// ExamController 考试控制器
type ExamController struct {
	examSvc service.ExamService
}

// NewExamController 创建考试控制器
func NewExamController(examSvc service.ExamService) *ExamController {
	return &ExamController{examSvc: examSvc}
}

// GetExams 获取考试安排
func (h *ExamController) GetExams(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	var req dto.ExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	exams, err := h.examSvc.GetAllExams(c.Request.Context(), uid.(int), req.Term)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取考试安排失败")
		}
		return
	}

	common.Success(c, exams)
}
