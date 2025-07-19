package embedding

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
)

const (
	OpenAIDefaultEmbeddingModel = openai.EmbeddingModelTextEmbeddingAda002
)

type OpenAIEmbedding struct {
	client *openai.Client

	options task.WithOption
}

func NewOpenAIEmbedding(options task.WithOption) *OpenAIEmbedding {
	client := openai.NewClient(conf.GetOpenAIOptions()...)
	return &OpenAIEmbedding{
		client:  &client,
		options: options,
	}
}

// Embedding
// Supporting models:
// 1. text-embedding-ada-002
// 2. text-embedding-3-small
// 3. text-embedding-3-large
func (o *OpenAIEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {

	embeddingParams := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model:          task.MustBindWithOption(o.options, "model", OpenAIDefaultEmbeddingModel),
		Dimensions:     openai.Int(task.MustBindWithOption[int64](o.options, "dimensions", 1024)),
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	}
	res, err := o.client.Embeddings.New(ctx, embeddingParams)
	if err != nil {
		return nil, err
	}
	return res.Data[0].Embedding, nil
}
