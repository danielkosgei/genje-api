-- The politician_websites table was never filled, so it is replaced by one table of cited links. A link is either a
-- profile page (on parliament.go.ke, for example) or the politician's own website. Each one names its publisher and
-- whether that publisher is official, so the app can say where a link came from.
DROP TABLE IF EXISTS politician_websites;

CREATE TABLE politician_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    politician_id UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('profile', 'website')),
    title TEXT NOT NULL,
    url TEXT NOT NULL CHECK (url LIKE 'https://%'),
    source_id UUID NOT NULL REFERENCES sources(id),
    publisher_type TEXT NOT NULL CHECK (publisher_type IN ('official', 'third_party')),
    checked_on DATE NOT NULL,
    check_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (politician_id, url)
);
CREATE INDEX idx_politician_links_politician ON politician_links (politician_id);
