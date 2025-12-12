package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"spider-go/internal/service"
	"spider-go/internal/shared"
	"strings"
	"time"
)

type Service interface {
	GetEvaluationInfo(ctx context.Context, uid int) (*[]EvaluationInfo, error)
	LoginAndCacheEvaluation(ctx context.Context, uid int, sid, spwd string) error
}

type evaluationService struct {
	userQuery       shared.UserQuery
	sessionService  service.SessionService
	evaluationCache cache.EvaluationCache
	// 教评系统相关 URL
	evaluationInfoURL string
	casRedirectURL    string // 教评系统 CAS 回调 URL（用于获取 ticket）
	doLoginURL        string // 教评系统 doLogin API
	timeout           time.Duration
	cacheExpire       time.Duration
}

func NewService(
	userQuery shared.UserQuery,
	sessionService service.SessionService,
	evaluationCache cache.EvaluationCache,
	evaluationInfoURL string,
	casRedirectURL string,
	doLoginURL string,
) Service {
	return &evaluationService{
		userQuery:         userQuery,
		sessionService:    sessionService,
		evaluationCache:   evaluationCache,
		evaluationInfoURL: evaluationInfoURL,
		casRedirectURL:    casRedirectURL,
		doLoginURL:        doLoginURL,
		timeout:           30 * time.Second,
		cacheExpire:       time.Hour,
	}
}

func (s *evaluationService) GetEvaluationInfo(ctx context.Context, uid int) (*[]EvaluationInfo, error) {
	user, err := s.userQuery.GetUserByUid(ctx, uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, "查询数据库错误")
	}

	accessToken, err := s.getAccessTokenOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, err
	}

	// 使用 accessToken 请求教评信息
	body, err := s.fetchWithAccessToken(ctx, "POST", s.evaluationInfoURL, accessToken, nil)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcRequestFailed, "发送教评请求失败")
	}
	defer body.Close()

	// TODO: 解析响应
	fmt.Println(body)

	return nil, errors.New("Not implemented")
}

// LoginAndCacheEvaluation 登录教评系统并缓存 accessToken
// 流程：复用 SessionService 登录获取带 TGC 的 client → 用 TGC 访问教评系统重定向链 → 获取 userToken → doLogin 获取 accessToken
func (s *evaluationService) LoginAndCacheEvaluation(ctx context.Context, uid int, sid, spwd string) error {
	// 1. 使用 SessionService 登录 CAS，获取带 TGC cookie 的 client
	client, err := s.sessionService.LoginAndGetClient(ctx, sid, spwd)
	if err != nil {
		return err
	}

	// 2. 用这个 client 访问教评系统的 CAS 重定向 URL
	// CAS 服务器会识别 TGC 并签发 ticket，然后重定向到教评系统
	return s.followRedirectsAndGetToken(ctx, client, s.casRedirectURL, uid)
}

// followRedirectsAndGetToken 跟随重定向链，获取 userToken 并调用 doLogin 获取 accessToken
func (s *evaluationService) followRedirectsAndGetToken(ctx context.Context, client *http.Client, startURL string, uid int) error {
	currentURL := startURL
	var userToken string

	// 跟随重定向，最多 10 次
	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", currentURL, nil)
		if err != nil {
			return common.NewAppError(common.CodeJwcLoginFailed, "构造请求失败")
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := client.Do(req)
		if err != nil {
			return common.NewAppError(common.CodeJwcLoginFailed, "请求失败")
		}

		// 检查是否是最终页面（包含 userToken 的重定向）
		location := resp.Header.Get("Location")

		// 检查当前 URL 或 Location 是否包含 userToken
		if strings.Contains(currentURL, "userToken=") {
			parsedURL, _ := url.Parse(currentURL)
			userToken = parsedURL.Query().Get("userToken")
		} else if strings.Contains(location, "userToken=") {
			parsedURL, _ := url.Parse(location)
			userToken = parsedURL.Query().Get("userToken")
		}

		resp.Body.Close()

		if userToken != "" {
			break
		}

		if resp.StatusCode/100 != 3 || location == "" {
			// 非重定向，尝试从响应中提取
			break
		}

		// 解析相对 URL
		base, _ := url.Parse(currentURL)
		next, _ := url.Parse(location)
		currentURL = base.ResolveReference(next).String()
	}

	if userToken == "" {
		return common.NewAppError(common.CodeJwcLoginFailed, "未能获取 userToken")
	}

	// 7. 调用 doLogin 获取 accessToken
	doLoginFullURL := fmt.Sprintf("%s?userToken=%s", s.doLoginURL, url.QueryEscape(userToken))

	req, err := http.NewRequest("POST", doLoginFullURL, nil)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "构造 doLogin 请求失败")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://jxzlpt.csuft.edu.cn")

	resp, err := client.Do(req)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "doLogin 请求失败")
	}
	defer resp.Body.Close()

	// 解析响应获取 accessToken
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return common.NewAppError(common.CodeJwcParseFailed, "读取 doLogin 响应失败")
	}

	var loginResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return common.NewAppError(common.CodeJwcParseFailed, fmt.Sprintf("解析 doLogin 响应失败: %v", err))
	}

	if loginResp.Data.AccessToken == "" {
		return common.NewAppError(common.CodeJwcLoginFailed, "未获取到 accessToken")
	}
	fmt.Println(loginResp.Data.AccessToken)
	// 8. 缓存 accessToken
	if err := s.evaluationCache.SetAccessToken(ctx, uid, loginResp.Data.AccessToken, s.cacheExpire); err != nil {
		return common.NewAppError(common.CodeCacheError, "缓存 accessToken 失败")
	}

	return nil
}

// getAccessTokenOrLogin 获取缓存的 accessToken 或登录
func (s *evaluationService) getAccessTokenOrLogin(ctx context.Context, uid int, sid, spwd string) (string, error) {
	// 先尝试从缓存中获取 accessToken
	accessToken, err := s.evaluationCache.GetAccessToken(ctx, uid)
	if err != nil {
		return "", common.NewAppError(common.CodeCacheError, "缓存错误")
	}

	if accessToken != "" {
		return accessToken, nil
	}

	// 如果没有缓存，则登录教评系统
	if err := s.LoginAndCacheEvaluation(ctx, uid, sid, spwd); err != nil {
		return "", err
	}

	// 重新获取 accessToken
	accessToken, err = s.evaluationCache.GetAccessToken(ctx, uid)
	if err != nil || accessToken == "" {
		return "", common.NewAppError(common.CodeJwcLoginFailed, "获取教评系统会话失败")
	}

	return accessToken, nil
}

// fetchWithAccessToken 使用 accessToken 发起请求
func (s *evaluationService) fetchWithAccessToken(ctx context.Context, method, targetURL string, accessToken string, formData url.Values) (io.ReadCloser, error) {
	var body io.Reader
	if formData != nil {
		body = strings.NewReader(formData.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer"+accessToken) // 关键：添加 accessToken 到请求头
	if formData != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	client := &http.Client{
		Timeout: s.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("响应状态码异常: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
