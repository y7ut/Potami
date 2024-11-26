package logger

import (
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// GenerateRotateLog 生成日志
func GenerateRotateLog(logPath string, filename string) *rotatelogs.RotateLogs {
	logFullPath := path.Join(logPath, filename)
	logwriter, err := rotatelogs.New(
		logFullPath+"_%Y%m%d.log",
		rotatelogs.WithLinkName(logFullPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithRotationCount(60),          // 文件最大保存份数
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		panic(err)
	}

	return logwriter
}
