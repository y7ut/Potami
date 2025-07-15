ALTER TABLE corpus ADD COLUMN embedding_provider TEXT;
ALTER TABLE corpus ADD COLUMN embedding_model TEXT;
ALTER TABLE corpus ADD COLUMN embedding_options TEXT;
ALTER TABLE corpus ADD COLUMN vector_dimension INTEGER;
ALTER TABLE corpus ADD COLUMN vector_metric_type TEXT;
ALTER TABLE corpus ADD COLUMN context_embed_prompt TEXT;