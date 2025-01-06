package retrieval

import (
	"context"
)

type SearchEngine interface {
	Search(ctx context.Context, query string) (DocumentCollection, error)
}

// SearchRetriever 搜索检索器
type SearchRetriever[T SearchEngine] struct {
	SearchEngine T
}

// Query Implements Retriever
func (sr *SearchRetriever[T]) Query(ctx context.Context, query string) (DocumentCollection, error) {
	return sr.SearchEngine.Search(ctx, query)
}

func NewSearchRetriever[T SearchEngine](searchEngine T) *SearchRetriever[T] {
	return &SearchRetriever[T]{SearchEngine: searchEngine}
}
