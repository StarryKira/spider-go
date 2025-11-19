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

// buildVerificationCodeHTML 构建验证码邮件 HTML 内容 下面的是死二次元
//func (s *emailServiceImpl) buildVerificationCodeHTML(code string) string {
//	return fmt.Sprintf(`<!DOCTYPE html>
//<html lang="zh-CN">
//<head>
//  <meta charset="UTF-8" />
//  <title>验证码邮件</title>
//</head>
//<body style="margin:0;padding:0;background-color:#f5f5f5;">
//  <table width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:#f5f5f5;padding:24px 0;">
//    <tr>
//      <td align="center">
//        <table width="520" cellpadding="0" cellspacing="0" border="0" style="background-color:#ffffff;border-radius:4px;border:1px solid #e1e1e1;font-family:Segoe UI, Arial, Helvetica, sans-serif;">
//
//          <!-- 顶部蓝条 -->
//          <tr>
//            <td style="background-color:#0078D4;height:4px;border-radius:4px 4px 0 0;font-size:0;line-height:0;">
//              &nbsp;
//            </td>
//          </tr>
//
//          <!-- Logo 与 标题 -->
//          <tr>
//            <td style="padding:24px 32px 8px 32px;">
//              <table width="100%%" cellpadding="0" cellspacing="0" border="0">
//                <tr>
//                  <td align="left">
//                    <!-- Logo（可选）-->
//                    <!-- 没有 Logo 可以删掉整个 img -->
//                    <img src="【公司Logo链接】" alt="【公司名称】" style="height:32px;display:block;">
//                  </td>
//                </tr>
//                <tr>
//                  <td style="padding-top:16px;">
//                    <h2 style="margin:0;font-size:22px;color:#323130;font-weight:600;">
//                      验证您的电子邮件地址
//                    </h2>
//                  </td>
//                </tr>
//              </table>
//            </td>
//          </tr>
//
//          <!-- 正文内容 -->
//          <tr>
//            <td style="padding:8px 32px 0 32px;">
//              <p style="margin:0 0 12px 0;font-size:14px;line-height:1.6;color:#323130;">
//                您好：
//              </p>
//              <p style="margin:0 0 12px 0;font-size:14px;line-height:1.6;color:#323130;">
//                您正在使用 <strong>【公司名称】</strong> 进行安全操作。为了保护您的帐户，我们需要验证这是您本人。
//              </p>
//              <p style="margin:0 0 12px 0;font-size:14px;line-height:1.6;color:#323130;">
//                请在验证页面输入以下验证码：
//              </p>
//            </td>
//          </tr>
//
//          <!-- 验证码块 -->
//          <tr>
//            <td align="center" style="padding:24px 32px 16px 32px;">
//              <table cellpadding="0" cellspacing="0" border="0">
//                <tr>
//                  <td style="
//                    padding:14px 32px;
//                    border-radius:4px;
//                    border:1px solid #0078D4;
//                    background-color:#f3f9ff;
//                  ">
//                    <span style="
//                      font-size:26px;
//                      letter-spacing:6px;
//                      font-weight:600;
//                      color:#005A9E;
//                      font-family:Segoe UI, Arial, Helvetica, sans-serif;
//                    ">
//                      %s
//                    </span>
//                  </td>
//                </tr>
//              </table>
//            </td>
//          </tr>
//
//          <!-- 有效期与说明 -->
//          <tr>
//            <td style="padding:0 32px 16px 32px;">
//              <p style="margin:0 0 8px 0;font-size:13px;line-height:1.6;color:#605e5c;">
//                验证码有效期为 <strong>5 分钟</strong>，请勿转发或泄露给他人。
//              </p>
//              <p style="margin:0 0 8px 0;font-size:13px;line-height:1.6;color:#605e5c;">
//                如果这不是您的操作，可能是其他人误输入了您的邮箱地址，您可以忽略本邮件。
//              </p>
//            </td>
//          </tr>
//
//          <!-- 底部信息 -->
//          <tr>
//            <td style="padding:16px 32px 24px 32px;border-top:1px solid #e1e1e1;">
//              <p style="margin:0 0 4px 0;font-size:12px;line-height:1.6;color:#898989;">
//                此邮件由系统自动发送，请勿直接回复。
//              </p>
//              <p style="margin:0;font-size:12px;line-height:1.6;color:#898989;">
//                © 【公司名称】 保留所有权利
//              </p>
//            </td>
//          </tr>
//
//        </table>
//      </td>
//    </tr>
//  </table>
//</body>
//</html>
//`, code)
//}

