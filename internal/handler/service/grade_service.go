package service

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
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
	Status   int     `json:"Status"`   //正常考试，重修为0，补考为1
	Property string  `json:"property"` // 课程性质/属性（见下方说明）
}

type GPA struct {
	AverageGPA   float64 `json:"averageGPA"`
	AverageScore float64 `json:"averageScore"`
	BasicScore   float64 `json:"basicScore"`
}
type LevelGrade struct {
	No         string `json:"no"`
	CourseName string `json:"CourseName"`
	LevGrade   string `json:"LevelGrade"`
	Time       string `json:"Time"`
}

func NewGradeService(uRepo repository.UserRepository) *GradeService {
	return &GradeService{uRepo: uRepo}
}

// 获取所有考试成绩
func (s *GradeService) GetAllGrade(uid int) ([]Grade, *GPA, error) {
	user, err := s.uRepo.GetUserByUid(uid)
	if err != nil {
		return nil, nil, err
	}
	if user.Sid == "" || user.Spwd == "" {
		return nil, nil, errors.New("请绑定教务系统")
	}

	// 1) 取/刷新 cookie
	isRedisHasCookie, err := app.Rdb.Exists(app.Ctx, strconv.Itoa(uid)).Result()
	if err != nil {
		return nil, nil, errors.New("Redis错误")
	}

	var cookies []*http.Cookie
	if isRedisHasCookie > 0 {
		// 直接从 redis 读回 cookie 数组
		data, _ := app.Rdb.Get(app.Ctx, strconv.Itoa(uid)).Bytes()
		if err := json.Unmarshal(data, &cookies); err != nil {
			return nil, nil, err
		}
	} else {
		// 登录 -> 跟随一次重定向拿 cookie -> 回读 redis
		err := utils.LoginAndStoreSession(uid, user.Sid, user.Spwd)
		if err != nil {
			return nil, nil, err
		}
		data, _ := app.Rdb.Get(app.Ctx, strconv.Itoa(uid)).Bytes()
		if err := json.Unmarshal(data, &cookies); err != nil {
			return nil, nil, err
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
		return nil, nil, err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	GradeList, err3 := parseGradesFromHTML(resp.Body)
	if err3 != nil {
		return nil, nil, err3
	}
	gpa, err4 := CalculateGPA(GradeList)
	if err4 != nil {
		return nil, nil, err4
	}
	return GradeList, gpa, nil
}

// 根据学期获取考试成绩
func (s *GradeService) GetGradeByTerm(uid int, term string) ([]Grade, *GPA, error) {
	//校验输入term规则
	re := regexp.MustCompile(`^\d{4}-\d{4}-[12]$`)
	if !re.MatchString(term) {
		return nil, nil, errors.New("请求不合法")
	}

	user, err := s.uRepo.GetUserByUid(uid)
	if err != nil {
		return nil, nil, err
	}
	if user.Sid == "" || user.Spwd == "" {
		return nil, nil, errors.New("请绑定教务系统")
	}

	// 1) 取/刷新 cookie
	isRedisHasCookie, err := app.Rdb.Exists(app.Ctx, strconv.Itoa(uid)).Result()
	if err != nil {
		return nil, nil, errors.New("Redis错误")
	}

	var cookies []*http.Cookie
	if isRedisHasCookie > 0 {
		// 直接从 redis 读回 cookie 数组
		data, _ := app.Rdb.Get(app.Ctx, strconv.Itoa(uid)).Bytes()
		if err := json.Unmarshal(data, &cookies); err != nil {
			return nil, nil, err
		}
	} else {
		// 登录 -> 跟随一次重定向拿 cookie -> 回读 redis
		err := utils.LoginAndStoreSession(uid, user.Sid, user.Spwd)
		if err != nil {
			return nil, nil, err
		}
		data, _ := app.Rdb.Get(app.Ctx, strconv.Itoa(uid)).Bytes()
		if err := json.Unmarshal(data, &cookies); err != nil {
			return nil, nil, err
		}
	}

	// 2) 构造请求
	form := url.Values{}
	form.Set("kksj", term)
	form.Set("kcxz", "")
	form.Set("kcmc", "")
	form.Set("xsfs", "all")

	req, err := http.NewRequest("POST", utils.Grade_url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	GradeList, err3 := parseGradesFromHTML(resp.Body)
	if err3 != nil {
		return nil, nil, err3
	}
	gpa, err4 := CalculateGPA(GradeList)
	if err4 != nil {
		return nil, nil, err4
	}
	return GradeList, gpa, nil
}

// 获取等级考试成绩，四六级那些
func (s *GradeService) GetLevelGrades(uid int) ([]LevelGrade, error) {

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
	req, err := http.NewRequest("GET", utils.Grade_level_url, nil)
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
	return parseLevelGradesFromHTML(resp.Body)

}
func parseLevelGradesFromHTML(html io.Reader) ([]LevelGrade, error) {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return nil, err
	}
	table := doc.Find("#dataList")
	if table.Length() == 0 {
		return nil, errors.New("未解析到datalist")
	}
	var LevelGrades []LevelGrade
	table.Find("tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 9 {
			return
		}
		trim := func(s string) string { return strings.ReplaceAll(s, "\u00A0", "") }
		No := trim(tds.Eq(0).Text())
		CourseName := trim(tds.Eq(1).Text())
		//处理分数类成绩和等级类成绩，只输出等级类成绩
		var LevGrade string = ""
		if trim(tds.Eq(4).Text()) == "" {
			LevGrade = trim(tds.Eq(7).Text())
		} else {
			LevGrade = trim(tds.Eq(4).Text())
		}
		Time := trim(tds.Eq(8).Text())
		LevelGrades = append(LevelGrades, LevelGrade{
			No:         No,
			CourseName: CourseName,
			LevGrade:   LevGrade,
			Time:       Time,
		})
	})
	return LevelGrades, nil
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

// 解析HTML 这个解析的是成绩的HTML不是等级考试的
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

		//处理status
		var statusNormalRegexp = regexp.MustCompile(`^正常考试$|.*重.*`)

		var status int
		if statusNormalRegexp.MatchString(trim(tds.Eq(10).Text())) {
			status = 0 // 正常或重修
		} else {
			status = 1
		} // 其它（如 补考/缓考 等）
		property := trim(tds.Eq(11).Text())
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
			Status:   status,
			Property: property,
		})
	})

	if len(grades) == 0 {
		return nil, errors.New("未解析到成绩，检查是否被重定向或选择器/编码不匹配")
	}
	return grades, nil
}

