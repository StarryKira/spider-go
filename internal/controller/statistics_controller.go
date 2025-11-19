package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
	"spider-go/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// StatisticsController 统计控制器
type StatisticsController struct {
	dauService service.DAUService
}

// NewStatisticsController 创建统计控制器
func NewStatisticsController(dauService service.DAUService) *StatisticsController {
	return &StatisticsController{dauService: dauService}
}

// GetDAUStatistics 获取日活统计
func (h *StatisticsController) GetDAUStatistics(c *gin.Context) {
	var req dto.DAURequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	// 如果没有传日期，默认查询今天
	if req.Date == "" {
		todayDAU, err := h.dauService.GetTodayDAU(c.Request.Context())
		if err != nil {
			if appErr, ok := err.(*common.AppError); ok {
				common.ErrorWithAppError(c, appErr)
			} else {
				common.Error(c, common.CodeInternalError, "获取日活失败")
			}
			return
		}

		common.Success(c, gin.H{
			"date": time.Now().Format("2006-01-02"),
			"dau":  todayDAU,
		})
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "日期格式错误，应为 YYYY-MM-DD")
		return
	}

	dau, err := h.dauService.GetDAUByDate(c.Request.Context(), date)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取日活失败")
		}
		return
	}

	common.Success(c, gin.H{
		"date": req.Date,
		"dau":  dau,
	})
}

// GetDAURange 获取指定日期范围的日活统计
func (h *StatisticsController) GetDAURange(c *gin.Context) {
	var req dto.DAURangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	// 解析起始日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "起始日期格式错误")
		return
	}

	// 解析结束日期
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "结束日期格式错误")
		return
	}

	// 验证日期范围
	if endDate.Before(startDate) {
		common.Error(c, common.CodeInvalidParams, "结束日期不能早于起始日期")
		return
	}

	// 限制查询范围不超过 90 天
	if endDate.Sub(startDate).Hours()/24 > 90 {
		common.Error(c, common.CodeInvalidParams, "查询范围不能超过90天")
		return
	}

	dauMap, err := h.dauService.GetDAURange(c.Request.Context(), startDate, endDate)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取日活统计失败")
		}
		return
	}

	common.Success(c, dauMap)
}
