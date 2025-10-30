package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"invoice-agent/internal/app/controllers"
	"invoice-agent/internal/app/services"
	"invoice-agent/internal/pkg/code"
	"invoice-agent/pkg/util"
)

func InvoiceStart(c *gin.Context) {
	services.AutoFillingServiceStart()
	controllers.Response(c, code.Success, "success", nil)
}

func InvoiceChat(c *gin.Context) {
	fileIds := make([]string, 0)
	fileIds = append(fileIds, util.InvoiceFiles[12].FileID)
	//fileIds = append(fileIds, util.InvoiceFiles[1].FileID)
	chat, err := services.ChatClient.Chat(c.Request.Context(), fileIds)
	fmt.Print("chat response: ", chat)
	if nil != err {
		log.Error(err.Error())
	}
	controllers.Response(c, code.Success, "success", chat)
}
func InvoiceList(c *gin.Context) {

	controllers.Response(c, code.Success, "success", nil)
}
