package op

import (
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/syncc"
)

var Tasks = new(syncc.Map[string, *task.Task])
