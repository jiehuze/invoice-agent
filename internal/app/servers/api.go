package servers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	"invoice-agent/internal/app/routers"
	"invoice-agent/pkg/config"
)

var (
	apiImpl *api
	apiOnce sync.Once
)

type api struct {
	server *http.Server
}

func ApiServer() IServer {
	apiOnce.Do(func() {
		apiImpl = &api{}
		apiImpl.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", config.AppConf.Port),
			Handler: routers.SetUp(),
		}
		log.Infoln("api init finish")
	})
	return apiImpl
}

func (r *api) Start() error {
	log.Infoln("api sever start")
	return r.server.ListenAndServe()
}

func (r *api) Stop() error {
	log.Infoln("api server stop")
	return r.server.Shutdown(context.Background())
}
