package requester

import (
	"strings"

	"github.com/imroc/req/v3"

	"invoice-agent/pkg/config"
)

func Req() *req.Client {
	if strings.EqualFold(config.GetRunMode(), "dev") {
		return req.DevMode()
	} else {
		return req.C()
	}
}
