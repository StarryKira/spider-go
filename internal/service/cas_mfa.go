package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"spider-go/internal/cache"
	"spider-go/internal/common"

	"golang.org/x/net/publicsuffix"
)

const mfaChallengeTTL = 10 * time.Minute

func (s *jwcSessionService) casAPIURL(path string, query url.Values) string {
	raw := s.loginURL
	if raw == "" {
		raw = s.mfaDetectURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	u.Path = path
	if query != nil {
		u.RawQuery = query.Encode()
	} else {
		u.RawQuery = ""
	}
	u.Fragment = ""
	return u.String()
}

func (s *jwcSessionService) requirePhoneMFA(
	ctx context.Context,
	client *http.Client,
	uid int,
	purpose, username, password, execution, fpVisitorID, mfaState, loginURL, redirectURL string,
) error {
	if uid <= 0 {
		return common.NewAppError(common.CodeJwcMFARequired, "需要输入绑定手机收到的验证码")
	}
	if s.mfaCache == nil {
		return common.NewAppError(common.CodeJwcMFARequired, "需要输入绑定手机收到的验证码")
	}

	phone, gid, attestURL, err := s.startSecurePhoneMFA(ctx, client, mfaState)
	if err != nil {
		return err
	}

	challenge := &cache.MFAChallenge{
		UID:         uid,
		Purpose:     purpose,
		SID:         username,
		Password:    password,
		Phone:       phone,
		State:       mfaState,
		GID:         gid,
		AttestURL:   attestURL,
		VisitorID:   fpVisitorID,
		Execution:   execution,
		LoginURL:    loginURL,
		RedirectURL: redirectURL,
		Cookies:     cookiesFromClient(client, loginURL),
	}
	if err := s.mfaCache.Set(ctx, uid, challenge, mfaChallengeTTL); err != nil {
		return common.NewAppError(common.CodeCacheError, "保存手机验证状态失败")
	}

	return common.NewAppErrorWithData(
		common.CodeJwcMFARequired,
		fmt.Sprintf("请输入手机 %s 收到的验证码", phone),
		map[string]interface{}{
			"mfa_required": true,
			"mfa_type":     "securephone",
			"phone":        phone,
		},
	)
}

func (s *jwcSessionService) startSecurePhoneMFA(ctx context.Context, client *http.Client, mfaState string) (phone, gid, attestURL string, err error) {
	initURL := s.casAPIURL("/cas/mfa/initByType/securephone", url.Values{"state": {mfaState}})
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, initURL, nil)
	if reqErr != nil {
		return "", "", "", common.NewAppError(common.CodeInternalError, "构造手机验证请求失败")
	}
	setChromeHeaders(req, s.loginURL)

	resp, doErr := client.Do(req)
	if doErr != nil {
		if isTimeoutError(doErr) {
			return "", "", "", common.NewAppError(common.CodeJwcLoginTimeout, "获取安全手机信息超时")
		}
		return "", "", "", common.NewAppError(common.CodeJwcRequestFailed, "获取安全手机信息失败")
	}
	defer resp.Body.Close()

	var initResp struct {
		Code int `json:"code"`
		Data struct {
			AttestServerURL string `json:"attestServerUrl"`
			GID             string `json:"gid"`
			SecurePhone     string `json:"securePhone"`
		} `json:"data"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&initResp); decodeErr != nil {
		return "", "", "", common.NewAppError(common.CodeJwcParseFailed, "解析安全手机信息失败")
	}
	if initResp.Code != 0 || initResp.Data.GID == "" {
		return "", "", "", common.NewAppError(common.CodeJwcMFARequired, "安全手机未绑定，无法发送验证码")
	}

	phone = strings.TrimSpace(initResp.Data.SecurePhone)
	if phone == "" {
		phone = "已绑定手机"
	}
	gid = initResp.Data.GID
	attestURL = strings.TrimRight(initResp.Data.AttestServerURL, "/")
	if attestURL == "" {
		return "", "", "", common.NewAppError(common.CodeJwcRequestFailed, "安全验证服务地址无效")
	}

	if sendErr := s.sendSecurePhoneCode(ctx, client, attestURL, gid); sendErr != nil {
		return "", "", "", sendErr
	}
	return phone, gid, attestURL, nil
}

func (s *jwcSessionService) sendSecurePhoneCode(ctx context.Context, client *http.Client, attestURL, gid string) error {
	payload, _ := json.Marshal(map[string]string{"gid": gid})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, attestURL+"/api/guard/securephone/send", bytes.NewReader(payload))
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "构造发送验证码请求失败")
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", chromeUserAgent)
	req.Header.Set("Origin", originFromURL(attestURL))
	req.Header.Set("Referer", s.loginURL)

	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return common.NewAppError(common.CodeJwcLoginTimeout, "发送手机验证码超时")
		}
		return common.NewAppError(common.CodeJwcRequestFailed, "发送手机验证码失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var sendResp struct {
		Code int `json:"code"`
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &sendResp); err != nil {
		return common.NewAppError(common.CodeJwcParseFailed, "解析发送验证码结果失败")
	}
	if sendResp.Code != 0 {
		if sendResp.Data.Result == "expired" {
			return common.NewAppError(common.CodeJwcMFARequired, "验证已过期，请重新绑定")
		}
		return common.NewAppError(common.CodeJwcRequestFailed, "发送手机验证码失败")
	}
	return nil
}

func (s *jwcSessionService) verifySecurePhoneCode(ctx context.Context, client *http.Client, attestURL, gid, code string) error {
	payload, _ := json.Marshal(map[string]string{"gid": gid, "code": strings.TrimSpace(code)})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, attestURL+"/api/guard/securephone/valid", bytes.NewReader(payload))
	if err != nil {
		return common.NewAppError(common.CodeInternalError, "构造校验验证码请求失败")
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", chromeUserAgent)
	req.Header.Set("Origin", originFromURL(attestURL))
	req.Header.Set("Referer", s.loginURL)

	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return common.NewAppError(common.CodeJwcLoginTimeout, "校验手机验证码超时")
		}
		return common.NewAppError(common.CodeJwcRequestFailed, "校验手机验证码失败")
	}
	defer resp.Body.Close()

	var validResp struct {
		Code int `json:"code"`
		Data struct {
			Status int `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&validResp); err != nil {
		return common.NewAppError(common.CodeJwcParseFailed, "解析校验验证码结果失败")
	}
	if validResp.Code != 0 || validResp.Data.Status != 2 {
		return common.NewAppError(common.CodeJwcLoginFailed, "手机验证码错误或已失效")
	}
	return nil
}

