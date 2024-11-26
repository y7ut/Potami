package stream

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/y7ut/potami/internal/db"
	"github.com/y7ut/potami/internal/schema"
)

// Apply 应用一个新的Stream
func Apply(ctx context.Context, stream *schema.Stream) error {
	StreamApplyed, err := db.GetQueries().GetStreamByName(ctx, stream.Name)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("find stream error, %s", err)
	}
	if transactionErr := db.Transaction(ctx, func(ctx context.Context, qtx *db.Queries) error {
		if err == sql.ErrNoRows {
			StreamApplyed, err = db.GetQueries().CreateStream(ctx, db.CreateStreamParams{
				Name:        stream.Name,
				Description: sql.NullString{String: stream.Description, Valid: true},
				Level:       sql.NullInt64{Int64: int64(stream.Level), Valid: true},
			})

			if err != nil {
				return fmt.Errorf("create stream error, %s", err)
			}
		} else {
			updateParams := &db.UpdateStreamParams{
				Description: sql.NullString{String: stream.Description, Valid: true},
				Level:       sql.NullInt64{Int64: int64(stream.Level), Valid: true},
				ID:          StreamApplyed.ID,
			}

			if _, err := qtx.UpdateStream(ctx, *updateParams); err != nil {
				return fmt.Errorf("update stream error, %s", err)
			}
		}

		JobsofStream, err := db.GetQueries().ListJobsByStreamID(ctx, StreamApplyed.ID)
		if err != nil {
			return fmt.Errorf("get jobs error, %s", err)
		}

		jobIds := make([]int64, 0)
		for _, job := range JobsofStream {
			jobIds = append(jobIds, job.ID)
		}

		if len(JobsofStream) > 0 {
			if err := qtx.DeleteJobs(ctx, jobIds); err != nil {
				return fmt.Errorf("delete jobs error, %s", err)
			}
		}

		for i, job := range stream.Jobs {
			var outputParses strings.Builder
			if job.OutputParses != nil {
				b, err := json.Marshal(job.OutputParses)
				if err != nil {
					return fmt.Errorf("json marshal output parses error, %s", err)
				}
				outputParses.Write(b)
			}

			var searchOptions strings.Builder
			if job.SearchOptions != nil {
				b, err := json.Marshal(job.SearchOptions)
				if err != nil {
					return fmt.Errorf("json marshal search options error, %s", err)
				}
				searchOptions.Write(b)
			}

			if _, err := qtx.CreateJob(ctx, db.CreateJobParams{
				StreamID:    StreamApplyed.ID,
				Sorted:      int64(i + 1),
				Name:        job.Name,
				Type:        job.Type,
				Description: sql.NullString{String: job.Description, Valid: true},
				LlmModel:    sql.NullString{String: job.LlmModel, Valid: true},
				SystemPrompt: sql.NullString{
					String: job.SystemPrompt,
					Valid:  job.SystemPrompt != "",
				},
				MaxTokens: sql.NullInt64{Int64: int64(job.MaxTokens), Valid: job.MaxTokens > 0},
				TopP:      sql.NullFloat64{Float64: job.TopP, Valid: job.TopP != 0},
				Temperature: sql.NullFloat64{
					Float64: job.Temperature,
					Valid:   job.Temperature != 0,
				},
				Template: sql.NullString{String: job.Template, Valid: true},
				Endpoint: sql.NullString{String: job.Endpoint, Valid: true},
				Method:   sql.NullString{String: job.Method, Valid: true},
				Params:   sql.NullString{String: strings.Join(job.Params, ","), Valid: true},
				Output:   sql.NullString{String: strings.Join(job.Output, ","), Valid: true},
				OutputParses: sql.NullString{
					String: outputParses.String(),
					Valid:  len(job.OutputParses) > 0,
				},
				SearchOptions: sql.NullString{
					String: searchOptions.String(),
					Valid:  len(job.SearchOptions) > 0,
				},
				SearchEngine: sql.NullString{String: job.SearchEngine, Valid: true},
				QueryField:   sql.NullString{String: job.QueryField, Valid: true},
				OutputField:  sql.NullString{String: job.OutputField, Valid: true},
			}); err != nil {
				return fmt.Errorf("create job error, %s", err)
			}
		}
		return nil
	}); transactionErr != nil {
		return transactionErr
	}
	return nil
}

func RemoveStream(ctx context.Context, name string) error {
	StreamApplyed, err := db.GetQueries().GetStreamByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("find stream error, %s", err)
	}

	if transactionErr := db.Transaction(ctx, func(ctx context.Context, qtx *db.Queries) error {
		if err := qtx.DeleteStream(ctx, StreamApplyed.Name); err != nil {
			return fmt.Errorf("delete stream error, %s", err)
		}
		if err := qtx.DeleteJobsByStreamID(ctx, StreamApplyed.ID); err != nil {
			return fmt.Errorf("delete jobs error, %s", err)
		}
		return nil

	}); transactionErr != nil {
		return transactionErr
	}
	return nil
}
