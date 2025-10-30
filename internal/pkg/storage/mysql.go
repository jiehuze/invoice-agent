package storage

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"invoice-agent/pkg/config"
)

var DB *gorm.DB

func initMysql() {
	if DB != nil {
		return
	}
	var err error
	DB, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&loc=Local",
		config.GetMysqlConf().Username, config.GetMysqlConf().Password, config.GetMysqlConf().Host, config.GetMysqlConf().DBName)))
	if err != nil {
		log.Errorf("db connect fail:%s", err.Error())
		panic("failed to connect database:")
	}
	sqlDb, _ := DB.DB()
	sqlDb.SetConnMaxLifetime(time.Hour * 6)
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetMaxOpenConns(20)
	if strings.Contains(config.GetRunMode(), "dev") {
		DB = DB.Debug()
	}
	log.Info("mysql connection success")
}
