package app

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App      Appconfig          `yaml:"app" mapstructure:"app"`
	CORS     CORSConfig         `yaml:"cors" mapstructure:"cors"`
	Database DatabaseConfig     `yaml:"database" mapstructure:"database"`
	Redis    RedisClusterConfig `yaml:"redis" mapstructure:"redis"`
	Jwc      JwcConfig          `yaml:"jwc" mapstructure:"jwc"`
	JWT      JWTConfig          `yaml:"jwt" mapstructure:"jwt"`
	Email    EmailConfig        `yaml:"email" mapstructure:"email"`
	Ocr      DdddOCRConfig      `yaml:"ocr" mapstructure:"ocr"`
}

type Appconfig struct {
	Port int `yaml:"port" mapstructure:"port"`
}

// CORSConfig CORS 跨域配置
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" mapstructure:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods" mapstructure:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers" mapstructure:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers" mapstructure:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" mapstructure:"max_age"` // 预检请求缓存时间（秒）
}

// JwcConfig 教务系统配置
type JwcConfig struct {
	Mode            string        `yaml:"mode" mapstructure:"mode"` // 模式：campus 或 webvpn
	Campus          JwcModeConfig `yaml:"campus" mapstructure:"campus"`
	Webvpn          JwcModeConfig `yaml:"webvpn" mapstructure:"webvpn"`
	GetRSAKeyURL    string        `yaml:"rsa_url" mapstructure:"rsa_url"`
	CaptchaURL      string        `yaml:"captcha_url" mapstructure:"captcha_url"`
	CaptchaImageURL string        `yaml:"captcha_image_url" mapstructure:"captcha_image_url"`
}

// JwcModeConfig 教务系统单个模式的配置
type JwcModeConfig struct {
	LoginURL      string `yaml:"login_url" mapstructure:"login_url"`
	RedirectURL   string `yaml:"redirect_url" mapstructure:"redirect_url"`
	CourseURL     string `yaml:"course_url" mapstructure:"course_url"`
	GradeURL      string `yaml:"grade_url" mapstructure:"grade_url"`
	GradeLevelURL string `yaml:"grade_level_url" mapstructure:"grade_level_url"`
	ExamURL       string `yaml:"exam_url" mapstructure:"exam_url"`
}

// GetCurrentModeConfig 获取当前模式的配置
func (c *JwcConfig) GetCurrentModeConfig() JwcModeConfig {
	if c.Mode == "webvpn" {
		return c.Webvpn
	}
	return c.Campus // 默认使用校园网模式
}

type JWTConfig struct {
	Secret string `yaml:"secret" mapstructure:"secret"`
	Issuer string `yaml:"issuer" mapstructure:"issuer"`
}

// EmailConfig 邮件服务配置
type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host" mapstructure:"smtp_host"` // SMTP 服务器地址
	SMTPPort int    `yaml:"smtp_port" mapstructure:"smtp_port"` // SMTP 端口
	Username string `yaml:"username" mapstructure:"username"`   // 发件人邮箱
	Password string `yaml:"password" mapstructure:"password"`   // SMTP 授权码
	FromName string `yaml:"from_name" mapstructure:"from_name"` // 发件人名称
}

type DatabaseConfig struct {
	Host string `yaml:"source" mapstructure:"source"`
	Port int    `yaml:"port" mapstructure:"port"`
	User string `yaml:"user" mapstructure:"user"`
	Pass string `yaml:"pass" mapstructure:"pass"`
	Name string `yaml:"name" mapstructure:"name"`
}

// RedisConfig 单个 Redis 数据库配置
type RedisConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
	Pass string `yaml:"pass" mapstructure:"pass"`
	DB   int    `yaml:"db" mapstructure:"db"`
}

// RedisClusterConfig Redis 集群配置（同一个 Redis 服务器的不同数据库）
type RedisClusterConfig struct {
	Session RedisConfig `yaml:"session" mapstructure:"session"` // DB 0: 用户会话缓存
	Captcha RedisConfig `yaml:"captcha" mapstructure:"captcha"` // DB 1: 验证码存储
}

// DdddOCRConfig OCR 服务配置
type DdddOCRConfig struct {
	Host string `yaml:"host" mapstructure:"host"` // OCR 服务地址
}

var Conf *Config

// LoadConfigFromPath 从指定路径加载配置
func LoadConfigFromPath(configPath string) (*Config, error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Load config failed: %s", err)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("Unmarshal config failed: %s", err)
	}

	return config, nil
}
