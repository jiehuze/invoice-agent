package config

import (
	"strings"

	"github.com/spf13/viper"
)

var mode string

func GetServerName() string {
	return AppConf.Servername
}

func GetRunMode() string {
	if len(mode) == 0 {
		return "dev"
	}
	return mode
}

func Init(runMode string) {
	mode = runMode
	var confPath string
	if strings.EqualFold(mode, "dev") {
		confPath = "./conf/dev"
	} else {
		confPath = "./conf/" + mode
	}
	v := viper.New()
	v.AddConfigPath(confPath)

	loadConf(v, "app", &AppConf)
	loadConf(v, "mysql", &mysqlConf)
	loadConf(v, "openai", &myopenai)
}
