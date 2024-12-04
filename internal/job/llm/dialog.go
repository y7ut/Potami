package llm

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/parser"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/message"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Dialog struct {
	task.JobHelper

	System      string
	Template    string
	Temperature float64
	TopP        float64
	Model       string

	Intput []string
	Output []string
}

func (d *Dialog) Handle(ctx context.Context) error {
	client := openai.NewClient(
		option.WithAPIKey(conf.OpenAI.APIKey),
		// option.WithBaseURL(conf.OpenAI.BaseURL),
	)

	param, err := d.generateParam()
	if err != nil {
		d.Logger().WithError(err).Error("generate param error")
		return err
	}

	chatCompletion, err := client.Chat.Completions.New(ctx, *param)

	if err != nil {
		d.Logger().WithError(err).Error("chat completion error")
		return err
	}

	outputAttributes, err := parser.XMLOutPutParser(chatCompletion.Choices[0].Message.Content, d.Output...)
	if err != nil {
		d.Logger().WithError(err).Error("output parse error")
		return err
	}
	d.SetAttributes(outputAttributes)
	d.Logger().WithFields(outputAttributes).Debug("dialog by prompt complete")
	return nil
}

// generateParam 生成参数
func (d *Dialog) generateParam() (*openai.ChatCompletionNewParams, error) {
	messageTemplate := message.NewPromptTemplate(
		message.NewSystemMessage(d.System),
		message.NewUserMessage(d.Template),
	)

	messages, err := messageTemplate.RenderMessages(d.GetAttributes(d.Intput...))
	if err != nil {
		d.Logger().WithError(err).Error("message template render error")
		return nil, err
	}
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

	return &openai.ChatCompletionNewParams{
		Model: openai.F(d.Model),
		// Temperature: openai.F(d.Temperature),
		// TopP:        openai.F(d.TopP),
		Messages: openai.F(openaiMessages),
	}, nil
}
