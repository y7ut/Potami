package stream

import (
	"context"

	"github.com/y7ut/potami/internal/db"
)

func List(ctx context.Context) ([]*db.Stream, error) {
	streams, err := db.GetQueries().ListStreams(context.Background())
	if err != nil {
		return nil, err
	}

	return streams, nil
}
