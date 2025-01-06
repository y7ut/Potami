package chat

import (
	"context"

	"github.com/y7ut/potami/internal/task"
)

type OutputParser interface {
	Parse(ctx context.Context, content string, options task.WithOption) (map[string]interface{}, error)
}
