package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"spider-go/internal/app"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

// 返回：显示名（如果页面可解析到），否则可能为空字符串
func LoginAndStoreSession(uid int, username, jwcpassword string) error {
	// 0) 同一个 client + cookiejar 贯穿全流程
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return err
	}
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		// 阻止自动重定向，便于拿到 Location；后续我们手动跟随
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 1) GET 登录页（初始 cookie 自动进 jar）
	res, err := client.Get(Jwc_url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("连接登录页失败: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	lt := doc.Find("input[name='lt']").AttrOr("value", "")
	dllt := doc.Find("input[name='dllt']").AttrOr("value", "")
	execution := doc.Find("input[name='execution']").AttrOr("value", "")
	eventID := doc.Find("input[name='_eventId']").AttrOr("value", "")
	salt := doc.Find("input[id='pwdDefaultEncryptSalt']").AttrOr("value", "")
	rmShown := doc.Find("input[name='rmShown']").AttrOr("value", "")

	if lt == "" || execution == "" || eventID == "" || salt == "" {
		return errors.New("登录页缺少必要字段，可能需要验证码或页面结构已变")
	}

	// 2) 加密密码 + 提交表单
	encryptedPwd := JsCrypto(jwcpassword, salt)

	form := url.Values{}
	form.Set("username", username)
	form.Set("password", encryptedPwd)
	form.Set("lt", lt)
	form.Set("dllt", dllt)
	form.Set("execution", execution)
	form.Set("_eventId", eventID)
	form.Set("rmShown", rmShown)

	req, err := http.NewRequest("POST", Jwc_url, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", Jwc_url)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 期望 3xx，拿 Location
	if resp.StatusCode/100 != 3 {
		return fmt.Errorf("登录提交失败，期望 3xx，得到 %d", resp.StatusCode)
	}
	loc, err := resp.Location()
	if err != nil {
		return errors.New("未获得重定向地址")
	}
	redirectURL := loc.String()

	// 3) 用同一个 client 手动跟随重定向直到落地（同时积累 cookie）
	finalResp, finalURL, err := followGET(client, redirectURL, 8) // 最多 8 跳
	if err != nil {
		return err
	}
	defer finalResp.Body.Close()

	// 5) 统一把 cookies 存入 Redis（1 小时）
	// 优先用最终落地页的域；兜底用根域
	uFinal, _ := url.Parse(finalURL)
	base := &url.URL{Scheme: uFinal.Scheme, Host: uFinal.Host, Path: "/"}
	cookies := client.Jar.Cookies(base)
	if len(cookies) == 0 {
		if u, e := url.Parse(redirectURL); e == nil {
			cookies = client.Jar.Cookies(u)
		}
	}
	data, err := json.Marshal(cookies)
	if err != nil {
		return err
	}
	if err := app.Rdb.Set(app.Ctx, strconv.Itoa(uid), data, time.Hour).Err(); err != nil {
		return err
	}

	return nil
}

// 手动 GET 跟随重定向（因为 client 禁用了自动跟随）
// 返回最终响应和最终 URL（调用方负责关闭 resp.Body）
func followGET(client *http.Client, start string, maxHops int) (*http.Response, string, error) {
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
