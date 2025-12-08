package user

import (
	"context"
	"errors"
	"spider-go/internal/cache"
	"spider-go/internal/service"
	"spider-go/internal/shared"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("用户不存在或密码错误")
	ErrEmailAlreadyExists = errors.New("邮箱已被注册")
	ErrInvalidCaptcha     = errors.New("验证码错误")
	ErrEmptyParams        = errors.New("参数不能为空")
)

// Service 用户服务接口
type Service interface {
	// 认证相关
	Register(ctx context.Context, req *RegisterRequest) error
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error

	// 用户信息
	GetUserInfo(ctx context.Context, uid int) (*UserResponse, error)

	// 教务系统绑定
	BindJwc(ctx context.Context, uid int, req *BindJwcRequest) error
	CheckIsBind(ctx context.Context, uid int) (bool, error)
}

// userService 用户服务实现
type userService struct {
	repo           Repository
	sessionService service.SessionService
	captchaService service.CaptchaService
	captchaCache   cache.CaptchaCache
	dauService     service.DAUService
	jwtSecret      []byte
	jwtIssuer      string
	jwtExpire      time.Duration
}

// NewService 创建用户服务
func NewService(
	repo Repository,
	sessionService service.SessionService,
	captchaService service.CaptchaService,
	captchaCache cache.CaptchaCache,
	dauService service.DAUService,
	jwtSecret string,
	jwtIssuer string,
) Service {
	return &userService{
		repo:           repo,
		sessionService: sessionService,
		captchaService: captchaService,
		captchaCache:   captchaCache,
		dauService:     dauService,
		jwtSecret:      []byte(jwtSecret),
		jwtIssuer:      jwtIssuer,
		jwtExpire:      168 * time.Hour, // 7天
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *RegisterRequest) error {
	// 检查用户是否已存在
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return err
	}
	if existing != nil {
		return ErrEmailAlreadyExists
	}

	// 验证验证码
	if err := s.captchaService.VerifyEmailCaptcha(ctx, req.Email, req.Captcha); err != nil {
		return ErrInvalidCaptcha
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建用户
	user := &User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	return s.repo.Create(ctx, user)
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 查找用户
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 记录DAU
	_ = s.dauService.RecordUserActivity(ctx, user.Uid)

	// 生成JWT token
	claims := shared.UserClaims{
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
		return nil, err
	}

	return &LoginResponse{
		Token: tokenString,
		User:  user.ToResponse(),
	}, nil
}

// ResetPassword 重置密码
func (s *userService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	// 查找用户
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return ErrUserNotFound
	}

	// 验证验证码
	if err := s.captchaService.VerifyEmailCaptcha(ctx, req.Email, req.Captcha); err != nil {
		return ErrInvalidCaptcha
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	return s.repo.UpdatePassword(ctx, user.Uid, string(hashedPassword))
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(ctx context.Context, uid int) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// BindJwc 绑定教务系统
func (s *userService) BindJwc(ctx context.Context, uid int, req *BindJwcRequest) error {
	if req.Sid == "" || req.Spwd == "" {
		return ErrEmptyParams
	}

	// 尝试登录教务系统验证账号
	if err := s.sessionService.LoginAndCache(ctx, uid, req.Sid, req.Spwd); err != nil {
		return errors.New("请绑定i中南林APP账号")
	}

	// 更新数据库
	if err := s.repo.UpdateJwc(ctx, uid, req.Sid, req.Spwd); err != nil {
		return err
	}

	// 清除旧的会话缓存
	_ = s.sessionService.InvalidateSession(ctx, uid)

	return nil
}

// CheckIsBind 检查是否绑定教务系统
func (s *userService) CheckIsBind(ctx context.Context, uid int) (bool, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return false, err
	}

	return user.Sid != "" && user.Spwd != "", nil
}
