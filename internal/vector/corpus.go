package vector

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/pkg/embedding"
	"github.com/y7ut/potami/pkg/llm"
	"github.com/y7ut/potami/pkg/message"
	"github.com/y7ut/potami/pkg/vector/milvus"
)

const (
	contentPrompt = `<document>
{{.doc_content}}
</document>`

	chunkPrompt = `Here is the chunk we want to situate within the whole document
<chunk>
{{.chunk_content}}
</chunk>

Please give a short succinct context using the language of the original document to situate this chunk within the overall document for the purposes of improving search retrieval of the chunk.
Answer only with the succinct context and nothing else.`
)

type Corpus struct {
	Topic           string
	Description     string
	CollectionName  string
	UseBm25Index    bool
	UseContextEmbed bool

	VectorDimension int64

	MilvusConnectionManager *milvus.MilvusConnectionPool
	EmbedProvider           embedding.Embed
	LLMProvider             llm.Provider
}

func (c *Corpus) Upsert(ctx context.Context, docs ...*document.Document) error {

	for _, doc := range docs {
		if c.UseContextEmbed {
			if err := c.contextEmbedding(ctx, doc); err != nil {
				return err
			}
		}
		vectors, err := c.EmbedProvider.Embed(ctx, doc.Text)
		if err != nil {
			return err
		}
		doc.Embed = vectors
	}

	milvusConn, err := c.MilvusConnectionManager.GetConnection(
		ctx,
		milvus.WithCollection(c.CollectionName),
		milvus.WithUseBM25(c.UseBm25Index),
		milvus.WithUseContextEmbed(c.UseContextEmbed),
		milvus.WithDimensions(c.VectorDimension),
	)

	if err != nil {
		return err
	}

	return milvusConn.Upsert(ctx, docs...)
}

func (c *Corpus) Search(ctx context.Context, query string, limit int) (document.DocumentCollection, error) {

	vector, err := c.EmbedProvider.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	milvusConn, err := c.MilvusConnectionManager.GetConnection(
		ctx,
		milvus.WithCollection(c.CollectionName),
		milvus.WithUseBM25(c.UseBm25Index),
		milvus.WithUseContextEmbed(c.UseContextEmbed),
		milvus.WithDimensions(c.VectorDimension),
	)
	if err != nil {
		return nil, err
	}
	return milvusConn.Search(ctx, query, vector, limit)
}

func (c *Corpus) Embedding(ctx context.Context, text string) ([]float64, error) {
	return c.EmbedProvider.Embed(ctx, text)
}

func (c *Corpus) contextEmbedding(ctx context.Context, doc *document.Document) error {

	messageTemplate := message.NewPromptTemplate(
		message.NewPromptCacheMessage(message.RoleUser, contentPrompt),
		message.NewUserMessage(chunkPrompt),
	)
	messages, err := messageTemplate.RenderMessages(map[string]interface{}{
		"chunk_content": doc.Text,
		"doc_content":   doc.Source.Content,
	})
	if err != nil {
		return fmt.Errorf("failed to render prompt: %v", err)
	}
	logrus.WithField("doc_id", doc.ID).Debug("Generating context embedding")

	context, err := c.LLMProvider.Complete(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to get context embedding: %v", err)
	}

	logrus.WithField("doc_id", doc.ID).Debug("Generating context succesfully: ", context)

	doc.Context = context
	doc.ContextEmbed, err = c.Embedding(ctx, context)
	if err != nil {
		return fmt.Errorf("failed to generate context embedding: %v", err)
	}

	logrus.WithField("doc_id", doc.ID).Debug("Generated context embedding succesfully")
	return nil
}
