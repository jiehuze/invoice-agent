package config

var mysqlConf Mysql

type Mysql struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbName"`
}

func GetMysqlConf() Mysql {
	return mysqlConf
}