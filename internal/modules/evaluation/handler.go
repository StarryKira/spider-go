package evaluation

import "github.com/gin-gonic/gin"

type Handler struct {
	service Service
}

// NewHandler 创建考试处理器
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	exams := r.Group("/exams")
	{
		exams.GET("", h.GetExams) // 获取考试安排
	}
}
