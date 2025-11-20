// internal/app/controllers/v1/invoice_file_controller.go
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
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type InvoiceFileController struct {
	invoiceFileService services.IInvoiceFile
}

func NewInvoiceFileController(service services.IInvoiceFile) *InvoiceFileController {
	return &InvoiceFileController{invoiceFileService: service}
}

// CreateInvoiceFile 创建发票文件
func (c *InvoiceFileController) CreateInvoiceFile(ctx *gin.Context) {
	var invoiceFile models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFile); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := c.invoiceFileService.CreateInvoiceFile(&invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "创建发票文件失败", err)
		return
	}

	controllers.Response(ctx, http.StatusOK, "创建成功", invoiceFile)
}

// CreateInvoiceFilesBatch 批量创建发票文件
func (c *InvoiceFileController) CreateInvoiceFilesBatch(ctx *gin.Context) {
	var invoiceFiles []models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFiles); err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := c.invoiceFileService.CreateInvoiceFilesBatch(invoiceFiles); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "创建发票文件失败", err)
		return
	}

	controllers.Response(ctx, http.StatusOK, "批量创建成功", nil)
}

// UpdateInvoiceFile 更新发票文件（支持部分字段更新）
func (c *InvoiceFileController) UpdateInvoiceFile(ctx *gin.Context) {
	var invoiceFile models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := c.invoiceFileService.UpdateInvoiceFile(invoiceFile.ID, &invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "更新发票文件失败", err)
		return
	}

	controllers.Response(ctx, http.StatusOK, "更新成功", invoiceFile)
}

// GetInvoiceFile 获取单个发票文件
func (c *InvoiceFileController) GetInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", "invalid id")
		return
	}

	invoiceFile, err := c.invoiceFileService.GetInvoiceFileByID(id)
	if err != nil {
		controllers.Response(ctx, http.StatusNotFound, "发票文件不存在", "invoice file not found")
		return
	}

	controllers.Response(ctx, http.StatusOK, "获取成功", invoiceFile)
}

// ListInvoiceFiles 获取发票文件列表
func (c *InvoiceFileController) ListInvoiceFiles(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	serviceType, _ := strconv.Atoi(ctx.DefaultQuery("service_type", "3"))
	sessionId := ctx.Query("session_id")
	invoiceFile := models.InvoiceFile{
		SessionId:   sessionId,
		ServiceType: models.ServiceType(serviceType),
	}

	invoiceFiles, err := c.invoiceFileService.ListInvoiceFilesByCont(invoiceFile, limit, offset)
	if err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "获取发票文件列表失败", err)
		return
	}

	controllers.Response(ctx, http.StatusOK, "获取成功", invoiceFiles)
}

// DeleteInvoiceFile 删除发票文件
func (c *InvoiceFileController) DeleteInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", "invalid id")
		return
	}

	invoiceFile, err1 := c.invoiceFileService.GetInvoiceFileByID(id)
	if err1 != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "不存在这个文件：", err1)
		return
	}

	if err = c.invoiceFileService.DeleteInvoiceFile(id); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "删除发票文件失败", err)
		return
	}

	err = services.FileClient.DeleteFile(ctx.Request.Context(), strings.TrimSpace(invoiceFile.FileID))
	if err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "删除发票文件失败", err)
		return
	}

	controllers.Response(ctx, http.StatusOK, "删除成功", nil)
}

func (c *InvoiceFileController) GetInvoiceFileInExpensive(ctx *gin.Context) {
	serviceType, _ := strconv.Atoi(ctx.DefaultQuery("service_type", "1"))
	sessionId := ctx.Query("session_id")
	invoiceFile := models.InvoiceFile{
		SessionId:   sessionId,
		ServiceType: models.ServiceType(serviceType),
	}
	invoiceFiles, err := services.InvoiceFile.ListInvoiceFilesByCont(invoiceFile, 100, 0)
	if err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "获取发票文件列表失败", err)
		return
	}

	category := models.StatByExpenseCategory(invoiceFiles)

	controllers.Response(ctx, http.StatusOK, "获取成功", category)
}

