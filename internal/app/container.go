package app

import (
	"context"
	"fmt"
	"log"
	"spider-go/internal/cache"
	"spider-go/internal/controller"
	"spider-go/internal/repository"
	"spider-go/internal/service"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	Config *Config
	DB     *gorm.DB

	// Redis 客户端（同一个 Redis 服务器的不同数据库）
	SessionRedis *redis.Client // 会话 Redis (DB 0)
	CaptchaRedis *redis.Client // 验证码 Redis (DB 1)

	// Repositories
	UserRepo   repository.UserRepository
	AdminRepo  repository.AdminRepository
	NoticeRepo repository.NoticeRepository

	// Caches
	SessionCache  cache.SessionCache
	CaptchaCache  cache.CaptchaCache
	DAUCache      cache.DAUCache
	ConfigCache   cache.ConfigCache
	UserDataCache cache.UserDataCache

	// Services
	RSAKeyService        service.RSAKeyService
	SessionService       service.SessionService
	CrawlerService       service.CrawlerService
	EmailService         service.EmailService
	CaptchaService       service.CaptchaService
	DAUService           service.DAUService
	AdminService         service.AdminService
	NoticeService        service.NoticeService
	UserService          service.UserService
	CourseService        service.CourseService
	GradeService         service.GradeService
	ExamService          service.ExamService
	GradeAnalysisService service.GradeAnalysisService
	TaskService          service.TaskService

	// Controllers
	UserController          *controller.UserController
	CourseController        *controller.CourseController
	GradeController         *controller.GradeController
	ExamController          *controller.ExamController
	CaptchaController       *controller.CaptchaController
	StatisticsController    *controller.StatisticsController
	AdminController         *controller.AdminController
	NoticeController        *controller.NoticeController
	GradeAnalysisController *controller.GradeAnalysisController
	ConfigController        *controller.ConfigController
}

// NewContainer 创建依赖注入容器
func NewContainer(configPath string) (*Container, error) {
	c := &Container{}

	// 加载配置
	if err := c.initConfig(configPath); err != nil {
		return nil, fmt.Errorf("初始化配置失败: %w", err)
	}

	// 初始化数据库
	if err := c.initDB(); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化 Redis
	if err := c.initRedis(); err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}

	// 初始化 Repositories
	c.initRepositories()

	// 初始化 Caches
	c.initCaches()

	// 初始化 Services
	c.initServices()

	// 初始化 Controllers
	c.initControllers()

	// 初始化 RSA 公钥（首次获取）
	if err := c.initRSAPublicKey(); err != nil {
		return nil, fmt.Errorf("初始化 RSA 公钥失败: %w", err)
	}

	// 初始化默认管理员（如果不存在）
	if err := c.AdminService.InitDefaultAdmin(); err != nil {
		return nil, fmt.Errorf("初始化默认管理员失败: %w", err)
	}

	return c, nil
}

// initConfig 初始化配置
func (c *Container) initConfig(configPath string) error {
	config, err := LoadConfigFromPath(configPath)
	if err != nil {
		return err
	}
	c.Config = config
	return nil
}

// initDB 初始化数据库
func (c *Container) initDB() error {
	db, err := InitDBWithConfig(c.Config)
	if err != nil {
		return err
	}
	c.DB = db
	return nil
}

// initRedis 初始化 Redis（同一个 Redis 服务器的不同数据库）
func (c *Container) initRedis() error {
	ctx := context.Background()

	// 初始化会话 Redis (DB 0) - 用于存储用户登录会话
	sessionClient, err := c.createRedisClient(c.Config.Redis.Session)
	if err != nil {
		return fmt.Errorf("初始化会话Redis失败: %w", err)
	}
	if err := sessionClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("会话Redis连接测试失败: %w", err)
	}
	c.SessionRedis = sessionClient

	// 初始化验证码 Redis (DB 1) - 用于存储验证码
	captchaClient, err := c.createRedisClient(c.Config.Redis.Captcha)
	if err != nil {
		return fmt.Errorf("初始化验证码Redis失败: %w", err)
	}
	if err := captchaClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("验证码Redis连接测试失败: %w", err)
	}
	c.CaptchaRedis = captchaClient

	return nil
}

// createRedisClient 创建 Redis 客户端（辅助函数）
func (c *Container) createRedisClient(config RedisConfig) (*redis.Client, error) {
	// 如果 Host 已经包含端口，直接使用；否则添加端口
	addr := config.Host
	if config.Port != 0 && !strings.Contains(addr, ":") {
		addr = addr + ":" + strconv.Itoa(config.Port)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Pass,
		DB:       config.DB,
	})

	return client, nil
}

// initRepositories 初始化 Repositories
func (c *Container) initRepositories() {
	c.UserRepo = repository.NewGormUserRepository(c.DB)
	c.AdminRepo = repository.NewGormAdminRepository(c.DB)
	c.NoticeRepo = repository.NewGormNoticeRepository(c.DB)
}

// initCaches 初始化 Caches
func (c *Container) initCaches() {
	// 会话缓存（DB 0）
	c.SessionCache = cache.NewRedisSessionCache(c.SessionRedis)
	// 验证码缓存（DB 1）
	c.CaptchaCache = cache.NewRedisCaptchaCache(c.CaptchaRedis)
	// 日活统计缓存（DB 0，与会话共用）
	c.DAUCache = cache.NewRedisDAUCache(c.SessionRedis)
	// 系统配置缓存（DB 0，与会话共用）
	c.ConfigCache = cache.NewRedisConfigCache(c.SessionRedis)
	// 用户数据缓存（DB 0，与会话共用）
	c.UserDataCache = cache.NewRedisUserDataCache(c.SessionRedis)
}

