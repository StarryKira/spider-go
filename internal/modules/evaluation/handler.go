package evaluation

import (
	"spider-go/internal/common"

	"github.com/gin-gonic/gin"
)

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
	evaluation := r.Group("/evaluation")
	evaluation.GET("/info", h.GetEvaluationInfo)
}

func (h *Handler) GetEvaluationInfo(c *gin.Context) {
	uidValue, ok := c.Get("uid")
	if !ok {
		common.Error(c, common.CodeUnauthorized, "未授权，请登录")
		return
	}

	uid := uidValue.(int)

	data, err := h.service.GetEvaluationInfo(c, uid)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}
	common.Success(c, data)
}
