package embedding

import (
	"context"
	"net/http"

	"github.com/ollama/ollama/api"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/param"
)

const (
	OllamaEmbeddingModel = "nomic-embed-text:latest"
)

type OllamaEmbedding struct {
	Client *api.Client

	Model         string
	OllamaOptions map[string]interface{}

	options task.WithOption
}

func NewOllamaEmbedding(options task.WithOption) *OllamaEmbedding {
	return &OllamaEmbedding{
		Client: api.NewClient(
			conf.Ollama.GetURL(),
			http.DefaultClient,
		),

		options: options,
	}
}

func (o *OllamaEmbedding) Embed(ctx context.Context, text string) ([]float64, error) {
	if err := o.applyParams(); err != nil {
		return nil, err
	}
	response, err := o.Client.Embeddings(ctx, &api.EmbeddingRequest{
		Prompt:  text,
		Model:   o.Model,
		Options: o.OllamaOptions,
	})
	if err != nil {
		return nil, err
	}
	
	return response.Embedding, nil
}

func (o *OllamaEmbedding) applyParams() error {
	if err := param.Assign(&o.Model, o.options.GetOptionWithDefault("model", OllamaEmbeddingModel)); err != nil {
		return err
	}
	return nil
}
