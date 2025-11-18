package dto

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Captcha  string `json:"captcha" binding:"required"`
}

// 绑定教务系统
type BindRequest struct {
	Sid  string `json:"sid" binding:"required"`
	Spwd string `json:"spwd" binding:"required"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha" binding:"required"`
}