func (s *jwcSessionService) CompletePhoneMFA(ctx context.Context, uid int, code string) (*cache.MFAChallenge, error) {
	if strings.TrimSpace(code) == "" {
		return nil, common.NewAppError(common.CodeInvalidParams, "请输入手机验证码")
	}
	if s.mfaCache == nil {
		return nil, common.NewAppError(common.CodeJwcMFARequired, "没有待验证的手机验证码")
	}

	challenge, err := s.mfaCache.Get(ctx, uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeCacheError, "读取手机验证状态失败")
	}
	if challenge == nil || challenge.GID == "" {
		return nil, common.NewAppError(common.CodeJwcMFARequired, "手机验证已过期，请重新登录或绑定")
	}

	client, err := clientFromChallenge(challenge, s.timeout)
	if err != nil {
		return nil, err
	}
	if err := s.verifySecurePhoneCode(ctx, client, challenge.AttestURL, challenge.GID, code); err != nil {
		return nil, err
	}

	encryptedPwd, err := s.encryptPassword(challenge.Password)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, fmt.Sprintf("密码加密失败: %v", err))
	}

	loginURL := challenge.LoginURL
	if loginURL == "" {
		loginURL = s.loginURL
	}
	form := s.casLoginForm(challenge.SID, encryptedPwd, challenge.Execution, challenge.VisitorID, challenge.State)
	form.Set("trustAgent", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, common.NewAppError(common.CodeInternalError, "构造登录请求失败")
	}
	setChromeHeaders(req, loginURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", originFromURL(loginURL))

	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return nil, common.NewAppError(common.CodeJwcLoginTimeout, "教务系统登录请求超时")
		}
		return nil, common.NewAppError(common.CodeJwcRequestFailed, "教务系统网络连接失败")
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "手机验证通过，但统一认证登录未成功")
	}

	if cacheErr := s.finishLoginAfterMFA(ctx, client, challenge, loginURL, resp.Header.Get("Location")); cacheErr != nil && challenge.Purpose == cache.MFAPurposeLogin {
		return nil, cacheErr
	}

	_ = s.mfaCache.Delete(ctx, uid)
	return challenge, nil
}

func (s *jwcSessionService) ResendPhoneMFA(ctx context.Context, uid int) error {
	if s.mfaCache == nil {
		return common.NewAppError(common.CodeJwcMFARequired, "没有待发送的手机验证码")
	}
	challenge, err := s.mfaCache.Get(ctx, uid)
	if err != nil {
		return common.NewAppError(common.CodeCacheError, "读取手机验证状态失败")
	}
	if challenge == nil || challenge.GID == "" {
		return common.NewAppError(common.CodeJwcMFARequired, "手机验证已过期，请重新登录或绑定")
	}
	client, err := clientFromChallenge(challenge, s.timeout)
	if err != nil {
		return err
	}
	return s.sendSecurePhoneCode(ctx, client, challenge.AttestURL, challenge.GID)
}

