
-- name: ListStreams :many
SELECT * FROM streams;


-- name: GetStreamByName :one
SELECT * FROM streams WHERE name = ?;


-- name: CreateStream :one
INSERT INTO streams (name, description, level)
VALUES (?, ?, ?)
RETURNING *;


-- name: UpdateStream :one
UPDATE streams
SET description = ?, level = ?
WHERE id = ?
RETURNING *;


-- name: DeleteStream :exec
DELETE FROM streams WHERE name = ?;