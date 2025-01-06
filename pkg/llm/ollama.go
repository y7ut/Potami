package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ollama/ollama/api"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/message"
	"github.com/y7ut/potami/pkg/param"
)

const (
	OllamaDefaultCompletionModel = "llama3.1:latest"
)

type OllamaProvider struct {
	Client *api.Client

	tracer task.Tracer
}

func NewOllamaProvider(tracer task.Tracer) *OllamaProvider {
	return &OllamaProvider{
		Client: api.NewClient(
			conf.Ollama.GetURL(),
			http.DefaultClient,
		),

		tracer: tracer,
	}
}

func (p *OllamaProvider) Complete(ctx context.Context, messages []*message.Message) (string, error) {
	req, err := p.buildRequest(messages)
	if err != nil {
		return "", err
	}
	var content string
	err = p.Client.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.Message.Content != "" {
			content += resp.Message.Content
		}
		return nil
	})

	return content, nil
}

func (p *OllamaProvider) buildRequest(messages []*message.Message) (*api.ChatRequest, error) {

	ollamaMessages := make([]api.Message, 0)

	for _, m := range messages {
		switch m.Role {
		case message.RoleUser:
			ollamaMessages = append(ollamaMessages, api.Message{Role: "user", Content: m.Content})
		case message.RoleAssistant:
			ollamaMessages = append(ollamaMessages, api.Message{Role: "assistant", Content: m.Content})
		case message.RoleSystem:
			ollamaMessages = append(ollamaMessages, api.Message{Role: "system", Content: m.Content})
		default:
			return nil, fmt.Errorf("message role is invalid")
		}
	}

	var model string
	if err := param.Assign(&model, p.tracer.GetOptionWithDefault("model", OllamaDefaultCompletionModel)); err != nil {
		return nil, err
	}

	stream := false
	ollamaOption := make(map[string]interface{})

	if temp, ok := p.tracer.GetOption("temperature"); ok {
		ollamaOption["temperature"] = temp
	}

	if TopP, ok := p.tracer.GetOption("top_p"); ok {
		ollamaOption["top_p"] = TopP
	}

	if maxToken, ok := p.tracer.GetOption("max_tokens"); ok {
		ollamaOption["num_ctx"] = maxToken
	}

	fmt.Println(ollamaOption)

	return &api.ChatRequest{
		Messages: ollamaMessages,
		Model:    model,
		Stream:   &stream,
		Options:  ollamaOption,
	}, nil
}