func CalculateGPA(gradeArray []Grade) (*GPA, error) {
	distinct := distinctGrades(gradeArray)

	var (
		sumScore  float64 // APF 分子（按门数）
		sumGp     float64 // GPA 分子（加权）
		sumCredit float64 // GPA 分母
		num2      int     // APF 分母（门数）

		sumScore2  float64 // BasicPoint 分子：∑(score * credit)
		sumCredit2 float64 // BasicPoint 分母：∑credit
	)

	for _, g := range distinct {
		// 只统计必修
		if g.Property != "必修" {
			continue
		}

		scoreText := g.Score

		// ---------- BasicPoint：只统计 必修 + Status==0 ----------
		if g.Status == 0 {
			gradeD := mapGradeToScoreForBasic(scoreText)
			sumScore2 += gradeD * g.Credit
			sumCredit2 += g.Credit
		}

		// ---------- GPA & APF ----------
		numericScore, isNum := parseNumeric(scoreText)

		if isNum && g.Status == 0 && numericScore >= 59.9 {
			// 正常考试 + 数字及格
			sumScore += numericScore

			gp := getCourseGp(g, scoreText)
			sumGp += gp * g.Credit
			sumCredit += g.Credit
			num2++
		} else {
			if g.Status == 0 && !isNum {
				// 正常考试 + 等级
				gp := getCourseGp(g, scoreText)
				score := gp*10.0 + 50.0 // 优=90 良=80 中=70 及格/合格=60 不及格/不合格=50

				sumScore += score
				sumGp += gp * g.Credit
				sumCredit += g.Credit
				num2++
			} else {
				// Status == 1（补考/重修等）
				if g.Status == 1 && isNum && numericScore >= 59.9 {
					// 补考/重修及格（数字）
					sumScore += 60.0 // APF 固定 60
					gp := getCourseGp(g, scoreText)
					sumGp += gp * 1.0 // ★ 与 Java 一致：只乘 1.0
					sumCredit += g.Credit
					num2++
				} else if g.Status == 1 && !isNum && (scoreText == "及格" || scoreText == "合格") {
					// 补考/重修及格（等级）
					gp := getCourseGp(g, scoreText)
					sumScore += 60.0
					sumGp += gp * 1.0 // ★ 只乘 1.0
					sumCredit += g.Credit
					num2++
				} else if g.Status == 1 && !isNum && (scoreText == "不及格" || scoreText == "不合格") {
					// 补考/重修不及格（等级）
					sumCredit += g.Credit
					num2++
				} else if g.Status == 1 && isNum && numericScore <= 59.9 {
					// 补考/重修不及格（数字）
					sumCredit += g.Credit
					num2++
				} else {
					// 兜底：正常或重修不及格的其它样式
					sumCredit += g.Credit
					num2++
					if isNum {
						sumScore += numericScore
					} else {
						log.Println("特殊成绩样式:", scoreText)
					}
				}
			}
		}
	}

	// ---------- 汇总 ----------
	var gpa, apf, basic float64
	if sumCredit != 0 {
		gpa = sumGp / sumCredit
	}
	if num2 != 0 {
		apf = sumScore / float64(num2)
	}
	if sumCredit2 != 0 {
		basic = sumScore2 / sumCredit2
	}

	if math.IsNaN(gpa) {
		gpa = 0
	}
	if math.IsNaN(apf) {
		apf = 0
	}
	if math.IsNaN(basic) {
		basic = 0
	}

	return &GPA{
		AverageGPA:   round3(gpa),
		AverageScore: round3(apf),
		BasicScore:   round3(basic),
	}, nil
}

