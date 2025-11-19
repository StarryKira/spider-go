package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/service"

	"github.com/gin-gonic/gin"
)

// GradeAnalysisController 成绩分析控制器
type GradeAnalysisController struct {
	gradeAnalysisSvc service.GradeAnalysisService
}

// NewGradeAnalysisController 创建成绩分析控制器
func NewGradeAnalysisController(gradeAnalysisSvc service.GradeAnalysisService) *GradeAnalysisController {
	return &GradeAnalysisController{gradeAnalysisSvc: gradeAnalysisSvc}
}

// GetRecentTermsAnalysis 获取最近三个学期的成绩分析
func (h *GradeAnalysisController) GetRecentTermsAnalysis(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	analysis, err := h.gradeAnalysisSvc.GetRecentTermsGrades(c.Request.Context(), uid.(int))
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取成绩分析失败")
		}
		return
	}

	common.Success(c, analysis)
}
