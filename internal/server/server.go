package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/conf"
)

var webServer *WebServer

type WebServer struct {
	e      *gin.Engine
	server *http.Server
}

func Initialized() {
	// 禁用控制台颜色，将日志写入文件时不需要控制台颜色。
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	engine := newWebEngine()

	server := &http.Server{
		Addr:         conf.HttpServer.Address(),
		Handler:      engine,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	webServer = &WebServer{
		e:      engine,
		server: server,
	}

	logrus.Info("server initialized")
}

func Start() {
	if webServer == nil {
		panic("please init web server first")
	}
	for _, routeRegister := range Routers {
		routeRegister(webServer.e)
	}
	logrus.Info("server started")
	if err := webServer.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Debugf("listen: %s\n", err)
	}
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := webServer.server.Shutdown(ctx); err != nil {
		logrus.Debugf("Server Shutdown: %s", err)
	}

	logrus.Info("Server shutdown gracefully")
}
