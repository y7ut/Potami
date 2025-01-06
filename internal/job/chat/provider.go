package chat

import (
	"context"

	"github.com/y7ut/potami/pkg/message"
)

type LLMProvider interface {
	Complete(ctx context.Context, messages []*message.Message) (string, error)
}
