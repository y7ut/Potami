package retrieval

import (
	"context"
	"io"

	"github.com/y7ut/potami/internal/document"
)

// Extractor 资源提取器
type Extractor interface {
	Extract(ctx context.Context, resource io.ReadSeekCloser) (*document.Resource, error)
}
