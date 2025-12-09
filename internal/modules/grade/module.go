package grade

import (
	"spider-go/internal/cache"
	"spider-go/internal/service"
	"spider-go/internal/shared"

	"github.com/gin-gonic/gin"
)

// Module 成绩模块
type Module struct {
	handler *Handler
	service Service
}

// NewModule 创建成绩模块
func NewModule(
	userQuery shared.UserQuery,
	sessionService service.SessionService,
	crawlerService service.CrawlerService,
	userDataCache cache.UserDataCache,
	gradeURL string,
	gradeLevelURL string,
) *Module {
	// 初始化各层：service -> handler
	svc := NewService(userQuery, sessionService, crawlerService, userDataCache, gradeURL, gradeLevelURL)
	handler := NewHandler(svc)

	return &Module{
		handler: handler,
		service: svc,
	}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	m.handler.RegisterRoutes(r)
}

// GetService 获取服务实例（用于跨模块调用）
func (m *Module) GetService() Service {
	return m.service
}
