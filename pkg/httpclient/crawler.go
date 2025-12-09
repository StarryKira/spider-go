package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Crawler HTTP 爬虫客户端接口
type Crawler interface {
	// FetchWithCookies 使用 cookies 发起请求
	FetchWithCookies(ctx context.Context, method, targetURL string, cookies []*http.Cookie, formData url.Values) (io.ReadCloser, error)
}

// crawler 爬虫客户端实现
type crawler struct {
	timeout time.Duration
}

// NewCrawler 创建 HTTP 爬虫客户端
func NewCrawler() Crawler {
	return &crawler{
		timeout: 30 * time.Second,
	}
}

// NewCrawlerWithTimeout 创建带自定义超时的 HTTP 爬虫客户端
func NewCrawlerWithTimeout(timeout time.Duration) Crawler {
	return &crawler{
		timeout: timeout,
	}
}

// FetchWithCookies 使用 cookies 发起请求
func (c *crawler) FetchWithCookies(ctx context.Context, method, targetURL string, cookies []*http.Cookie, formData url.Values) (io.ReadCloser, error) {
	var body io.Reader
	if formData != nil {
		body = strings.NewReader(formData.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加 cookies
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	if formData != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
