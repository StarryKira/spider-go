package service

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"spider-go/internal/common"
	"spider-go/internal/repository"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExamService 考试服务接口
type ExamService interface {
	// GetAllExams 获取考试安排
	GetAllExams(ctx context.Context, uid int, term string) ([]ExamArrangement, error)
}

// ExamArrangement 考试安排
type ExamArrangement struct {
	SerialNo  string `json:"serial_no"`
	ClassNo   string `json:"class_no"`
	ClassName string `json:"class_name"`
	Time      string `json:"time"`
	Place     string `json:"place"`
	Execution string `json:"execution"`
}

// examServiceImpl 考试服务实现
type examServiceImpl struct {
	userRepo       repository.UserRepository
	sessionService SessionService
	crawlerService CrawlerService
	examURL        string
}

// NewExamService 创建考试服务
func NewExamService(
	userRepo repository.UserRepository,
	sessionService SessionService,
	crawlerService CrawlerService,
	examURL string,
) ExamService {
	return &examServiceImpl{
		userRepo:       userRepo,
		sessionService: sessionService,
		crawlerService: crawlerService,
		examURL:        examURL,
	}
}

// GetAllExams 获取考试安排
func (s *examServiceImpl) GetAllExams(ctx context.Context, uid int, term string) ([]ExamArrangement, error) {
	// 1. 校验参数
	re := regexp.MustCompile(`^\d{4}-\d{4}-[12]$`)
	if !re.MatchString(term) {
		return nil, common.NewAppError(common.CodeJwcInvalidParams, "学期格式错误")
	}

	// 2. 获取用户信息
	user, err := s.userRepo.GetUserByUid(uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}

	if user.Sid == "" || user.Spwd == "" {
		return nil, common.NewAppError(common.CodeJwcNotBound, "")
	}

	// 3. 获取会话
	cookies, err := s.getCookiesOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, err
	}

	// 4. 构造请求
	form := url.Values{}
	form.Add("xnxqid", term)

	// 5. 发起请求
	body, err := s.crawlerService.FetchWithCookies(ctx, "POST", s.examURL, cookies, form)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// 6. 解析响应
	return s.parseExamArrangementFromHTML(body)
}

// getCookiesOrLogin 获取缓存的 cookies 或登录
func (s *examServiceImpl) getCookiesOrLogin(ctx context.Context, uid int, sid, spwd string) ([]*http.Cookie, error) {
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
func (s *examServiceImpl) parseExamArrangementFromHTML(r io.Reader) ([]ExamArrangement, error) {
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
