package extractor

import (
	"context"
	"io"

	"github.com/y7ut/potami/internal/document"
)

const (
	HTMLContentType = "text/html"
)

type HTMLExtractor struct {
}

// Extract 从 HTML 资源中提取文本
func (he *HTMLExtractor) Extract(ctx context.Context, resource io.ReadSeekCloser) (*document.Resource, error) {
	// 假设使用 goquery 库来解析 HTML
	// 这里需要实现 HTML 文本提取逻辑
	return &document.Resource{MineType: HTMLContentType}, nil
}
