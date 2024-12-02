package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/y7ut/potami/internal/op"
)

func Look(c *gin.Context) {

	tastInfo, ok := op.Tasks.Load(c.Param("uuid"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if tastInfo.GetCompleteness() == 1 || tastInfo.Health() == 0 {
		c.JSON(200, gin.H{
			"task": tastInfo,
		})
		return
	}

	taskBoardCast, ok := op.TaskBoardCasts.Load(c.Param("uuid"))
	if taskBoardCast == nil || !ok {
		c.JSON(400, gin.H{"error": "sse error"})
		return
	}

	if err := taskBoardCast.Listen(c.Writer, c.Request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

}