func (s *jwcSessionService) finishLoginAfterMFA(ctx context.Context, client *http.Client, challenge *cache.MFAChallenge, loginURL, loginRedirect string) error {
	if s.mode == "webvpn" && s.webVPNTokenURL != "" && loginRedirect != "" {
		if err := s.completeWebVPNFinish(ctx, client, challenge.VisitorID, loginRedirect); err != nil {
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

	redirectURL := challenge.RedirectURL
	if redirectURL == "" {
		redirectURL = s.redirectURL
	}
	finalResp, finalURL, err := s.followGET(client, redirectURL, 12)
	if err == nil {
		finalResp, finalURL, err = s.completeJwglTicketLogin(client, finalResp, finalURL)
	}
	if err != nil {
		return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("进入教务系统失败: %v", err))
	}
	defer finalResp.Body.Close()

	var cookies []*http.Cookie
	seen := make(map[string]bool)
	for _, rawURL := range []string{finalURL, redirectURL, loginRedirect, loginURL} {
		u, parseErr := url.Parse(rawURL)
		if parseErr != nil || u.Host == "" {
			continue
		}
		base := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/"}
		for _, cookie := range client.Jar.Cookies(base) {
			key := cookie.Name + "|" + cookie.Domain + "|" + cookie.Path
			if !seen[key] {
				seen[key] = true
				cookies = append(cookies, cookie)
			}
		}
	}
	if len(cookies) == 0 {
		return common.NewAppError(common.CodeJwcLoginFailed, "登录成功但未获取到会话")
	}
	return s.sessionCache.SetCookies(ctx, challenge.UID, cookies, s.cacheExpire)
}

func (s *jwcSessionService) webVPNOrigin() string {
	if origin := originFromURL(s.webVPNTokenURL); origin != "" {
		return origin
	}
	return "https://webvpn.csuft.edu.cn"
}

func (s *jwcSessionService) prepareWebVPNSession(ctx context.Context, client *http.Client) error {
	origin := s.webVPNOrigin()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/", nil)
	if err != nil {
		return common.NewAppError(common.CodeJwcRequestFailed, "构造 WebVPN 会话请求失败")
	}
	setChromeHeaders(req, "")
	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return common.NewAppError(common.CodeJwcLoginTimeout, "连接 WebVPN 超时")
		}
		return common.NewAppError(common.CodeJwcRequestFailed, "连接 WebVPN 失败")
	}
	_ = resp.Body.Close()
	return nil
}

func (s *jwcSessionService) webVPNExternalID() string {
	u, err := url.Parse(s.loginURL)
	if err != nil {
		return ""
	}
	service := u.Query().Get("service")
	if service == "" {
		return ""
	}
	svc, err := url.Parse(service)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(svc.Path, "/callback/cas/")
}

func (s *jwcSessionService) lookupWebVPNCasExternalID(ctx context.Context, client *http.Client) (string, error) {
	origin := s.webVPNOrigin()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/api/access/authentication/list?type=0", nil)
	if err != nil {
		return "", common.NewAppError(common.CodeJwcRequestFailed, "构造 WebVPN 认证方式请求失败")
	}
	setChromeHeaders(req, origin+"/auth/login")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	resp, err := client.Do(req)
	if err != nil {
		return "", common.NewAppError(common.CodeJwcRequestFailed, "获取 WebVPN 认证方式失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var listResp struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				AuthType   int    `json:"authType"`
				ExternalID string `json:"externalId"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		return "", common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("解析 WebVPN 认证方式失败: %s", trimForError(body)))
	}
	for _, item := range listResp.Data.List {
		if item.AuthType == 4 && item.ExternalID != "" {
			return item.ExternalID, nil
		}
	}
	if len(listResp.Data.List) == 1 && listResp.Data.List[0].ExternalID != "" {
		return listResp.Data.List[0].ExternalID, nil
	}
	return "", nil
}

func (s *jwcSessionService) startWebVPNAuth(ctx context.Context, client *http.Client) (string, error) {
	if err := s.prepareWebVPNSession(ctx, client); err != nil {
		return "", err
	}
	externalID, err := s.lookupWebVPNCasExternalID(ctx, client)
	if err != nil {
		return "", err
	}
	if externalID == "" {
		externalID = s.webVPNExternalID()
	}
	if externalID == "" {
		return s.loginURL, nil
	}
	origin := s.webVPNOrigin()
	callbackURL := origin + "/callback/cas/" + externalID
	payload, _ := json.Marshal(map[string]string{
		"externalId": externalID,
		"data":       fmt.Sprintf(`{"callbackUrl":%q}`, callbackURL),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, origin+"/api/access/auth/start", bytes.NewReader(payload))
	if err != nil {
		return "", common.NewAppError(common.CodeJwcRequestFailed, "构造 WebVPN 认证启动请求失败")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", chromeUserAgent)
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", origin+"/auth/login")
	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return "", common.NewAppError(common.CodeJwcLoginTimeout, "启动 WebVPN 认证超时")
		}
		return "", common.NewAppError(common.CodeJwcRequestFailed, "启动 WebVPN 认证失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var startResp struct {
		Code int `json:"code"`
		Data struct {
			Action struct {
				LoginURL string `json:"login_url"`
			} `json:"action"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &startResp); err != nil || startResp.Data.Action.LoginURL == "" {
		return "", common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("WebVPN 未返回登录地址: %s", trimForError(body)))
	}
	return startResp.Data.Action.LoginURL, nil
}

