package api

import (
	"net/http"
	"spider-go/internal/app"
	"spider-go/internal/middleware"
	"spider-go/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, container *app.Container) {
	// JWT secret
	secret := []byte(container.Config.JWT.Secret)

	// API 根路由组
	api := r.Group("/api")
	{
		// ========== 公开接口 ==========
		// 验证码接口（公开）
		captchaGroup := api.Group("/captcha")
		{
			captchaGroup.POST("/send", sendEmailCaptchaHandler(container.CaptchaService))
		}

		// 系统配置（公开 - 只读）
		// TODO: 需要重新实现这些端点以匹配实际的服务接口
		// configGroup := api.Group("/config")
		// {
		// 	configGroup.GET("/term", getCurrentTermHandler(container.ConfigCache))
		// 	configGroup.GET("/semester-dates", getSemesterDatesHandler(container.ConfigCache))
		// }

		// ========== 用户路由 ==========
		// 需要认证的用户接口
		userAuth := api.Group("/user")
		userAuth.Use(middleware.AuthMiddleWare(secret, container.DAUService))

		// 注册用户模块路由（包含公开和认证路由）
		container.UserModule.RegisterRoutes(api, userAuth)

		// ========== 管理员路由 ==========
		// 需要管理员认证的接口
		adminAuth := api.Group("/admin")
		adminAuth.Use(middleware.AdminAuthMiddleware(secret))

		// 注册管理员模块路由（包含公开和认证路由）
		container.AdminModule.RegisterRoutes(api, adminAuth)

		// 管理员专属功能 - 统计和配置
		// TODO: 需要重新实现这些端点以匹配实际的服务接口
		// adminAuth.GET("/statistics/dau", getDAUStatisticsHandler(container.DAUService))
		// adminAuth.GET("/statistics/dau/range", getDAURangeHandler(container.DAUService))
		// adminAuth.POST("/config/term", setCurrentTermHandler(container.ConfigCache))
		// adminAuth.POST("/config/semester-dates", setSemesterDatesHandler(container.ConfigCache))

		// ========== 业务模块路由 ==========
		// 成绩模块
		container.GradeModule.RegisterRoutes(userAuth)

		// 课程模块
		container.CourseModule.RegisterRoutes(userAuth)

		// 考试模块
		container.ExamModule.RegisterRoutes(userAuth)

		// 通知模块（包含公开和管理员路由）
		container.NoticeModule.RegisterRoutes(api, adminAuth)
	}
}

// ========== 验证码处理器 ==========

type sendEmailCaptchaRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func sendEmailCaptchaHandler(captchaService service.CaptchaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req sendEmailCaptchaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		if err := captchaService.SendEmailCaptcha(c.Request.Context(), req.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证码失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
	}
}

// ========== DAU 统计处理器 ==========
// TODO: 需要重新实现以匹配实际的服务接口

// ========== 系统配置处理器 ==========
// TODO: 需要重新实现以匹配实际的服务接口
