package app

import (
	"context"
	"fmt"
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
	Redis  *redis.Client

	// Repositories
	UserRepo repository.UserRepository

	// Caches
	SessionCache cache.SessionCache

	// Services
	SessionService service.SessionService
	CrawlerService service.CrawlerService
	UserService    service.UserService
	CourseService  service.CourseService
	GradeService   service.GradeService
	ExamService    service.ExamService

	// Controllers
	UserController   *controller.UserController
	CourseController *controller.CourseController
	GradeController  *controller.GradeController
	ExamController   *controller.ExamController
}

// NewContainer 创建依赖注入容器
func NewContainer(configPath string) (*Container, error) {
	c := &Container{}

	// 1. 加载配置
	if err := c.initConfig(configPath); err != nil {
		return nil, fmt.Errorf("初始化配置失败: %w", err)
	}

	// 2. 初始化数据库
	if err := c.initDB(); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 3. 初始化 Redis
	if err := c.initRedis(); err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}

	// 4. 初始化 Repositories
	c.initRepositories()

	// 5. 初始化 Caches
	c.initCaches()

	// 6. 初始化 Services
	c.initServices()

	// 7. 初始化 Controllers
	c.initControllers()

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

// initRedis 初始化 Redis
func (c *Container) initRedis() error {
	redisConfig := c.Config.Redis
	// 如果 Host 已经包含端口，直接使用；否则添加端口
	addr := redisConfig.Host
	if redisConfig.Port != 0 && !strings.Contains(addr, ":") {
		addr = addr + ":" + strconv.Itoa(redisConfig.Port)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redisConfig.Pass,
		DB:       redisConfig.DB,
	})

	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		return err
	}

	c.Redis = client
	return nil
}

// initRepositories 初始化 Repositories
func (c *Container) initRepositories() {
	c.UserRepo = repository.NewGormUserRepository(c.DB)
}

// initCaches 初始化 Caches
func (c *Container) initCaches() {
	c.SessionCache = cache.NewRedisSessionCache(c.Redis)
}

// initServices 初始化 Services
func (c *Container) initServices() {
	// Session Service
	c.SessionService = service.NewJwcSessionService(
		c.SessionCache,
		c.Config.Jwc.LoginURL,
	)

	// Crawler Service
	c.CrawlerService = service.NewHttpCrawlerService()

	// User Service
	c.UserService = service.NewUserService(
		c.UserRepo,
		c.SessionService,
		c.Config.JWT.Secret,
		c.Config.JWT.Issuer,
	)

	// Course Service
	c.CourseService = service.NewCourseService(
		c.UserRepo,
		c.SessionService,
		c.CrawlerService,
		c.Config.Jwc.CourseURL,
	)

	// Grade Service
	c.GradeService = service.NewGradeService(
		c.UserRepo,
		c.SessionService,
		c.CrawlerService,
		c.Config.Jwc.GradeURL,
		c.Config.Jwc.GradeLevelURL,
	)

	// Exam Service
	c.ExamService = service.NewExamService(
		c.UserRepo,
		c.SessionService,
		c.CrawlerService,
		c.Config.Jwc.ExamURL,
	)
}

// initControllers 初始化 Controllers
func (c *Container) initControllers() {
	c.UserController = controller.NewUserController(c.UserService)
	c.CourseController = controller.NewCourseController(c.CourseService)
	c.GradeController = controller.NewGradeController(c.GradeService)
	c.ExamController = controller.NewExamController(c.ExamService)
}

// Close 关闭资源
func (c *Container) Close() error {
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			return err
		}
	}

	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
	}

	return nil
}
