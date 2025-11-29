package service

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"spider-go/internal/cache"
	"spider-go/internal/common"
	"spider-go/internal/repository"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CourseService 课程服务接口
type CourseService interface {
	// GetCourseTableByWeek 获取指定周的课程表
	GetCourseTableByWeek(ctx context.Context, week int, term string, uid int) (*WeekSchedule, error)
}

// courseServiceImpl 课程服务实现
type courseServiceImpl struct {
	userRepo       repository.UserRepository
	sessionService SessionService
	crawlerService CrawlerService
	userDataCache  cache.UserDataCache
	courseURL      string
}

// NewCourseService 创建课程服务
func NewCourseService(
	userRepo repository.UserRepository,
	sessionService SessionService,
	crawlerService CrawlerService,
	userDataCache cache.UserDataCache,
	courseURL string,
) CourseService {
	return &courseServiceImpl{
		userRepo:       userRepo,
		sessionService: sessionService,
		crawlerService: crawlerService,
		userDataCache:  userDataCache,
		courseURL:      courseURL,
	}
}

// DaySchedule 一天的课程安排
type DaySchedule struct {
	Weekday int      `json:"weekday"` // 值为1-7，表示周一到周日
	Courses []Course `json:"courses"` // 当天课程，没有课则为nil
}

// WeekSchedule 一周的课程安排
type WeekSchedule struct {
	WeekNo    int           `json:"weekno"`
	Starttime string        `json:"starttime"`
	Endtime   string        `json:"endtime"`
	Days      []DaySchedule `json:"days"`
}

// Course 课程信息
type Course struct {
	Name        string `json:"name"`         // 课程名称
	Teacher     string `json:"teacher"`      // 任课老师
	Classroom   string `json:"classroom"`    // 教室
	Weekday     int    `json:"weekday"`      // 周几：1~7
	StartPeriod int    `json:"start_period"` // 开始节次
	EndPeriod   int    `json:"end_period"`   // 结束节次
}

// GetCourseTableByWeek 获取指定周的课程表
func (s *courseServiceImpl) GetCourseTableByWeek(ctx context.Context, week int, term string, uid int) (*WeekSchedule, error) {
	// 1. 校验参数
	if week > 20 || week < 1 {
		return nil, common.NewAppError(common.CodeJwcInvalidParams, "周次必须在1-20之间")
	}

	if term == "" {
		return nil, common.NewAppError(common.CodeJwcInvalidParams, "学期不能为空")
	}

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

	// 3. 先查询缓存
	var cachedSchedule WeekSchedule
	if err := s.userDataCache.GetCourseTable(ctx, uid, term, week, &cachedSchedule); err == nil {
		return &cachedSchedule, nil
	}

	// 4. 获取或创建会话
	cookies, err := s.getCookiesOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "获取cookie失败")
	}

	// 5. 构造请求
	form := url.Values{}
	form.Add("zc", strconv.Itoa(week))
	form.Add("xnxq01id", term)

	// 6. 发起请求
	body, err := s.crawlerService.FetchWithCookies(ctx, "POST", s.courseURL, cookies, form)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// 7. 解析响应
	schedule, err := s.parseCourseTableFromHTML(body)
	if err != nil {
		return nil, err
	}

	// 8. 写入缓存（1小时过期）
	_ = s.userDataCache.CacheCourseTable(ctx, uid, term, week, schedule, time.Hour)

	return schedule, nil
}

// getCookiesOrLogin 获取缓存的 cookies 或登录
func (s *courseServiceImpl) getCookiesOrLogin(ctx context.Context, uid int, sid, spwd string) ([]*http.Cookie, error) {
	// 先尝试从缓存获取
	cookies, err := s.sessionService.GetCachedCookies(ctx, uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeCacheError, "缓存错误")
	}

	// 如果有缓存，直接返回
	if len(cookies) > 0 {
		return cookies, nil
	}

	// 没有缓存，需要登录
	if err := s.sessionService.LoginAndCache(ctx, uid, sid, spwd); err != nil {
		return nil, err
	}

	// 重新获取 cookies
	cookies, err = s.sessionService.GetCachedCookies(ctx, uid)
	if err != nil || len(cookies) == 0 {
		return nil, common.NewAppError(common.CodeJwcLoginFailed, "获取会话失败")
	}

	return cookies, nil
}

