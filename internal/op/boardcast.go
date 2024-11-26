package op

import (
	"github.com/y7ut/potami/internal/boardcast"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/syncc"
)

var TaskBoardCasts = new(syncc.Map[string, *boardcast.Boardcast[*task.Task]])

func BoardCastCreateOrLoad(t *task.Task, selfCheck func(t *task.Task) string) *boardcast.Boardcast[*task.Task] {
	b, _ := TaskBoardCasts.LoadOrStore(t.ID, boardcast.NewBoardCast(t, selfCheck))
	return b
}
