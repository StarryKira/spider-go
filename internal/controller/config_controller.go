package controller

import (
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"spider-go/internal/dto"

	"github.com/gin-gonic/gin"
)

// ConfigController 系统配置控制器
type ConfigController struct {
	configCache cache.ConfigCache
}

// NewConfigController 创建系统配置控制器
func NewConfigController(configCache cache.ConfigCache) *ConfigController {
	return &ConfigController{configCache: configCache}
}

// GetCurrentTerm 获取当前学期（公开）
func (h *ConfigController) GetCurrentTerm(c *gin.Context) {
	term, err := h.configCache.GetCurrentTerm(c.Request.Context())
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, gin.H{
		"current_term": term,
	})
}

// SetCurrentTerm 设置当前学期（管理员）
func (h *ConfigController) SetCurrentTerm(c *gin.Context) {
	var req dto.SetCurrentTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, common.CodeInvalidParams, "参数错误")
		return
	}

	if err := h.configCache.SetCurrentTerm(c.Request.Context(), req.Term); err != nil {
		common.Error(c, common.CodeInvalidParams, err.Error())
		return
	}

	common.Success(c, gin.H{
		"message":      "设置成功",
		"current_term": req.Term,
	})
}