// parseCourseTableFromHTML 解析课程表 HTML
func (s *courseServiceImpl) parseCourseTableFromHTML(r io.Reader) (*WeekSchedule, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "解析HTML失败")
	}

	title := strings.TrimSpace(doc.Find("title").Text())
	if title != "学期理论课表" {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "页面错误")
	}

	// 1. 解析当前周次
	weekNo := 0
	if opt := doc.Find("select#zc option[selected]"); opt.Length() > 0 {
		val, _ := opt.Attr("value")
		if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			weekNo = v
		}
	}

	// 2. 初始化 7 天
	days := make([]DaySchedule, 7)
	for i := 0; i < 7; i++ {
		days[i] = DaySchedule{
			Weekday: i + 1,
			Courses: nil,
		}
	}

	// 3. 遍历课表行
	doc.Find("#kbtable tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return // 跳过表头
		}

		thText := strings.TrimSpace(tr.Find("th").First().Text())
		if thText == "" || strings.HasPrefix(thText, "备注") {
			return
		}

		startP, endP := s.parsePeriodRange(thText)
		if startP == 0 && endP == 0 {
			return
		}

		// 遍历一行的 7 列（周一到周日）
		tr.Find("td").Each(func(col int, td *goquery.Selection) {
			weekday := col + 1

			td.Find("div.kbcontent").Each(func(_ int, div *goquery.Selection) {
				name := s.extractCourseName(div)
				if name == "" || name == "&nbsp;" {
					return
				}

				var teacher, classroom, weeksStr string
				div.Find("font").Each(func(_ int, f *goquery.Selection) {
					title, _ := f.Attr("title")
					text := strings.TrimSpace(f.Text())
					switch {
					case strings.Contains(title, "老师"):
						teacher = text
					case strings.Contains(title, "周次"):
						weeksStr = text
					case strings.Contains(title, "教室"):
						classroom = text
					}
				})

				// 按周次过滤
				if weekNo > 0 && weeksStr != "" && !s.weekInWeeks(weekNo, weeksStr) {
					return
				}

				c := Course{
					Name:        name,
					Teacher:     teacher,
					Classroom:   classroom,
					Weekday:     weekday,
					StartPeriod: startP,
					EndPeriod:   endP,
				}

				days[weekday-1].Courses = append(days[weekday-1].Courses, c)
			})
		})
	})

	return &WeekSchedule{
		WeekNo:    weekNo,
		Starttime: "",
		Endtime:   "",
		Days:      days,
	}, nil
}

// parsePeriodRange 解析节次范围
func (s *courseServiceImpl) parsePeriodRange(text string) (int, int) {
	text = strings.TrimSpace(text)
	re := regexp.MustCompile(`\d+`)
	nums := re.FindAllString(text, -1)
	if len(nums) == 0 {
		return 0, 0
	}
	start, _ := strconv.Atoi(nums[0])
	end := start
	if len(nums) > 1 {
		end, _ = strconv.Atoi(nums[len(nums)-1])
	}
	return start, end
}

// extractCourseName 提取课程名称
func (s *courseServiceImpl) extractCourseName(div *goquery.Selection) string {
	name := ""
	div.Contents().EachWithBreak(func(i int, sel *goquery.Selection) bool {
		if goquery.NodeName(sel) == "#text" {
			t := strings.TrimSpace(sel.Text())
			if t != "" {
				name = t
				return false
			}
		}
		if goquery.NodeName(sel) == "br" {
			return false
		}
		return true
	})
	return name
}

// weekInWeeks 判断某周是否在周次范围内
func (s *courseServiceImpl) weekInWeeks(weekNo int, weeksStr string) bool {
	// 去掉 "(周)" 后缀
	if idx := strings.Index(weeksStr, "("); idx >= 0 {
		weeksStr = weeksStr[:idx]
	}
	weeksStr = strings.TrimSpace(weeksStr)
	if weeksStr == "" {
		return true
	}

	parts := strings.Split(weeksStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			se := strings.SplitN(part, "-", 2)
			if len(se) != 2 {
				continue
			}
			start, err1 := strconv.Atoi(strings.TrimSpace(se[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(se[1]))
			if err1 != nil || err2 != nil {
				continue
			}
			if weekNo >= start && weekNo <= end {
				return true
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			if weekNo == n {
				return true
			}
		}
	}
	return false
}
