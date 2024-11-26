package logger

import (
	"github.com/y7ut/potami/internal/conf"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/pkg/logger"
)

// InitGlobalLogger 初始化全局日志
func InitGlobalLogger() {
	logrus.SetLevel(conf.Log.TransformLevel())
	logrus.AddHook(logger.NewUIDHook("id"))
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	if conf.Log.Path != "" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(logger.GenerateRotateLog(conf.Log.Path, conf.Log.Name))
	}
	if conf.Log.ZineURL != "" {
		zincHook, err := logger.NewZincLogHook(conf.Log.ZineURL, conf.Log.ZineIndex)
		if err != nil {
			panic(err)
		}
		logrus.AddHook(zincHook)
	}
}
