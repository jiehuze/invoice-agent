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
	// 设置响应头支持流式输出
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	progressChan := make(chan string, 100)
	go services.AutoFillingServiceStart(progressChan)
	// 监听自动填充服务的进度
	for progress := range progressChan {
		c.Writer.WriteString("AI执行: " + progress + "\n")
		c.Writer.Flush()
	}
	//controllers.Response(c, code.Success, "success", nil)
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

	// 第二阶段：启动自动填充服务，将收集到的内容作为参数
	progressChan := make(chan string, 100)
	go services.AutoFillingServiceStart(progressChan)

	// 监听自动填充服务的进度
	for progress := range progressChan {
		c.Writer.WriteString("data: " + progress + "\n\n")
		c.Writer.Flush()
	}
}
func InvoiceList(c *gin.Context) {

	controllers.Response(c, code.Success, "success", nil)
}
