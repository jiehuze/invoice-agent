package routers

import (
	"invoice-agent/internal/app/controllers"
	v1 "invoice-agent/internal/app/controllers/v1"
	"invoice-agent/internal/app/services"

	"sync"

	"github.com/gin-gonic/gin"
)

var apiOnce sync.Once
var g *gin.Engine

func SetUp() *gin.Engine {
	apiOnce.Do(func() {
		g = gin.Default()

		// 跨域中间件
		// g.Use(corsMiddleware())

		mainGroup := g.Group("/invoice/")
		mainGroup.GET("/health", controllers.Health)

		chatGroup := mainGroup.Group("/chat")
		{
			controller := v1.NewInvoiceChatController(services.ChatClient)
			chatGroup.POST("/:id", controller.Chat)
		}

		invoiceFileGroup := mainGroup.Group("/files")
		{
			controller := v1.NewInvoiceFileController(services.InvoiceFile)
			invoiceFileGroup.POST("/add", controller.CreateInvoiceFile)
			invoiceFileGroup.POST("/batch", controller.CreateInvoiceFilesBatch)
			invoiceFileGroup.GET("/list", controller.ListInvoiceFiles)
			invoiceFileGroup.GET("/:id", controller.GetInvoiceFile)
			invoiceFileGroup.PUT("/update/:id", controller.UpdateInvoiceFile)
			invoiceFileGroup.DELETE("/:id", controller.DeleteInvoiceFile)
			invoiceFileGroup.POST("/upload", controller.UploadInvoiceFile)
		}

		autoFillGroup := mainGroup.Group("/filling")
		{
			autoFillingController := v1.NewAutoFillingController(services.AutoFilling)
			autoFillGroup.GET("/start", autoFillingController.InvoiceStart)
			autoFillGroup.GET("/file", autoFillingController.InvoiceChat)
		}
	})

	return g
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}
