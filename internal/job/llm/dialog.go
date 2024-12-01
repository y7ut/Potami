package llm

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/task"

	"github.com/openai/openai-go"
)


type Dialog struct {
	task.JobHelper

	Params []string
	Output []string
}

func (d *Dialog) Handle(ctx context.Context) error {
	logrus.WithFields(d.GetAttributes()).WithField("task_id", d.GetTask().ID).Debug("prompt start")

	client := openai.NewClient()
	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say this is a test"),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		panic(err.Error())
	}
	println(chatCompletion.Choices[0].Message.Content)

	logrus.WithFields(d.GetAttributes()).WithField("task_id", d.GetTask().ID).Debug("prompt finish")
	return nil
}
