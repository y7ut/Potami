
-- 创建 streams 表
CREATE TABLE streams (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    level INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX streams_unique_name ON streams(name);


-- 创建 jobs 表
CREATE TABLE jobs (
    id INTEGER NOT NULL PRIMARY KEY,
    stream_id INTEGER NOT NULL,
    sorted INTEGER NOT NULL DEFAULT 1,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    llm_model TEXT,
    temperature REAL,
    top_p REAL,
    max_tokens INTEGER,
    template TEXT,
    system_prompt TEXT,
    method TEXT,
    endpoint TEXT,
    params TEXT,  -- 以逗号分隔的参数列表
    output TEXT,  -- 以逗号分隔的输出字段
    output_parses TEXT,  -- JSON 格式的输出解析
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX jobs_unique_name ON jobs(stream_id, name);
