package dto

// SendCaptchaRequest 发送验证码请求
type SendCaptchaRequest struct {
	Email string `json:"email" binding:"required,email"` // 邮箱地址
}

// VerifyCaptchaRequest 验证验证码请求
type VerifyCaptchaRequest struct {
	Email string `json:"email" binding:"required,email"` // 邮箱地址
	Code  string `json:"code" binding:"required,len=6"`  // 验证码（6位数字）
}
