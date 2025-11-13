package service

import (
	"encoding/json"
	"errors"
	"io"
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

type GPA struct {
	AverageGPA   float64 `json:"averageGPA"`
	AverageScore float64 `json:"averageScore"`
	BasicScore   float64 `json:"basicScore"`
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
		err := utils.LoginAndStoreSession(uid, user.Sid, user.Spwd)
		if err != nil {
			return nil, err
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

	return parseGradesFromHTML(resp.Body)
}

// 兼容神人教务系统用的
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

// 解析HTML
func parseGradesFromHTML(r io.Reader) ([]Grade, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	table := doc.Find("#dataList")
	if table.Length() == 0 {
		return nil, errors.New("未找到 #dataList，可能未登录或被重定向到登录页")
	}

	var grades []Grade
	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 13 {
			return
		}
		trim := func(s string) string {
			return strings.TrimSpace(strings.ReplaceAll(s, "\u00A0", ""))
		}

		serialNo := trim(tds.Eq(0).Text())
		term := trim(tds.Eq(1).Text())
		code := trim(tds.Eq(2).Text())
		subject := trim(tds.Eq(3).Text())
		score := trim(tds.Eq(4).Text()) // <font> 或 <a> 里的文本都会被 .Text() 拿到

		credit := parseFloatSafe(trim(tds.Eq(5).Text()))
		gpa := parseFloatSafe(trim(tds.Eq(7).Text()))

		property := trim(tds.Eq(12).Text())
		if property == "" {
			property = trim(tds.Eq(11).Text())
		}

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
		return nil, errors.New("未解析到成绩，检查是否被重定向或选择器/编码不匹配")
	}
	return grades, nil
}

func (s *GradeService) calculateGPA(gradeArray []Grade) {
	var avarageScore float64
	var avarageGPA float64
	var basicScore float64
	//遍历gradeArray计算平均值
	for _, val := range gradeArray {
		int
	}
}
