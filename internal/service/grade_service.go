package service

import (
	"context"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"spider-go/internal/common"
	"spider-go/internal/repository"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GradeService 成绩服务接口
type GradeService interface {
	// GetAllGrade 获取所有成绩
	GetAllGrade(ctx context.Context, uid int) ([]Grade, *GPA, error)
	// GetGradeByTerm 根据学期获取成绩
	GetGradeByTerm(ctx context.Context, uid int, term string) ([]Grade, *GPA, error)
	// GetLevelGrades 获取等级考试成绩
	GetLevelGrades(ctx context.Context, uid int) ([]LevelGrade, error)
}

// Grade 成绩信息
type Grade struct {
	SerialNo string  `json:"serialNo"`
	Term     string  `json:"Year"`
	Code     string  `json:"Code"`
	Subject  string  `json:"subject"`
	Score    string  `json:"score"`
	Credit   float64 `json:"credit"`
	Gpa      float64 `json:"gpa"`
	Status   int     `json:"Status"`
	Property string  `json:"property"`
}

// GPA 绩点信息
type GPA struct {
	AverageGPA   float64 `json:"averageGPA"`
	AverageScore float64 `json:"averageScore"`
	BasicScore   float64 `json:"basicScore"`
}

// LevelGrade 等级考试成绩
type LevelGrade struct {
	No         string `json:"no"`
	CourseName string `json:"CourseName"`
	LevGrade   string `json:"LevelGrade"`
	Time       string `json:"Time"`
}

// gradeServiceImpl 成绩服务实现
type gradeServiceImpl struct {
	userRepo       repository.UserRepository
	sessionService SessionService
	crawlerService CrawlerService
	gradeURL       string
	gradeLevelURL  string
}

// NewGradeService 创建成绩服务
func NewGradeService(
	userRepo repository.UserRepository,
	sessionService SessionService,
	crawlerService CrawlerService,
	gradeURL string,
	gradeLevelURL string,
) GradeService {
	return &gradeServiceImpl{
		userRepo:       userRepo,
		sessionService: sessionService,
		crawlerService: crawlerService,
		gradeURL:       gradeURL,
		gradeLevelURL:  gradeLevelURL,
	}
}

// GetAllGrade 获取所有成绩
func (s *gradeServiceImpl) GetAllGrade(ctx context.Context, uid int) ([]Grade, *GPA, error) {
	// 1. 获取用户信息
	user, err := s.userRepo.GetUserByUid(uid)
	if err != nil {
		return nil, nil, common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}

	if user.Sid == "" || user.Spwd == "" {
		return nil, nil, common.NewAppError(common.CodeJwcNotBound, "")
	}

	// 2. 获取会话
	cookies, err := s.getCookiesOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, nil, err
	}

	// 3. 构造请求
	form := url.Values{}
	form.Set("kksj", "")
	form.Set("kcxz", "")
	form.Set("kcmc", "")
	form.Set("xsfs", "all")

	// 4. 发起请求
	body, err := s.crawlerService.FetchWithCookies(ctx, "POST", s.gradeURL, cookies, form)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	// 5. 解析成绩
	gradeList, err := s.parseGradesFromHTML(body)
	if err != nil {
		return nil, nil, err
	}

	// 6. 计算 GPA
	gpa, err := s.calculateGPA(gradeList)
	if err != nil {
		return nil, nil, err
	}

	return gradeList, gpa, nil
}

// GetGradeByTerm 根据学期获取成绩
func (s *gradeServiceImpl) GetGradeByTerm(ctx context.Context, uid int, term string) ([]Grade, *GPA, error) {
	// 1. 校验参数
	re := regexp.MustCompile(`^\d{4}-\d{4}-[12]$`)
	if !re.MatchString(term) {
		return nil, nil, common.NewAppError(common.CodeJwcInvalidParams, "学期格式错误")
	}

	// 2. 获取用户信息
	user, err := s.userRepo.GetUserByUid(uid)
	if err != nil {
		return nil, nil, common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}

	if user.Sid == "" || user.Spwd == "" {
		return nil, nil, common.NewAppError(common.CodeJwcNotBound, "")
	}

	// 3. 获取会话
	cookies, err := s.getCookiesOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, nil, err
	}

	// 4. 构造请求
	form := url.Values{}
	form.Set("kksj", term)
	form.Set("kcxz", "")
	form.Set("kcmc", "")
	form.Set("xsfs", "all")

	// 5. 发起请求
	body, err := s.crawlerService.FetchWithCookies(ctx, "POST", s.gradeURL, cookies, form)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()

	// 6. 解析成绩
	gradeList, err := s.parseGradesFromHTML(body)
	if err != nil {
		return nil, nil, err
	}

	// 7. 计算 GPA
	gpa, err := s.calculateGPA(gradeList)
	if err != nil {
		return nil, nil, err
	}

	return gradeList, gpa, nil
}

