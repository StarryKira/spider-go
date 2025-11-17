package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
	"spider-go/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	userSvc service.UserService
}

// NewUserController 创建用户控制器
func NewUserController(userSvc service.UserService) *UserController {
	return &UserController{userSvc: userSvc}
}

// Login 用户登录
func (h *UserController) Login(c *gin.Context) {
	var req dto.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	token, err := h.userSvc.UserLogin(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "登录失败")
		}
		return
	}

	// 设置 Cookie
	maxAge := int((168 * time.Hour).Seconds())
	c.SetCookie("access_token", token, maxAge, "/", "", true, true)

	common.Success(c, gin.H{
		"token": token,
	})
}

// Register 用户注册
func (h *UserController) Register(c *gin.Context) {
	var req dto.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	if err := h.userSvc.Register(c.Request.Context(), req.Name, req.Email, req.Password); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "注册失败")
		}
		return
	}

	common.Success(c, gin.H{"message": "注册成功"})
}

// BindJwcAccount 绑定教务系统账号
func (h *UserController) BindJwcAccount(c *gin.Context) {
	var req dto.BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	// 从中间件获取 uid
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	if err := h.userSvc.Bind(c.Request.Context(), uid.(int), req.Sid, req.Spwd); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "绑定失败")
		}
		return
	}

	common.Success(c, gin.H{"message": "绑定成功"})
}

// GetUserInfo 获取用户信息
func (h *UserController) GetUserInfo(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	user, err := h.userSvc.GetUserInfo(c.Request.Context(), uid.(int))
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "获取用户信息失败")
		}
		return
	}

	common.Success(c, user)
}
