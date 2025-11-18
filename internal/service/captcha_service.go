package service

import (
	"context"
	"fmt"
	"math/rand"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"time"
)

// CaptchaService 验证码服务接口
type CaptchaService interface {
	// SendEmailCaptcha 发送邮箱验证码
	SendEmailCaptcha(ctx context.Context, email string) error

	// VerifyEmailCaptcha 验证邮箱验证码
	VerifyEmailCaptcha(ctx context.Context, email string, code string) error
}

// captchaServiceImpl 验证码服务实现
type captchaServiceImpl struct {
	captchaCache cache.CaptchaCache
	emailService EmailService
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(captchaCache cache.CaptchaCache, emailService EmailService) CaptchaService {
	return &captchaServiceImpl{
		captchaCache: captchaCache,
		emailService: emailService,
	}
}

// SendEmailCaptcha 发送邮箱验证码
func (s *captchaServiceImpl) SendEmailCaptcha(ctx context.Context, email string) error {
	// 1. 生成 6 位数字验证码
	code := s.generateCode(6)

	// 2. 存储到 Redis，5 分钟过期
	if err := s.captchaCache.SetCaptcha(ctx, email, code, 5*time.Minute); err != nil {
		return common.NewAppError(common.CodeCacheError, "存储验证码失败")
	}

	// 3. 发送邮件
	if err := s.emailService.SendVerificationCode(ctx, email, code); err != nil {
		// 发送失败，删除已存储的验证码
		_ = s.captchaCache.DeleteCaptcha(ctx, email)
		return err
	}

	return nil
}

// VerifyEmailCaptcha 验证邮箱验证码
func (s *captchaServiceImpl) VerifyEmailCaptcha(ctx context.Context, email string, code string) error {
	// 使用原子操作验证并删除验证码
	valid, err := s.captchaCache.VerifyAndDelete(ctx, email, code)
	if err != nil {
		return common.NewAppError(common.CodeInvalidParams, err.Error())
	}

	if !valid {
		return common.NewAppError(common.CodeInvalidParams, "验证码错误")
	}

	return nil
}

// generateCode 生成指定位数的数字验证码
func (s *captchaServiceImpl) generateCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}
