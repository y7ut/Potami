package api

import "github.com/gin-gonic/gin"

func RegisterRouter(router *gin.Engine) {
	// v1
	api := router.Group("/api")
	{
		streamsRoute := api.Group("/stream")
		{
			streamsRoute.GET("", StreamsList)
			streamsRoute.POST("", StreamApply)
			streamsRoute.GET("/:stream", StreamInfo)
			streamsRoute.DELETE("/:stream", StreamRemove)
		}

		tasksRoute := api.Group("/task")
		{
			tasksRoute.GET("", TaskList)
			tasksRoute.GET("/:uuid", TaskInfo)
			tasksRoute.GET("/:uuid/look", TaskLook)
		}

		api.POST("/complete/:stream", CompleteStream)
	}
}