func (c *InvoiceFileController) DeleteUploadedInvoiceFile(ctx *gin.Context) {
	// 从请求体中读取原始数据
	rawData, err := ctx.GetRawData()
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "无法读取请求数据", err)
		return
	}

	// 将字节数据转换为字符串并清理特殊字符
	rawString := strings.TrimSpace(string(rawData))

	// 移除空格、换行等特殊字符，只保留逗号分隔的ID
	//cleanString := strings.Map(func(r rune) rune {
	//	if r == ',' || (r >= '0' && r <= '9') {
	//		return r
	//	}
	//	return -1
	//}, rawString)

	// 按逗号分割ID
	fileIds := strings.Split(rawString, ",")

	for _, id := range fileIds {
		log.Infoln("---------delete file: ", id)
		err := services.FileClient.DeleteFile(ctx.Request.Context(), strings.TrimSpace(id))
		if err != nil {
			log.Infoln("---------delete file err: ", err)
			controllers.Response(ctx, http.StatusInternalServerError, "删除发票文件失败", err)
			return
		}
	}

	controllers.Response(ctx, http.StatusOK, "删除成功", nil)
}

/*
UploadInvoiceFileDirect 先保存文件到本地，然后上传到OpenAI
*/
func (c *InvoiceFileController) UploadInvoiceFile(ctx *gin.Context) {
	// 获取上传的文件
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Errorf("文件上传失败: %v", err)
		controllers.Response(ctx, http.StatusBadRequest, "文件上传失败", err)
		return
	}
	sessionId := ctx.PostForm("session_id")
	if sessionId == "" {
		log.Errorf("请创建一个新的对话")
		controllers.Response(ctx, http.StatusBadRequest, "请创建一个新的对话...", nil)
		return
	}

	md5 := ctx.PostForm("md5")
	if md5 == "" {
		log.Errorf("请上传文件的md5")
		controllers.Response(ctx, http.StatusBadRequest, "请上传文件的md5...", nil)
		return
	}

	tmp := models.InvoiceFile{
		SessionId: sessionId,
		MD5:       md5,
	}
	//去重出去，重复的数据要进行删除
	files, _ := services.InvoiceFile.ListInvoiceFilesByCont(tmp, 10, 0)
	if files != nil && len(files) > 0 {
		log.Errorf("文件重复, file: %s, md5: %s", fileHeader.Filename, md5)
		controllers.Response(ctx, http.StatusBadRequest, "重复的文件...", tmp)
		return
	}

	// 创建目录路径
	dirPath := filepath.Join("/app/output/uploads")
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		log.Errorf("本地目录:%s, 创建失败", dirPath)
		controllers.Response(ctx, http.StatusInternalServerError, "创建本地目录失败", tmp)
		return
	}

	// 保存文件到本地时间命名的目录
	localFilePath := filepath.Join(dirPath, fileHeader.Filename)
	if err := ctx.SaveUploadedFile(fileHeader, localFilePath); err != nil {
		log.Errorf("保存文件到本地失败:%s", fileHeader.Filename)
		controllers.Response(ctx, http.StatusInternalServerError, "保存文件到本地失败", tmp)
		return
	}

	// 计算文件MD5值
	//md5Hash, err := util.CalculateFileMD5(localFilePath)
	//if err != nil {
	//	// 如果计算MD5失败，删除已保存的本地文件
	//	os.Remove(localFilePath)
	//	controllers.Response(ctx, http.StatusInternalServerError, "计算文件MD5值失败", err)
	//	return
	//}

	// 从本地文件上传到OpenAI
	fileId, err := services.FileClient.UploadFile(
		ctx.Request.Context(),
		localFilePath,
		"file-extract", // 或其他适当的purpose
	)
	if err != nil {
		// 如果上传OpenAI失败，删除已保存的本地文件
		_ = os.Remove(localFilePath)
		controllers.Response(ctx, http.StatusInternalServerError, fmt.Sprintf("文件上传到OpenAI失败:%s", err), tmp)
		return
	}

	// 创建发票文件记录
	invoiceFile := models.InvoiceFile{
		SessionId: sessionId,
		FileName:  fileHeader.Filename,
		FilePath:  localFilePath,
		FileID:    fileId, // 使用OpenAI返回的文件ID
		MD5:       md5,
	}

	// 保存到数据库
	if err := c.invoiceFileService.CreateInvoiceFile(&invoiceFile); err != nil {
		// 如果数据库保存失败，删除已保存的本地文件和OpenAI文件
		_ = os.Remove(localFilePath)
		_ = services.FileClient.DeleteFile(ctx.Request.Context(), fileId)
		controllers.Response(ctx, http.StatusInternalServerError, "创建发票文件记录失败", invoiceFile)
		return
	}

	controllers.Response(ctx, http.StatusOK, "文件上传和保存成功", invoiceFile)
}

