package routers

import (
	"invoice-agent/internal/app/controllers"
	v1 "invoice-agent/internal/app/controllers/v1"

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

		mainGroup.GET("/info/list", v1.InvoiceList)
		mainGroup.GET("/start", v1.InvoiceStart)
		mainGroup.GET("/chat", v1.InvoiceChat)

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
