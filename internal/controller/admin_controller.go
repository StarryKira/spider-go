package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
	"spider-go/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminController 管理员控制器
type AdminController struct {
	adminSvc service.AdminService
}

// NewAdminController 创建管理员控制器
func NewAdminController(adminSvc service.AdminService) *AdminController {
	return &AdminController{adminSvc: adminSvc}
}

// Login 管理员登录
func (h *AdminController) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.NewAppError(common.CodeInvalidParams, "参数错误")
		return
	}

	token, err := h.adminSvc.AdminLogin(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "登录失败")
		}
		return
	}

	// 设置 Cookie
	maxAge := int((24 * time.Hour).Seconds())
	c.SetCookie("admin_token", token, maxAge, "/", "", true, true)

	common.Success(c, gin.H{
		"token": token,
	})
}

// GetInfo 获取管理员信息
func (h *AdminController) GetInfo(c *gin.Context) {
	adminUid, ok := c.Get("admin_uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	admin, err := h.adminSvc.GetAdminInfo(c.Request.Context(), adminUid.(int))
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取管理员信息失败")
		}
		return
	}

	common.Success(c, admin)
}

func (h *AdminController) ChangePwd(c *gin.Context) {
	req := dto.AdminLoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
	}
	_, ok := c.Get("admin_uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
	}
	err := h.adminSvc.ResetPassword(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, gin.H{"message": "重置成功"})
	return
}
