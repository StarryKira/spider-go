package evaluation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"spider-go/internal/service"
	"spider-go/internal/shared"
)

type Service interface {
	GetEvaluationInfo(ctx context.Context, uid int) (*[]EvaluationInfo, error)
	LoginAndCacheEvaluation(ctx context.Context, uid int, sid, spwd string) error
}

type evaluationService struct {
	userQuery         shared.UserQuery
	sessionService    service.SessionService
	crawlerService    service.CrawlerService
	evaluationCache   cache.EvaluationCache
	evaluationInfoURL string
	loginURL          string
	redirectURL       string
}

func NewService(
	userQuery shared.UserQuery,
	sessionService service.SessionService,
	crawlerService service.CrawlerService,
	evaluationCache cache.EvaluationCache,
	evaluationInfoURL string,
	loginURL string,
	redirectURL string,
) Service {
	return &evaluationService{
		userQuery:         userQuery,
		sessionService:    sessionService,
		crawlerService:    crawlerService,
		evaluationCache:   evaluationCache,
		evaluationInfoURL: evaluationInfoURL,
		loginURL:          loginURL,
		//教评系统重定向链接
		redirectURL: redirectURL,
	}
}

func (s *evaluationService) GetEvaluationInfo(ctx context.Context, uid int) (*[]EvaluationInfo, error) {

	user, err := s.userQuery.GetUserByUid(ctx, uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, "查询数据库错误")
	}
	cookies, err := s.getCookiesOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "获取cookie失败")
	}
	formData := url.Values{}
	body, err := s.crawlerService.FetchWithCookies(ctx, "POST", s.evaluationInfoURL, cookies, formData)

	if err != nil {
		return nil, common.NewAppError(common.CodeJwcRequestFailed, "发送教评请求失败")
	}

	defer body.Close()

	fmt.Println(body)

	return nil, errors.New("Not implemented")
}

// LoginAndCacheEvaluation 登录教评系统并缓存会话
func (s *evaluationService) LoginAndCacheEvaluation(ctx context.Context, uid int, sid, spwd string) error {
	// 使用 SessionService 的通用登录方法，传入教评系统的 URL 和缓存
	return s.sessionService.LoginAndCacheWithConfig(ctx, uid, sid, spwd, s.loginURL, s.redirectURL, s.evaluationCache)
}

// getCookiesOrLogin 获取缓存的教评系统 cookies 或登录
// 这个方法在实现具体功能时会被用到
func (s *evaluationService) getCookiesOrLogin(ctx context.Context, uid int, sid, spwd string) ([]*http.Cookie, error) {
	// 先尝试从教评缓存中获取 cookies
	cookies, err := s.evaluationCache.GetCookies(ctx, uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeCacheError, "缓存错误")
	}

	if len(cookies) > 0 {
		return cookies, nil
	}

	// 如果没有缓存，则登录教评系统
	if err := s.LoginAndCacheEvaluation(ctx, uid, sid, spwd); err != nil {
		return nil, err
	}

	// 重新获取 cookies
	cookies, err = s.evaluationCache.GetCookies(ctx, uid)
	if err != nil || len(cookies) == 0 {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "获取教评系统会话失败")
	}

	return cookies, nil
}
