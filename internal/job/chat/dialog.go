package chat

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/parser"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/message"

	"github.com/y7ut/potami/pkg/llm"
)

// Dialog 对话
// 目前可用的options:
// - model
// - temperature
// - top_p
type Dialog struct {
	task.JobHelper

	Provider llm.Provider

	System   string
	Template string

	Intput []string
	Output []string
}

func (d *Dialog) Handle(ctx context.Context) (err error) {
	defer func() {
		if recoverError := recover(); recoverError != nil {
			err = fmt.Errorf("dialog panic [%v]", recoverError)
		}
		if err != nil {
			d.Logger().WithError(err).Error("dialog failed")
			d.SetError(err)
		}
	}()

	messages, err := d.generateMessage()
	if err != nil {
		err = fmt.Errorf("generate message error: %v", err)
		return
	}

	chatCompletion, err := d.Provider.Complete(ctx, messages)
	if err != nil {
		return
	}

	d.Logger().WithFields(d.GetAttributes()).Debug("llm output: \n", chatCompletion)

	parser := parser.NewXMLOutPutParser(d.Output...)
	outputAttributes, err := parser.Parse(ctx, chatCompletion)
	if err != nil {
		return
	}

	d.SetAttributes(outputAttributes)
	d.Logger().WithFields(outputAttributes).Debug("dialog complete")

	return
}

// generateParam 生成参数
func (d *Dialog) generateMessage() ([]*message.Message, error) {
	messageTemplate := message.NewPromptTemplate(
		message.NewSystemMessage(d.System),
		message.NewUserMessage(d.Template),
	)

	return messageTemplate.RenderMessages(d.GetAttributes(d.Intput...))
}
