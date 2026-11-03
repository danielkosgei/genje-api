CREATE TABLE pollsters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE polls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pollster_id UUID NOT NULL REFERENCES pollsters(id),
    race TEXT NOT NULL CHECK (race IN ('president', 'governor', 'senate', 'party', 'other')),
    fieldwork_start DATE,
    fieldwork_end DATE,
    sample_size INT CHECK (sample_size IS NULL OR sample_size > 0),
    margin_of_error NUMERIC(4,2) CHECK (margin_of_error IS NULL OR (margin_of_error >= 0 AND margin_of_error <= 100)),
    published_on DATE NOT NULL,
    source_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (fieldwork_end IS NULL OR fieldwork_start IS NULL OR fieldwork_end >= fieldwork_start)
);

CREATE TABLE poll_results (
    poll_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    candidate_name TEXT NOT NULL,
    party TEXT,
    share NUMERIC(5,2) NOT NULL CHECK (share >= 0 AND share <= 100),
    PRIMARY KEY (poll_id, candidate_name)
);

CREATE INDEX idx_polls_published ON polls (published_on DESC);
