package vector

import (
	"context"

	"github.com/y7ut/potami/internal/document"
)

type Store interface {
	Search(ctx context.Context, vector []float64) (document.DocumentCollection, error)
	Upsert(ctx context.Context, vector []float64, doc document.Document) error
}
