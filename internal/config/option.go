package config

type RdbOpts struct {
	Driver       string       `mapstructure:"driver"`     //
	MysqlOpts    MysqlOpts    `mapstructure:"mysql"`      //
	PostgresOpts PostgresOpts `mapstructure:"postgresql"` //
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

type PostgresOpts struct {
	Host     string `mapstructure:"host"`     //
	Port     int    `mapstructure:"port"`     //
	User     string `mapstructure:"user"`     //
	Password string `mapstructure:"password"` //
	DBName   string `mapstructure:"dbname"`   //
}
