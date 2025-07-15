package embedding

import "context"

type Embed interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}
