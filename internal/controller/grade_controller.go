package controller

import (
	"spider-go/internal/common"
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

// GetGrades 获取成
// 如果传递 term 参数则查询指定学期，否则查询所有成绩
func (h *GradeController) GetGrades(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	// 从 query params 获取学期参数（RESTful 规范）
	term := c.Query("term")

	var grades []service.Grade
	var gpa *service.GPA
	var err error

	if term != "" {
		// 查询指定学期的成绩
		grades, gpa, err = h.gradeSvc.GetGradeByTerm(c.Request.Context(), uid.(int), term)
	} else {
		// 查询所有成绩
		grades, gpa, err = h.gradeSvc.GetAllGrade(c.Request.Context(), uid.(int))
	}

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
