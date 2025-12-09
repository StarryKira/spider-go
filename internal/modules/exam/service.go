package exam

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"spider-go/internal/service"
	"spider-go/internal/shared"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Service 考试服务接口
type Service interface {
	GetAllExams(ctx context.Context, uid int, term string) ([]ExamArrangement, error)
}

// examService 考试服务实现
type examService struct {
	userQuery      shared.UserQuery
	sessionService service.SessionService
	crawlerService service.CrawlerService
	userDataCache  cache.UserDataCache
	examURL        string
}

// NewService 创建考试服务
func NewService(
	userQuery shared.UserQuery,
	sessionService service.SessionService,
	crawlerService service.CrawlerService,
	userDataCache cache.UserDataCache,
	examURL string,
) Service {
	return &examService{
		userQuery:      userQuery,
		sessionService: sessionService,
		crawlerService: crawlerService,
		userDataCache:  userDataCache,
		examURL:        examURL,
	}
}

// GetAllExams 获取考试安排
func (s *examService) GetAllExams(ctx context.Context, uid int, term string) ([]ExamArrangement, error) {
	// 校验参数
	re := regexp.MustCompile(`^\d{4}-\d{4}-[12]$`)
	if !re.MatchString(term) {
		return nil, common.NewAppError(common.CodeJwcInvalidParams, "学期格式错误")
	}

	// 获取用户信息
	user, err := s.userQuery.GetUserByUid(ctx, uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}

	if user.Sid == "" || user.Spwd == "" {
		return nil, common.NewAppError(common.CodeJwcNotBound, "")
	}

	// 先查询缓存
	var cachedExams []ExamArrangement
	if err := s.userDataCache.GetExams(ctx, uid, term, &cachedExams); err == nil {
		return cachedExams, nil
	}

	// 获取会话
	cookies, err := s.getCookiesOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, err
	}

	// 构造请求
	form := url.Values{}
	form.Add("xnxqid", term)

	// 发起请求
	body, err := s.crawlerService.FetchWithCookies(ctx, "POST", s.examURL, cookies, form)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// 解析响应
	exams, err := s.parseExamArrangementFromHTML(body)
	if err != nil {
		return nil, err
	}

	// 写入缓存（1小时过期）
	_ = s.userDataCache.CacheExams(ctx, uid, term, exams, time.Hour)

	return exams, nil
}

// getCookiesOrLogin 获取缓存的 cookies 或登录
func (s *examService) getCookiesOrLogin(ctx context.Context, uid int, sid, spwd string) ([]*http.Cookie, error) {
	cookies, err := s.sessionService.GetCachedCookies(ctx, uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeCacheError, "缓存错误")
	}

	if len(cookies) > 0 {
		return cookies, nil
	}

	if err := s.sessionService.LoginAndCache(ctx, uid, sid, spwd); err != nil {
		return nil, err
	}

	cookies, err = s.sessionService.GetCachedCookies(ctx, uid)
	if err != nil || len(cookies) == 0 {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "获取会话失败")
	}

	return cookies, nil
}

// parseExamArrangementFromHTML 解析考试安排 HTML
func (s *examService) parseExamArrangementFromHTML(r io.Reader) ([]ExamArrangement, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "解析HTML失败")
	}

	title := strings.TrimSpace(doc.Find("title").Text())
	if title != "我的考试 - 考试安排查询" {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "页面错误")
	}

	table := doc.Find("#dataList")
	if table.Length() == 0 {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "未找到考试安排数据")
	}

	rows := table.Find("tr")
	if rows.Length() <= 1 {
		return nil, nil // 只有表头，无数据
	}

	// 检查是否显示"未查询到数据"
	if strings.Contains(rows.Eq(1).Text(), "未查询到数据") {
		return nil, nil
	}

	var exams []ExamArrangement

	rows.Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return // 跳过表头
		}

		tds := tr.Find("td")
		if tds.Length() < 9 {
			return
		}

		trim := func(s string) string {
			s = strings.TrimSpace(s)
			s = strings.ReplaceAll(s, "\u00A0", "")
			return s
		}

		exams = append(exams, ExamArrangement{
			SerialNo:  trim(tds.Eq(0).Text()),
			ClassNo:   trim(tds.Eq(2).Text()),
			ClassName: trim(tds.Eq(3).Text()),
			Time:      trim(tds.Eq(4).Text()),
			Place:     trim(tds.Eq(5).Text()),
			Execution: trim(tds.Eq(8).Text()),
		})
	})

	return exams, nil
}
