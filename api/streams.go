package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/y7ut/potami/internal/op"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/internal/service/stream"
)

func StreamsList(c *gin.Context) {
	c.JSON(200, op.GetStreamList())
}

func StreamInfo(c *gin.Context) {
	streamname := c.Param("stream")
	stream, ok := op.GetStream(streamname)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	accept := c.Request.Header.Get("Accept")

	if accept == "application/x-yaml" || accept == "text/yaml" {
		c.YAML(http.StatusOK, stream)
	} else {
		c.JSON(http.StatusOK, stream)
	}

}

func StreamRemove(c *gin.Context) {
	streamname := c.Param("stream")

	if err := stream.RemoveStream(c, streamname); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	op.RemoveStream(streamname)
	op.ReloadStreamFactory()

	c.AbortWithStatus(http.StatusNoContent)
}

func StreamApply(c *gin.Context) {
	var ApplyStream schema.Stream

	contentType := c.ContentType()
	accept := c.Request.Header.Get("Accept")

	if contentType == "application/json" {
		if err := c.ShouldBindJSON(&ApplyStream); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else if contentType == "application/x-yaml" || contentType == "text/yaml" {
		if err := c.ShouldBindBodyWithYAML(&ApplyStream); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Unsupported content type"})
		return
	}

	if err := schema.ValidateStream(ApplyStream); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("validate stream error: %s", err.Error())})
		return
	}

	if err := stream.Apply(c, &ApplyStream); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("apply stream error: %s", err.Error())})
	}

	op.UpdateStream(&ApplyStream)
	op.ReloadStreamFactory()
	if accept == "application/x-yaml" || accept == "text/yaml" {
		c.YAML(http.StatusOK, ApplyStream)
	} else {
		c.JSON(http.StatusOK, ApplyStream)
	}

}
