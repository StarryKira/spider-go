package service

import (
	"context"
	"spider-go/internal/common"
	"spider-go/internal/model"
	"spider-go/internal/repository"
	"time"
)

// NoticeService 通知服务接口
type NoticeService interface {
	// CreateNotice 创建通知（管理员）
	CreateNotice(ctx context.Context, content, noticeType string, isShow, isTop, isHtml bool) error
	// UpdateNotice 更新通知（管理员）
	UpdateNotice(ctx context.Context, nid int, content, noticeType string, isShow, isTop, isHtml bool) error
	// DeleteNotice 删除通知（管理员）
	DeleteNotice(ctx context.Context, nid int) error
	// GetAllNotices 获取所有通知（管理员）
	GetAllNotices(ctx context.Context) ([]model.Notice, error)
	// GetVisibleNotices 获取可见通知（普通用户）
	GetVisibleNotices(ctx context.Context) ([]model.Notice, error)
	// GetNoticeByNid 根据 NID 获取通知
	GetNoticeByNid(ctx context.Context, nid int) (*model.Notice, error)
}

// noticeServiceImpl 通知服务实现
type noticeServiceImpl struct {
	noticeRepo repository.NoticeRepository
}

// NewNoticeService 创建通知服务
func NewNoticeService(noticeRepo repository.NoticeRepository) NoticeService {
	return &noticeServiceImpl{
		noticeRepo: noticeRepo,
	}
}

// CreateNotice 创建通知
func (s *noticeServiceImpl) CreateNotice(ctx context.Context, content, noticeType string, isShow, isTop, isHtml bool) error {
	if content == "" {
		return common.NewAppError(common.CodeInvalidParams, "通知内容不能为空")
	}

	notice := &model.Notice{
		Content:    content,
		NoticeType: noticeType,
		IsShow:     isShow,
		IsTop:      isTop,
		IsHtml:     isHtml,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := s.noticeRepo.CreateNotice(notice); err != nil {
		return common.NewAppError(common.CodeInternalError, "创建通知失败")
	}

	return nil
}

// UpdateNotice 更新通知
func (s *noticeServiceImpl) UpdateNotice(ctx context.Context, nid int, content, noticeType string, isShow, isTop, isHtml bool) error {
	if content == "" {
		return common.NewAppError(common.CodeInvalidParams, "通知内容不能为空")
	}

	// 获取现有通知
	notice, err := s.noticeRepo.GetNoticeByNid(nid)
	if err != nil {
		return common.NewAppError(common.CodeNotFound, "通知不存在")
	}

	// 更新字段
	notice.Content = content
	notice.NoticeType = noticeType
	notice.IsShow = isShow
	notice.IsTop = isTop
	notice.IsHtml = isHtml
	notice.UpdateTime = time.Now()

	if err := s.noticeRepo.UpdateNotice(notice); err != nil {
		return common.NewAppError(common.CodeInternalError, "更新通知失败")
	}

	return nil
}

// DeleteNotice 删除通知
func (s *noticeServiceImpl) DeleteNotice(ctx context.Context, nid int) error {
	if err := s.noticeRepo.DeleteNotice(nid); err != nil {
		return common.NewAppError(common.CodeInternalError, "删除通知失败")
	}
	return nil
}

// GetAllNotices 获取所有通知（管理员）
func (s *noticeServiceImpl) GetAllNotices(ctx context.Context) ([]model.Notice, error) {
	notices, err := s.noticeRepo.GetAllNotices()
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, "获取通知列表失败")
	}
	return notices, nil
}

// GetVisibleNotices 获取可见通知（普通用户）
func (s *noticeServiceImpl) GetVisibleNotices(ctx context.Context) ([]model.Notice, error) {
	notices, err := s.noticeRepo.GetVisibleNotices()
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, "获取通知列表失败")
	}
	return notices, nil
}

// GetNoticeByNid 根据 NID 获取通知
func (s *noticeServiceImpl) GetNoticeByNid(ctx context.Context, nid int) (*model.Notice, error) {
	notice, err := s.noticeRepo.GetNoticeByNid(nid)
	if err != nil {
		return nil, common.NewAppError(common.CodeNotFound, "通知不存在")
	}
	return notice, nil
}
