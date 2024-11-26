
-- name: ListJobs :many
SELECT * FROM jobs;


-- name: GetJobByID :one
SELECT * FROM jobs WHERE id = ?;


-- name: ListJobsByStreamID :many
SELECT * FROM jobs WHERE stream_id = ? ORDER BY sorted;


-- name: CreateJob :one
INSERT INTO jobs (stream_id, sorted, name, type, description, llm_model, system_prompt, max_tokens, top_p, temperature, template, method, endpoint, params, output, output_parses, search_engine, search_options, query_field, output_field)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;


-- name: UpdateJob :one
UPDATE jobs
SET name = ?, type = ?, sorted = ?, description = ?, llm_model = ?, system_prompt = ?, max_tokens = ?, top_p = ?, temperature = ?, template = ?, method = ?, endpoint = ?, params = ?, output = ?, output_parses = ?, search_engine = ?, search_options = ?, query_field = ?, output_field = ?
WHERE id = ?
RETURNING *;


-- name: DeleteJob :exec
DELETE FROM jobs WHERE id = ?;


-- name: DeleteJobs :exec
DELETE FROM jobs WHERE id IN (sqlc.slice('ids'));

-- name: DeleteJobsByStreamID :exec
DELETE FROM jobs WHERE stream_id = ?;