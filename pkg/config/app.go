package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var AppConf App

// App viper的yaml格式使用mapstructure进行序列化和反序列化，tag为yaml实际不生效，使用默认小写
type App struct {
	Servername string `mapstructure:"servername"`
	Port       int    `mapstructure:"port"`
}

func GetAppConf() App {
	return AppConf
}

func loadConf(v *viper.Viper, confName string, confContainer interface{}) {
	v.SetConfigName(confName)
	err := v.ReadInConfig()
	if nil != err {
		panic(err)
	}
	err = v.Unmarshal(confContainer)
	if err != nil {
		panic(err)
	}
	log.Infoln("load ", confName, " finish")
}
