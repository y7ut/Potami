// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: streams.sql

package db

import (
	"context"
	"database/sql"
)

const createStream = `-- name: CreateStream :one
INSERT INTO streams (name, description, level)
VALUES (?, ?, ?)
RETURNING id, name, description, level, created_at
`

type CreateStreamParams struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Level       sql.NullInt64  `json:"level"`
}

func (q *Queries) CreateStream(ctx context.Context, arg CreateStreamParams) (*Stream, error) {
	row := q.queryRow(ctx, q.createStreamStmt, createStream, arg.Name, arg.Description, arg.Level)
	var i Stream
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.CreatedAt,
	)
	return &i, err
}

const deleteStream = `-- name: DeleteStream :exec
DELETE FROM streams WHERE name = ?
`

func (q *Queries) DeleteStream(ctx context.Context, name string) error {
	_, err := q.exec(ctx, q.deleteStreamStmt, deleteStream, name)
	return err
}

const getStreamByName = `-- name: GetStreamByName :one
SELECT id, name, description, level, created_at FROM streams WHERE name = ?
`

func (q *Queries) GetStreamByName(ctx context.Context, name string) (*Stream, error) {
	row := q.queryRow(ctx, q.getStreamByNameStmt, getStreamByName, name)
	var i Stream
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.CreatedAt,
	)
	return &i, err
}

const listStreams = `-- name: ListStreams :many
SELECT id, name, description, level, created_at FROM streams
`

func (q *Queries) ListStreams(ctx context.Context) ([]*Stream, error) {
	rows, err := q.query(ctx, q.listStreamsStmt, listStreams)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Stream
	for rows.Next() {
		var i Stream
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Level,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateStream = `-- name: UpdateStream :one
UPDATE streams
SET description = ?, level = ?
WHERE id = ?
RETURNING id, name, description, level, created_at
`

type UpdateStreamParams struct {
	Description sql.NullString `json:"description"`
	Level       sql.NullInt64  `json:"level"`
	ID          int64          `json:"id"`
}

func (q *Queries) UpdateStream(ctx context.Context, arg UpdateStreamParams) (*Stream, error) {
	row := q.queryRow(ctx, q.updateStreamStmt, updateStream, arg.Description, arg.Level, arg.ID)
	var i Stream
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.CreatedAt,
	)
	return &i, err
}