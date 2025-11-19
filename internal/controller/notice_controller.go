package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
	"spider-go/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NoticeController 通知控制器
type NoticeController struct {
	noticeSvc service.NoticeService
}

// NewNoticeController 创建通知控制器
func NewNoticeController(noticeSvc service.NoticeService) *NoticeController {
	return &NoticeController{noticeSvc: noticeSvc}
}

// CreateNotice 创建通知（管理员）
func (h *NoticeController) CreateNotice(c *gin.Context) {
	var req dto.NoticeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	if err := h.noticeSvc.CreateNotice(
		c.Request.Context(),
		req.Content,
		req.NoticeType,
		req.IsShow,
		req.IsTop,
		req.IsHtml,
	); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "创建通知失败")
		}
		return
	}

	common.Success(c, gin.H{"message": "创建成功"})
}

// UpdateNotice 更新通知（管理员）
func (h *NoticeController) UpdateNotice(c *gin.Context) {
	nid, err := strconv.Atoi(c.Param("nid"))
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "通知ID格式错误")
		return
	}

	var req dto.NoticeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	if err := h.noticeSvc.UpdateNotice(
		c.Request.Context(),
		nid,
		req.Content,
		req.NoticeType,
		req.IsShow,
		req.IsTop,
		req.IsHtml,
	); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "更新通知失败")
		}
		return
	}

	common.Success(c, gin.H{"message": "更新成功"})
}

// DeleteNotice 删除通知（管理员）
func (h *NoticeController) DeleteNotice(c *gin.Context) {
	nid, err := strconv.Atoi(c.Param("nid"))
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "通知ID格式错误")
		return
	}

	if err := h.noticeSvc.DeleteNotice(c.Request.Context(), nid); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "删除通知失败")
		}
		return
	}

	common.Success(c, gin.H{"message": "删除成功"})
}

// GetAllNotices 获取所有通知（管理员）
func (h *NoticeController) GetAllNotices(c *gin.Context) {
	notices, err := h.noticeSvc.GetAllNotices(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取通知列表失败")
		}
		return
	}

	common.Success(c, notices)
}

// GetVisibleNotices 获取可见通知（普通用户）
func (h *NoticeController) GetVisibleNotices(c *gin.Context) {
	notices, err := h.noticeSvc.GetVisibleNotices(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取通知列表失败")
		}
		return
	}

	common.Success(c, notices)
}
