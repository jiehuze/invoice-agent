package services

import (
	"fmt"
	"invoice-agent/pkg/config"
	"sync"

	"invoice-agent/internal/app/models"
	"invoice-agent/internal/pkg/code"
)

var initOnce sync.Once
var ()

func Init() {
	initOnce.Do(func() {
		NewOpenAIClient(config.GetOpenaiConf().ApiKey)
		NewChantClient(config.GetOpenaiConf().ApiKey)
		fmt.Println("=========prompt: ", config.GetOpenaiConf().Prompt)
	})
}

var successInfo = models.RespInfo{
	Code: code.Success,
	Msg:  code.MsgSuccess,
}
