package config

type RdbOpts struct {
	Driver    string    `mapstructure:"driver"` //
	MysqlOpts MysqlOpts `mapstructure:"mysql"`  //
}

type MysqlOpts struct {
	Address         string `mapstructure:"address"`           //
	UserName        string `mapstructure:"username"`          //
	Password        string `mapstructure:"password"`          //
	DBName          string `mapstructure:"dbname"`            //
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`    //
	MaxOpenConns    int    `mapstructure:"max_open_conns"`    //
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` //
}
