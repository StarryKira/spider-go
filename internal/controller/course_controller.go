package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CourseController 课程控制器
type CourseController struct {
	courseSvc service.CourseService
}

// NewCourseController 创建课程控制器
func NewCourseController(courseSvc service.CourseService) *CourseController {
	return &CourseController{courseSvc: courseSvc}
}

// GetCourses 获取课程表（RESTful 规范）
// 使用 query params 传递参数
func (h *CourseController) GetCourses(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	// 从 query params 获取参数（RESTful 规范）
	weekStr := c.Query("week")
	term := c.Query("term")

	// 验证必填参数
	if weekStr == "" {
		common.Error(c, common.CodeInvalidParams, "week 参数不能为空")
		return
	}
	if term == "" {
		common.Error(c, common.CodeInvalidParams, "term 参数不能为空")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "week 格式错误")
		return
	}

	courseTable, err := h.courseSvc.GetCourseTableByWeek(c.Request.Context(), week, term, uid.(int))
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取课程表失败")
		}
		return
	}

	common.Success(c, courseTable)
}
