package services

import (
	"invoice-agent/pkg/config"
	"sync"

	"invoice-agent/internal/app/models"
	"invoice-agent/internal/pkg/code"
)

var initOnce sync.Once

var (
	InvoiceFile IInvoiceFile
	AutoFilling *AutoFillingService
)

func Init() {
	initOnce.Do(func() {
		NewOpenAIClient(config.GetOpenaiConf().ApiKey)
		NewChantClient(config.GetOpenaiConf().ApiKey)
		//fmt.Println("=========prompt: ", config.GetOpenaiConf().Prompt)
		InvoiceFile = NewInvoiceFileService()
		AutoFilling = NewAutoFillingService()
	})
}

var successInfo = models.RespInfo{
	Code: code.Success,
	Msg:  code.MsgSuccess,
}
