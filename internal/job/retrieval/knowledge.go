package retrieval

import (
	"context"

	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/internal/vector"
)

type KnowledgeBaseBuilder struct {
	Corpus vector.Corpus
}

func NewKnowledgeBaseBuilder(corpus vector.Corpus) *KnowledgeBaseBuilder {
	return &KnowledgeBaseBuilder{Corpus: corpus}
}

func (c *KnowledgeBaseBuilder) Store(ctx context.Context, doc document.Document) error {
	vectors, err := c.Corpus.Embedding(ctx, doc.Text)
	if err != nil {
		return err
	}
	doc.Embed = vectors

	return c.Corpus.Upsert(ctx, doc)
}
