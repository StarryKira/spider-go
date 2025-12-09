package user

import "time"

// User 用户模型
type User struct {
	Uid       int       `gorm:"primary_key;AUTO_INCREMENT" json:"uid"`
	Email     string    `gorm:"unique" json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"`   // 不序列化
	Sid       string    `json:"sid"` // 学号
	Spwd      string    `json:"-"`   // 教务系统密码（不序列化）
	CreatedAt time.Time `json:"created_at"`
	Avatar    string    `json:"avatar"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Captcha  string `json:"captcha" binding:"required"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// BindJwcRequest 绑定教务系统请求
type BindJwcRequest struct {
	Sid  string `json:"sid" binding:"required"`  // 学号
	Spwd string `json:"spwd" binding:"required"` // 教务系统密码
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Captcha  string `json:"captcha" binding:"required"`
}

// UserResponse 用户响应（不包含敏感信息）
type UserResponse struct {
	Uid       int       `json:"uid"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Sid       string    `json:"sid"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	IsBind    bool      `json:"is_bind"` // 是否绑定教务系统
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		Uid:       u.Uid,
		Email:     u.Email,
		Name:      u.Name,
		Sid:       u.Sid,
		Avatar:    u.Avatar,
		CreatedAt: u.CreatedAt,
		IsBind:    u.Sid != "" && u.Spwd != "",
	}
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}