// ----------------- helpers -----------------

func distinctGrades(grades []Grade) []Grade {
	m := make(map[string]Grade)
	for _, g := range grades {
		key := g.SerialNo + "|" + g.Code + "|" + g.Term
		m[key] = g
	}
	res := make([]Grade, 0, len(m))
	for _, g := range m {
		res = append(res, g)
	}
	return res
}

func parseNumeric(s string) (float64, bool) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func mapGradeToScoreForBasic(scoreText string) float64 {
	switch scoreText {
	case "不及格", "不合格":
		return 50.0
	case "及格", "合格":
		return 60.0
	case "中":
		return 70.0
	case "良":
		return 80.0
	case "优":
		return 90.0
	default:
		if v, ok := parseNumeric(scoreText); ok {
			return v
		}
		return 0
	}
}

// ★ 修正点：GPA 为 0 也视为“有值”（不再要求 >0）
func getCourseGp(g Grade, scoreText string) float64 {
	if !math.IsNaN(g.Gpa) {
		return g.Gpa
	}
	return handelGp(scoreText)
}

// 与 Java 的 handelGp 一致
func handelGp(scoreText string) float64 {
	switch scoreText {
	case "不及格", "不合格":
		return 0
	case "及格", "合格":
		return 1.0
	case "中":
		return 2.0
	case "良":
		return 3.0
	case "优":
		return 4.0
	}
	// 数字分
	score, ok := parseNumeric(scoreText)
	if !ok {
		log.Println("额外成绩样式:", scoreText)
		return 0
	}
	raw := (score - 50.0) / 10.0
	raw = round3(raw)
	if raw <= 0.1 {
		return 0
	}
	return raw
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}
