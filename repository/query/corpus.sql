
-- name: ListCorpus :many
SELECT * FROM corpus;


-- name: GetCorpus :one
SELECT * FROM corpus WHERE name = ?;


-- name: CreateCorpus :one
INSERT INTO corpus (name, description, collection_name, rerank_fusion_method, use_bm25_index, use_context_embed, embedding_provider, embedding_model, embedding_options, vector_store, vector_dimension, vector_metric_type, context_embed_prompt)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateCorpus :one
UPDATE corpus
SET description = ?, collection_name = ?, rerank_fusion_method = ?, use_bm25_index = ?, use_context_embed = ?, embedding_provider = ?, embedding_model = ?, embedding_options = ?, vector_store = ?, vector_dimension = ?, vector_metric_type = ?, context_embed_prompt = ?
WHERE name = ?
RETURNING *;

-- name: DeleteCorpus :exec
DELETE FROM corpus WHERE name = ?;