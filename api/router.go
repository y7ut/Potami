package api

import "github.com/gin-gonic/gin"

func RegisterRouter(router *gin.Engine) {
	// v1
	api := router.Group("/api")
	{
		api.GET("/stream", StreamsList)
		api.POST("/stream", StreamApply)
		api.GET("/stream/:stream", StreamInfo)
		api.DELETE("/stream/:stream", StreamRemove)

		api.POST("/complete/:stream", Push)

		api.GET("/info/:uuid", Info)
		api.GET("/look/:uuid", Look)

		api.GET("/list", List)
	}
}
