package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
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
	sessionCache    cache.SessionCache
	rsaKeyService   RSAKeyService
	loginURL        string
	redirectURL     string
	captchaURL      string
	ocrURL          string
	captchaImageURL string
	timeout         time.Duration
	cacheExpire     time.Duration
}

// NewJwcSessionService 创建教务系统会话服务
func NewJwcSessionService(
	sessionCache cache.SessionCache,
	rsaKeyService RSAKeyService,
	loginURL string,
	redirectURL string,
	captchaURL string,
	captchaImageURL string,
	ocrURL string,
) SessionService {
	return &jwcSessionService{
		sessionCache:    sessionCache,
		rsaKeyService:   rsaKeyService,
		loginURL:        loginURL,
		redirectURL:     redirectURL,
		captchaURL:      captchaURL,
		captchaImageURL: captchaImageURL,
		ocrURL:          ocrURL,
		timeout:         30 * time.Second,
		cacheExpire:     time.Hour,
	}
}

// LoginAndCache 登录教务系统并缓存会话（带重试机制）
func (s *jwcSessionService) LoginAndCache(ctx context.Context, uid int, username, password string) error {
	var err error
	// 重试 3 次
	for i := 0; i < 3; i++ {
		if err = s.loginAndCacheOnce(ctx, uid, username, password); err == nil {
			return nil
		}
		// 重试间隔
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("登录失败，已重试3次: %v", err))
}

// loginAndCacheOnce 单次登录逻辑
func (s *jwcSessionService) loginAndCacheOnce(ctx context.Context, uid int, username, password string) error {
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

	// 1. 请求登录页获取 execution
	res, err := client.Get(s.loginURL)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "连接教务系统失败")
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

	fpVisitorId, err := s.GenerateRandomFingerPrintHash()
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "生成设备指纹失败")
	}

	form := url.Values{
		"username":    {username},
		"password":    {encryptedPwd},
		"execution":   {execution},
		"fpVisitorId": {fpVisitorId},
		"_eventId":    {"submit"},
		"failN":       {"0"},
		"submit1":     {"login1"},
	}

	// 3. 构造 POST 请求
	req, err := http.NewRequest("POST", s.loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "构造登录请求失败")
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", s.loginURL)

	resp, err := client.Do(req)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "登录失败")
	}

	resp.Body.Close()

	if resp.StatusCode != 302 {
		return common.NewAppError(common.CodeJwcLoginFailed, "重定向并非302")
	}

	//直接不处理重定向，用这个tgc的cookie去get教务系统，触发下一条重定向链，get全自动重定向

	finalResp, finalURL, err := s.followGET(client, s.redirectURL, 8)
	if err != nil {
		return common.NewAppError(common.CodeJwcLoginFailed, "跟随重定向失败")
	}
	defer finalResp.Body.Close()

	// 6. 提取并缓存 cookies
	uFinal, _ := url.Parse(finalURL)
	base := &url.URL{Scheme: uFinal.Scheme, Host: uFinal.Host, Path: "/"}
	cookies := client.Jar.Cookies(base)

	if len(cookies) == 0 {
		if u, e := url.Parse(s.redirectURL); e == nil {
			cookies = client.Jar.Cookies(u)
		}
	}

	// 7. 存入缓存
	if err := s.sessionCache.SetCookies(ctx, uid, cookies, s.cacheExpire); err != nil {
		return common.NewAppError(common.CodeCacheError, "缓存会话失败")
	}

	return nil
}

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
			return nil, cur, fmt.Errorf("location 无法解析: %v", err)
		}

		cur = lastReqURL.ResolveReference(locURL).String()
		lastReqURL = locURL
	}

	return nil, cur, errors.New("重定向层级过多")
}

// GetCachedCookies 获取缓存的 cookies
func (s *jwcSessionService) GetCachedCookies(ctx context.Context, uid int) ([]*http.Cookie, error) {
	return s.sessionCache.GetCookies(ctx, uid)
}

// InvalidateSession 清除会话缓存
func (s *jwcSessionService) InvalidateSession(ctx context.Context, uid int) error {
	return s.sessionCache.DeleteCookies(ctx, uid)
}

func (s *jwcSessionService) HandleCaptcha(picBase64 string) (string, error) {
	data := url.Values{}
	data.Set("image", picBase64)
	data.Set("probability", "false")
	data.Set("png_fix", "false")

	resp, err := http.PostForm(s.ocrURL, data)
	if err != nil {
		return "", common.NewAppError(common.CodeInternalError, err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body), nil
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

// GenerateRandomFingerPrintHash 随机生成32位设备指纹hash
func (s *jwcSessionService) GenerateRandomFingerPrintHash() (string, error) {
	// 生成 32 字节随机数
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// 计算 SHA256
	h := sha256.Sum256(b)

	// 转成 hex 字符串返回
	return hex.EncodeToString(h[:]), nil
}

// ReplaceClientID 替换clientID
func (s *jwcSessionService) ReplaceClientID(rawURL, newClientID string) (string, error) {
	// 解析外层 URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 解析外层 URL 的查询参数
	q := u.Query()

	// 获取 service 参数（内层 URL）
	serviceRaw := q.Get("service")
	if serviceRaw == "" {
		return "", fmt.Errorf("service parameter not found")
	}

	// 解析 service URL
	serviceURL, err := url.Parse(serviceRaw)
	if err != nil {
		return "", err
	}

	// 解析 service 内层查询参数
	serviceQ := serviceURL.Query()

	// 替换 client_id
	serviceQ.Set("client_id", newClientID)
	serviceURL.RawQuery = serviceQ.Encode()

	// 替换回外层的 service 参数
	q.Set("service", serviceURL.String())
	u.RawQuery = q.Encode()

	return u.String(), nil
}
