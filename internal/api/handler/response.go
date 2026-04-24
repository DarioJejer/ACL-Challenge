package handler

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Envelope{
		Success: true,
		Data:    data,
	})
}
