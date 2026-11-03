-- A photo is stored by the API and linked to the official source it came from.
CREATE TABLE politician_photos (
    politician_id UUID PRIMARY KEY REFERENCES politicians(id) ON DELETE CASCADE,
    source_id UUID REFERENCES sources(id),
    source_url TEXT NOT NULL,
    file_path TEXT NOT NULL,
    sha256 TEXT NOT NULL CHECK (char_length(sha256) = 64),
    bytes INT NOT NULL CHECK (bytes > 0),
    content_type TEXT NOT NULL,
    retrieved_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A politician's official website, found on an official listing and linked to that listing.
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
