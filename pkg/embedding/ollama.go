package embedding

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ollama/ollama/api"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
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

	response, err := o.Client.Embeddings(ctx, &api.EmbeddingRequest{
		Prompt:  text,
		Model:   task.MustBindWithOption(o.options, "embedding_model", OllamaEmbeddingModel),
		Options: o.OllamaOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %v", err)
	}

	return response.Embedding, nil
}
