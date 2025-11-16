package controller

import (
	"spider-go/internal/dto"
	"spider-go/internal/handler/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseController struct {
	courseSvc *service.CourseService
}

func NewCourseController(courseSvc *service.CourseService) *CourseController {
	return &CourseController{courseSvc: courseSvc}
}

func (h *CourseController) GetCourseTable(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		dto.BadRequest(c, 114514, "invalid token")
		return
	}
	req := dto.CourseRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, 114514, "invalid request")
		return
	}

	week, err := strconv.Atoi(c.Param("week"))
	if err != nil {
		dto.BadRequest(c, 114514, "invalid week")
		return
	}
	courseTable, err := h.courseSvc.GetCourseTableByWeek(week, req.Term, uid.(int))
	if err != nil {
		dto.BadRequest(c, 114514, err.Error())
		return
	}
	dto.Success(c, courseTable)
	return
}
