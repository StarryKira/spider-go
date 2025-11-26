package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"spider-go/internal/utils"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

// SessionService 会话服务接口
type SessionService interface {
	// LoginAndCache 登录教务系统并缓存会话
	LoginAndCache(ctx context.Context, uid int, username, password string) error
	// GetCachedCookies 获取缓存的 cookies
	GetCachedCookies(ctx context.Context, uid int) ([]*http.Cookie, error)
	// InvalidateSession 清除会话缓存
	InvalidateSession(ctx context.Context, uid int) error
}

// jwcSessionService 教务系统会话服务实现
type jwcSessionService struct {
	sessionCache cache.SessionCache
	jwcURL       string
	captchaURL   string
	timeout      time.Duration
	cacheExpire  time.Duration
}

// NewJwcSessionService 创建教务系统会话服务
func NewJwcSessionService(sessionCache cache.SessionCache, jwcURL string, captchaURL string) SessionService {
	return &jwcSessionService{
		sessionCache: sessionCache,
		jwcURL:       jwcURL,
		captchaURL:   captchaURL,
		timeout:      30 * time.Second,
		cacheExpire:  time.Hour,
	}
}

// LoginAndCache 登录教务系统并缓存会话
func (s *jwcSessionService) LoginAndCache(ctx context.Context, uid int, username, password string) error {
	// 创建 cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "创建会话失败")
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: s.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 1. 获取登录页面
	res, err := client.Get(s.jwcURL)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "连接教务系统失败")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("教务系统响应异常: %d", res.StatusCode))
	}

	// 2. 解析登录表单
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return common.NewAppError(common.CodeJwcParseFailed, "解析登录页面失败")
	}

	lt := doc.Find("input[name='lt']").AttrOr("value", "")
	dllt := doc.Find("input[name='dllt']").AttrOr("value", "")
	execution := doc.Find("input[name='execution']").AttrOr("value", "")
	eventID := doc.Find("input[name='_eventId']").AttrOr("value", "")
	salt := doc.Find("input[id='pwdDefaultEncryptSalt']").AttrOr("value", "")
	rmShown := doc.Find("input[name='rmShown']").AttrOr("value", "")

	if lt == "" || execution == "" || eventID == "" || salt == "" {
		return common.NewAppError(common.CodeJwcLoginFailed, "登录页缺少必要字段")
	}

	//构造请求体
	encryptedPwd := utils.JsCrypto(password, salt)

	form := url.Values{}
	form.Set("username", username)
	form.Set("password", encryptedPwd)
	form.Set("lt", lt)
	form.Set("dllt", dllt)
	form.Set("execution", execution)
	form.Set("_eventId", eventID)
	form.Set("rmShown", rmShown)

	//处理验证码
	//TODO 验证码处理
	isNeedCaptcha, err := client.Get(s.captchaURL + "username=" + username + "&pwdEncrypt2=pwdEncryptSalt" + "&_=" + strconv.FormatInt(time.Now().UnixMilli(), 10))
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "获取验证码失败")
	}
	isNeedCaptcha.Body.Close()
	length := isNeedCaptcha.ContentLength
	body := make([]byte, length)
	_, err = isNeedCaptcha.Body.Read(body)
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "获取是否需要验证码失败")
	}

	if string(body) == "true" {
		//验证码处理逻辑
	}

	//提交请求
	req, err := http.NewRequest("POST", s.jwcURL, strings.NewReader(form.Encode()))
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "构造登录请求失败")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", s.jwcURL)

	resp, err := client.Do(req)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "登录请求失败")
	}
	defer resp.Body.Close()

	// 4. 检查重定向
	if resp.StatusCode/100 != 3 {
		return common.NewAppError(common.CodeJwcLoginFailed, "登录失败，请检查账号密码")
	}

	loc, err := resp.Location()
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "获取重定向地址失败")
	}

	// 5. 跟随重定向获取完整 cookies
	finalResp, finalURL, err := s.followGET(client, loc.String(), 8)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "跟随重定向失败")
	}
	defer finalResp.Body.Close()

	// 6. 提取并缓存 cookies
	uFinal, _ := url.Parse(finalURL)
	base := &url.URL{Scheme: uFinal.Scheme, Host: uFinal.Host, Path: "/"}
	cookies := client.Jar.Cookies(base)

	if len(cookies) == 0 {
		if u, e := url.Parse(loc.String()); e == nil {
			cookies = client.Jar.Cookies(u)
		}
	}

	// 7. 存入缓存
	if err := s.sessionCache.SetCookies(ctx, uid, cookies, s.cacheExpire); err != nil {
		return common.NewAppError(common.CodeCacheError, "缓存会话失败")
	}

	return nil
}

// GetCachedCookies 获取缓存的 cookies
func (s *jwcSessionService) GetCachedCookies(ctx context.Context, uid int) ([]*http.Cookie, error) {
	return s.sessionCache.GetCookies(ctx, uid)
}

// InvalidateSession 清除会话缓存
func (s *jwcSessionService) InvalidateSession(ctx context.Context, uid int) error {
	return s.sessionCache.DeleteCookies(ctx, uid)
}

// followGET 手动跟随 GET 重定向
func (s *jwcSessionService) followGET(client *http.Client, start string, maxHops int) (*http.Response, string, error) {
	cur := start
	var lastReqURL *url.URL

	for i := 0; i < maxHops; i++ {
		req, _ := http.NewRequest("GET", cur, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := client.Do(req)
		if err != nil {
			return nil, cur, err
		}

		// 非 3xx：落地
		if resp.StatusCode/100 != 3 {
			return resp, cur, nil
		}

		// 3xx：读取 Location 手动跳转
		loc := resp.Header.Get("Location")
		_ = resp.Body.Close()

		if loc == "" {
			return nil, cur, errors.New("重定向缺少 Location")
		}

		// 解析相对跳转
		if lastReqURL == nil {
			lastReqURL, _ = url.Parse(cur)
		}

		locURL, err := url.Parse(loc)
		if err != nil {
			return nil, cur, fmt.Errorf("Location 无法解析: %v", err)
		}

		cur = lastReqURL.ResolveReference(locURL).String()
		lastReqURL = locURL
	}

	return nil, cur, errors.New("重定向层级过多")
}
