package extractor

import (
	"context"
	"io"

	"github.com/y7ut/potami/internal/document"
)

const (
	PDFContentType = "application/pdf"
)

type PDFExtractor struct {
}

// Extract 从 PDF 资源中提取文本
func (pe *PDFExtractor) Extract(ctx context.Context, resource io.ReadSeekCloser) (*document.Resource, error) {
	// 假设使用 pdfcpu 库来解析 PDF
	// 这里需要实现 PDF 文本提取逻辑
	return &document.Resource{MineType: PDFContentType}, nil
}
