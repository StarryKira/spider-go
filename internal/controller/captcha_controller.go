package controller

import (
	"spider-go/internal/common"
	"spider-go/internal/dto"
	"spider-go/internal/service"

	"github.com/gin-gonic/gin"
)

// CaptchaController 验证码控制器
type CaptchaController struct {
	captchaSvc service.CaptchaService
}

// NewCaptchaController 创建验证码控制器
func NewCaptchaController(captchaSvc service.CaptchaService) *CaptchaController {
	return &CaptchaController{captchaSvc: captchaSvc}
}

// SendEmailCaptcha 发送邮箱验证码
func (h *CaptchaController) SendEmailCaptcha(c *gin.Context) {
	var req dto.SendCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	// 发送验证码
	if err := h.captchaSvc.SendEmailCaptcha(c.Request.Context(), req.Email); err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.ErrorWithAppError(c, appErr)
		} else {
			common.Error(c, common.CodeInternalError, "发送验证码失败")
		}
		return
	}

	common.Success(c, gin.H{
		"message": "验证码已发送，请查收邮件",
	})
}
