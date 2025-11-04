package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"invoice-agent/internal/app/controllers"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/services"
	"net/http"
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

	chat, err := services.ChatClient.Chat(ctx, req.Input, req.History)
	if err != nil {
		controllers.Response(ctx, http.StatusInternalServerError, "自动填开发票失败", gin.H{})
		return
	}

	log.Infoln("提取的信息为：\n", *chat)

	controllers.Response(ctx, http.StatusOK, "开始自动填开发票", *chat)
}
