package vector

import (
	"context"

	"github.com/y7ut/potami/internal/document"
)

type VectorStore interface {
	Query(ctx context.Context, query string) (document.DocumentCollection, error)
	Upsert(ctx context.Context, doc document.Document) error
}
