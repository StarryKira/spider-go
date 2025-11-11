package service

import (
	"errors"
	"net/url"
	"spider-go/internal/app"
	"spider-go/internal/repository"
	"spider-go/internal/utils"
	"strconv"
)

type GradeService struct {
	uRepo repository.UserRepository
}
type Grade struct {
	SerialNo string  `json:"serialNo"`
	Term     string  `json:"Year"`
	Code     string  `json:"Code"`
	Subject  string  `json:"subject"`
	Score    float64 `json:"score"`  //成绩
	Credit   float64 `json:"credit"` //学分
	Gpa      float64 `json:"gpa"`
	Property string  `json:"property"` //课程类型，必修选修这种
}

func NewGradeService() *GradeService {
	return &GradeService{}
}

func (s *GradeService) GetAllGrade(uid int) ([]Grade, error) {
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
	//构造拿成绩的请求体
	form := url.Values{}
	form.Set("kksj", "")
	form.Set("kcxz", "")
	form.Set("kcmc", "")
	form.Set("xsfs", "all")

	//client := &http.Client{}
	//request, err := http.NewRequest("POST", utils.Grade_url, strings.NewReader(form.Encode()))
	//redis里存在
	if isRedisHasCookie > 0 {
		//cookie没过期，直接从redis里拿cookie，添加到请求里面去
		session := app.Rdb.Get(app.Ctx, strconv.Itoa(uid))
		println(session)
	} else {
		redirectlink, err1 := utils.Jwclogin(user.Sid, user.Spwd)
		if err1 != nil {
			return nil, err1
		}
		_, err := utils.HandleRedirect(uid, redirectlink)
		if err != nil {
			return nil, err
		}
		session := app.Rdb.Get(app.Ctx, strconv.Itoa(uid))
		println(session)
		//cookie过期，执行一次教务系统登录，然后再去redis里拿cookie
	}
	return nil, nil
}
