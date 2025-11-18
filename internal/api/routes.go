package api

import (
	"spider-go/internal/app"
	"spider-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, container *app.Container) {
	// JWT secret
	secret := []byte(container.Config.JWT.Secret)

	api := r.Group("/api")
	{
		// 公开接口
		api.POST("/login", container.UserController.Login)
		api.POST("/register", container.UserController.Register)
		api.POST("/reset", container.UserController.ResetPassword)

		// 验证码接口（公开）
		api.POST("/captcha/send", container.CaptchaController.SendEmailCaptcha) // 发送验证码
	}

	// 需要认证的接口
	user := api.Group("/user")
	user.Use(middleware.AuthMiddleWare(secret))
	{
		user.POST("/bind", container.UserController.BindJwcAccount)          // 绑定教务系统账号
		user.GET("/info", container.UserController.GetUserInfo)              // 获取用户信息
		user.GET("/grades/all", container.GradeController.GetAllGrade)       // 获取所有成绩
		user.GET("/grades/term", container.GradeController.GetGradeByTerm)   // 根据学期获取成绩
		user.GET("/grades/level", container.GradeController.GetLevelGrade)   // 获取等级考试成绩
		user.GET("/course/:week", container.CourseController.GetCourseTable) // 获取课程表
		user.GET("/exam", container.ExamController.GetExams)                 // 获取考试安排
	}
}
