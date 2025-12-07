package dto

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// BroadcastEmailRequest 群发邮件请求
type BroadcastEmailRequest struct {
	Subject string `json:"subject" binding:"required"` // 邮件主题
	Body    string `json:"body" binding:"required"`    // 邮件内容（HTML格式）
}
