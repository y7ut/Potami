package op

import (
	"github.com/y7ut/potami/internal/boardcast"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/internal/task/tracker"
	"github.com/y7ut/potami/pkg/syncc"
)

// TaskBoardCasts 任务的boardcast的Map
var TaskBoardCasts = new(syncc.Map[string, *boardcast.Boardcast[*task.Task]])

// BoardCastCreateOrLoad 创建或者加载boardcast
func BoardCastCreateOrLoad(t *task.Task) *boardcast.Boardcast[*task.Task] {
	b, _ := TaskBoardCasts.LoadOrStore(t.ID, boardcast.NewBoardCast(t, tracker.AnalyzeTaskEventState))
	return b
}
