// internal/app/controllers/v1/invoice_file_controller.go
package v1

import (
	"invoice-agent/internal/app/controllers"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/services"
)

type InvoiceFileController struct {
	service services.IInvoiceFile
}

func NewInvoiceFileController(service services.IInvoiceFile) *InvoiceFileController {
	return &InvoiceFileController{service: service}
}

// CreateInvoiceFile 创建发票文件
func (c *InvoiceFileController) CreateInvoiceFile(ctx *gin.Context) {
	var invoiceFile models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	if err := c.service.CreateInvoiceFile(&invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "创建发票文件失败", err.Error())
		return
	}

	controllers.Response(ctx, http.StatusCreated, "创建发票文件成功", invoiceFile)
}

// UpdateInvoiceFile 更新发票文件
func (c *InvoiceFileController) UpdateInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	var invoiceFile models.InvoiceFile
	if err := ctx.ShouldBindJSON(&invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	if err := c.service.UpdateInvoiceFile(id, &invoiceFile); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "更新发票文件失败", err.Error())
		return
	}

	controllers.Response(ctx, http.StatusOK, "更新发票文件成功", invoiceFile)
}

// GetInvoiceFile 获取单个发票文件
func (c *InvoiceFileController) GetInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	invoiceFile, err := c.service.GetInvoiceFileByID(id)
	if err != nil {
		controllers.Response(ctx, http.StatusNotFound, "发票文件不存在", err.Error())
		return
	}

	controllers.Response(ctx, http.StatusOK, "获取发票文件成功", invoiceFile)
}

// ListInvoiceFiles 获取发票文件列表
func (c *InvoiceFileController) ListInvoiceFiles(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	invoiceFiles, err := c.service.ListInvoiceFiles(limit, offset)
	if err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "获取发票文件列表失败", err.Error())
		return
	}

	controllers.Response(ctx, http.StatusOK, "获取发票文件列表成功", invoiceFiles)
}

// DeleteInvoiceFile 删除发票文件
func (c *InvoiceFileController) DeleteInvoiceFile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		controllers.Response(ctx, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	if err := c.service.DeleteInvoiceFile(id); err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "删除发票文件失败", err.Error())
		return
	}

	controllers.Response(ctx, http.StatusOK, "删除发票文件成功", nil)
}
