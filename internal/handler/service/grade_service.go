package service

import (
	"errors"
	"spider-go/internal/app"
	"spider-go/internal/repository"
	"spider-go/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GradeService struct {
	uRepo repository.UserRepository
}
type Grade struct {
	SerialNo string  `json:"serialNo"`
	Term     string  `json:"Year"`
	Code     string  `json:"Code"`
	Subject  string  `json:"subject"`
	Score    float64 `json:"score"`
	Credit   float64 `json:"credit"`
}

func NewGradeService() *GradeService {
	return &GradeService{}
}

func (s *GradeService) GetAllGrade(c *gin.Context) ([]Grade, error) {
	uid := c.GetInt("uid")
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

	//redis里存在
	if isRedisHasCookie > 0 {

	}
}
