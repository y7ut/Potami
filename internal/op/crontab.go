package op

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/task"
)

const (
	DefaultCrontabInterval = 24 * time.Hour
	TaskMaxSaveDruration   = 7
)

func StartCrontab() {
	go func() {
		ticker := time.NewTicker(DefaultCrontabInterval)
		for {
			<-ticker.C
			crontab()
		}
	}()
}

func crontab() {
	// TODO: 持久化到一个search index中
	logrus.Info("op crontab start")
	count := 0
	Tasks.Range(func(key string, value *task.Task) bool {
		if time.Since(value.StartAt) > time.Duration(TaskMaxSaveDruration)*DefaultCrontabInterval {
			Tasks.Delete(key)
		}
		count++
		return true
	})

	logrus.Infof("op clean old tasks: %d", count)
	logrus.Info("op crontab end")
}
