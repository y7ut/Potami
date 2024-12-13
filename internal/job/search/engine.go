package search

import (
	"context"
)

// SearchEngine 搜索引擎
type Engine interface {
	Search(ctx context.Context, param *SearchParam) ([]Document, error)
}
