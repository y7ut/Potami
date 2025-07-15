package llm

import (
	"context"

	"github.com/y7ut/potami/pkg/message"
)

type Provider interface {
	Complete(ctx context.Context, messages []*message.Message) (string, error)
}
