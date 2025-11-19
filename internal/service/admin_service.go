package service

import (
	"context"
	"spider-go/internal/common"
	"spider-go/internal/model"
	"spider-go/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminService 管理员服务接口
type AdminService interface {
	// AdminLogin 管理员登录
	AdminLogin(ctx context.Context, email, password string) (string, error)
	// GetAdminInfo 获取管理员信息
	GetAdminInfo(ctx context.Context, uid int) (*model.Administrator, error)
	// InitDefaultAdmin 初始化默认管理员（如果不存在）
	InitDefaultAdmin() error
}

// adminServiceImpl 管理员服务实现
type adminServiceImpl struct {
	adminRepo repository.AdminRepository
	jwtSecret []byte
	jwtIssuer string
	jwtExpire time.Duration
}

// NewAdminService 创建管理员服务
func NewAdminService(
	adminRepo repository.AdminRepository,
	jwtSecret string,
	jwtIssuer string,
) AdminService {
	return &adminServiceImpl{
		adminRepo: adminRepo,
		jwtSecret: []byte(jwtSecret),
		jwtIssuer: jwtIssuer,
		jwtExpire: 24 * time.Hour, // 管理员 token 24小时
	}
}

// AdminClaims 管理员 JWT Claims
type AdminClaims struct {
	Uid     int    `json:"uid"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"` // 标识为管理员
	jwt.RegisteredClaims
}

// AdminLogin 管理员登录
func (s *adminServiceImpl) AdminLogin(ctx context.Context, email, password string) (string, error) {
	admin, err := s.adminRepo.GetAdminByEmail(email)
	if err != nil {
		return "", common.NewAppError(common.CodeUserNotFound, "管理员不存在或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", common.NewAppError(common.CodeInvalidPassword, "管理员不存在或密码错误")
	}

	// 生成 JWT token
	claims := AdminClaims{
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
		return "", common.NewAppError(common.CodeInternalError, "生成令牌失败")
	}

	return tokenString, nil
}

// GetAdminInfo 获取管理员信息
func (s *adminServiceImpl) GetAdminInfo(ctx context.Context, uid int) (*model.Administrator, error) {
	admin, err := s.adminRepo.GetAdminByUid(uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeUserNotFound, "管理员不存在")
	}

	// 清除敏感信息
	admin.Password = ""

	return admin, nil
}

// InitDefaultAdmin 初始化默认管理员（email: admin@spider-go.com, password: 123456）
func (s *adminServiceImpl) InitDefaultAdmin() error {
	// 检查是否已存在管理员
	exists, err := s.adminRepo.CheckAdminExists()
	if err != nil {
		return err
	}

	if exists {
		return nil // 已存在管理员，不需要初始化
	}

	// 创建默认管理员
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.Administrator{
		Email:     "admin@spider-go.com",
		Name:      "管理员",
		Password:  string(passwordHash),
		CreatedAt: time.Now(),
	}

	return s.adminRepo.CreateAdmin(admin)
}
