package controller

import "spider-go/internal/handler/service"

// 获取基本信息
type NormalController struct {
	normSvc *service.NormalService
}

func NewNormalController(normSvc *service.NormalService) *NormalController {
	return &NormalController{normSvc: normSvc}
}