func (s *emailServiceImpl) buildVerificationCodeHTML(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
 <meta charset="UTF-8">
 <title>来自爱莉希雅的温柔提醒</title>
</head>
<body style="margin:0; padding:0; background:#fff1f8; font-family:'HarmonyOS Sans','PingFang SC','Microsoft YaHei',sans-serif;">
 <table width="100%%" style="background:#fff1f8; padding:32px 0;" align="center">
   <tr>
     <td align="center">
       <table width="600" style="max-width:600px; background:#ffffff; border-radius:16px; box-shadow:0 8px 24px rgba(249,168,212,0.3); overflow:hidden;">

         <!-- 顶部封面 -->
         <tr>
           <td align="center" style="padding:24px; background:linear-gradient(90deg,#fbcfe8,#e9d5ff);">
             <img src="https://i.imgur.com/l90rM5L.png" alt="Elysia Emblem" width="64" style="display:block; margin-bottom:12px;">
             <h1 style="margin:0; font-size:22px; color:#d946ef;">一封来自爱莉希雅的小信笺</h1>
             <p style="margin:8px 0 0 0; color:#a855f7; font-size:14px;">
               「我喜欢一切美好的事物，包括...现在打开这封信的你♡」
             </p>
           </td>
         </tr>

         <!-- 虚线分隔 -->
         <tr>
           <td><hr style="border:none; border-top:1px dashed #fcd5f5; margin:0;"></td>
         </tr>

         <!-- 正文 -->
         <tr>
           <td style="padding:28px; color:#6b21a8; font-size:15px; line-height:1.8;">
             <p style="margin-top:0;">
               嗨，可爱的女孩子🎶～
             </p>
             <p>
               这是来自「真我」的刻印哦 ✨<br>
             </p>

             <!-- 验证码卡片 -->
             <div style="margin:24px auto; max-width:360px; background:#fef6ff; border:2px dashed #f0abfc; border-radius:12px; padding:20px; text-align:center;">
               <p style="margin:0; font-size:13px; color:#c026d3;">专属于你的魔法印记：</p>
               <div style="font-size:32px; font-weight:800; color:#ec4899; letter-spacing:8px; margin:12px 0;">
                 %s
               </div>
               <p style="margin:0; font-size:12px; color:#a855f7;">请在 {{expireMinutes}} 分钟内使用，魔法不会永远停留哦～</p>
             </div>

             <p>
               如果你并没有请求这个验证码……
               那也许，是有人对你太感兴趣了呢？<br />
               不用担心，爱莉希雅一直在这里，回应你的期待
             </p>
           </td>
         </tr>

         <!-- 结尾与签名 -->
         <tr>
           <td style="padding:0 28px 24px 28px; font-size:13px; color:#7e22ce;">
             <p>
               妖精的魔法要结束啦
             </p>
             <p style="margin-top:20px; text-align:right; font-style:italic;">
               如飞花般绚丽的少女 <br />
               <strong>爱莉希雅 ♡</strong>
             </p>
           </td>
         </tr>

         <!-- 底部说明 -->
         <tr>
           <td style="padding:20px; text-align:center; background:#faf5ff; font-size:11px; color:#a78bfa;">
             <p style="margin:0;">
               ✦ 这是妖精的魔法哦，无需回复 ✦<br />
               如果你不清楚这封邮件的来源，建议忽略并删除～
             </p>
           </td>
         </tr>

       </table>
     </td>
   </tr>
 </table>
</body>
</html>
`, code)
}
