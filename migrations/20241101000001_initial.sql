-- Create news_sources table
CREATE TABLE news_sources (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    base_url VARCHAR NOT NULL,
    rss_url VARCHAR,
    scrape_config JSONB,
    source_type JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_fetched TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    language VARCHAR NOT NULL,
    country VARCHAR NOT NULL,
    region VARCHAR,
    credibility_score REAL
);

-- Create articles table
CREATE TABLE articles (
    id UUID PRIMARY KEY,
    title VARCHAR NOT NULL,
    content TEXT,
    summary TEXT,
    url VARCHAR NOT NULL UNIQUE,
    author VARCHAR,
    published_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    source_id UUID NOT NULL REFERENCES news_sources(id),
    category VARCHAR,
    tags TEXT[] DEFAULT '{}',
    is_trending BOOLEAN NOT NULL DEFAULT false,
    view_count BIGINT NOT NULL DEFAULT 0,
    language VARCHAR NOT NULL,
    region VARCHAR
);

-- Create indexes for better query performance
CREATE INDEX idx_articles_published_at ON articles(published_at DESC);
CREATE INDEX idx_articles_source_id ON articles(source_id);
CREATE INDEX idx_articles_category ON articles(category);
CREATE INDEX idx_articles_language ON articles(language);
CREATE INDEX idx_articles_region ON articles(region);
CREATE INDEX idx_articles_is_trending ON articles(is_trending);
CREATE INDEX idx_articles_view_count ON articles(view_count DESC);
CREATE INDEX idx_articles_url ON articles(url);

-- Create full-text search index
CREATE INDEX idx_articles_search ON articles USING gin(to_tsvector('english', title || ' ' || COALESCE(summary, '') || ' ' || COALESCE(content, '')));

-- Create indexes on news_sources
CREATE INDEX idx_news_sources_is_active ON news_sources(is_active);
CREATE INDEX idx_news_sources_country ON news_sources(country);
CREATE INDEX idx_news_sources_language ON news_sources(language); 