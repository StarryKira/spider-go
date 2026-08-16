package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

// CookieCache 通用的 Cookie 缓存接口，用于统一 SessionCache 和 EvaluationCache
type CookieCache interface {
	GetCookies(ctx context.Context, uid int) ([]*http.Cookie, error)
	SetCookies(ctx context.Context, uid int, cookies []*http.Cookie, expiration time.Duration) error
	DeleteCookies(ctx context.Context, uid int) error
	HasCookies(ctx context.Context, uid int) (bool, error)
}

// SessionService 会话服务接口
type SessionService interface {
	// LoginAndCache 登录教务系统并缓存会话
	LoginAndCache(ctx context.Context, uid int, username, password string) error
	// GetCachedCookies 获取缓存的 cookies
	GetCachedCookies(ctx context.Context, uid int) ([]*http.Cookie, error)
	// InvalidateSession 清除会话缓存
	InvalidateSession(ctx context.Context, uid int) error
	// LoginAndCacheWithConfig 通用登录方法，支持自定义 URL 和缓存
	LoginAndCacheWithConfig(ctx context.Context, uid int, username, password string, loginURL, redirectURL string, cookieCache CookieCache) error
	// LoginAndGetClient 登录 CAS 并返回带 TGC cookie 的 client，供其他系统复用
	LoginAndGetClient(ctx context.Context, username, password string) (*http.Client, error)
	//LoginCheck 模拟一次登录以校验绑定的是否正确
	LoginCheck(ctx context.Context, uid int, username, password string) error
	// CompletePhoneMFA 提交安全手机短信验证码并完成待处理的 CAS 登录
	CompletePhoneMFA(ctx context.Context, uid int, code string) (*cache.MFAChallenge, error)
	// ResendPhoneMFA 重新发送安全手机短信验证码
	ResendPhoneMFA(ctx context.Context, uid int) error
}

// jwcSessionService 教务系统会话服务实现
type jwcSessionService struct {
	sessionCache    cache.SessionCache
	mfaCache        cache.MFAChallengeCache
	rsaKeyService   RSAKeyService
	mode            string // 登录模式：campus 或 webvpn
	loginURL        string
	webVPNTokenURL  string
	redirectURL     string
	mfaDetectURL    string
	captchaURL      string
	captchaImageURL string
	timeout         time.Duration
	cacheExpire     time.Duration
	tgcExpire       time.Duration // TGC cookie 的过期时间
}

// NewJwcSessionService 创建教务系统会话服务
func NewJwcSessionService(
	sessionCache cache.SessionCache,
	mfaCache cache.MFAChallengeCache,
	rsaKeyService RSAKeyService,
	mode string,
	loginURL string,
	webVPNTokenURL string,
	redirectURL string,
	mfaDetectURL string,
	captchaURL string,
	captchaImageURL string,
) SessionService {
	return &jwcSessionService{
		sessionCache:    sessionCache,
		mfaCache:        mfaCache,
		rsaKeyService:   rsaKeyService,
		mode:            mode,
		loginURL:        loginURL,
		webVPNTokenURL:  webVPNTokenURL,
		redirectURL:     redirectURL,
		mfaDetectURL:    mfaDetectURL,
		captchaURL:      captchaURL,
		captchaImageURL: captchaImageURL,
		timeout:         30 * time.Second,
		cacheExpire:     time.Hour,
	}
}

// LoginAndCache 登录教务系统并缓存会话（带重试机制，根据模式选择登录方法）
func (s *jwcSessionService) LoginAndCache(ctx context.Context, uid int, username, password string) error {
	var err error
	// 重试 1 次
	for i := 0; i < 1; i++ {
		// 根据模式选择登录函数
		if s.mode == "webvpn" {
			err = s.loginAndCacheOnceByWebVPN(ctx, uid, username, password)
		} else {
			err = s.loginAndCacheOnce(ctx, uid, username, password)
		}

		if err == nil {
			return nil
		}

		// 重试间隔
		time.Sleep(time.Second * time.Duration(i+1))
	}

	// 保留原始错误码，不要强制改为 CodeJwcLoginFailed
	// 如果是 AppError，保留原始错误码；否则包装为 CodeJwcLoginFailed
	if appErr, ok := err.(*common.AppError); ok {
		// 保留原始错误码，只修改错误消息提示重试次数
		return common.NewAppError(appErr.Code, fmt.Sprintf("%s (已重试)", appErr.Message))
	}

	// 非 AppError 的错误，包装为 CodeJwcLoginFailed
	return common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("登录失败，请重试: %v", err))
}

