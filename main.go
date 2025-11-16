package main

import (
	"log"
	"spider-go/internal/api"
	"spider-go/internal/app"
	"spider-go/internal/handler/controller"
	"spider-go/internal/handler/service"
	"spider-go/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化你的 DB/Repo（这里假设你已有 db 连接）
	userRepo := repository.NewGormUserRepository(app.InitDB())
	userSvc := service.NewUserService(userRepo)
	userController := controller.NewUserController(userSvc)
	gradeSvc := service.NewGradeService(userRepo)
	gradeController := controller.NewGradeController(gradeSvc)
	courseSvc := service.NewCourseService(userRepo)
	courseController := controller.NewCourseController(courseSvc)
	r := gin.Default()
	api.SetupRoutes(r, userController, gradeController, courseController)

	if err := app.LoadConfig(); err != nil {
		log.Fatalf("config error: %v \n", err)
	}
	app.InitDB()
	app.InitRedis()
	err := r.Run(":" + strconv.Itoa(app.Conf.App.Port))
	if err != nil {
		return
	}

}
