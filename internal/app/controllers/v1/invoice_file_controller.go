// internal/app/controllers/v1/invoice_file_controller.go
package v1

import (
	"github.com/gin-gonic/gin"
	"invoice-agent/internal/app/controllers"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/services"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
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
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", "invalid id")
		return
	}

	var invoiceFile models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := c.invoiceFileService.UpdateInvoiceFile(id, &invoiceFile); err != nil {
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
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	documentNumber := ctx.Query("document_number")
	invoiceFile := models.InvoiceFile{
		DocumentNumber: documentNumber,
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

	if err := c.invoiceFileService.DeleteInvoiceFile(id); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "删除发票文件失败", err)
		return
	}

	controllers.Response(ctx, http.StatusOK, "删除成功", nil)
}

// / UploadInvoiceFileDirect 先保存文件到本地，然后上传到OpenAI
func (c *InvoiceFileController) UploadInvoiceFile(ctx *gin.Context) {
	// 获取上传的文件
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "文件上传失败", err)
		return
	}

	// 生成时间格式的目录名
	timestampDir := time.Now().Format("20060102150405")

	// 创建目录路径
	dirPath := filepath.Join("uploads", timestampDir)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "创建本地目录失败", err)
		return
	}

	// 保存文件到本地时间命名的目录
	localFilePath := filepath.Join(dirPath, fileHeader.Filename)
	if err := ctx.SaveUploadedFile(fileHeader, localFilePath); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "保存文件到本地失败", err)
		return
	}

	// 从本地文件上传到OpenAI
	fileId, err := services.FileClient.UploadFile(
		ctx.Request.Context(),
		localFilePath,
		"file-extract", // 或其他适当的purpose
	)
	if err != nil {
		// 如果上传OpenAI失败，删除已保存的本地文件
		os.Remove(localFilePath)
		controllers.Response(ctx, http.StatusInternalServerError, "文件上传到OpenAI失败", err)
		return
	}

	// 创建发票文件记录
	invoiceFile := models.InvoiceFile{
		FileName: fileHeader.Filename,
		FilePath: localFilePath,
		FileID:   fileId, // 使用OpenAI返回的文件ID
	}

	// 保存到数据库
	if err := c.invoiceFileService.CreateInvoiceFile(&invoiceFile); err != nil {
		// 如果数据库保存失败，删除已保存的本地文件和OpenAI文件
		os.Remove(localFilePath)
		services.FileClient.DeleteFile(ctx.Request.Context(), fileId)
		controllers.Response(ctx, http.StatusInternalServerError, "创建发票文件记录失败", err)
		return
	}

	controllers.Response(ctx, http.StatusOK, "文件上传和保存成功", gin.H{
		"file_id":    fileId,
		"file_name":  fileHeader.Filename,
		"local_path": localFilePath,
	})
}
