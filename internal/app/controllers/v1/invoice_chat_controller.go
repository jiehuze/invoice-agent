package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"invoice-agent/internal/app/controllers"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/services"
	"invoice-agent/pkg/util"
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

	log.Infoln("====== request parse: ", req.Parse)
	log.Infoln("====== request session_id: : ", req.SessionId)
	log.Infoln("====== request input: : ", req.Input)
	log.Infoln("====== request history: ", req.History)

	// 设置响应头支持流式输出
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	contentChan, errorChan := services.ChatClient.ChatStream(ctx, req)

	if req.Parse {
		_ = util.WriteAppendText(ctx.Writer, "## 报销信息")
		_ = util.WriteAppendText(ctx.Writer, "\n```json")
		_ = util.WriteAppendText(ctx.Writer, "\n")
	}

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
				_ = util.WriteAppendText(ctx.Writer, "\n```")
				_ = util.WriteAppendText(ctx.Writer, "\n")
				//所有信息收集完成，开始进行数据解析和填报发票单流程
				if req.Parse {
					c.fillingStart(ctx, &req, &contentStr)
				}
				//结束
				util.WriteDone(ctx.Writer)
				return
			}
			// 实时处理内容
			fmt.Print(content)
			fullContent.WriteString(content)
			_ = util.WriteAppendText(ctx.Writer, content)
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

func (c *InvoiceChatController) StartFilling(ctx *gin.Context) {
	// 设置响应头支持流式输出
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	var req models.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", gin.H{"error": err.Error()})
		return
	}

	c.fillingStart(ctx, &req, &req.History)
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
		_ = util.WriteAppendText(ctx.Writer, "\nAI助手: 填开发票失败")
		return
	} // 获取进度通道

	progressChan, exists := services.AutoFilling.GetTaskProgressChan(autoFillingRequest.SessionId)
	if !exists {
		_ = util.WriteAppendText(ctx.Writer, "\nAI助手: 填开发票失败")
		return
	}
	// 设置监听条件
	timeout := time.After(30 * time.Minute) // 30分钟超时
	clientGone := ctx.Request.Context().Done()

	// 监听自动填充服务的进度
	for {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// 通道关闭，任务完成
				_ = util.WriteAppendText(ctx.Writer, "\n## ✅ 报销服务已完成，请查看OA系统，审核后提交。。。")
				return
			}
			_ = util.WriteAppendText(ctx.Writer, progress)

		case <-clientGone:
			// 客户端断开连接
			log.Warning("\n> 客户端连接断开，任务已取消")
			_ = util.WriteAppendText(ctx.Writer, "\nAI助手: 客户端连接断开，任务已取消")
			//c.service.CancelTask(taskID)
			return

		case <-timeout:
			// 超时
			log.Warning("\n> 任务执行超时，已取消\n")
			_ = util.WriteAppendText(ctx.Writer, "\nAI助手: 任务执行超时，已取消")
			services.AutoFilling.CancelTask(autoFillingRequest.SessionId)
			return

		case <-time.After(30 * time.Second):
			// 心跳检测，保持连接活跃
			_ = util.WriteAppendText(ctx.Writer, "\n> 检测到任务执行中...")
		}
	}
}
