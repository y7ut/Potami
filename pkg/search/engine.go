package search

import (
	"context"

	"github.com/y7ut/potami/internal/document"
)

type SearchEngine interface {
	Search(ctx context.Context, query string) (document.DocumentCollection, error)
}
