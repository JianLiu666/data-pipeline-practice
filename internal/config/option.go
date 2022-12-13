package config

type MysqlOpts struct {
	Address  string `mapstructure:"address"`  //
	UserName string `mapstructure:"username"` //
	Password string `mapstructure:"password"` //
	DBName   string `mapstructure:"dbname"`   //
}
