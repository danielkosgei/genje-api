CREATE TABLE claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    politician_id UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    claim TEXT NOT NULL,
    rating TEXT NOT NULL CHECK (rating IN ('confirmed', 'unproven', 'contested')),
    rated_on DATE,
    source_title TEXT,
    source_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claims_politician ON claims (politician_id, rated_on DESC NULLS LAST);
