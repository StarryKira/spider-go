package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
	"spider-go/internal/service"

	"github.com/gin-gonic/gin"
)

// GradeController 成绩控制器
type GradeController struct {
	gradeSvc service.GradeService
}

// NewGradeController 创建成绩控制器
func NewGradeController(gradeSvc service.GradeService) *GradeController {
	return &GradeController{gradeSvc: gradeSvc}
}

// GetAllGrade 获取所有成绩
func (h *GradeController) GetAllGrade(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	grades, gpa, err := h.gradeSvc.GetAllGrade(c.Request.Context(), uid.(int))
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取成绩失败")
		}
		return
	}

	common.Success(c, gin.H{
		"grades": grades,
		"gpa":    gpa,
	})
}

// GetGradeByTerm 根据学期获取成绩
func (h *GradeController) GetGradeByTerm(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	var req dto.GradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	grades, gpa, err := h.gradeSvc.GetGradeByTerm(c.Request.Context(), uid.(int), req.Term)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取成绩失败")
		}
		return
	}

	common.Success(c, gin.H{
		"grades": grades,
		"gpa":    gpa,
	})
}

// GetLevelGrade 获取等级考试成绩
func (h *GradeController) GetLevelGrade(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	grades, err := h.gradeSvc.GetLevelGrades(c.Request.Context(), uid.(int))
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取等级考试成绩失败")
		}
		return
	}

	common.Success(c, grades)
}
