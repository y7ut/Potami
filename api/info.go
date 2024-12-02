package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/y7ut/potami/internal/op"
)

func Info(c *gin.Context) {

	tastInfo, ok := op.Tasks.Load(c.Param("uuid"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(200, gin.H{
		"task": tastInfo,
	})
}