func trimForError(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 240 {
		return s[:240]
	}
	if s == "" {
		return "(empty)"
	}
	return s
}

func (s *jwcSessionService) completeWebVPNFinish(ctx context.Context, client *http.Client, fpVisitorId, loginRedirect string) error {
	if err := s.prepareWebVPNSession(ctx, client); err != nil {
		return err
	}

	origin := s.webVPNOrigin()
	callbackURL, parseErr := url.Parse(loginRedirect)
	if parseErr != nil {
		return common.NewAppError(common.CodeJwcRequestFailed, "WebVPN 回调地址无效")
	}
	if callbackURL.Host == "" || strings.Contains(callbackURL.Host, "cas") {
		base, _ := url.Parse(origin + "/")
		callbackURL = base.ResolveReference(&url.URL{Path: callbackURL.Path, RawQuery: callbackURL.RawQuery})
	}
	ticket := callbackURL.Query().Get("ticket")
	externalID := strings.TrimPrefix(callbackURL.Path, "/callback/cas/")
	callbackURL.RawQuery = ""
	if ticket == "" || externalID == "" {
		return common.NewAppError(common.CodeJwcRequestFailed, "WebVPN 回调缺少认证参数")
	}

	// 官方前端会先打开回调页再 POST finish，需要先落到 webvpn 域拿 cookie。
	landingReq, landingErr := http.NewRequestWithContext(ctx, http.MethodGet, callbackURL.String()+"?ticket="+url.QueryEscape(ticket), nil)
	if landingErr == nil {
		setChromeHeaders(landingReq, origin+"/")
		if landingResp, doErr := client.Do(landingReq); doErr == nil {
			_ = landingResp.Body.Close()
		}
	}

	deviceID := fpVisitorId
	if len(deviceID) > 32 {
		deviceID = deviceID[:32]
	}
	dataPayload, _ := json.Marshal(map[string]string{
		"callbackUrl": callbackURL.String(),
		"ticket":      ticket,
		"deviceId":    deviceID,
	})
	finishPayload, _ := json.Marshal(map[string]string{
		"externalId": externalID,
		"data":       string(dataPayload),
	})
	finishReq, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, s.webVPNTokenURL, strings.NewReader(string(finishPayload)))
	if requestErr != nil {
		return common.NewAppError(common.CodeJwcRequestFailed, "构造 WebVPN 认证请求失败")
	}
	finishReq.Header.Set("Content-Type", "application/json")
	finishReq.Header.Set("Accept", "application/json, text/plain, */*")
	finishReq.Header.Set("Referer", loginRedirect)
	finishReq.Header.Set("User-Agent", chromeUserAgent)
	finishReq.Header.Set("Origin", originFromURL(s.webVPNTokenURL))
	finishResp, finishErr := client.Do(finishReq)
	if finishErr != nil {
		if isTimeoutError(finishErr) {
			return common.NewAppError(common.CodeJwcLoginTimeout, "WebVPN 认证请求超时")
		}
		return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("WebVPN 认证请求失败: %v", finishErr))
	}
	finishBody, _ := io.ReadAll(finishResp.Body)
	_ = finishResp.Body.Close()
	if finishResp.StatusCode != http.StatusOK {
		return common.NewAppError(common.CodeJwcRequestFailed, fmt.Sprintf("WebVPN 认证失败: %d %s", finishResp.StatusCode, trimForError(finishBody)))
	}
	return nil
}

func cookiesFromClient(client *http.Client, rawURL string) []*http.Cookie {
	if client == nil || client.Jar == nil {
		return nil
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return nil
	}
	base := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/"}
	return client.Jar.Cookies(base)
}

func clientFromChallenge(challenge *cache.MFAChallenge, timeout time.Duration) (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "创建会话失败")
	}
	if u, parseErr := url.Parse(challenge.LoginURL); parseErr == nil && u.Host != "" && len(challenge.Cookies) > 0 {
		jar.SetCookies(&url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/"}, challenge.Cookies)
	}
	return &http.Client{
		Jar:     jar,
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}
