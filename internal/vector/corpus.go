package vector

import (
	"context"
	"time"

	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/pkg/embedding"
	"github.com/y7ut/potami/pkg/vector/milvus"
)

type Corpus struct {
	Topic           string
	Description     string
	CollectionName  string
	UseBm25Index    bool
	UseContextEmbed bool

	Dimension int

	VectorStore   *milvus.PooledConnection
	EmbedProvider embedding.Embed

	LastUsedAt int64
}

func CreateCorpusFromSchema(corpus *schema.Corpus) (*Corpus, error) {
	c := &Corpus{
		Topic:           corpus.Name,
		Description:     corpus.Description,
		CollectionName:  corpus.CollectionName,
		UseBm25Index:    corpus.UseBm25Index,
		UseContextEmbed: corpus.UseContextEmbed,
	}

	// var err error
	// c.VectorStore, err = milvus.NewMilvus(
	// 	milvus.WithCollection(corpus.CollectionName),
	// 	milvus.WithUseBM25(corpus.UseBm25Index),
	// 	milvus.WithUseBM25(corpus.UseContextEmbed),
	// )

	// if err != nil {
	// 	return nil, err
	// }

	return c, nil
}

func (c *Corpus) Upsert(ctx context.Context, doc document.Document) error {
	defer func() {
		c.LastUsedAt = time.Now().Unix()
	}()

	return c.VectorStore.Upsert(ctx, doc)
}

func (c *Corpus) Search(ctx context.Context, query string, vector []float64, limit int) (document.DocumentCollection, error) {
	return c.VectorStore.Search(ctx, query, vector, limit)
}

func (c *Corpus) Embedding(ctx context.Context, text string) ([]float64, error) {
	return c.EmbedProvider.Embed(ctx, text)
}
