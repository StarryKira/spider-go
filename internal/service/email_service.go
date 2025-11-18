package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"spider-go/internal/common"

	"gopkg.in/gomail.v2"
)

// EmailService 邮件服务接口
type EmailService interface {
	// SendVerificationCode 发送验证码邮件
	SendVerificationCode(ctx context.Context, to string, code string) error

	// SendEmail 发送普通邮件
	SendEmail(ctx context.Context, to string, subject string, body string) error
}

// emailServiceImpl 邮件服务实现
type emailServiceImpl struct {
	smtpHost string
	smtpPort int
	username string
	password string
	fromName string
}

// NewEmailService 创建邮件服务
func NewEmailService(smtpHost string, smtpPort int, username, password, fromName string) EmailService {
	return &emailServiceImpl{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		fromName: fromName,
	}
}

// SendVerificationCode 发送验证码邮件
func (s *emailServiceImpl) SendVerificationCode(ctx context.Context, to string, code string) error {
	subject := "您的验证码"
	body := s.buildVerificationCodeHTML(code)
	return s.SendEmail(ctx, to, subject, body)
}

// SendEmail 发送邮件
func (s *emailServiceImpl) SendEmail(ctx context.Context, to string, subject string, body string) error {
	m := gomail.NewMessage()

	// 设置发件人
	m.SetHeader("From", m.FormatAddress(s.username, s.fromName))

	// 设置收件人
	m.SetHeader("To", to)

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置邮件正文（HTML 格式）
	m.SetBody("text/html", body)

	// 创建 SMTP 拨号器
	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.username, s.password)

	// 跳过证书验证（如果需要的话）
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return common.NewAppError(common.CodeInternalError, fmt.Sprintf("发送邮件失败: %v", err))
	}

	return nil
}

// buildVerificationCodeHTML 构建验证码邮件 HTML 内容
func (s *emailServiceImpl) buildVerificationCodeHTML(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .code-box { 
            background: #f4f4f4; 
            padding: 15px; 
            text-align: center; 
            font-size: 24px; 
            font-weight: bold; 
            color: #2196F3;
            letter-spacing: 5px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer { font-size: 12px; color: #999; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <h2>验证码</h2>
        <p>您好，</p>
        <p>您正在进行邮箱验证，验证码为：</p>
        <div class="code-box">%s</div>
        <p>验证码 5 分钟内有效，请及时使用。</p>
        <p>如果这不是您本人的操作，请忽略此邮件。</p>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`, code)
}
