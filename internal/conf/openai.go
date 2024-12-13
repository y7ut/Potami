package conf

import (
	"github.com/openai/openai-go/option"
)

func GetOpenAIOptions() []option.RequestOption {
	options := []option.RequestOption{
		option.WithAPIKey(OpenAI.APIKey),
	}
	if OpenAI.BaseURL != "" {
		options = append(options, option.WithBaseURL(OpenAI.BaseURL))
	}
	return options
}
