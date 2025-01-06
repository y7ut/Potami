package retrieval

import (
	"context"
)

// VectorStore Interface
// you can implement your own vector store.
type VectorStore interface {
	Query(ctx context.Context, vector []float64) (DocumentCollection, error)
}

// VectorRetriever
// Retrieval with VectorStore, EmbeddingProvider
type VectorRetriever[T VectorStore, E EmbeddingProvider] struct {
	VectorStore       T
	EmbeddingProvider E
}

func (vr *VectorRetriever[T, E]) Query(ctx context.Context, query string) ([]Document, error) {
	vector, err := vr.EmbeddingProvider.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	return vr.VectorStore.Query(ctx, vector)
}

func NewVectorRetriever[T VectorStore, E EmbeddingProvider](vectorStore T, embeddingProvider E) *VectorRetriever[T, E] {
	return &VectorRetriever[T, E]{
		VectorStore:       vectorStore,
		EmbeddingProvider: embeddingProvider,
	}
}
