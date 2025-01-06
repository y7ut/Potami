package retrieval

import (
	"context"
)

type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}
