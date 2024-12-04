package op

import (
	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/boardcast"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/internal/task/tracker"
	"github.com/y7ut/potami/pkg/syncc"
)

// TaskBoardCasts 任务的boardcast的Map
var TaskBoardCasts = new(syncc.Map[string, *boardcast.Boardcast[*task.Task]])

// BoardCastLoader 创建或者加载boardcast
func BoardCastLoader(t *task.Task) *boardcast.Boardcast[*task.Task] {

	bc, ok := TaskBoardCasts.Load(t.ID)
	if !ok {
		logrus.WithField("task_id", t.ID).Info("boardcast created")
		bc = boardcast.NewBoardCast(t, tracker.AnalyzeTaskEventState)
		TaskBoardCasts.Store(t.ID, bc)
	}
	return bc
}
