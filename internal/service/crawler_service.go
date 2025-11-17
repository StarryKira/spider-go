package service

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"spider-go/internal/common"
	"strings"
	"time"
)

// CrawlerService 爬虫服务接口
type CrawlerService interface {
	// FetchWithCookies 使用 cookies 发起请求
	FetchWithCookies(ctx context.Context, method, targetURL string, cookies []*http.Cookie, formData url.Values) (io.ReadCloser, error)
}

// httpCrawlerService HTTP 爬虫服务实现
type httpCrawlerService struct {
	timeout time.Duration
}

// NewHttpCrawlerService 创建 HTTP 爬虫服务
func NewHttpCrawlerService() CrawlerService {
	return &httpCrawlerService{
		timeout: 30 * time.Second,
	}
}

// FetchWithCookies 使用 cookies 发起请求
func (s *httpCrawlerService) FetchWithCookies(ctx context.Context, method, targetURL string, cookies []*http.Cookie, formData url.Values) (io.ReadCloser, error) {
	var body io.Reader
	if formData != nil {
		body = strings.NewReader(formData.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcRequestFailed, "构造请求失败")
	}

	// 添加 cookies
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if formData != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	client := &http.Client{
		Timeout: s.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcRequestFailed, "请求失败")
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, common.NewAppError(common.CodeJwcRequestFailed, "教务系统响应异常")
	}

	return resp.Body, nil
}
