package service

import (
	"context"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"spider-go/internal/model"
	"spider-go/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务接口
type UserService interface {
	// UserLogin 用户登录
	UserLogin(ctx context.Context, email, password string) (string, error)
	// Register 用户注册
	Register(ctx context.Context, name, email, captcha, password string) error
	// Bind 绑定教务系统账号
	Bind(ctx context.Context, uid int, sid, spwd string) error
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, uid int) (*model.User, error)
	// ResetPassword 使用验证码修改密码
	ResetPassword(ctx context.Context, email string, sid, password string) error
	// CheckIsBind 检查用户是否绑定教务系统账号
	CheckIsBind(ctx context.Context, uid int) (bool, error)
}

// userServiceImpl 用户服务实现
type userServiceImpl struct {
	userRepo       repository.UserRepository
	sessionService SessionService
	captchaService CaptchaService
	captchaCache   cache.CaptchaCache
	dauService     DAUService
	jwtSecret      []byte
	jwtIssuer      string
	jwtExpire      time.Duration
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo repository.UserRepository,
	sessionService SessionService,
	captchaService CaptchaService,
	captchaCache cache.CaptchaCache,
	dauService DAUService,
	jwtSecret string,
	jwtIssuer string,
) UserService {
	return &userServiceImpl{
		userRepo:       userRepo,
		sessionService: sessionService,
		captchaService: captchaService,
		captchaCache:   captchaCache,
		dauService:     dauService,
		jwtSecret:      []byte(jwtSecret),
		jwtIssuer:      jwtIssuer,
		jwtExpire:      168 * time.Hour, // 7天
	}
}

// Claims JWT Claims
type Claims struct {
	Uid  int    `json:"user_id"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

// UserLogin 用户登录
func (s *userServiceImpl) UserLogin(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", common.NewAppError(common.CodeUserNotFound, "用户不存在或密码错误")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", common.NewAppError(common.CodeInvalidPassword, "用户不存在或密码错误")
	}

	// 记录用户活跃（日活统计）
	_ = s.dauService.RecordUserActivity(ctx, user.Uid)

	// 生成 JWT token
	claims := Claims{
		Uid:  user.Uid,
		Name: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpire)),
			Issuer:    s.jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", common.NewAppError(common.CodeInternalError, "生成令牌失败")
	}

	return tokenString, nil
}

// Register 用户注册
func (s *userServiceImpl) Register(ctx context.Context, name, email, captcha, password string) error {
	// 检查用户是否已存在
	_, err := s.userRepo.GetUserByEmail(email)
	if err == nil {
		return common.NewAppError(common.CodeUserAlreadyExists, "邮箱已被注册")
	}
	//验证验证码
	err = s.captchaService.VerifyEmailCaptcha(ctx, email, captcha)
	if err != nil {
		return common.NewAppError(common.CodeCaptchaInvalid, err.Error())
	}
	// 加密密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "密码加密失败")
	}

	// 创建用户
	u := &model.User{
		Name:      name,
		Email:     email,
		Password:  string(passwordHash),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(u); err != nil {
		return common.NewAppError(common.CodeInternalError, "创建用户失败")
	}

	return nil
}

// Bind 绑定教务系统账号
func (s *userServiceImpl) Bind(ctx context.Context, uid int, sid, spwd string) error {
	if sid == "" || spwd == "" {
		return common.NewAppError(common.CodeInvalidParams, "学号和密码不能为空")
	}
	if s.sessionService.LoginAndCache(ctx, uid, sid, spwd) != nil {
		return common.NewAppError(common.CodeJwcInvalidParams, "请绑定i中南林APP账号")
	}
	// 更新数据库
	if err := s.userRepo.UpdateJwc(uid, sid, spwd); err != nil {
		return common.NewAppError(common.CodeInternalError, "更新绑定信息失败")
	}

	// 清除旧的会话缓存
	if err := s.sessionService.InvalidateSession(ctx, uid); err != nil {
		// 即使清除失败也不影响绑定流程，只记录错误
		// 可以在这里添加日志
	}

	return nil
}

// GetUserInfo 获取用户信息
func (s *userServiceImpl) GetUserInfo(ctx context.Context, uid int) (*model.User, error) {
	user, err := s.userRepo.GetUserByUid(uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}

	// 清除敏感信息
	user.Password = ""
	user.Spwd = ""

	return user, nil
}

func (s *userServiceImpl) ResetPassword(ctx context.Context, email string, password string, captcha string) error {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}
	err = s.captchaService.VerifyEmailCaptcha(ctx, email, captcha)
	if err != nil {
		return common.NewAppError(common.CodeInternalError, err.Error())
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "密码加密失败")
	}
	if err := s.userRepo.UpdatePassword(user.Uid, string(passwordHash)); err != nil {
		return common.NewAppError(common.CodeInternalError, "修改密码失败")
	}
	return nil
}

// CheckIsBind 检查用户是否绑定教务系统账号
func (s *userServiceImpl) CheckIsBind(ctx context.Context, uid int) (bool, error) {
	user, err := s.userRepo.GetUserByUid(uid)
	if err != nil {
		return false, common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}

	// 判断是否绑定：Sid 和 Spwd 都不为空
	isBind := user.Sid != "" && user.Spwd != ""
	return isBind, nil
}
