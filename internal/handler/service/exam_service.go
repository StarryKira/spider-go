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

type ExamService struct {
	uRepo repository.UserRepository
}

func NewExamService(uRepo repository.UserRepository) *ExamService {
	return &ExamService{uRepo: uRepo}
}

type ExamArrangement struct {
	SerialNo  string `json:"serial_no"`
	ClassNo   string `json:"class_no"`
	ClassName string `json:"class_name"`
	Time      string `json:"time"`
	Place     string `json:"place"`
	Execution string `json:"execution"`
}

func (s *ExamService) GetAllExams(uid int, term string) ([]ExamArrangement, error) {

	re := regexp.MustCompile(`^\d{4}-\d{4}-[12]$`)
	if !re.MatchString(term) {
		return nil, errors.New("请求不合法")
	}
	//和所有service一样，构造请求
	//开始构造请求
	user, err := s.uRepo.GetUserByUid(uid)
	if err != nil {
		return nil, errors.New("数据库错误")
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
	form.Add("xnxqid", term)

	req, err := http.NewRequest("POST", utils.Exam_url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.New("教务系统炸了")
	}
	//构造cookie
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

	return parseExamArrangementFromHtml(resp.Body)
}

func parseExamArrangementFromHtml(r io.Reader) ([]ExamArrangement, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.New("HTML无效或HTML为空")
	}

	title := strings.TrimSpace(doc.Find("title").Text())
	if title != "我的考试 - 考试安排查询" {
		return nil, errors.New("页面错误，并非考试页面，请检查是否登录")
	}

	// table id 是 dataList （注意 D 大写）
	table := doc.Find("#dataList")
	if table.Length() == 0 {
		return nil, errors.New("未找到考试安排表格")
	}

	rows := table.Find("tr")
	if rows.Length() <= 1 {
		return nil, nil // 只有表头视为无数据
	}

	// 如果第二行含有“未查询到数据”则返回空
	if strings.Contains(rows.Eq(1).Text(), "未查询到数据") {
		return nil, nil
	}

	var exams []ExamArrangement

	rows.Each(func(i int, tr *goquery.Selection) {
		// 跳过表头
		if i == 0 {
			return
		}

		tds := tr.Find("td")
		// 这页每行一定是 9 列
		if tds.Length() < 9 {
			return
		}

		trim := func(s string) string {
			s = strings.TrimSpace(s)
			s = strings.ReplaceAll(s, "\u00A0", "") // 清除 &nbsp;
			return s
		}

		exams = append(exams, ExamArrangement{
			SerialNo: trim(tds.Eq(0).Text()), // 序号
			// tds[1] 是考试场次（你原代码不要，仍不取）
			ClassNo:   trim(tds.Eq(2).Text()), // 课程编号
			ClassName: trim(tds.Eq(3).Text()), // 课程名称
			Time:      trim(tds.Eq(4).Text()), // 考试时间
			Place:     trim(tds.Eq(5).Text()), // 考场（可能多个）

			Execution: trim(tds.Eq(8).Text()), // 操作（随堂考试 / 正考 / 补考）
		})
	})

	return exams, nil
}
