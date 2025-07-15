package retrieval

import (
	"context"

	"github.com/y7ut/potami/internal/document"
)

// Splitter 文本分割器
type Splitter interface {
	Split(ctx context.Context, resource *document.Resource) document.DocumentCollection
}
