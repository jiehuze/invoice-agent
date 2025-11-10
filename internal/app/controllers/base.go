package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"invoice-agent/internal/app/models"
)

func Response(c *gin.Context, code int, message string, data interface{}) {
	if nil == data {
		data = struct {
		}{}
	}
	resp := &models.RespValue{
		Code: code,
		Msg:  message,
		Data: data,
	}
	c.JSON(http.StatusOK, resp)
}

func SSEPush(c *gin.Context, data interface{}) {
	_, _ = c.Writer.WriteString("data: " + data.(string) + "\n\n")
	c.Writer.Flush()
}

func ResponseWithErr(c *gin.Context, code int, message string, err string, data interface{}) {
	if nil == data {
		data = struct {
		}{}
	}
	resp := &models.RespValue{
		Code: code,
		Msg:  message,
		Err:  err,
		Data: data,
	}
	c.JSON(http.StatusOK, resp)
}

func Health(c *gin.Context) {
	Response(c, 200, "Success", "")
}
