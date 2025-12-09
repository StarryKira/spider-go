package user

import (
	"spider-go/internal/cache"
	"spider-go/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module 用户模块
type Module struct {
	handler        *Handler
	service        Service
	captchaService CaptchaService
}

// NewModule 创建用户模块
func NewModule(
	db *gorm.DB,
	sessionService service.SessionService,
	captchaCache cache.CaptchaCache,
	emailService service.EmailService,
	dauService service.DAUService,
	jwtSecret string,
	jwtIssuer string,
) *Module {
	repo := NewRepository(db)
	captchaService := NewCaptchaService(captchaCache, emailService)
	svc := NewService(repo, sessionService, captchaService, dauService, jwtSecret, jwtIssuer)
	handler := NewHandler(svc, captchaService)

	return &Module{
		handler:        handler,
		service:        svc,
		captchaService: captchaService,
	}
}

// RegisterRoutes 注册路由
// public: 公开路由组（无需认证）
// authenticated: 需要认证的路由组
func (m *Module) RegisterRoutes(public *gin.RouterGroup, authenticated *gin.RouterGroup) {
	m.handler.RegisterRoutes(public, authenticated)
}

// GetService 获取服务实例（用于跨模块调用）
func (m *Module) GetService() Service {
	return m.service
}
