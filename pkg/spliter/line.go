package spliter

import (
	"context"
	"strings"

	"github.com/y7ut/potami/internal/document"
)

// JsonlineSplitter jsonline分割器
type LineSplitter struct {
	LineBreaks string
}

func (ls *LineSplitter) Split(ctx context.Context, resource *document.Resource) document.DocumentCollection {
	texts := strings.Split(resource.Content, ls.LineBreaks)
	documents := make([]document.Document, 0)
	for _, text := range texts {
		documents = append(documents, document.Document{
			Text: text,
			Name: resource.Name,
			Source: &document.Resource{
				Name:     resource.Name,
				Address:  resource.Address,
				MineType: resource.MineType,
			},
		})
	}
	return documents
}