// loginAndCacheOnce 单次登录逻辑
func (s *jwcSessionService) loginAndCacheOnce(ctx context.Context, uid int, username, password string) error {
	return s.LoginAndCacheWithConfig(ctx, uid, username, password, s.loginURL, s.redirectURL, s.sessionCache)
}

func (s *jwcSessionService) followGET(client *http.Client, start string, maxHops int) (*http.Response, string, error) {
	cur := start
	var lastReqURL *url.URL

	for i := 0; i < maxHops; i++ {
		req, _ := http.NewRequest("GET", cur, nil)
		req.Header.Set("User-Agent", chromeUserAgent)

		resp, err := client.Do(req)
		if err != nil {
			// 网络错误，不是认证问题
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
			// 重定向配置问题，不是认证问题
			return nil, cur, common.NewAppError(common.CodeJwcRequestFailed, "重定向缺少 Location")
		}

		// 解析相对跳转
		if lastReqURL == nil {
			lastReqURL, _ = url.Parse(cur)
		}

		locURL, err := url.Parse(loc)
		if err != nil {
			return nil, cur, common.NewAppError(common.CodeJwcParseFailed, "location 无法解析")
		}

		cur = lastReqURL.ResolveReference(locURL).String()
		lastReqURL = locURL
	}

	// 重定向层级过多，可能是配置问题或系统异常，不是认证问题
	return nil, cur, common.NewAppError(common.CodeJwcRequestFailed, "重定向层级过多")
}

func (s *jwcSessionService) completeJwglTicketLogin(client *http.Client, resp *http.Response, current string) (*http.Response, string, error) {
	if resp == nil {
		return resp, current, nil
	}
	u, err := url.Parse(current)
	if err != nil || !strings.Contains(u.Path, "LoginToXk") {
		return resp, current, nil
	}
	ticket := u.Query().Get("ticket")
	if ticket == "" {
		return resp, current, nil
	}

	_ = resp.Body.Close()
	form := url.Values{
		"method": {"jwxt"},
		"ticket": {ticket},
	}
	postURL := u.Scheme + "://" + u.Host + u.Path
	req, err := http.NewRequest(http.MethodPost, postURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, current, err
	}
	req.Header.Set("User-Agent", chromeUserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", current)
	postResp, err := client.Do(req)
	if err != nil {
		return nil, current, err
	}
	if postResp.StatusCode/100 == 3 {
		_ = postResp.Body.Close()
		loc := postResp.Header.Get("Location")
		if loc != "" {
			next := u.ResolveReference(mustParseURL(loc)).String()
			return s.followGET(client, next, 8)
		}
	}

	mainURL := u.Scheme + "://" + u.Host + "/jsxsd/framework/xsMain.jsp"
	_ = postResp.Body.Close()
	return s.followGET(client, mainURL, 6)
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		return &url.URL{Path: raw}
	}
	return u
}

// GetCachedCookies 获取缓存的 cookies
func (s *jwcSessionService) GetCachedCookies(ctx context.Context, uid int) ([]*http.Cookie, error) {
	return s.sessionCache.GetCookies(ctx, uid)
}

// InvalidateSession 清除会话缓存
func (s *jwcSessionService) InvalidateSession(ctx context.Context, uid int) error {
	return s.sessionCache.DeleteCookies(ctx, uid)
}

