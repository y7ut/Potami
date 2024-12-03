package tracker

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/boardcast"
	"github.com/y7ut/potami/internal/task"
)

type BoardCastTracker struct {
	boardcastLoader func(t *task.Task) *boardcast.Boardcast[*task.Task]
}

func AnalyzeTaskEventState(t *task.Task) string {
	event := boardcast.HeartbeatEvent
	if t.Health() == 0 {
		event = boardcast.DeadEvent
		return event
	}

	if t.StartAt == (time.Time{}) {
		event = boardcast.WaitEvent
	}
	return event
}

func NewBoardCastTracker(loader func(t *task.Task) *boardcast.Boardcast[*task.Task]) *BoardCastTracker {
	return &BoardCastTracker{
		boardcastLoader: loader,
	}
}

// Tracker BoardCastTracker
func (bct *BoardCastTracker) Notice(ctx context.Context, complete float64, t *task.Task) {
	if bct.boardcastLoader == nil {
		logrus.WithField("task_id", t.ID).Warning("boardcast loader not init")
		return
	}

	bc := bct.boardcastLoader(t)
	eventTranslation := func(complete float64) string {
		if complete >= 1 {
			return boardcast.FinishEvent
		}
		if complete == 0 {
			return boardcast.StartEvent
		}
		return "update"
	}

	bc.Send(eventTranslation(complete))
	if complete >= 1 || t.Arrived == t.JobsPipline.Len() {
		// 任务完成
		go func() {
			logrus.WithField("task_id", t.ID).Info("boardcast tracker exit")
			time.Sleep(5 * time.Second)
			bc.Stop()
		}()
	}
}
