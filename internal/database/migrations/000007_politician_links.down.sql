DROP TABLE IF EXISTS politician_links;
CREATE TABLE politician_websites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    politician_id UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    source_id UUID REFERENCES sources(id),
    source_url TEXT NOT NULL,
    retrieved_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (politician_id, url)
);
