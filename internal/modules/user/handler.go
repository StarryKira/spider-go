package user

import (
	"net/http"
	"spider-go/internal/common"

	"github.com/gin-gonic/gin"
)

// Handler 用户HTTP处理器
type Handler struct {
	service        Service
	captchaService CaptchaService
}

// NewHandler 创建用户处理器
func NewHandler(service Service, captchaService CaptchaService) *Handler {
	return &Handler{
		service:        service,
		captchaService: captchaService,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(public *gin.RouterGroup, authenticated *gin.RouterGroup) {
	// 公开路由（无需认证）
	publicUser := public.Group("/user")
	{
		publicUser.POST("/register", h.Register)            // 用户注册
		publicUser.POST("/login", h.Login)                  // 用户登录
		publicUser.POST("/reset-password", h.ResetPassword) // 重置密码
	}

	// 验证码路由（公开）
	captcha := public.Group("/captcha")
	{
		captcha.POST("/send", h.SendEmailCaptcha) // 发送邮箱验证码
	}

	authenticated.GET("/info", h.GetUserInfo)    // 获取用户信息
	authenticated.POST("/bind", h.BindJwc)       // 绑定教务系统
	authenticated.GET("/is-bind", h.CheckIsBind) // 检查绑定状态
}

// Register 用户注册
// @Summary 用户注册
// @Tags User
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求"
// @Success 200 {object} gin.H
// @Router /user/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	token, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.Name, req.Captcha)
	if err != nil {
		if err == ErrEmailAlreadyExists {
			common.Error(c, common.CodeUserAlreadyExists, err.Error())
		} else if err == ErrInvalidCaptcha {
			common.Error(c, common.CodeCaptchaInvalid, err.Error())
		} else {
			common.Error(c, common.CodeInternalError, "注册失败")
		}
		return
	}

	common.Success(c, gin.H{
		"token":   token,
		"message": "注册成功",
	})
}

// Login 用户登录
// @Summary 用户登录
// @Tags User
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} LoginResponse
// @Router /user/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	token, user, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == ErrInvalidCredentials {
			common.Error(c, common.CodeInvalidPassword, err.Error())
		} else {
			common.Error(c, common.CodeInternalError, "登录失败")
		}
		return
	}

	common.Success(c, LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	})
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Tags User
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} gin.H
// @Router /user/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Email, req.Password, req.Captcha); err != nil {
		if err == ErrUserNotFound {
			common.Error(c, common.CodeUserNotFound, err.Error())
		} else if err == ErrInvalidCaptcha {
			common.Error(c, common.CodeCaptchaInvalid, err.Error())
		} else {
			common.Error(c, common.CodeInternalError, "重置密码失败")
		}
		return
	}

	common.Success(c, gin.H{"message": "密码重置成功"})
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Tags User
// @Produce json
// @Success 200 {object} UserResponse
// @Router /user/info [get]
func (h *Handler) GetUserInfo(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	user, err := h.service.GetUserInfo(c.Request.Context(), uid.(int))
	if err != nil {
		common.Error(c, common.CodeUserNotFound, "获取用户信息失败")
		return
	}

	common.Success(c, user.ToResponse())
}

// BindJwc 绑定教务系统
// @Summary 绑定教务系统
// @Tags User
// @Accept json
// @Produce json
// @Param request body BindJwcRequest true "绑定请求"
// @Success 200 {object} gin.H
// @Router /user/bind [post]
func (h *Handler) BindJwc(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	var req BindJwcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	if err := h.service.BindJwc(c.Request.Context(), uid.(int), req.Sid, req.Spwd); err != nil {
		if err == ErrEmptyParams {
			common.Error(c, common.CodeInvalidParams, err.Error())
		} else {
			common.Error(c, common.CodeJwcInvalidParams, err.Error())
		}
		return
	}

	common.Success(c, gin.H{"message": "绑定成功"})
}

// CheckIsBind 检查绑定状态
// @Summary 检查绑定状态
// @Tags User
// @Produce json
// @Success 200 {object} gin.H
// @Router /user/is-bind [get]
func (h *Handler) CheckIsBind(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		common.Error(c, common.CodeUnauthorized, "未授权")
		return
	}

	isBind, err := h.service.CheckIsBind(c.Request.Context(), uid.(int))
	if err != nil {
		common.Error(c, common.CodeUserNotFound, "获取绑定状态失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_bind": isBind})
}

// SendEmailCaptcha 发送邮箱验证码
// @Summary 发送邮箱验证码
// @Tags Captcha
// @Accept json
// @Produce json
// @Param request body SendEmailCaptchaRequest true "发送验证码请求"
// @Success 200 {object} gin.H
// @Router /captcha/send [post]
func (h *Handler) SendEmailCaptcha(c *gin.Context) {
	var req SendEmailCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	if err := h.captchaService.SendEmailCaptcha(c.Request.Context(), req.Email); err != nil {
		common.Error(c, common.CodeInternalError, "发送验证码失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}
