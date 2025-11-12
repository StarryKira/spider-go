package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"spider-go/internal/app"
	"spider-go/internal/repository"
	"spider-go/internal/utils"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type GradeService struct {
	uRepo repository.UserRepository
}

type Grade struct {
	SerialNo string  `json:"serialNo"`
	Term     string  `json:"Year"`
	Code     string  `json:"Code"`
	Subject  string  `json:"subject"`
	Score    string  `json:"score"`    // 成绩
	Credit   float64 `json:"credit"`   // 学分
	Gpa      float64 `json:"gpa"`      // 绩点
	Property string  `json:"property"` // 课程性质/属性（见下方说明）
}

func NewGradeService(uRepo repository.UserRepository) *GradeService {
	return &GradeService{uRepo: uRepo}
}

func (s *GradeService) GetAllGrade(uid int) ([]Grade, error) {
	user, err := s.uRepo.GetUserByUid(uid)
	if err != nil {
		return nil, err
	}
	if user.Sid == "" || user.Spwd == "" {
		return nil, errors.New("请绑定教务系统")
	}

	// 1) 取/刷新 cookie
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
		redirectlink, err := utils.Jwclogin(user.Sid, user.Spwd)
		if err != nil {
			return nil, err
		}
		if name, err := utils.HandleRedirect(uid, redirectlink); err != nil {
			return nil, errors.New(name)
		}
		data, _ := app.Rdb.Get(app.Ctx, strconv.Itoa(uid)).Bytes()
		if err := json.Unmarshal(data, &cookies); err != nil {
			return nil, err
		}
	}

	// 2) 构造请求
	form := url.Values{}
	form.Set("kksj", "")
	form.Set("kcxz", "")
	form.Set("kcmc", "")
	form.Set("xsfs", "all")

	req, err := http.NewRequest("POST", utils.Grade_url, strings.NewReader(form.Encode()))
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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// 4) 解析表格
	var grades []Grade
	doc.Find("#dataList tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		// 至少 13 列：序号 开课学期 课程编号 课程名称 成绩 学分 总学时 绩点 成绩标志 考核方式 考试性质 课程属性 课程性质
		if tds.Length() < 13 {
			return
		}

		trim := func(s string) string {
			// 去除常见空白和全角空格
			return strings.TrimSpace(strings.ReplaceAll(s, "\u00A0", ""))
		}

		serialNo := trim(tds.Eq(0).Text())
		term := trim(tds.Eq(1).Text())
		code := trim(tds.Eq(2).Text())
		subject := trim(tds.Eq(3).Text())
		score := trim(tds.Eq(4).Text())

		creditStr := trim(tds.Eq(5).Text()) // 学分
		gpaStr := trim(tds.Eq(7).Text())    // 绩点

		// “课程属性/课程性质”在第 11/12 列（下标 11/12），你可以按需取其一
		property := trim(tds.Eq(12).Text())
		if property == "" {
			property = trim(tds.Eq(11).Text())
		}

		// 可能出现“及格/不及格”等字符串，无需转数值
		credit := parseFloatSafe(creditStr)
		gpa := parseFloatSafe(gpaStr)

		// 跳过明显空行
		if subject == "" && score == "" {
			return
		}

		grades = append(grades, Grade{
			SerialNo: serialNo,
			Term:     term,
			Code:     code,
			Subject:  subject,
			Score:    score,
			Credit:   credit,
			Gpa:      gpa,
			Property: property,
		})
	})

	if len(grades) == 0 {
		// 友好提示：可能是 cookie 失效或页面需要登录
		return nil, errors.New("未解析到成绩，请检查是否登录成功（cookie 是否有效）以及页面编码/选择器是否正确")
	}

	return grades, nil
}

func parseFloatSafe(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// 兼容“3.0 ”、“3,0”之类
	s = strings.ReplaceAll(s, ",", "")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
