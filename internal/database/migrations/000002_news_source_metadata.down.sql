DROP INDEX IF EXISTS idx_news_sources_name;
ALTER TABLE news_sources DROP COLUMN IF EXISTS publisher;
ALTER TABLE news_sources DROP COLUMN IF EXISTS default_category;