// encryptPassword 使用 RSA 公钥加密密码
func (s *jwcSessionService) encryptPassword(password string) (string, error) {
	// 从 RSA Key Service 获取公钥
	publicKey := s.rsaKeyService.GetPublicKey()
	if publicKey == "" {
		return "", common.NewAppError(common.CodeInternalError, "RSA 公钥未初始化")
	}

	// 1. 解析 PEM 公钥
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return "", common.NewAppError(common.CodeJwcLoginFailed, "RSA 公钥格式无效")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("解析 RSA 公钥失败: %v", err))
	}

	pub := pubInterface.(*rsa.PublicKey)

	// 2. 执行 RSA 加密（PKCS1v15 —— 和 JSEncrypt 完全一致）
	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(password))
	if err != nil {
		return "", common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("RSA 加密失败: %v", err))
	}

	// 3. 输出 Base64（JSEncrypt 默认也是 Base64）
	return "__RSA__" + base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// chromeUserAgent 与 CAS 页面 FingerprintJS 采集时的桌面 Chrome 一致。
const chromeUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

func setChromeHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", chromeUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
}

func originFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// GenerateRandomFingerPrintHash 生成 FingerprintJS v3 格式的 visitorId（32 位小写 hex）。
func (s *jwcSessionService) GenerateRandomFingerPrintHash() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// chromeVisitorID 模拟浏览器 localStorage.fpVisitorId：同一账号保持同一 FingerprintJS visitorId。
func (s *jwcSessionService) chromeVisitorID(username string) string {
	sum := sha256.Sum256([]byte("fpjs3:chrome:" + username))
	return hex.EncodeToString(sum[:16])
}

func (s *jwcSessionService) casLoginForm(username, encryptedPwd, execution, fpVisitorID, mfaState string) url.Values {
	return url.Values{
		"username":    {username},
		"password":    {encryptedPwd},
		"captcha":     {""},
		"execution":   {execution},
		"fpVisitorId": {fpVisitorID},
		"rememberMe":  {"on"},
		"_eventId":    {"submit"},
		"failN":       {"0"},
		"currentMenu": {"1"},
		"mfaState":    {mfaState},
		"geolocation": {""},
		"trustAgent":  {""},
	}
}