// 发票分析
func (c *InvoiceFileController) FileParseChat(ctx *gin.Context) {
	var req models.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Errorf("参数错误: %v", err)
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", gin.H{"error": err.Error()})
		return
	}
	if req.FileIds == nil || len(req.FileIds) == 0 {
		log.Errorf("请上传发票文件")
		controllers.Response(ctx, http.StatusBadRequest, "请上传发票文件", nil)
		return
	}

	// 设置响应头支持流式输出
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	contentChan, errorChan := services.ChatClient.FileParseStream(ctx.Request.Context(), req)
	_ = util.WriteAppendText(ctx.Writer, "\n## 所有单据信息")
	_ = util.WriteAppendText(ctx.Writer, "\n```json")
	_ = util.WriteAppendText(ctx.Writer, "\n")
	// 收集完整的流式数据
	var fullContent strings.Builder
	for {
		select {
		case content, ok := <-contentChan:
			if !ok {
				// 流结束，开始解析数据
				log.Info("===== 流结束，开始解析数据")
				_ = util.WriteAppendText(ctx.Writer, "\n```")
				_ = util.WriteAppendText(ctx.Writer, "\n")
				// 获取完整内容
				contentStr := fullContent.String()
				var invoiceFiles []models.InvoiceFile
				// 解析JSON数据
				if err := json.Unmarshal([]byte(contentStr), &invoiceFiles); err != nil {
					errorMsg := fmt.Sprintf("AI助手: JSON解析失败: %v\n", err)
					_ = util.WriteAppendText(ctx.Writer, "\n")
					_ = util.WriteAppendText(ctx.Writer, errorMsg)
					return
				}

				for _, invoiceFile := range invoiceFiles {
					err := services.InvoiceFile.UpdateInvoiceFileByFileId(invoiceFile.FileID, &invoiceFile)
					if err != nil {
						errorMsg := fmt.Sprintf("AI助手: 更新发票文件失败: %v\n", err)
						_ = util.WriteAppendText(ctx.Writer, "\n")
						_ = util.WriteAppendText(ctx.Writer, errorMsg)
					}
				}

				// 发送解析结果
				resultMsg := fmt.Sprintf("## 📋AI助手: 共解析%d条发票记录\n", len(invoiceFiles))
				_ = util.WriteAppendText(ctx.Writer, "\n")
				_ = util.WriteAppendText(ctx.Writer, resultMsg)

				// 发送每条记录的详细信息
				for i, invoice := range invoiceFiles {
					detailMsg := fmt.Sprintf("- 票据%d: %s (%s) - %.2f元, %s\n",
						i+1, invoice.InvoiceType, invoice.ItemName, invoice.TotalAmount, invoice.FileName)
					_ = util.WriteAppendText(ctx.Writer, "\n")
					_ = util.WriteAppendText(ctx.Writer, detailMsg)
				}
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
}
