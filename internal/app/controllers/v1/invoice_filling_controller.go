package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"invoice-agent/internal/app/controllers"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/services"
	"invoice-agent/internal/pkg/code"
	"invoice-agent/pkg/util"
	"strings"
	"time"
)

type AutoFillingController struct {
	service *services.AutoFillingService
}

func NewAutoFillingController(service *services.AutoFillingService) *AutoFillingController {
	return &AutoFillingController{service: service}
}

func (c *AutoFillingController) InvoiceStart(ctx *gin.Context) {
	// 设置响应头支持流式输出
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	// 生成唯一任务ID
	taskID := uuid.New().String()
	basicItem := models.BasicItem{
		Category:   "日常报销",
		Title:      "主题xxxxxxxxxxxxxxxxx",
		UrgentType: "紧急",
		Comment:    "xxxxxxxxxxxxxxxxxxxxxxx项目出差",
	}
	payItem := models.PayItem{
		BusinessDept: "智能业务部/智能业务部-IT外包项目",
		BudgetDept:   "智能业务部/智能业务部-IT外包项目",
		PayDept:      "天宇正清科技有限公司",
		ProjectType:  "成本中心",
		Project:      "成本中心",
	}

	autoFillingRequest := &models.AutoFillingRequest{
		BasicInfo: basicItem,
		PayInfo:   payItem,
		Username:  "tyzq-wangmeng5",
		Password:  "tyzq123456",
	}
	autoFillingRequest.FilePaths = append(autoFillingRequest.FilePaths, "invoice2.pdf")
	autoFillingRequest.FilePaths = append(autoFillingRequest.FilePaths, "invoice3.pdf")
	// 填写报销明细
	item1 := models.CostItem{
		Category:   "团建费",
		Name:       "本部门团建",
		Comment:    "费用说明xxxxxxxxxxxxxxxxxx",
		Cost:       "1500",
		BillNumber: "3",
	}
	item2 := models.CostItem{
		Category:   "办公费",
		Name:       "办公用电",
		Comment:    "办公费用说明",
		Cost:       "800",
		BillNumber: "4",
	}
	autoFillingRequest.CostItems = append(autoFillingRequest.CostItems, item1)
	autoFillingRequest.CostItems = append(autoFillingRequest.CostItems, item2)

	err := c.service.StartAutoFilling(taskID, autoFillingRequest)
	if err != nil {
		controllers.Response(ctx, code.HTTPStatusErr, "自动填报失败", nil)
		return
	} // 获取进度通道
	progressChan, exists := c.service.GetTaskProgressChan(taskID)
	if !exists {
		ctx.Writer.WriteString("错误: 无法获取任务进度通道\n")
		ctx.Writer.Flush()
		return
	}
	// 设置监听条件
	timeout := time.After(30 * time.Minute) // 30分钟超时
	clientGone := ctx.Request.Context().Done()

	// 开始监听进度
	ctx.Writer.WriteString("AI执行: 任务已启动，任务ID: " + taskID + "\n")
	ctx.Writer.Flush()

	// 监听自动填充服务的进度
	for {
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// 通道关闭，任务完成
				ctx.Writer.WriteString("AI执行: 任务已完成\n")
				ctx.Writer.Flush()
				return
			}
			ctx.Writer.WriteString("AI执行: " + progress + "\n")
			ctx.Writer.Flush()

		case <-clientGone:
			// 客户端断开连接
			ctx.Writer.WriteString("AI执行: 客户端连接断开，任务已取消\n")
			ctx.Writer.Flush()
			//c.service.CancelTask(taskID)
			return

		case <-timeout:
			// 超时
			ctx.Writer.WriteString("AI执行: 任务执行超时，已取消\n")
			ctx.Writer.Flush()
			c.service.CancelTask(taskID)
			return

		case <-time.After(30 * time.Second):
			// 心跳检测，保持连接活跃
			ctx.Writer.WriteString("AI执行: 任务执行中...\n")
			ctx.Writer.Flush()
		}
	}
}

func (c *AutoFillingController) InvoiceChat(ctx *gin.Context) {
	fileIds := make([]string, 0)
	fileIds = append(fileIds, util.InvoiceFiles[9].FileID)
	fileIds = append(fileIds, util.InvoiceFiles[13].FileID)
	// 设置响应头支持流式输出
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	contentChan, errorChan := services.ChatClient.ChatStream(ctx.Request.Context(), fileIds)
	// 收集完整的流式数据
	var fullContent strings.Builder
	for {
		select {
		case content, ok := <-contentChan:
			if !ok {
				// 流结束，开始解析数据
				log.Info("===== 流结束，开始解析数据")
				// 获取完整内容
				contentStr := fullContent.String()
				var invoiceFiles []models.InvoiceFile
				// 解析JSON数据
				if err := json.Unmarshal([]byte(contentStr), &invoiceFiles); err != nil {
					errorMsg := fmt.Sprintf("AI执行: JSON解析失败: %v\n\n", err)
					ctx.Writer.WriteString(errorMsg)
					ctx.Writer.Flush()
					return
				}
				// 解析JSON数据

				// 发送解析结果
				resultMsg := fmt.Sprintf("AI执行: 解析成功，共解析%d条发票记录\n\n", len(invoiceFiles))
				ctx.Writer.WriteString(resultMsg)

				// 发送每条记录的详细信息
				for i, invoice := range invoiceFiles {
					detailMsg := fmt.Sprintf("AI执行: 发票%d: %s (%s) - %.2f元\n\n",
						i+1, invoice.InvoiceType, invoice.InvoiceCode, invoice.TotalAmount)
					ctx.Writer.WriteString(detailMsg)
				}

				ctx.Writer.Flush()
				//services.InvoiceFile.CreateInvoiceFilesBatch(invoiceFiles)
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
}
func InvoiceList(c *gin.Context) {

	controllers.Response(c, code.Success, "success", nil)
}