// LoginAndCacheWithConfig 通用登录方法，支持自定义 URL 和缓存
func (s *jwcSessionService) LoginAndCacheWithConfig(ctx context.Context, uid int, username, password string, loginURL, redirectURL string, cookieCache CookieCache) error {
	// 创建 cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "创建会话失败")
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: s.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止自动跳转（CAS 必须手动）
		},
	}

	if s.mode == "webvpn" {
		startedURL, startErr := s.startWebVPNAuth(ctx, client)
		if startErr != nil {
			return startErr
		}
		if startedURL != "" {
			loginURL = startedURL
		}
	}

	// 1. 请求登录页获取 execution
	loginPageReq, err := http.NewRequestWithContext(ctx, http.MethodGet, loginURL, nil)
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "构造登录页请求失败")
	}
	setChromeHeaders(loginPageReq, "")
	res, err := client.Do(loginPageReq)
	if err != nil {
		// 检查是否是超时错误
		if isTimeoutError(err) {
			return common.NewAppError(common.CodeJwcLoginTimeout, "教务系统连接超时，请稍后重试")
		}
		return common.NewAppError(common.CodeJwcLoginFailed, "连接系统失败")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("响应异常: %d", res.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return common.NewAppError(common.CodeJwcParseFailed, "解析登录页面失败")
	}

	execution := doc.Find("input[name='execution']").AttrOr("value", "")
	if execution == "" {
		return common.NewAppError(common.CodeJwcLoginFailed, "找不到 execution")
	}

	// 2. 密码加密
	encryptedPwd, err := s.encryptPassword(password)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("密码加密失败: %v", err))
	}

	fpVisitorId := s.chromeVisitorID(username)

	// 2.5 MFA 检测
	needMFA, mfaState, err := s.detectMFA(ctx, client, username, password, fpVisitorId)
	if err != nil {
		return err
	}
	if needMFA {
		return s.requirePhoneMFA(ctx, client, uid, cache.MFAPurposeLogin, username, password, execution, fpVisitorId, mfaState, loginURL, redirectURL)
	}

	form := s.casLoginForm(username, encryptedPwd, execution, fpVisitorId, mfaState)

	// 3. 构造 POST 请求
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "构造登录请求失败")
	}

	setChromeHeaders(req, loginURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", originFromURL(loginURL))

	resp, err := client.Do(req)
	if err != nil {
		// 检查是否是超时错误
		if isTimeoutError(err) {
			return common.NewAppError(common.CodeJwcLoginTimeout, "教务系统登录请求超时，请稍后重试")
		}
		// 网络错误，不是认证错误
		return common.NewAppError(common.CodeJwcRequestFailed, "教务系统网络连接失败，请检查网络")
	}

	resp.Body.Close()
	loginRedirect := resp.Header.Get("Location")

	// CAS 登录的预期行为：
	// - 302: 登录成功，跳转到目标系统
	// - 200: 登录失败，返回登录表单页面（用户名或密码错误）
	// - 5xx: 服务器错误
	// - 其他: 异常情况
	if resp.StatusCode == 200 {
		// 返回 200 表示登录失败（表单页面），通常是用户名或密码错误
		return common.NewAppError(common.CodeJwcLoginFailed, "用户名或密码错误")
	} else if resp.StatusCode == 302 {
		// 登录成功，继续后续流程
	} else if resp.StatusCode >= 500 {
		// 服务器错误，不是认证问题
		return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("教务系统服务器错误: %d", resp.StatusCode))
	} else {
		// 其他状态码（4xx 等），可能是系统配置问题或异常
		return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("教务系统返回异常状态码: %d", resp.StatusCode))
	}

	// 4. 提取并缓存 CAS TGC cookie（登录成功后立即保存）
	if s.mode == "webvpn" && s.webVPNTokenURL != "" && loginRedirect != "" {
		if err := s.completeWebVPNFinish(ctx, client, fpVisitorId, loginRedirect); err != nil {
			return err
		}
	}

	if loginRedirect != "" {
		redirectBase, parseErr := url.Parse(loginURL)
		redirectTarget, targetErr := url.Parse(loginRedirect)
		if parseErr == nil && targetErr == nil {
			loginRedirect = redirectBase.ResolveReference(redirectTarget).String()
		}

		callbackResp, _, followErr := s.followGET(client, loginRedirect, 8)
		if followErr != nil {
			return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("统一认证回调失败: %v", followErr))
		}
		_ = callbackResp.Body.Close()
	}

	casURL, _ := url.Parse(loginURL)
	casCookies := client.Jar.Cookies(casURL)
	for _, cookie := range casCookies {
		if cookie.Name == "TGC" {
			// 缓存 TGC cookie，使用与 session 相同的过期时间
			_ = s.sessionCache.SetTGC(ctx, uid, cookie, s.cacheExpire)
			break
		}
	}

	// 直接不处理重定向，用这个tgc的cookie去get系统，触发下一条重定向链，get全自动重定向
	finalResp, finalURL, err := s.followGET(client, redirectURL, 12)
	if err == nil {
		finalResp, finalURL, err = s.completeJwglTicketLogin(client, finalResp, finalURL)
	}
	if err != nil {
		// 重定向失败可能是网络问题或目标系统不可达，不是认证问题
		// 检查是否是超时错误
		if appErr, ok := err.(*common.AppError); ok {
			return appErr // 保留原始错误
		}
		if isTimeoutError(err) {
			return common.NewAppError(common.CodeJwcRequestFailed, "跟随重定向超时，教务系统响应缓慢")
		}
		return common.NewAppError(common.CodeJwcRequestFailed, "跟随重定向失败，教务系统可能暂时不可用")
	}
	defer finalResp.Body.Close()

	// 5. 提取并缓存 cookies
	var cookies []*http.Cookie
	seenCookies := make(map[string]bool)
	for _, rawURL := range []string{finalURL, redirectURL, loginRedirect, loginURL} {
		u, parseErr := url.Parse(rawURL)
		if parseErr != nil || u.Host == "" {
			continue
		}
		base := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/"}
		for _, cookie := range client.Jar.Cookies(base) {
			key := cookie.Name + "|" + cookie.Domain + "|" + cookie.Path
			if !seenCookies[key] {
				seenCookies[key] = true
				cookies = append(cookies, cookie)
			}
		}
	}

	// 6. 存入缓存
	if err := cookieCache.SetCookies(ctx, uid, cookies, s.cacheExpire); err != nil {
		return common.NewAppError(common.CodeCacheError, "缓存会话失败")
	}

	return nil
}