// GetLevelGrades 获取等级考试成绩
func (s *gradeServiceImpl) GetLevelGrades(ctx context.Context, uid int) ([]LevelGrade, error) {
	// 1. 获取用户信息
	user, err := s.userRepo.GetUserByUid(uid)
	if err != nil {
		return nil, common.NewAppError(common.CodeUserNotFound, "用户不存在")
	}

	if user.Sid == "" || user.Spwd == "" {
		return nil, common.NewAppError(common.CodeJwcNotBound, "")
	}

	// 2. 获取会话
	cookies, err := s.getCookiesOrLogin(ctx, uid, user.Sid, user.Spwd)
	if err != nil {
		return nil, err
	}

	// 3. 发起请求
	body, err := s.crawlerService.FetchWithCookies(ctx, "GET", s.gradeLevelURL, cookies, nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// 4. 解析成绩
	return s.parseLevelGradesFromHTML(body)
}

// getCookiesOrLogin 获取缓存的 cookies 或登录
func (s *gradeServiceImpl) getCookiesOrLogin(ctx context.Context, uid int, sid, spwd string) ([]*http.Cookie, error) {
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

// parseGradesFromHTML 解析成绩 HTML
func (s *gradeServiceImpl) parseGradesFromHTML(r io.Reader) ([]Grade, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "解析HTML失败")
	}

	table := doc.Find("#dataList")
	if table.Length() == 0 {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "未找到成绩数据")
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
		score := trim(tds.Eq(4).Text())
		credit := parseFloatSafe(trim(tds.Eq(5).Text()))
		gpa := parseFloatSafe(trim(tds.Eq(7).Text()))

		// 处理 status
		statusNormalRegexp := regexp.MustCompile(`^正常考试$|.*重.*`)
		var status int
		if statusNormalRegexp.MatchString(trim(tds.Eq(10).Text())) {
			status = 0
		} else {
			status = 1
		}

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
		return nil, common.NewAppError(common.CodeJwcParseFailed, "未解析到成绩")
	}

	return grades, nil
}

// parseLevelGradesFromHTML 解析等级考试成绩 HTML
func (s *gradeServiceImpl) parseLevelGradesFromHTML(r io.Reader) ([]LevelGrade, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "解析HTML失败")
	}

	table := doc.Find("#dataList")
	if table.Length() == 0 {
		return nil, common.NewAppError(common.CodeJwcParseFailed, "未找到等级考试数据")
	}

	var levelGrades []LevelGrade
	table.Find("tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 9 {
			return
		}

		trim := func(s string) string {
			return strings.ReplaceAll(s, "\u00A0", "")
		}

		no := trim(tds.Eq(0).Text())
		courseName := trim(tds.Eq(1).Text())

		// 处理分数类和等级类成绩
		var levGrade string
		if trim(tds.Eq(4).Text()) == "" {
			levGrade = trim(tds.Eq(7).Text())
		} else {
			levGrade = trim(tds.Eq(4).Text())
		}

		time := trim(tds.Eq(8).Text())

		levelGrades = append(levelGrades, LevelGrade{
			No:         no,
			CourseName: courseName,
			LevGrade:   levGrade,
			Time:       time,
		})
	})

	return levelGrades, nil
}

// calculateGPA 计算 GPA
func (s *gradeServiceImpl) calculateGPA(gradeArray []Grade) (*GPA, error) {
	distinct := s.distinctGrades(gradeArray)

	var (
		sumScore   float64
		sumGp      float64
		sumCredit  float64
		num2       int
		sumScore2  float64
		sumCredit2 float64
	)

	for _, g := range distinct {
		if g.Property != "必修" {
			continue
		}

		scoreText := g.Score

		// BasicPoint
		if g.Status == 0 {
			gradeD := mapGradeToScoreForBasic(scoreText)
			sumScore2 += gradeD * g.Credit
			sumCredit2 += g.Credit
		}

		// GPA & APF
		numericScore, isNum := parseNumeric(scoreText)

		if isNum && g.Status == 0 && numericScore >= 59.9 {
			sumScore += numericScore
			gp := s.getCourseGp(g, scoreText)
			sumGp += gp * g.Credit
			sumCredit += g.Credit
			num2++
		} else {
			if g.Status == 0 && !isNum {
				gp := s.getCourseGp(g, scoreText)
				score := gp*10.0 + 50.0
				sumScore += score
				sumGp += gp * g.Credit
				sumCredit += g.Credit
				num2++
			} else {
				if g.Status == 1 && isNum && numericScore >= 59.9 {
					sumScore += 60.0
					gp := s.getCourseGp(g, scoreText)
					sumGp += gp * 1.0
					sumCredit += g.Credit
					num2++
				} else if g.Status == 1 && !isNum && (scoreText == "及格" || scoreText == "合格") {
					gp := s.getCourseGp(g, scoreText)
					sumScore += 60.0
					sumGp += gp * 1.0
					sumCredit += g.Credit
					num2++
				} else if g.Status == 1 && !isNum && (scoreText == "不及格" || scoreText == "不合格") {
					sumCredit += g.Credit
					num2++
				} else if g.Status == 1 && isNum && numericScore <= 59.9 {
					sumCredit += g.Credit
					num2++
				} else {
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

// distinctGrades 去重成绩
func (s *gradeServiceImpl) distinctGrades(grades []Grade) []Grade {
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

// getCourseGp 获取课程绩点
func (s *gradeServiceImpl) getCourseGp(g Grade, scoreText string) float64 {
	if !math.IsNaN(g.Gpa) {
		return g.Gpa
	}
	return handelGp(scoreText)
}

// 辅助函数
func parseFloatSafe(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", "")
	v, _ := strconv.ParseFloat(s, 64)
	return v
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
