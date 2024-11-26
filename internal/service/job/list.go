package job

import (
	"context"

	"github.com/y7ut/potami/internal/db"
)

func List(ctx context.Context) ([]*db.Job, error) {
	jobs, err := db.GetQueries().ListJobs(context.Background())
	if err != nil {
		return nil, err
	}

	return jobs, nil
}
