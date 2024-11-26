
-- 创建 streams 表
CREATE TABLE streams (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    level INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX streams_unique_name ON streams(name);

INSERT INTO  "streams" ("name", "description") VALUES ('company_qa', '企业问答');


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

INSERT INTO  "jobs" ("stream_id", "sorted", "name", "type", "description", "llm_model", "temperature", "top_p", "template", "params", "output") VALUES ('1', '1', 'company_question_analyze', 'prompt', '公司问答分析', 'gpt-4o', '0.7', '1', '分析所提供的问题，返回问题本身所提问的公司名，并从以下几个维度中返回要解答问题所需要的信息类型。
        注意！请不要回答问题，也不要返回下列信息维度类型以外的类型。
        1. 研发实力
        2. 组织结构
        3. 行业影响
        4. 关联企业
        5. 融资进程
        6. 主要产品
        
        问题如下：
        {{.question}}

        输出：
        请对问题进行分析，在适当的 XML 标记内输出结果：

        <company>
        [在此处插入问题的所提问的公司]
        </company>

        <chapter>
        [在此处插入解答问题所需要的维度信息类型]
        </chapter>', 'question,companyId', 'company,chapter');

INSERT INTO  "jobs" ("stream_id", "sorted", "name", "type", "description", "method", "endpoint", "params", "output_parses") VALUES ('1', '2', 'company_detail', 'api_tool', '根据维度信息获取公司详情', 'GET', 'https://eip.ijiwei.com/api-free/eip/companyIntro?type=1', 'companyId,company,chapter', '{"document": "$.data.[1].content"}');

INSERT INTO  "jobs" ("stream_id", "sorted", "name", "type", "description", "llm_model", "template", "params", "output") VALUES ('1', '3', 'company_rag', 'prompt', '公司问答解答', 'gpt-4o', '根据已知的有关信息回答问题，已知信息不包含问题所需的信息时，请自行检索问题的答案：
        已知信息如下：
        {{.document}}
        问题如下：
        {{.question}}

        输出：
        根据要求回答问题，在适当的 XML 标记内输出结果：

        <output>
        [在此处插入问题的回答]
        </output>', 'document,question', 'output');