package llm

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/parser"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/message"

	"github.com/openai/openai-go"
)

var prizeMap = map[string]func(inputToken, outputToken int64) float64{
	"gpt-4o": func(inputToken, outputToken int64) float64 {
		return float64(inputToken)*2.5/1000000 + float64(outputToken)*10/1000000
	},
}

// Dialog 对话
// 目前可用的options:
// - model
// - temperature
// - top_p
type Dialog struct {
	task.JobHelper

	System   string
	Template string

	Intput []string
	Output []string
}

func (d *Dialog) Handle(ctx context.Context) error {
	client := openai.NewClient(conf.GetOpenAIOptions()...)

	param, err := d.buildRequestParam()
	if err != nil {
		d.Logger().WithError(err).Error("generate param error")
		return err
	}

	chatCompletion, err := client.Chat.Completions.New(ctx, *param)
	if err != nil {
		d.Logger().WithError(err).Error("chat completion error")
		return err
	}

	d.billingWithTokenizer(chatCompletion)

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
func (d *Dialog) buildRequestParam() (*openai.ChatCompletionNewParams, error) {
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

	p := &openai.ChatCompletionNewParams{
		Model:    openai.Raw[string](d.GetOptionWithDefault("model", "gpt-4o")),
		Messages: openai.F(openaiMessages),
	}

	if temperatur, ok := d.GetOption("temperature"); ok {
		p.Temperature = openai.Raw[float64](temperatur)
	} else if TopP, ok := d.GetOption("top_p"); ok {
		p.TopP = openai.Raw[float64](TopP)
	} else {
		p.Temperature = openai.Raw[float64](1.0)
	}

	return p, nil
}

// billingWithTokenizer 计算费用
func (d *Dialog) billingWithTokenizer(chatCompletion *openai.ChatCompletion) (int64, int64, float64) {
	inputToken := chatCompletion.Usage.PromptTokens
	outputToken := chatCompletion.Usage.CompletionTokens

	model, _ := d.GetOption("model")
	getPrize, ok := prizeMap[model.(string)]
	if !ok {
		getPrize = prizeMap["gpt-4o"]
	}
	prize := getPrize(inputToken, outputToken)
	d.Billing(prize)
	d.Logger().Debug(fmt.Sprintf("chat completion finish with input token: %d, output token: %d, prize: %f", inputToken, outputToken, prize))
	return inputToken, outputToken, prize
}
