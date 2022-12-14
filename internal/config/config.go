package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfg *Config

type Config struct {
	MySQL MysqlOpts `mapstructure:"mysql"`
}

func NewFromViper() *Config {
	if err := viper.ReadInConfig(); err != nil {
		logrus.Errorf("failed to load configuration file from disk: %v", err)
		logrus.Warnln("used default configuration")
		return NewFromDefault()
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		logrus.Errorf("failed to unmarshal configuration into struct: %v", err)
		logrus.Warnln("used default configuration")
		return NewFromDefault()
	}

	return cfg
}

func NewFromDefault() *Config {
	mysql := MysqlOpts{
		Address:  "mysql:3306",
		UserName: "root",
		Password: "0",
		DBName:   "development",
	}
	cfg = &Config{
		MySQL: mysql,
	}
	return cfg
}
