package spliter

import (
	"context"

	"github.com/y7ut/potami/internal/document"
)

type LengthSplitter struct {
	LenFunc   func(text string) int
	ChunkSize int
}

func (ls *LengthSplitter) Split(ctx context.Context, resource *document.Resource) document.DocumentCollection {
	texts := splitByLength(resource.Content, ls.LenFunc, ls.ChunkSize)

	documents := make([]*document.Document, 0)
	for _, text := range texts {
		documents = append(documents, &document.Document{
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

func splitByLength(text string, lenFunc func(text string) int, chunkSize int) []string {
	var chunks []string
	var start int
	var end int

	for {
		textLength := lenFunc(text)
		if start >= textLength {
			break
		}
		end = min(start+chunkSize, textLength)
		chunks = append(chunks, text[start:end])
		start = end
	}

	return chunks
}
