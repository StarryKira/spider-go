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

		// 通知接口（公开 - 查看）
		api.GET("/notices", container.NoticeController.GetVisibleNotices) // 获取可见通知

		// 系统配置（公开 - 查看）
		api.GET("/config/term", container.ConfigController.GetCurrentTerm)             // 获取当前学期
		api.GET("/config/semester-dates", container.ConfigController.GetSemesterDates) // 获取学期开学和放假时间
	}

	// 需要认证的接口（普通用户）
	user := api.Group("/user")
	user.Use(middleware.AuthMiddleWare(secret, container.DAUService))
	{
		user.POST("/bind", container.UserController.BindJwcAccount) // 绑定教务系统账号
		user.GET("/info", container.UserController.GetUserInfo)     // 获取用户信息
		user.GET("/isbind", container.UserController.CheckIsBind)   // 检查是否绑定教务系统账号

		// 成绩相关接口（RESTful 规范）
		user.GET("/grades", container.GradeController.GetGrades)                               // 获取成绩（query: term 可选）
		user.GET("/grades/level", container.GradeController.GetLevelGrade)                     // 获取等级考试成绩
		user.GET("/grades/analysis", container.GradeAnalysisController.GetRecentTermsAnalysis) // 获取成绩分析

		// 课程和考试（RESTful 规范）
		user.GET("/courses", container.CourseController.GetCourses) // 获取课程表（query: week, term）
		user.GET("/exams", container.ExamController.GetExams)       // 获取考试安排（query: term）
	}

	// 管理员接口
	admin := api.Group("/admin")
	{
		// 管理员登录（公开）
		admin.POST("/login", container.AdminController.Login)

		// 需要管理员认证的接口
		adminAuth := admin.Group("")
		adminAuth.Use(middleware.AdminAuthMiddleware(secret))
		{
			// 管理员信息
			adminAuth.GET("/info", container.AdminController.GetInfo)
			adminAuth.POST("/reset", container.AdminController.ChangePwd)
			// 通知管理
			adminAuth.POST("/notices", container.NoticeController.CreateNotice)        // 创建通知
			adminAuth.PUT("/notices/:nid", container.NoticeController.UpdateNotice)    // 更新通知
			adminAuth.DELETE("/notices/:nid", container.NoticeController.DeleteNotice) // 删除通知
			adminAuth.GET("/notices", container.NoticeController.GetAllNotices)        // 获取所有通知

			// 日活统计（仅管理员）
			adminAuth.GET("/statistics/dau", container.StatisticsController.GetDAUStatistics)  // 获取日活统计
			adminAuth.GET("/statistics/dau/range", container.StatisticsController.GetDAURange) // 获取日活范围统计

			// 系统配置（仅管理员）
			adminAuth.POST("/config/term", container.ConfigController.SetCurrentTerm)             // 设置当前学期
			adminAuth.POST("/config/semester-dates", container.ConfigController.SetSemesterDates) // 设置学期开学和放假时间
		}
	}
}
