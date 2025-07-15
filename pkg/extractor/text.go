package extractor

import (
	"context"
	"io"
	"strings"

	"github.com/y7ut/potami/internal/document"
)

const (
	TextContentType = "text/plain"
)

type TextExtractor struct{}

// Extract 从资源中提取文本
func (te *TextExtractor) Extract(ctx context.Context, resource io.ReadSeekCloser) (*document.Resource, error) {
	// 将资源内容读取到缓冲区
	buf := new(strings.Builder)
	_, err := io.Copy(buf, resource)
	if err != nil {
		return nil, err
	}

	// 返回提取的文本
	return &document.Resource{Content: buf.String(), MineType: TextContentType}, nil
}
