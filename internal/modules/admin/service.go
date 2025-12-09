package admin

import (
	"context"
	"errors"
	"spider-go/internal/service"
	"spider-go/internal/shared"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("管理员不存在或密码错误")
	ErrInvalidPassword    = errors.New("原密码错误")
)

// Service 管理员服务接口
type Service interface {
	// 认证相关
	Login(ctx context.Context, email, password string) (token string, admin *Admin, err error)
	GetAdminInfo(ctx context.Context, uid int) (*Admin, error)
	ChangePassword(ctx context.Context, uid int, oldPassword, newPassword string) error

	// 系统管理
	InitDefaultAdmin(ctx context.Context) error
	BroadcastEmail(ctx context.Context, subject, content string) (successCount, failCount, totalCount int, err error)
}

// adminService 管理员服务实现
type adminService struct {
	repo         Repository
	userQuery    shared.UserQuery
	emailService service.EmailService
	jwtSecret    []byte
	jwtIssuer    string
	jwtExpire    time.Duration
}

// NewService 创建管理员服务
func NewService(
	repo Repository,
	userQuery shared.UserQuery,
	emailService service.EmailService,
	jwtSecret string,
	jwtIssuer string,
) Service {
	return &adminService{
		repo:         repo,
		userQuery:    userQuery,
		emailService: emailService,
		jwtSecret:    []byte(jwtSecret),
		jwtIssuer:    jwtIssuer,
		jwtExpire:    24 * time.Hour, // 管理员token 24小时
	}
}

// Login 管理员登录
func (s *adminService) Login(ctx context.Context, email, password string) (string, *Admin, error) {
	// 查找管理员
	admin, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// 生成JWT token
	claims := shared.AdminClaims{
		Uid:     admin.Uid,
		Name:    admin.Name,
		IsAdmin: true,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpire)),
			Issuer:    s.jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, admin, nil
}

// GetAdminInfo 获取管理员信息
func (s *adminService) GetAdminInfo(ctx context.Context, uid int) (*Admin, error) {
	admin, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// ChangePassword 修改密码
func (s *adminService) ChangePassword(ctx context.Context, uid int, oldPassword, newPassword string) error {
	// 查找管理员
	admin, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return ErrAdminNotFound
	}

	// 验证原密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	return s.repo.UpdatePassword(ctx, uid, string(hashedPassword))
}

// InitDefaultAdmin 初始化默认管理员
func (s *adminService) InitDefaultAdmin(ctx context.Context) error {
	// 检查是否已存在管理员
	exists, err := s.repo.CheckExists(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil // 已存在管理员，不需要初始化
	}

	// 创建默认管理员
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &Admin{
		Email:     "admin@spider-go.com",
		Name:      "Haruka",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	return s.repo.Create(ctx, admin)
}

// BroadcastEmail 群发邮件给所有用户
func (s *adminService) BroadcastEmail(ctx context.Context, subject, content string) (int, int, int, error) {
	// 获取所有用户的邮箱
	emails, err := s.userQuery.GetAllUserEmails(ctx)
	if err != nil {
		return 0, 0, 0, errors.New("获取用户邮箱列表失败")
	}

	if len(emails) == 0 {
		return 0, 0, 0, errors.New("没有用户可以发送邮件")
	}

	// 群发邮件
	successCount := 0
	failCount := 0

	for _, email := range emails {
		err := s.emailService.SendEmail(ctx, email, subject, content)
		if err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	return successCount, failCount, len(emails), nil
}
