package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"invoice-agent/internal/app/controllers"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/services"
	"net/http"
	"strings"
	"time"
)

type InvoiceChatController struct {
	service *services.QwenLongClient
}

func NewInvoiceChatController(service *services.QwenLongClient) *InvoiceChatController {
	return &InvoiceChatController{service: service}
}

func (c *InvoiceChatController) Chat(ctx *gin.Context) {
	var req models.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", gin.H{"error": err.Error()})
		return
	}

	// 设置响应头支持流式输出
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	contentChan, errorChan := services.ChatClient.ChatStream(ctx, req)
	var fullContent strings.Builder
	for {
		select {
		case content, ok := <-contentChan:
			if !ok {
				// 流结束，开始解析数据
				log.Info("===== 流结束，开始解析数据")
				// 获取完整内容
				contentStr := fullContent.String()
				log.Infoln("提取的信息为：\n", contentStr)
				//所有信息收集完成，开始进行数据解析和填报发票单流程
				if req.Parse {
					c.fillingStart(ctx, &req, &contentStr)
				}
				return
			}
			// 实时处理内容
			fmt.Print(content)
			fullContent.WriteString(content)
			ctx.Writer.WriteString(content)
			ctx.Writer.Flush()
		case err, ok := <-errorChan:
			if ok && err != nil {
				// 处理错误
				fmt.Printf("Error: %v\n", err)
				return
			}
		}
	}
	//controllers.Response(ctx, http.StatusOK, "开始自动填开发票", *chat)
}

func (c *InvoiceChatController) fillingStart(ctx *gin.Context, req *models.ChatRequest, chat *string) {
	var autoFillingRequest models.AutoFillingRequest
	// 解析JSON数据
	if err := json.Unmarshal([]byte(*chat), &autoFillingRequest); err != nil {
		errorMsg := fmt.Sprintf("AI执行: JSON解析失败: %v\n\n", err)
		controllers.Response(ctx, http.StatusInternalServerError, errorMsg, nil)
		return
	}
	invoiceFile := models.InvoiceFile{
		SessionId:   req.SessionId,
		ServiceType: models.ServiceType(3),
	}
	files, err := services.InvoiceFile.ListInvoiceFilesByCont(invoiceFile, 100, 0)
	if err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "自动填开发票失败", gin.H{})
		return
	}
	autoFillingRequest.SessionId = req.SessionId
	autoFillingRequest.Username = "tyzq-wangmeng5"
	autoFillingRequest.Password = "tyzq123456"
	autoFillingRequest.CostItems = models.StatByExpenseCategory(files)
	_ = models.CollateFile(files, &autoFillingRequest)
	//controllers.Response(ctx, http.StatusOK, "开始自动填开发票", autoFillingRequest)
	//return
	log.Infoln("1---: ", autoFillingRequest)
	err = services.AutoFilling.StartAutoFilling(autoFillingRequest.SessionId, &autoFillingRequest)
	if err != nil {
		log.Infoln("2---: ")
		_, _ = ctx.Writer.WriteString("AI助手: 误差为你启动助手\n")
		ctx.Writer.Flush()
		return
	} // 获取进度通道
	log.Infoln("3---: ")

	progressChan, exists := services.AutoFilling.GetTaskProgressChan(autoFillingRequest.SessionId)
	if !exists {
		_, _ = ctx.Writer.WriteString("AI助手: 无法获取任务进度通道\n")
		ctx.Writer.Flush()
		return
	}
	// 设置监听条件
	timeout := time.After(30 * time.Minute) // 30分钟超时
	clientGone := ctx.Request.Context().Done()

	// 开始监听进度
	_, _ = ctx.Writer.WriteString("AI助手: 任务已启动，任务ID: " + autoFillingRequest.SessionId + "\n")
	ctx.Writer.Flush()

	// 监听自动填充服务的进度
	for {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// 通道关闭，任务完成
				_, _ = ctx.Writer.WriteString("AI助手: 任务已完成\n")
				ctx.Writer.Flush()
				return
			}
			_, _ = ctx.Writer.WriteString("AI助手: " + progress + "\n")
			ctx.Writer.Flush()

		case <-clientGone:
			// 客户端断开连接
			_, _ = ctx.Writer.WriteString("AI助手: 客户端连接断开，任务已取消\n")
			ctx.Writer.Flush()
			//c.service.CancelTask(taskID)
			return

		case <-timeout:
			// 超时
			_, _ = ctx.Writer.WriteString("AI助手: 任务执行超时，已取消\n")
			ctx.Writer.Flush()
			services.AutoFilling.CancelTask(autoFillingRequest.SessionId)
			return

		case <-time.After(30 * time.Second):
			// 心跳检测，保持连接活跃
			_, _ = ctx.Writer.WriteString("AI助手: 任务执行中...\n")
			ctx.Writer.Flush()
		}
	}
}
