package api

import (
	"spider-go/internal/handler/controller"
	"spider-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

var secret = []byte("Haruka")

func SetupRoutes(r *gin.Engine, uc *controller.UserController, gc *controller.GradeController, cc *controller.CourseController, ec *controller.ExamController) {

	api := r.Group("/api")
	// api路由
	api.POST("/login", uc.Login)
	api.POST("/register", uc.Register)

	user := api.Group("/user")
	user.Use(middleware.AuthMiddleWare(secret))
	user.POST("/bind", uc.BindJwcAccount) //绑定校园网账号
	user.GET("/info", uc.GetUserInfo)     //获取用户信息

	user.GET("/grades/all", gc.GetAllGrade)     //获取全部成绩
	user.GET("/grades/term", gc.GetGradeByTerm) //根据学期获取成绩
	user.GET("/grades/level", gc.GetLevelGrade) //获取等级考试成绩

	user.GET("/course/:week", cc.GetCourseTable) //获取第 week 周的课程表

	user.GET("/exam", ec.GetExams) //根据学期获取考试安排
}
