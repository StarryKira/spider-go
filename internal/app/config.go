package app

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App      Appconfig      `yaml:"app" mapstructure:"app"`
	Database DatabaseConfig `yaml:"database" mapstructure:"database"`
}

type Appconfig struct {
	Port int `yaml:"port" mapstructure:"port"`
}

type DatabaseConfig struct {
	Host string `yaml:"source" mapstructure:"source"`
	Port int    `yaml:"port" mapstructure:"port"`
	User string `yaml:"user" mapstructure:"user"`
	Pass string `yaml:"pass" mapstructure:"pass"`
	Name string `yaml:"name" mapstructure:"name"`
}

var Conf *Config

func LoadConfig() error {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Load config failed: %s \n", err)
	}
	Conf = &Config{}
	err = viper.Unmarshal(Conf)
	if err != nil {
		return fmt.Errorf("Unmarshal config failed: %s \n", err)
	}

	return nil
}

// GetDSN MySQL
func GetDSN() string {
	db := Conf.Database
	// 格式: user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.User, db.Pass, db.Host, db.Port, db.Name)
}
