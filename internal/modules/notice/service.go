package notice

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmptyContent = errors.New("通知内容不能为空")
)

// Service 通知服务接口
type Service interface {
	Create(ctx context.Context, req *CreateNoticeRequest) (*Notice, error)
	Update(ctx context.Context, nid int, req *UpdateNoticeRequest) (*Notice, error)
	Delete(ctx context.Context, nid int) error
	GetByID(ctx context.Context, nid int) (*Notice, error)
	GetAll(ctx context.Context) ([]*Notice, error)
	GetVisible(ctx context.Context) ([]*Notice, error)
}

// service 通知服务实现
type service struct {
	repo Repository
}

// NewService 创建通知服务
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// Create 创建通知
func (s *service) Create(ctx context.Context, req *CreateNoticeRequest) (*Notice, error) {
	if req.Content == "" {
		return nil, ErrEmptyContent
	}

	notice := &Notice{
		Content:    req.Content,
		NoticeType: req.NoticeType,
		IsShow:     req.IsShow,
		IsTop:      req.IsTop,
		IsHtml:     req.IsHtml,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := s.repo.Create(ctx, notice); err != nil {
		return nil, err
	}

	return notice, nil
}

// Update 更新通知
func (s *service) Update(ctx context.Context, nid int, req *UpdateNoticeRequest) (*Notice, error) {
	if req.Content == "" {
		return nil, ErrEmptyContent
	}

	// 获取现有通知
	notice, err := s.repo.FindByID(ctx, nid)
	if err != nil {
		return nil, err
	}

	// 更新字段
	notice.Content = req.Content
	notice.NoticeType = req.NoticeType
	notice.IsShow = req.IsShow
	notice.IsTop = req.IsTop
	notice.IsHtml = req.IsHtml
	notice.UpdateTime = time.Now()

	if err := s.repo.Update(ctx, notice); err != nil {
		return nil, err
	}

	return notice, nil
}

// Delete 删除通知
func (s *service) Delete(ctx context.Context, nid int) error {
	return s.repo.Delete(ctx, nid)
}

// GetByID 根据ID获取通知
func (s *service) GetByID(ctx context.Context, nid int) (*Notice, error) {
	return s.repo.FindByID(ctx, nid)
}

// GetAll 获取所有通知（管理员）
func (s *service) GetAll(ctx context.Context) ([]*Notice, error) {
	return s.repo.FindAll(ctx)
}

// GetVisible 获取可见通知（普通用户）
func (s *service) GetVisible(ctx context.Context) ([]*Notice, error) {
	return s.repo.FindVisible(ctx)
}
