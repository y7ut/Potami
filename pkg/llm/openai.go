package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared/constant"
	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/message"
)

const (
	OpenAIDefaultCompletionModel = openai.ChatModelGPT4o
)

var cacheControlFields = map[string]interface{}{"cache_control": map[string]interface{}{"type": "ephemeral"}}

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

type OpenAIUsageDetail struct {
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	CacheTokens      int64
	ExtraFields      map[string]interface{}
}

func (u OpenAIUsageDetail) Format() map[string]interface{} {
	usageDetail := map[string]interface{}{
		"prompt_tokens":     u.PromptTokens,
		"completion_tokens": u.CompletionTokens,
		"total_tokens":      u.TotalTokens,
		"cache_tokens":      u.CacheTokens,
	}

	for k, v := range u.ExtraFields {
		usageDetail[k] = v
	}
	return usageDetail
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
	resp, err := p.OpenAIClient.Chat.Completions.New(ctx, *param, option.WithJSONSet("usage", map[string]interface{}{"include": true}), option.WithJSONSet("include_reasoning", true))
	if err != nil {
		return "", err
	}

	getPrize, ok := prizeMap[resp.Model]
	if !ok {
		getPrize = prizeMap["gpt-4o"]
	}

	p.tracer.Billing(getPrize(resp.Usage.PromptTokens, resp.Usage.CompletionTokens))
	p.tracer.Usage(OpenAIUsageDetail{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
		CacheTokens:      resp.Usage.PromptTokensDetails.CachedTokens,
	})
	tid := p.tracer.GetCurrentTraceID()
	logrus.WithField("trace_id", tid).WithFields(
		logrus.Fields{
			"prompt_tokens":     resp.Usage.PromptTokens,
			"completion_tokens": resp.Usage.CompletionTokens,
			"total_token":       resp.Usage.TotalTokens,
			"cache_token":       resp.Usage.PromptTokensDetails.CachedTokens,
		},
	).Info("OpenAI completion usage")

	return resp.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) buildParams(messages []*message.Message) (*openai.ChatCompletionNewParams, error) {
	openaiMessages, err := MessagesToOpenAI(messages)
	if err != nil {
		return nil, err
	}
	req := &openai.ChatCompletionNewParams{
		Model:       task.MustBindWithOption(p.tracer, "model", OpenAIDefaultCompletionModel),
		Messages:    openaiMessages,
		Temperature: openai.Float(task.MustBindWithOption(p.tracer, "temperatur", 0.7)),
		TopP:        openai.Float(task.MustBindWithOption(p.tracer, "top_p", 1.0)),
	}

	return req, nil

}

func MessagesToOpenAI(messages []*message.Message) ([]openai.ChatCompletionMessageParamUnion, error) {
	if len(messages) == 0 {
		return []openai.ChatCompletionMessageParamUnion{}, nil
	}

	var result []openai.ChatCompletionMessageParamUnion
	var currentGroup []*message.Message
	var currentRole string

	// 处理每个消息，按角色分组
	for _, msg := range messages {
		if msg.Role != currentRole {
			// 角色变化，处理之前累积的消息组
			if len(currentGroup) > 0 {
				openaiMsg, err := createOpenAIMessage(currentGroup)
				if err != nil {
					return nil, err
				}
				result = append(result, openaiMsg)
			}
			// 开始新的消息组
			currentGroup = []*message.Message{msg}
			currentRole = msg.Role
		} else {
			// 相同角色，添加到当前组
			currentGroup = append(currentGroup, msg)
		}
	}

	// 处理最后一组消息
	if len(currentGroup) > 0 {
		openaiMsg, err := createOpenAIMessage(currentGroup)
		if err != nil {
			return nil, err
		}
		result = append(result, openaiMsg)
	}

	return result, nil
}

// createOpenAIMessage 创建 OpenAI 消息的辅助函数
func createOpenAIMessage(messageGroup []*message.Message) (openai.ChatCompletionMessageParamUnion, error) {
	if len(messageGroup) == 0 {
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("empty message group")
	}

	role := messageGroup[0].Role

	if len(messageGroup) == 1 {
		// 单条消息，直接创建
		return createSingleOpenAIMessage(messageGroup[0])
	}

	// 多条消息，需要合并内容
	return createMergedOpenAIMessage(messageGroup, role)
}

// createSingleOpenAIMessage 创建单条 OpenAI 消息
func createSingleOpenAIMessage(msg *message.Message) (openai.ChatCompletionMessageParamUnion, error) {
	switch msg.Role {
	case message.RoleUser:
		openaiMessage := openai.UserMessage(msg.Content)
		if msg.CacheControl {
			openaiMessage.OfUser.SetExtraFields(cacheControlFields)
		}
		return openaiMessage, nil

	case message.RoleAssistant:
		openaiMessage := openai.AssistantMessage(msg.Content)
		if msg.CacheControl {
			openaiMessage.OfAssistant.SetExtraFields(cacheControlFields)
		}
		return openaiMessage, nil

	case message.RoleSystem:
		openaiMessage := openai.SystemMessage(msg.Content)
		if msg.CacheControl {
			openaiMessage.OfSystem.SetExtraFields(cacheControlFields)
		}
		return openaiMessage, nil

	default:
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("message role is invalid: %s", msg.Role)
	}
}

// createMergedOpenAIMessage 创建合并的 OpenAI 消息
func createMergedOpenAIMessage(messageGroup []*message.Message, role string) (openai.ChatCompletionMessageParamUnion, error) {
	switch role {
	case message.RoleUser:
		contents := make([]openai.ChatCompletionContentPartUnionParam, 0, len(messageGroup))
		for _, msg := range messageGroup {
			textParam := &openai.ChatCompletionContentPartTextParam{
				Text: msg.Content,
				Type: constant.Text("text"),
			}
			if msg.CacheControl {
				textParam.SetExtraFields(cacheControlFields)
			}
			contentPart := openai.ChatCompletionContentPartUnionParam{
				OfText: textParam,
			}
			contents = append(contents, contentPart)
		}
		return openai.UserMessage(contents), nil

	case message.RoleAssistant:
		contents := make([]openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion, 0, len(messageGroup))
		for _, msg := range messageGroup {
			textParam := &openai.ChatCompletionContentPartTextParam{
				Text: msg.Content,
				Type: constant.Text("text"),
			}
			if msg.CacheControl {
				textParam.SetExtraFields(cacheControlFields)
			}
			contentPart := openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{
				OfText: textParam,
			}
			contents = append(contents, contentPart)
		}
		return openai.AssistantMessage(contents), nil

	case message.RoleSystem:
		contents := make([]openai.ChatCompletionContentPartTextParam, 0, len(messageGroup))
		for _, msg := range messageGroup {
			textParam := openai.ChatCompletionContentPartTextParam{
				Text: msg.Content,
				Type: constant.Text("text"),
			}
			if msg.CacheControl {
				textParam.SetExtraFields(cacheControlFields)
			}
			contents = append(contents, textParam)
		}
		return openai.SystemMessage(contents), nil

	default:
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("message role is invalid: %s", role)
	}
}