// initServices 初始化 Services
func (c *Container) initServices() {
	// 获取当前模式的配置
	currentMode := c.Config.Jwc.GetCurrentModeConfig()
	log.Printf("教务系统模式: %s", c.Config.Jwc.Mode)

	// RSA Key Service（RSA 公钥服务）
	c.RSAKeyService = service.NewRSAKeyService(c.Config.Jwc.GetRSAKeyURL)

	// Session Service（根据配置模式注入对应的 URL）
	c.SessionService = service.NewJwcSessionService(
		c.SessionCache,
		c.RSAKeyService,
		c.Config.Jwc.Mode, // 注入当前模式
		currentMode.LoginURL,
		currentMode.RedirectURL,
		c.Config.Jwc.CaptchaURL,
		c.Config.Jwc.CaptchaImageURL,
		c.Config.Ocr.Host,
	)

	// Crawler Service
	c.CrawlerService = service.NewHttpCrawlerService()

	// Email Service（邮件服务）
	c.EmailService = service.NewEmailService(
		c.Config.Email.SMTPHost,
		c.Config.Email.SMTPPort,
		c.Config.Email.Username,
		c.Config.Email.Password,
		c.Config.Email.FromName,
	)

	// Captcha Service（验证码服务）
	c.CaptchaService = service.NewCaptchaService(
		c.CaptchaCache,
		c.EmailService,
	)

	// DAU Service（日活统计服务）
	c.DAUService = service.NewDAUService(c.DAUCache)

	// Admin Service（管理员服务）
	c.AdminService = service.NewAdminService(
		c.AdminRepo,
		c.Config.JWT.Secret,
		c.Config.JWT.Issuer,
	)

	// Notice Service（通知服务）
	c.NoticeService = service.NewNoticeService(c.NoticeRepo)

	// User Service
	c.UserService = service.NewUserService(
		c.UserRepo,
		c.SessionService,
		c.CaptchaService,
		c.CaptchaCache,
		c.DAUService,
		c.Config.JWT.Secret,
		c.Config.JWT.Issuer,
	)

	// Course Service（使用当前模式的 URL）
	c.CourseService = service.NewCourseService(
		c.UserRepo,
		c.SessionService,
		c.CrawlerService,
		c.UserDataCache,
		currentMode.CourseURL,
	)

	// Grade Service（使用当前模式的 URL）
	c.GradeService = service.NewGradeService(
		c.UserRepo,
		c.SessionService,
		c.CrawlerService,
		c.UserDataCache,
		currentMode.GradeURL,
		currentMode.GradeLevelURL,
	)

	// Exam Service（使用当前模式的 URL）
	c.ExamService = service.NewExamService(
		c.UserRepo,
		c.SessionService,
		c.CrawlerService,
		c.UserDataCache,
		currentMode.ExamURL,
	)

	// Grade Analysis Service（成绩分析服务）
	c.GradeAnalysisService = service.NewGradeAnalysisService(
		c.GradeService,
		c.ConfigCache,
	)

	// Task Service（定时任务服务）
	c.TaskService = service.NewTaskService(
		c.UserRepo,
		c.DAUCache,
		c.SessionService,
		c.CourseService,
		c.GradeService,
		c.ExamService,
		c.UserDataCache,
		c.ConfigCache,
	)
}

// initControllers 初始化 Controllers
func (c *Container) initControllers() {
	c.UserController = controller.NewUserController(c.UserService)
	c.CourseController = controller.NewCourseController(c.CourseService)
	c.GradeController = controller.NewGradeController(c.GradeService)
	c.ExamController = controller.NewExamController(c.ExamService)
	c.CaptchaController = controller.NewCaptchaController(c.CaptchaService)
	c.StatisticsController = controller.NewStatisticsController(c.DAUService)
	c.AdminController = controller.NewAdminController(c.AdminService)
	c.NoticeController = controller.NewNoticeController(c.NoticeService)
	c.GradeAnalysisController = controller.NewGradeAnalysisController(c.GradeAnalysisService)
	c.ConfigController = controller.NewConfigController(c.ConfigCache)
}

// initRSAPublicKey 初始化 RSA 公钥
func (c *Container) initRSAPublicKey() error {
	log.Println("正在获取 RSA 公钥...")
	if err := c.RSAKeyService.FetchAndUpdate(); err != nil {
		return fmt.Errorf("首次获取 RSA 公钥失败: %w", err)
	}
	log.Println("RSA 公钥初始化成功")
	return nil
}

// Close 关闭资源
func (c *Container) Close() error {
	// 关闭会话 Redis
	if c.SessionRedis != nil {
		if err := c.SessionRedis.Close(); err != nil {
			return fmt.Errorf("关闭会话Redis失败: %w", err)
		}
	}

	// 关闭验证码 Redis
	if c.CaptchaRedis != nil {
		if err := c.CaptchaRedis.Close(); err != nil {
			return fmt.Errorf("关闭验证码Redis失败: %w", err)
		}
	}

	// 关闭数据库
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return fmt.Errorf("获取数据库连接失败: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("关闭数据库失败: %w", err)
		}
	}

	return nil
}
