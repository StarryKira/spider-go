package admin

import (
	"net/http"
	"spider-go/internal/common"

	"github.com/gin-gonic/gin"
)

// Handler 管理员HTTP处理器
type Handler struct {
	service Service
}

// NewHandler 创建管理员处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(public *gin.RouterGroup, authenticated *gin.RouterGroup) {
	// 公开路由（无需认证）
	publicAdmin := public.Group("/admin")
	{
		publicAdmin.POST("/login", h.Login) // 管理员登录
	}

	// 需要认证的路由
	authAdmin := authenticated.Group("/admin")
	{
		authAdmin.GET("/info", h.GetInfo)                    // 获取管理员信息
		authAdmin.POST("/reset", h.ChangePassword)           // 修改密码
		authAdmin.POST("/broadcast-email", h.BroadcastEmail) // 群发邮件
	}
}

// Login 管理员登录
// @Summary 管理员登录
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} LoginResponse
// @Router /admin/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if err == ErrInvalidCredentials {
			common.Error(c, common.CodeInvalidPassword, err.Error())
		} else {
			common.Error(c, common.CodeInternalError, "登录失败")
		}
		return
	}

	common.Success(c, resp)
}

// GetInfo 获取管理员信息
// @Summary 获取管理员信息
// @Tags Admin
// @Produce json
// @Success 200 {object} AdminResponse
// @Router /admin/info [get]
func (h *Handler) GetInfo(c *gin.Context) {
	aid, exists := c.Get("aid")
	if !exists {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	adminInfo, err := h.service.GetAdminInfo(c.Request.Context(), aid.(int))
	if err != nil {
		common.Error(c, common.CodeUserNotFound, "获取管理员信息失败")
		return
	}

	common.Success(c, adminInfo)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body ChangePwdRequest true "修改密码请求"
// @Success 200 {object} gin.H
// @Router /admin/reset [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	aid, exists := c.Get("aid")
	if !exists {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	var req ChangePwdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), aid.(int), &req); err != nil {
		if err == ErrInvalidPassword {
			common.Error(c, common.CodeInvalidPassword, err.Error())
		} else {
			common.Error(c, common.CodeInternalError, "修改密码失败")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// BroadcastEmail 群发邮件
// @Summary 群发邮件给所有用户
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body BroadcastEmailRequest true "群发邮件请求"
// @Success 200 {object} BroadcastEmailResponse
// @Router /admin/broadcast-email [post]
func (h *Handler) BroadcastEmail(c *gin.Context) {
	var req BroadcastEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	resp, err := h.service.BroadcastEmail(c.Request.Context(), &req)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, resp)
}
