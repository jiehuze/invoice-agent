package storage

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	once     sync.Once
)

func Init() {
	once.Do(func() {
        initMysql()

	})
	log.Infoln("Init storage ok")
}
