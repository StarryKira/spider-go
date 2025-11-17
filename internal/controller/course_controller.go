package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
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

// GetCourseTable 获取课程表
func (h *CourseController) GetCourseTable(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	var req dto.CourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	week, err := strconv.Atoi(c.Param("week"))
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "周次格式错误")
		return
	}

	courseTable, err := h.courseSvc.GetCourseTableByWeek(c.Request.Context(), week, req.Term, uid.(int))
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
