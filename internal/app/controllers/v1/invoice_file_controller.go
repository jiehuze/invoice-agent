// internal/app/controllers/v1/invoice_file_controller.go
package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/services"
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
		return
	}

	if err := c.invoiceFileService.CreateInvoiceFile(&invoiceFile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": invoiceFile})
}

// CreateInvoiceFilesBatch 批量创建发票文件
func (c *InvoiceFileController) CreateInvoiceFilesBatch(ctx *gin.Context) {
	var invoiceFiles []models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFiles); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.invoiceFileService.CreateInvoiceFilesBatch(invoiceFiles); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "batch create successful"})
}

// UpdateInvoiceFile 更新发票文件（支持部分字段更新）
func (c *InvoiceFileController) UpdateInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var invoiceFile models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFile); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.invoiceFileService.UpdateInvoiceFile(id, &invoiceFile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "update successful"})
}

// GetInvoiceFile 获取单个发票文件
func (c *InvoiceFileController) GetInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	invoiceFile, err := c.invoiceFileService.GetInvoiceFileByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invoice file not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": invoiceFile})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": invoiceFiles})
}

// DeleteInvoiceFile 删除发票文件
func (c *InvoiceFileController) DeleteInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.invoiceFileService.DeleteInvoiceFile(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "delete successful"})
}
