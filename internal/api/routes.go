package api

import (
	"spider-go/internal/handler/controller"
	"spider-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

var secret = []byte("Haruka")

func SetupRoutes(r *gin.Engine, uc *controller.UserController, gc *controller.GradeController) {
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello world",
		})
	})
	api := r.Group("/api")

	// api路由
	api.POST("/login", uc.Login)
	api.POST("/register", uc.Register)

	user := api.Group("/user")
	user.Use(middleware.AuthMiddleWare(secret))
	user.POST("/bind", uc.BindJwcAccount)   //绑定校园网账号
	user.GET("/info", uc.GetUserInfo)       //获取用户信息
	user.GET("/grades/all", gc.GetAllGrade) //获取全部成绩

}
