package controller

import (
	"spider-go/internal/common"
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

// GetExams 获取考试安排（RESTful 规范）
// 使用 query params 传递参数
func (h *ExamController) GetExams(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	// 从 query params 获取学期参数（RESTful 规范）
	term := c.Query("term")
	if term == "" {
		common.Error(c, common.CodeInvalidParams, "term 参数不能为空")
		return
	}

	exams, err := h.examSvc.GetAllExams(c.Request.Context(), uid.(int), term)
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
