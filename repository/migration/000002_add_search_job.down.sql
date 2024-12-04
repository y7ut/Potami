
ALTER TABLE jobs DROP COLUMN search_engine;

ALTER TABLE jobs DROP COLUMN search_options;

ALTER TABLE jobs DROP COLUMN query_field;

ALTER TABLE jobs DROP COLUMN output_field;

DELETE FROM streams WHERE name = 'rag_with_web_search';

DELETE FROM jobs WHERE name = 'chat_completion';

DELETE FROM jobs WHERE name = 'prompt';