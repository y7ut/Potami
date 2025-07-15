-- 创建 corpus 表
CREATE TABLE corpus (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    collection_name TEXT,
    rerank_fusion_method TEXT,
    vector_store TEXT,
    use_bm25_index bool,
    use_context_embed bool,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated_at DATETIME
);

CREATE UNIQUE INDEX corpus_unique_name ON corpus(name);