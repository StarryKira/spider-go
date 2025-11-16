package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"spider-go/internal/app"
	"spider-go/internal/repository"
	"spider-go/internal/utils"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type CourseService struct {
	uRepo repository.UserRepository
}

func NewCourseService(uRepo repository.UserRepository) *CourseService {
	return &CourseService{uRepo: uRepo}
}

type DaySchedule struct {
	Weekday int      `json:"weekday"` //值为1-7，表示周一到周日
	Courses []Course `json:"courses"` //当天课程 没有课则结构体则数组为nil
}

type WeekSchedule struct {
	WeekNo    int           `json:"weekno"`
	Starttime string        `json:"starttime"`
	Endtime   string        `json:"endtime"`
	Days      []DaySchedule `json:"days"`
}

type Course struct {
	Name        string `json:"name"`         // 课程名称
	Teacher     string `json:"teacher"`      // 任课老师
	Classroom   string `json:"classroom"`    // 教室：A1-203
	Weekday     int    `json:"weekday"`      // 周几：1~7 表示周一~周日
	StartPeriod int    `json:"start_period"` // 第几节开始：1 表示第一节
	EndPeriod   int    `json:"end_period"`   // 第几节结束：2 表示上到第二节
}

func (s *CourseService) GetCourseTableByWeek(week int, term string, uid int) (*WeekSchedule, error) {
	//校验请求体

	if week > 20 || week < 1 {
		return nil, errors.New("请求体无效")
	}

	if term == "" {
		return nil, errors.New("请求体无效")
	}

	re := regexp.MustCompile(`^\d{4}-\d{4}-[12]$`)
	if !re.MatchString(term) {
		return nil, errors.New("请求不合法")
	}

	//开始构造请求
	user, err := s.uRepo.GetUserByUid(uid)
	if err != nil {
		return nil, err
	}
	if user.Sid == "" || user.Spwd == "" {
		return nil, errors.New("请绑定教务系统")
	}
	isRedisHasCookie, err := app.Rdb.Exists(app.Ctx, strconv.Itoa(uid)).Result()
	if err != nil {
		return nil, errors.New("Redis错误")
	}

	var cookies []*http.Cookie
	if isRedisHasCookie > 0 {
		// 直接从 redis 读回 cookie 数组
		data, _ := app.Rdb.Get(app.Ctx, strconv.Itoa(uid)).Bytes()
		if err := json.Unmarshal(data, &cookies); err != nil {
			return nil, err
		}
	} else {
		// 登录 -> 跟随一次重定向拿 cookie -> 回读 redis
		err := utils.LoginAndStoreSession(uid, user.Sid, user.Spwd)
		if err != nil {
			return nil, err
		}
		data, _ := app.Rdb.Get(app.Ctx, strconv.Itoa(uid)).Bytes()
		if err := json.Unmarshal(data, &cookies); err != nil {
			return nil, err
		}
	}
	//构造请求体
	form := url.Values{}
	form.Add("zc", strconv.Itoa(week))
	form.Add("xnxq01id", term)

	req, err := http.NewRequest("POST", utils.Course_url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseCourseTableFromHTML(resp.Body)

}

func parseCourseTableFromHTML(r io.Reader) (*WeekSchedule, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.New("解析HTML失败")
	}
	title := strings.TrimSpace(doc.Find("title").Text())
	if title != "学期理论课表" {
		return nil, errors.New("解析HTML失败，HTML并非课表")
	}

	// 1. 解析当前周次 weekNo
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
		// 第 0 行是表头
		if i == 0 {
			return
		}

		thText := strings.TrimSpace(tr.Find("th").First().Text())
		if thText == "" {
			return
		}
		if strings.HasPrefix(thText, "备注") {
			// 最后一行备注
			return
		}

		startP, endP := parsePeriodRange(thText)
		if startP == 0 && endP == 0 {
			return
		}

		// 一行 7 列：周一到周日
		tr.Find("td").Each(func(col int, td *goquery.Selection) {
			weekday := col + 1 // 1~7

			// 每个 td 里有两个 div：kbcontent1(简略) 和 kbcontent(详细)
			td.Find("div.kbcontent").Each(func(_ int, div *goquery.Selection) {
				name := extractCourseName(div)
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

				// 按照当前周次过滤课程
				if weekNo > 0 && weeksStr != "" && !weekInWeeks(weekNo, weeksStr) {
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

	ws := &WeekSchedule{
		WeekNo:    weekNo,
		Starttime: "", // HTML 里没有周起止日期，这里先留空
		Endtime:   "",
		Days:      days,
	}

	return ws, nil
}

// 解析 "第1，2节" / "第3,4节" 这种行头，返回起止节次
func parsePeriodRange(s string) (int, int) {
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(`\d+`)
	nums := re.FindAllString(s, -1)
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

// 从 div.kbcontent 中获取课程名（即第一个文本节点，遇到 <br> 之前）
func extractCourseName(div *goquery.Selection) string {
	name := ""
	div.Contents().EachWithBreak(func(i int, s *goquery.Selection) bool {
		if goquery.NodeName(s) == "#text" {
			t := strings.TrimSpace(s.Text())
			if t != "" {
				name = t
				return false
			}
		}
		if goquery.NodeName(s) == "br" {
			return false
		}
		return true
	})
	return name
}

// 判断某一周 weekNo 是否包含在 "1-8,10-17(周)" 这种周次字符串里
func weekInWeeks(weekNo int, weeksStr string) bool {
	// 去掉 "(周)" 以及后面的任何内容
	if idx := strings.Index(weeksStr, "("); idx >= 0 {
		weeksStr = weeksStr[:idx]
	}
	weeksStr = strings.TrimSpace(weeksStr)
	if weeksStr == "" {
		return true // 没写周次就默认都上
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
