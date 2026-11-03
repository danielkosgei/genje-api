CREATE TABLE nominee_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,
    party TEXT NOT NULL,
    heading TEXT NOT NULL,
    nominee_name TEXT NOT NULL,
    county TEXT,
    source_page INT NOT NULL CHECK (source_page > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source, nominee_name, source_page, heading)
);

CREATE INDEX idx_nominee_lists_party ON nominee_lists (party);
CREATE INDEX idx_nominee_lists_county ON nominee_lists (county);
