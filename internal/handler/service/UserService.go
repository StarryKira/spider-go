package service

import (
	"errors"
	"spider-go/internal/model"
	"spider-go/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	uRepo repository.UserRepository
}

func NewUserService(repository repository.UserRepository) *UserService {
	return &UserService{uRepo: repository}
}

type Claims struct {
	Uid  int    `json:"user_id"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

func (s *UserService) UserLogin(Email, password string) (string, error) {
	user, err := s.uRepo.GetUserByEmail(Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", errors.New("invalid email or password")
	}

	claims := Claims{
		Uid:  user.Uid,
		Name: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 168)),
			Issuer:    "Haruka",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("Haruka"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *UserService) Register(name, email, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := model.User{Name: name, Email: email, Password: string(passwordHash), CreatedAt: time.Now()}
	err = s.uRepo.CreateUser(&u)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) Bind(c *gin.Context, sid, spwd string) (string, error) {

	if sid == "" || spwd == "" {
		return "", errors.New("invalid params")
	}
	uid, ok := c.Get("uid")
	if !ok {
		return "", errors.New("user id not found in context")
	}
	err := s.uRepo.UpdateJwc(uid.(int), sid, spwd)

	if err != nil {
		return "", errors.New("invalid sid or password")
	}

	return "success", nil

}
