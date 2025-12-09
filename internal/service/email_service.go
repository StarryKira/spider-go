package service

import (
	"context"
	pkgemail "spider-go/pkg/email"
	pkgerrors "spider-go/pkg/errors"
)

// EmailService 邮件服务接口（适配层）
type EmailService interface {
	// SendVerificationCode 发送验证码邮件
	SendVerificationCode(ctx context.Context, to string, code string) error

	// SendEmail 发送普通邮件
	SendEmail(ctx context.Context, to string, subject string, body string) error
}

// emailServiceAdapter 邮件服务适配器
type emailServiceAdapter struct {
	emailService pkgemail.EmailService
}

// NewEmailService 创建邮件服务（适配 pkg/email）
func NewEmailService(smtpHost string, smtpPort int, username, password, fromName string) EmailService {
	return &emailServiceAdapter{
		emailService: pkgemail.NewEmailService(smtpHost, smtpPort, username, password, fromName),
	}
}

// SendVerificationCode 发送验证码邮件
func (a *emailServiceAdapter) SendVerificationCode(ctx context.Context, to string, code string) error {
	if err := a.emailService.SendVerificationCode(ctx, to, code); err != nil {
		return pkgerrors.NewAppError(pkgerrors.CodeInternalError, err.Error())
	}
	return nil
}

// SendEmail 发送普通邮件
func (a *emailServiceAdapter) SendEmail(ctx context.Context, to string, subject string, body string) error {
	if err := a.emailService.SendEmail(ctx, to, subject, body); err != nil {
		return pkgerrors.NewAppError(pkgerrors.CodeInternalError, err.Error())
	}
	return nil
}
