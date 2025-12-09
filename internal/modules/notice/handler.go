package notice

import (
	"net/http"
	"spider-go/internal/common"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 通知HTTP处理器
type Handler struct {
	service Service
}

// NewHandler 创建通知处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup, adminGroup *gin.RouterGroup) {
	// 公开接口 - 获取可见通知
	notices := r.Group("/notices")
	{
		notices.GET("", h.GetVisibleNotices) // 获取可见通知
		notices.GET("/:id", h.GetNoticeByID) // 获取通知详情
	}

	// 管理员接口
	if adminGroup != nil {
		adminNotices := adminGroup.Group("/notices")
		{
			adminNotices.GET("", h.GetAllNotices)       // 获取所有通知
			adminNotices.POST("", h.CreateNotice)       // 创建通知
			adminNotices.PUT("/:id", h.UpdateNotice)    // 更新通知
			adminNotices.DELETE("/:id", h.DeleteNotice) // 删除通知
		}
	}
}

// GetVisibleNotices 获取可见通知（普通用户）
// @Summary 获取可见通知
// @Tags Notice
// @Produce json
// @Success 200 {array} Notice
// @Router /notices [get]
func (h *Handler) GetVisibleNotices(c *gin.Context) {
	notices, err := h.service.GetVisible(c.Request.Context())
	if err != nil {
		common.Error(c, common.CodeInternalError, "获取通知列表失败")
		return
	}

	common.Success(c, notices)
}

// GetNoticeByID 获取通知详情
// @Summary 获取通知详情
// @Tags Notice
// @Produce json
// @Param id path int true "通知ID"
// @Success 200 {object} Notice
// @Router /notices/{id} [get]
func (h *Handler) GetNoticeByID(c *gin.Context) {
	nid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "无效的通知ID")
		return
	}

	notice, err := h.service.GetByID(c.Request.Context(), nid)
	if err != nil {
		if err == ErrNoticeNotFound {
			common.Error(c, common.CodeNotFound, "通知不存在")
		} else {
			common.Error(c, common.CodeInternalError, "获取通知失败")
		}
		return
	}

	common.Success(c, notice)
}

// GetAllNotices 获取所有通知（管理员）
// @Summary 获取所有通知
// @Tags Admin/Notice
// @Produce json
// @Success 200 {array} Notice
// @Router /admin/notices [get]
func (h *Handler) GetAllNotices(c *gin.Context) {
	notices, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		common.Error(c, common.CodeInternalError, "获取通知列表失败")
		return
	}

	common.Success(c, notices)
}

// CreateNotice 创建通知（管理员）
// @Summary 创建通知
// @Tags Admin/Notice
// @Accept json
// @Produce json
// @Param request body CreateNoticeRequest true "创建通知请求"
// @Success 200 {object} Notice
// @Router /admin/notices [post]
func (h *Handler) CreateNotice(c *gin.Context) {
	var req CreateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	notice, err := h.service.Create(c.Request.Context(), req.Content, req.NoticeType, req.IsShow, req.IsTop, req.IsHtml)
	if err != nil {
		if err == ErrEmptyContent {
			common.Error(c, common.CodeInvalidParams, err.Error())
		} else {
			common.Error(c, common.CodeInternalError, "创建通知失败")
		}
		return
	}

	common.Success(c, notice)
}

// UpdateNotice 更新通知（管理员）
// @Summary 更新通知
// @Tags Admin/Notice
// @Accept json
// @Produce json
// @Param id path int true "通知ID"
// @Param request body UpdateNoticeRequest true "更新通知请求"
// @Success 200 {object} Notice
// @Router /admin/notices/{id} [put]
func (h *Handler) UpdateNotice(c *gin.Context) {
	nid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "无效的通知ID")
		return
	}

	var req UpdateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	notice, err := h.service.Update(c.Request.Context(), nid, req.Content, req.NoticeType, req.IsShow, req.IsTop, req.IsHtml)
	if err != nil {
		if err == ErrNoticeNotFound {
			common.Error(c, common.CodeNotFound, "通知不存在")
		} else if err == ErrEmptyContent {
			common.Error(c, common.CodeInvalidParams, err.Error())
		} else {
			common.Error(c, common.CodeInternalError, "更新通知失败")
		}
		return
	}

	common.Success(c, notice)
}

// DeleteNotice 删除通知（管理员）
// @Summary 删除通知
// @Tags Admin/Notice
// @Param id path int true "通知ID"
// @Success 200 {object} gin.H
// @Router /admin/notices/{id} [delete]
func (h *Handler) DeleteNotice(c *gin.Context) {
	nid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.Error(c, common.CodeInvalidParams, "无效的通知ID")
		return
	}

	if err := h.service.Delete(c.Request.Context(), nid); err != nil {
		common.Error(c, common.CodeInternalError, "删除通知失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知删除成功"})
}
