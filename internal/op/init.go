package op

import (
	"time"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/dispatch"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/internal/task/tracker"
)

const (
	TaskQueueSize      = 200                    // 队列缓冲区中允许存放的最大任务数
	DisPatcherInterval = 500 * time.Millisecond // 这个是发现的最小间隔时间
)

func Initialized() {
	TaskQueue = task.NewTaskQueue(TaskQueueSize)
	TaskPool = preparePool()
	Dispatcher = dispatch.NewDispatchKernal(
		DisPatcherInterval,
		TaskQueue,
		TaskPool,
		[]task.TaskTracker{
			tracker.NewBoardCastTracker(BoardCastLoader),
			tracker.NewRedisTracker(conf.GetRedisClient()),
		},
	)
	InitStreams()
	InitStreamFactory()
	StartCrontab()
	logrus.Info("op initialized")
}
