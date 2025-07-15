package corpus

import (
	"context"

	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/pkg/vector"
)

// Corpus 知识库
type Corpus struct {
	Topic          string
	Description    string
	CollectionName string

	// 是否使用BM25权重索引
	UseBm25Index bool
	// 是否使用上下文嵌入
	UseContextEmbed bool

	// RerankFusionFunc func(docs DocumentCollection) (DocumentCollection, error)

	// VectorStore
	// 用于向量查询, 如 Milvus
	VectorStore vector.Store
}

func (c *Corpus) Embedding(ctx context.Context, text string) ([]float64, error) {
	return nil, nil
}

func (c *Corpus) Upsert(ctx context.Context, vector []float64, doc document.Document) error {
	return nil
}

func (c *Corpus) Insert(ctx context.Context, vector []float64, doc document.Document) error {
	return nil
}
