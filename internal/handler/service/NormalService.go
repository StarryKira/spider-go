package service

import (
	"spider-go/internal/repository"
)

type NormalService struct {
	uRepo repository.UserRepository
	nRepo repository.NoticeRepository
}

func NewNormalService(uRepo repository.UserRepository) *NormalService {
	return &NormalService{uRepo: uRepo}
}

type BasicInfo struct {
}

// 主页通知
type Notice struct {
	Message string `json:"message"`
	IsShow  bool   `json:"is_show"`
	Nid     int    `json:"nid"` //notice id
}

func (s *NormalService) GetBasicInfo(uid int) {

}
