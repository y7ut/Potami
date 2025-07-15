package embedding

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/param"
)

const (
	OpenAIDefaultEmbeddingModel = openai.EmbeddingModelTextEmbeddingAda002
)

type OpenAIEmbedding struct {
	client     *openai.Client
	Model      string
	Dimensions int64

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

	if err := o.applyParams(); err != nil {
		return nil, err
	}
	embeddingParams := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model:          o.Model,
		Dimensions:     openai.Int(o.Dimensions),
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	}
	res, err := o.client.Embeddings.New(ctx, embeddingParams)
	if err != nil {
		return nil, err
	}
	return res.Data[0].Embedding, nil
}

func (o *OpenAIEmbedding) applyParams() error {

	if err := param.Assign(&o.Model, o.options.GetOptionWithDefault("model", OpenAIDefaultEmbeddingModel)); err != nil {
		return err
	}

	if err := param.Assign(&o.Dimensions, o.options.GetOptionWithDefault("dimensions", 1024)); err != nil {
		return err
	}
	return nil
}
