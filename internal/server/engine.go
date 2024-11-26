package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var Routers []func(e *gin.Engine)

func newWebEngine() *gin.Engine {
	e := gin.New()
	e.HandleMethodNotAllowed = true
	e.Use(gin.Logger())
	e.Use(gin.Recovery())

	// 跨域配置
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AddAllowHeaders("Authorization")
	e.Use(cors.New(corsConfig))
	// 默认404
	e.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	// 默认405
	e.NoMethod(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	return e
}

func Route(routeRegister func(e *gin.Engine)) {
	Routers = append(Routers, routeRegister)
}
