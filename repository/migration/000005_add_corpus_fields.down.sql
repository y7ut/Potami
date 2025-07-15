ALTER TABLE corpus DROP COLUMN embedding_provider;
ALTER TABLE corpus DROP COLUMN embedding_model;
ALTER TABLE corpus DROP COLUMN embedding_options;
ALTER TABLE corpus DROP COLUMN vector_dimension;
ALTER TABLE corpus DROP COLUMN vector_metric_type;
ALTER TABLE corpus DROP COLUMN context_embed_prompt;
ALTER TABLE corpus RENAME COLUMN vector_store_old TO vector_driver;