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
	fileIds = append(fileIds, util.InvoiceFiles[10].FileID)
	// 设置响应头支持流式输出
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	contentChan, errorChan := services.ChatClient.ChatStream(c.Request.Context(), fileIds)

	for {
		select {
		case content, ok := <-contentChan:
			if !ok {
				// 流结束
				log.Info("------------流结束")
				return
			}
			// 实时处理内容
			fmt.Print(content)
			c.Writer.WriteString(content)
			c.Writer.Flush()
		case err, ok := <-errorChan:
			if ok && err != nil {
				// 处理错误
				fmt.Printf("Error: %v\n", err)
				return
			}
		}
	}
}
func InvoiceList(c *gin.Context) {

	controllers.Response(c, code.Success, "success", nil)
}
