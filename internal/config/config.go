package config

import "github.com/spf13/viper"

var cfg *Config

type Config struct {
	MySQL MysqlOpts `mapstructure:"mysql"`
}

func NewFromViper() *Config {
	if err := viper.ReadInConfig(); err != nil {
		return NewFromDefault()
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
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
