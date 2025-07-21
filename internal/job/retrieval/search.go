package retrieval

import (
	"context"
	"log"

	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/internal/vector"
	"github.com/y7ut/potami/pkg/search"
)

// Retriever SearchRetriever implements Retriever
var _ Retrieval = (*WebSearchRetriever[search.SearchEngine])(nil)

// Retriever CorpusSearchRetriever implements Retriever
var _ Retrieval = (*KnowledgeBaseSearchRetriever)(nil)

// WebSearchRetriever 联网搜索检索器
type WebSearchRetriever[T search.SearchEngine] struct {
	SearchEngine T
}

// Query Implements Retriever
func (sr *WebSearchRetriever[T]) Query(ctx context.Context, query string) (document.DocumentCollection, error) {
	return sr.SearchEngine.Search(ctx, query)
}

func NewWebSearchRetriever[T search.SearchEngine](searchEngine T) *WebSearchRetriever[T] {
	return &WebSearchRetriever[T]{SearchEngine: searchEngine}
}

// KnowledgeBaseSearchRetriever 语料库检索器
type KnowledgeBaseSearchRetriever struct {
	Corpus *vector.Corpus

	options task.WithOption
}

// Query Implements Retriever
func (sr *KnowledgeBaseSearchRetriever) Query(ctx context.Context, query string) (document.DocumentCollection, error) {
	limit := task.MustBindWithOption(sr.options, "query_size", 3)
	log.Println("corpus", sr.Corpus)
	docs, err := sr.Corpus.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func NewKnowledgeBaseSearchRetriever(corpus *vector.Corpus, options task.WithOption) *KnowledgeBaseSearchRetriever {
	return &KnowledgeBaseSearchRetriever{Corpus: corpus, options: options}
}