func (s *jwcSessionService) loginAndCacheOnceByWebVPN(ctx context.Context, uid int, username, password string) error {
	return s.LoginAndCacheWithConfig(ctx, uid, username, password, s.loginURL, s.redirectURL, s.sessionCache)
}

// LoginAndGetClient 登录 CAS 并返回带 TGC cookie 的 client
func (s *jwcSessionService) LoginAndGetClient(ctx context.Context, username, password string) (*http.Client, error) {
	// 创建 cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "创建会话失败")
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: s.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止自动跳转
		},
	}

	// 请求登录页获取 execution
	loginPageReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.loginURL, nil)
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, "构造登录页请求失败")
	}
	setChromeHeaders(loginPageReq, "")
	res, err := client.Do(loginPageReq)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "连接系统失败")
	}
	defer res.Body.Close()

	// 如果已经 302，说明有 TGC，直接返回
	if res.StatusCode == 302 {
		return client, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("响应异常: %d", res.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "解析登录页面失败")
	}

	execution := doc.Find("input[name='execution']").AttrOr("value", "")
	if execution == "" {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "找不到 execution")
	}

	// 密码加密
	encryptedPwd, err := s.encryptPassword(password)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("密码加密失败: %v", err))
	}

	fpVisitorId := s.chromeVisitorID(username)

	// MFA 检测
	needMFA, mfaState, err := s.detectMFA(ctx, client, username, password, fpVisitorId)
	if err != nil {
		return nil, err
	}
	if needMFA {
		return nil, common.NewAppError(common.CodeJwcMFARequired, "需要手机短信验证，请先在网页端完成绑定")
	}

	form := s.casLoginForm(username, encryptedPwd, execution, fpVisitorId, mfaState)

	// 构造 POST 请求
	req, err := http.NewRequest("POST", s.loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "构造登录请求失败")
	}

	setChromeHeaders(req, s.loginURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", originFromURL(s.loginURL))

	resp, err := client.Do(req)
	if err != nil {
		// 检查是否是超时错误
		if isTimeoutError(err) {
			return nil, common.NewAppError(common.CodeJwcLoginTimeout, "教务系统登录请求超时")
		}
		return nil, common.NewAppError(common.CodeJwcRequestFailed, "教务系统网络连接失败")
	}
	resp.Body.Close()

	if resp.StatusCode == 200 {
		// 返回 200 表示登录失败（表单页面）
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "用户名或密码错误")
	} else if resp.StatusCode == 302 {
		// 登录成功
	} else if resp.StatusCode >= 500 {
		return nil, common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("教务系统服务器错误: %d", resp.StatusCode))
	} else {
		return nil, common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("教务系统返回异常状态码: %d", resp.StatusCode))
	}

	// TGC cookie 已在 jar 中，返回 client
	return client, nil
}

