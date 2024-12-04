package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/y7ut/collection"
	"github.com/y7ut/potami/internal/op"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/internal/task"
)

type StreamTask struct {
	Name     string                 `json:"name,omitempty"`
	Params   map[string]interface{} `json:"params"`
	Callback *task.CallBackConfig   `json:"callback"`
	Mode     string                 `json:"mode,omitempty" validate:"oneof=sync async stream"`
}

func CompleteStream(c *gin.Context) {
	var streamTaskReq StreamTask

	if err := c.ShouldBindJSON(&streamTaskReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	StreamName := c.Param("stream")
	streams := collection.New(op.GetStreamList()).Filter(func(item *schema.HumanFriendlyStreamConfig) bool {
		return item.Name == StreamName
	})

	if streams.Len() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream not found"})
		return
	}

	requiredParams := streams.Peek(0).RequiredParams

	// 如果没有提供全部的必要参数，返回错误
	for _, param := range requiredParams {
		if _, ok := streamTaskReq.Params[param]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing required param: %s", param)})
			return
		}
	}

	currentTask, err := createTask(StreamName, streamTaskReq.Name, streamTaskReq.Params, streamTaskReq.Callback)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bc := op.BoardCastLoader(currentTask)

	_, err = op.TaskQueue.PushTask(currentTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	op.Tasks.LoadOrStore(currentTask.ID, currentTask)

	if streamTaskReq.Mode == "stream" {
		if err := bc.Listen(c.Writer, c.Request); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		return
	}

	if streamTaskReq.Mode == "sync" {
		<-currentTask.Done()
		c.JSON(http.StatusOK, currentTask)
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": currentTask.ID})
}

func createTask(streamName, taskName string, params map[string]interface{}, callbackConfig *task.CallBackConfig) (*task.Task, error) {
	task, err := op.GetStreamFactory().Create(streamName, taskName)
	if err != nil {
		return nil, fmt.Errorf("create task error: %v", err)
	}
	task.MetaData = params
	task.CallBack = callbackConfig
	if callbackConfig != nil {
		task.CallBack.MetaDataBuilder = draftDefaultCallback()
	}
	return task, nil
}

func draftDefaultCallback() func(matedata map[string]interface{}, args []string) map[string]interface{} {
	return func(matedata map[string]interface{}, args []string) map[string]interface{} {
		return matedata
	}
}
