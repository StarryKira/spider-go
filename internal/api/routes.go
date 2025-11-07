package api

import (
	"spider-go/internal/handler/controller"
	"spider-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

var secret = []byte("Haruka")

func SetupRoutes(r *gin.Engine, uc *controller.UserController) {
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello world",
		})
	})
	api := r.Group("/api")

	//test
	api.POST("/jwclogin", controller.Jwclogin)
	//endtest

	api.POST("login", uc.Login)
	api.POST("/register", uc.Register)
	user := api.Group("/user")
	user.Use(middleware.AuthMiddleWare(secret))
}
