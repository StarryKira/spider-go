package app

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App      Appconfig      `yaml:"app" mapstructure:"app"`
	Database DatabaseConfig `yaml:"database" mapstructure:"database"`
	Redis    RedisConfig    `yaml:"redis" mapstructure:"redis"`
	Jwc      JwcConfig      `yaml:"jwc" mapstructure:"jwc"`
	JWT      JWTConfig      `yaml:"jwt" mapstructure:"jwt"`
}

type Appconfig struct {
	Port int `yaml:"port" mapstructure:"port"`
}

type JwcConfig struct {
	LoginURL      string `yaml:"login_url" mapstructure:"login_url"`
	CourseURL     string `yaml:"course_url" mapstructure:"course_url"`
	GradeURL      string `yaml:"grade_url" mapstructure:"grade_url"`
	GradeLevelURL string `yaml:"grade_level_url" mapstructure:"grade_level_url"`
	ExamURL       string `yaml:"exam_url" mapstructure:"exam_url"`
}

type JWTConfig struct {
	Secret string `yaml:"secret" mapstructure:"secret"`
	Issuer string `yaml:"issuer" mapstructure:"issuer"`
}

type DatabaseConfig struct {
	Host string `yaml:"source" mapstructure:"source"`
	Port int    `yaml:"port" mapstructure:"port"`
	User string `yaml:"user" mapstructure:"user"`
	Pass string `yaml:"pass" mapstructure:"pass"`
	Name string `yaml:"name" mapstructure:"name"`
}

type RedisConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
	Pass string `yaml:"pass" mapstructure:"pass"`
	DB   int    `yaml:"db" mapstructure:"db"`
}

var Conf *Config

// LoadConfig 加载配置（兼容旧代码）
func LoadConfig() error {
	config, err := LoadConfigFromPath("./config")
	if err != nil {
		return err
	}
	Conf = config
	return nil
}

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

// GetDSN MySQL
func GetDSN() string {
	db := Conf.Database
	// 格式: user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.User, db.Pass, db.Host, db.Port, db.Name)
}