// LoginCheck 检查账号是否能被教务系统绑定
func (s *jwcSessionService) LoginCheck(ctx context.Context, uid int, username, password string) error {
	// 创建 cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "创建会话失败")
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: s.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止自动跳转
		},
	}

	// 请求登录页获取 execution
	loginPageReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.loginURL, nil)
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "构造登录页请求失败")
	}
	setChromeHeaders(loginPageReq, "")
	res, err := client.Do(loginPageReq)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "连接系统失败")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("响应异常: %d", res.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return common.NewAppError(common.CodeJwcParseFailed, "解析登录页面失败")
	}

	execution := doc.Find("input[name='execution']").AttrOr("value", "")
	if execution == "" {
		return common.NewAppError(common.CodeJwcLoginFailed, "找不到 execution")
	}

	// 密码加密
	encryptedPwd, err := s.encryptPassword(password)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("密码加密失败: %v", err))
	}

	fpVisitorId := s.chromeVisitorID(username)

	// MFA 检测
	needMFA, mfaState, err := s.detectMFA(ctx, client, username, password, fpVisitorId)
	if err != nil {
		return err
	}
	if needMFA {
		return s.requirePhoneMFA(ctx, client, uid, cache.MFAPurposeBind, username, password, execution, fpVisitorId, mfaState, s.loginURL, s.redirectURL)
	}

	form := s.casLoginForm(username, encryptedPwd, execution, fpVisitorId, mfaState)

	// 构造 POST 请求
	req, err := http.NewRequest("POST", s.loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "构造登录请求失败")
	}

	setChromeHeaders(req, s.loginURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", originFromURL(s.loginURL))

	resp, err := client.Do(req)
	if err != nil {
		// 检查是否是超时错误
		if isTimeoutError(err) {
			return common.NewAppError(common.CodeJwcLoginTimeout, "教务系统登录请求超时")
		}
		return common.NewAppError(common.CodeJwcRequestFailed, "教务系统网络连接失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// CAS 登录失败时会重新返回表单页，优先透传页面上的具体错误原因。
		doc, parseErr := goquery.NewDocumentFromReader(resp.Body)
		if parseErr == nil {
			selectors := []string{"#msg", "#errorMsg", ".login-error", ".errors", ".error", ".alert-danger"}
			for _, selector := range selectors {
				if message := strings.TrimSpace(doc.Find(selector).First().Text()); message != "" {
					return common.NewAppError(common.CodeJwcLoginFailed, message)
				}
			}
		}
		return common.NewAppError(common.CodeJwcLoginFailed, "统一认证未通过，可能需要验证码或多因素认证")
	} else if resp.StatusCode == 302 {
		// 登录成功
	} else if resp.StatusCode >= 500 {
		return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("教务系统服务器错误: %d", resp.StatusCode))
	} else {
		return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("教务系统返回异常状态码: %d", resp.StatusCode))
	}

	// TGC cookie 已在 jar 中，返回 client
	return nil
}

// isTimeoutError 判断错误是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是 url.Error（http.Client.Do 返回的错误类型）
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// 检查是否是超时错误
		if urlErr.Timeout() {
			return true
		}
		// 检查底层错误是否是网络超时
		var netErr net.Error
		if errors.As(urlErr.Err, &netErr) && netErr.Timeout() {
			return true
		}
	}

	// 直接检查是否是 net.Error 类型的超时
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}

// detectMFA 检测是否需要多因素认证
// 返回 (needMFA, state, error)
func (s *jwcSessionService) detectMFA(ctx context.Context, client *http.Client, username, password, fpVisitorID string) (bool, string, error) {
	detectURL := s.casAPIURL("/cas/mfa/detect", nil)
	if detectURL == "" {
		detectURL = s.mfaDetectURL
	}
	if detectURL == "" {
		return false, "", nil
	}

	// 加密密码
	encryptedPwd, err := s.encryptPassword(password)
	if err != nil {
		return false, "", err
	}

	// 构造请求体 (application/x-www-form-urlencoded)
	formData := url.Values{
		"username":    {username},
		"password":    {encryptedPwd},
		"fpVisitorId": {fpVisitorID},
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", detectURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return false, "", common.NewAppError(common.CodeInternalError, "创建 MFA 检测请求失败")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", chromeUserAgent)
	req.Header.Set("Origin", originFromURL(detectURL))
	req.Header.Set("Referer", s.loginURL)

	if client == nil {
		client = &http.Client{Timeout: s.timeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return false, "", common.NewAppError(common.CodeJwcLoginTimeout, "MFA 检测请求超时")
		}
		return false, "", common.NewAppError(common.CodeJwcRequestFailed, "MFA 检测请求失败")
	}
	defer resp.Body.Close()

	// 解析响应
	var mfaResponse struct {
		Code int `json:"code"`
		Data struct {
			Need  bool   `json:"need"`
			State string `json:"state"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mfaResponse); err != nil {
		return false, "", common.NewAppError(common.CodeJwcParseFailed, "解析 MFA 检测响应失败")
	}

	// 检查响应状态
	if mfaResponse.Code != 0 {
		return false, "", common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("MFA 检测失败，错误码: %d", mfaResponse.Code))
	}

	return mfaResponse.Data.Need, mfaResponse.Data.State, nil
}
