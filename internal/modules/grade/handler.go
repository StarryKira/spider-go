package grade

import (
	"spider-go/internal/common"

	"github.com/gin-gonic/gin"
)

// Handler 成绩HTTP处理器
type Handler struct {
	service Service
}

// NewHandler 创建成绩处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	grades := r.Group("/grades")
	{
		grades.GET("", h.GetGrades)            // 获取成绩（可选term参数）
		grades.GET("/level", h.GetLevelGrades) // 获取等级考试成绩
	}
}

// GetGrades 获取成绩
// @Summary 获取成绩
// @Tags Grade
// @Produce json
// @Param term query string false "学期" example(2024-2025-1)
// @Success 200 {object} GradesResponse
// @Router /grades [get]
func (h *Handler) GetGrades(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	// 从 query params 获取学期参数
	term := c.Query("term")

	var grades []Grade
	var gpa *GPA
	var err error

	if term != "" {
		// 查询指定学期的成绩
		grades, gpa, err = h.service.GetGradesByTerm(c.Request.Context(), uid.(int), term)
	} else {
		// 查询所有成绩
		grades, gpa, err = h.service.GetAllGrades(c.Request.Context(), uid.(int))
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

// GetLevelGrades 获取等级考试成绩
// @Summary 获取等级考试成绩
// @Tags Grade
// @Produce json
// @Success 200 {array} LevelGrade
// @Router /grades/level [get]
func (h *Handler) GetLevelGrades(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	grades, err := h.service.GetLevelGrades(c.Request.Context(), uid.(int))
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
