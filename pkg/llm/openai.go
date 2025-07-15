package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/message"
)

const (
	OpenAIDefaultCompletionModel = openai.ChatModelGPT4o
)

var prizeMap = map[string]func(inputToken, outputToken int64) float64{
	"gpt-4o": func(inputToken, outputToken int64) float64 {
		return float64(inputToken)*2.5/1000000 + float64(outputToken)*10/1000000
	},
}

var _ Provider = (*OpenAIProvider)(nil)

type OpenAIProvider struct {
	OpenAIClient *openai.Client

	tracer task.Tracer
}

func NewOpenAIProvider(tracer task.Tracer) *OpenAIProvider {
	client := openai.NewClient(conf.GetOpenAIOptions()...)
	return &OpenAIProvider{
		OpenAIClient: &client,
		tracer:       tracer,
	}
}

func (p *OpenAIProvider) Complete(ctx context.Context, messages []*message.Message) (string, error) {

	param, err := p.buildParams(messages)
	if err != nil {
		return "", err
	}
	resp, err := p.OpenAIClient.Chat.Completions.New(ctx, *param)
	if err != nil {
		return "", err
	}

	getPrize, ok := prizeMap[resp.Model]
	if !ok {
		getPrize = prizeMap["gpt-4o"]
	}
	p.tracer.Billing(getPrize(resp.Usage.PromptTokens, resp.Usage.CompletionTokens))

	return resp.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) buildParams(messages []*message.Message) (*openai.ChatCompletionNewParams, error) {

	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0)
	for _, m := range messages {
		switch m.Role {
		case message.RoleUser:
			openaiMessages = append(openaiMessages, openai.UserMessage(m.Content))
		case message.RoleAssistant:
			openaiMessages = append(openaiMessages, openai.AssistantMessage(m.Content))
		case message.RoleSystem:
			openaiMessages = append(openaiMessages, openai.SystemMessage(m.Content))
		default:
			return nil, fmt.Errorf("message role is invalid")
		}
	}

	req := &openai.ChatCompletionNewParams{
		Model:       task.MustBindWithOption(p.tracer, "model", OpenAIDefaultCompletionModel),
		Messages:    openaiMessages,
		Temperature: openai.Float(task.MustBindWithOption(p.tracer, "temperatur", 0.7)),
		TopP:        openai.Float(task.MustBindWithOption(p.tracer, "top_p", 1.0)),
	}

	return req, nil

}
