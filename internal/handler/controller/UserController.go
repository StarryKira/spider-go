package controller

import (
	"spider-go/internal/dto"
	"spider-go/internal/handler/service"
	"time"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userSvc *service.UserService
}

func NewUserController(userSvc *service.UserService) *UserController {
	return &UserController{userSvc: userSvc}
}

func (h *UserController) Login(c *gin.Context) {
	var req dto.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 这里必须传 &req，否则绑定不到值
		dto.BadRequest(c, 40001, err.Error())
		return
	}

	token, err := h.userSvc.UserLogin(req.Email, req.Password)
	if err != nil {
		// 不暴露细节，统一提示
		dto.Unauthorized(c, 40101, "invalid email or password")
		return
	}

	maxAge := int((168 * time.Hour).Seconds()) // 与 service 里 7 天过期一致
	c.SetCookie("access_token", token, maxAge, "/", "", true, true)

	dto.Success(c, gin.H{
		"token": token,
	})
}

func (h *UserController) Register(c *gin.Context) {
	var req dto.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, 40002, err.Error())
		return
	}

	if err := h.userSvc.Register(req.Name, req.Email, req.Password); err != nil {
		dto.BadRequest(c, 40003, err.Error())
		return
	}

	dto.Success(c, gin.H{"message": "registered"})
}

func (h *UserController) BindJwcAccount(c *gin.Context) {
	var req dto.BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, 40002, err.Error())
		return
	}
	ret, err := h.userSvc.Bind(c, req.Sid, req.Spwd)
	if err != nil {
		dto.BadRequest(c, 40004, err.Error())
		return
	}
	dto.Success(c, gin.H{
		"msg": ret,
	})
}

func (h *UserController) GetUserInfo(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		dto.BadRequest(c, 114514, "invalid token")
		return
	}
	user, err := h.userSvc.GetUserInfo(uid.(int))
	if err != nil {
		dto.BadRequest(c, 40005, err.Error())
		return
	}
	dto.Success(c, user)
}
