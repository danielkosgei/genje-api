CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================================
-- Provenance: every fact links back to a verifiable source
-- ============================================================
CREATE TABLE sources (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    url         TEXT,
    type        TEXT NOT NULL CHECK (type IN ('iebc','eacc','gazette','hansard','news','social','court','other')),
    reliability TEXT NOT NULL DEFAULT 'unverified' CHECK (reliability IN ('official','trusted','unverified')),
    last_accessed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Core political entities
-- ============================================================
CREATE TABLE politicians (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug            TEXT NOT NULL UNIQUE,
    first_name      TEXT NOT NULL,
    last_name       TEXT NOT NULL,
    other_names     TEXT,
    date_of_birth   DATE,
    gender          TEXT CHECK (gender IN ('male','female','other')),
    bio             TEXT,
    photo_url       TEXT,
    education       JSONB DEFAULT '[]'::jsonb,
    career_history  JSONB DEFAULT '[]'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_politicians_slug ON politicians(slug);
CREATE INDEX idx_politicians_name ON politicians(last_name, first_name);
CREATE INDEX idx_politicians_search ON politicians USING gin (
    (first_name || ' ' || last_name || ' ' || COALESCE(other_names, '')) gin_trgm_ops
);

CREATE TABLE political_parties (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL,
    abbreviation    TEXT,
    slug            TEXT NOT NULL UNIQUE,
    logo_url        TEXT,
    founded_date    DATE,
    leader_id       UUID REFERENCES politicians(id) ON DELETE SET NULL,
    ideology        TEXT,
    website         TEXT,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','dissolved','suspended')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_parties_slug ON political_parties(slug);

CREATE TABLE coalitions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                TEXT NOT NULL,
    slug                TEXT NOT NULL UNIQUE,
    formed_date         DATE,
    dissolved_date      DATE,
    principal_party_id  UUID REFERENCES political_parties(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coalitions_slug ON coalitions(slug);

CREATE TABLE coalition_members (
    coalition_id UUID NOT NULL REFERENCES coalitions(id) ON DELETE CASCADE,
    party_id     UUID NOT NULL REFERENCES political_parties(id) ON DELETE CASCADE,
    joined_at    DATE,
    left_at      DATE,
    PRIMARY KEY (coalition_id, party_id)
);

CREATE TABLE party_memberships (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    party_id        UUID NOT NULL REFERENCES political_parties(id) ON DELETE CASCADE,
    joined_date     DATE,
    left_date       DATE,
    role            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_party_memberships_politician ON party_memberships(politician_id);
CREATE INDEX idx_party_memberships_party ON party_memberships(party_id);

-- ============================================================
-- Electoral geography
-- ============================================================
CREATE TABLE counties (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE constituencies (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    county_id           UUID NOT NULL REFERENCES counties(id) ON DELETE CASCADE,
    code                TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    slug                TEXT NOT NULL UNIQUE,
    registered_voters   INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_constituencies_county ON constituencies(county_id);

CREATE TABLE wards (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    constituency_id     UUID NOT NULL REFERENCES constituencies(id) ON DELETE CASCADE,
    code                TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    slug                TEXT NOT NULL UNIQUE,
    registered_voters   INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wards_constituency ON wards(constituency_id);

CREATE TABLE polling_stations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ward_id             UUID NOT NULL REFERENCES wards(id) ON DELETE CASCADE,
    code                TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    latitude            DOUBLE PRECISION,
    longitude           DOUBLE PRECISION,
    registered_voters   INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_polling_stations_ward ON polling_stations(ward_id);

-- ============================================================
-- Elections and candidacies
-- ============================================================
CREATE TABLE elections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL,
    election_date   DATE,
    type            TEXT NOT NULL CHECK (type IN ('general','by_election','referendum')),
    status          TEXT NOT NULL DEFAULT 'upcoming' CHECK (status IN ('upcoming','ongoing','completed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE elective_positions (
    id      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title   TEXT NOT NULL CHECK (title IN ('president','deputy_president','governor','senator','mp','woman_rep','mca')),
    level   TEXT NOT NULL CHECK (level IN ('national','county','constituency','ward')),
    county_id       UUID REFERENCES counties(id) ON DELETE SET NULL,
    constituency_id UUID REFERENCES constituencies(id) ON DELETE SET NULL,
    ward_id         UUID REFERENCES wards(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE candidacies (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id       UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    election_id         UUID NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    position_id         UUID NOT NULL REFERENCES elective_positions(id) ON DELETE CASCADE,
    party_id            UUID REFERENCES political_parties(id) ON DELETE SET NULL,
    status              TEXT NOT NULL DEFAULT 'declared' CHECK (status IN ('declared','cleared','disqualified','withdrew','elected','lost')),
    declaration_date    DATE,
    clearance_date      DATE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidacies_politician ON candidacies(politician_id);
CREATE INDEX idx_candidacies_election ON candidacies(election_id);
CREATE INDEX idx_candidacies_position ON candidacies(position_id);

CREATE TABLE election_results (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidacy_id        UUID NOT NULL REFERENCES candidacies(id) ON DELETE CASCADE,
    polling_station_id  UUID REFERENCES polling_stations(id) ON DELETE SET NULL,
    votes               INT NOT NULL DEFAULT 0,
    is_final            BOOLEAN NOT NULL DEFAULT false,
    level               TEXT NOT NULL CHECK (level IN ('station','ward','constituency','county','national')),
    reported_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_results_candidacy ON election_results(candidacy_id);
CREATE INDEX idx_results_station ON election_results(polling_station_id);

-- ============================================================
-- Manifesto, promises, policy
-- ============================================================
CREATE TABLE manifestos (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    election_id     UUID REFERENCES elections(id) ON DELETE SET NULL,
    title           TEXT NOT NULL,
    summary         TEXT,
    document_url    TEXT,
    published_date  DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_manifestos_politician ON manifestos(politician_id);

CREATE TABLE policy_positions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    manifesto_id    UUID NOT NULL REFERENCES manifestos(id) ON DELETE CASCADE,
    sector          TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    source_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policy_positions_manifesto ON policy_positions(manifesto_id);
CREATE INDEX idx_policy_positions_sector ON policy_positions(sector);

CREATE TABLE promises (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    description     TEXT NOT NULL,
    sector          TEXT,
    made_date       DATE,
    deadline        DATE,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','fulfilled','broken','partially_fulfilled')),
    evidence        TEXT,
    source_url      TEXT,
    source_id       UUID REFERENCES sources(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_promises_politician ON promises(politician_id);
CREATE INDEX idx_promises_status ON promises(status);

-- ============================================================
-- Integrity and accountability
-- ============================================================
CREATE TABLE court_cases (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    case_number     TEXT,
    court_name      TEXT,
    case_type       TEXT NOT NULL CHECK (case_type IN ('criminal','civil','election_petition','corruption','economic_crime','other')),
    title           TEXT NOT NULL,
    description     TEXT,
    filing_date     DATE,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','ongoing','convicted','acquitted','dismissed','appealed')),
    outcome         TEXT,
    source_url      TEXT,
    source_id       UUID REFERENCES sources(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_court_cases_politician ON court_cases(politician_id);
CREATE INDEX idx_court_cases_status ON court_cases(status);

CREATE TABLE asset_declarations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id       UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    declaration_year    INT NOT NULL,
    total_assets        NUMERIC(15,2),
    total_liabilities   NUMERIC(15,2),
    details             JSONB DEFAULT '{}'::jsonb,
    source_url          TEXT,
    source_id           UUID REFERENCES sources(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_asset_declarations_politician ON asset_declarations(politician_id);

CREATE TABLE integrity_flags (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    flag_type       TEXT NOT NULL CHECK (flag_type IN ('chapter6','eacc_investigation','lifestyle_audit','tax_compliance','other')),
    description     TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','resolved','dismissed')),
    source_url      TEXT,
    source_id       UUID REFERENCES sources(id) ON DELETE SET NULL,
    flagged_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integrity_flags_politician ON integrity_flags(politician_id);

-- ============================================================
-- Parliamentary record
-- ============================================================
CREATE TABLE voting_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    bill_name       TEXT NOT NULL,
    bill_number     TEXT,
    vote            TEXT NOT NULL CHECK (vote IN ('aye','nay','abstain','absent')),
    vote_date       DATE NOT NULL,
    session         TEXT,
    source_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_voting_records_politician ON voting_records(politician_id);
CREATE INDEX idx_voting_records_date ON voting_records(vote_date);

CREATE TABLE parliamentary_attendance (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    session_date    DATE NOT NULL,
    present         BOOLEAN NOT NULL DEFAULT false,
    source_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attendance_politician ON parliamentary_attendance(politician_id);

CREATE TABLE achievements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,
    category        TEXT,
    date            DATE,
    source_url      TEXT,
    source_id       UUID REFERENCES sources(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_achievements_politician ON achievements(politician_id);

CREATE TABLE controversies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,
    category        TEXT,
    date            DATE,
    severity        TEXT NOT NULL DEFAULT 'medium' CHECK (severity IN ('low','medium','high','critical')),
    source_url      TEXT,
    source_id       UUID REFERENCES sources(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_controversies_politician ON controversies(politician_id);

-- ============================================================
-- Affiliations graph
-- ============================================================
CREATE TABLE affiliations (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id           UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    related_politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    relationship_type       TEXT NOT NULL CHECK (relationship_type IN ('political_ally','family','business_partner','mentor','rival','other')),
    description             TEXT,
    source_url              TEXT,
    source_id               UUID REFERENCES sources(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT no_self_affiliation CHECK (politician_id != related_politician_id)
);

CREATE INDEX idx_affiliations_politician ON affiliations(politician_id);
CREATE INDEX idx_affiliations_related ON affiliations(related_politician_id);

-- ============================================================
-- News and social media
-- ============================================================
CREATE TABLE news_sources (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    url         TEXT NOT NULL,
    feed_url    TEXT,
    type        TEXT NOT NULL DEFAULT 'rss' CHECK (type IN ('rss','scraper','api')),
    outlet      TEXT CHECK (outlet IN ('nation','standard','star','citizen','ktn','other')),
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE news_articles (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id           UUID REFERENCES news_sources(id) ON DELETE SET NULL,
    title               TEXT NOT NULL,
    content             TEXT,
    summary             TEXT,
    url                 TEXT NOT NULL UNIQUE,
    author              TEXT,
    image_url           TEXT,
    published_at        TIMESTAMPTZ,
    scraped_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    category            TEXT,
    is_election_related BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_articles_published ON news_articles(published_at DESC);
CREATE INDEX idx_articles_source ON news_articles(source_id);
CREATE INDEX idx_articles_election ON news_articles(is_election_related) WHERE is_election_related = true;
CREATE INDEX idx_articles_search ON news_articles USING gin (title gin_trgm_ops);

CREATE TABLE article_politician_mentions (
    article_id      UUID NOT NULL REFERENCES news_articles(id) ON DELETE CASCADE,
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    sentiment_score NUMERIC(3,2),
    PRIMARY KEY (article_id, politician_id)
);

CREATE TABLE social_posts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    platform            TEXT NOT NULL CHECK (platform IN ('twitter','facebook')),
    platform_post_id    TEXT,
    author_handle       TEXT,
    author_name         TEXT,
    content             TEXT NOT NULL,
    url                 TEXT,
    posted_at           TIMESTAMPTZ,
    scraped_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    engagement          JSONB DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_social_posts_platform ON social_posts(platform, posted_at DESC);

CREATE TABLE social_post_mentions (
    post_id         UUID NOT NULL REFERENCES social_posts(id) ON DELETE CASCADE,
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    sentiment_score NUMERIC(3,2),
    PRIMARY KEY (post_id, politician_id)
);

-- ============================================================
-- Sentiment and analytics
-- ============================================================
CREATE TABLE sentiment_snapshots (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    platform        TEXT NOT NULL DEFAULT 'overall' CHECK (platform IN ('overall','twitter','facebook','news')),
    score           NUMERIC(4,2),
    sample_size     INT NOT NULL DEFAULT 0,
    positive_pct    NUMERIC(5,2),
    negative_pct    NUMERIC(5,2),
    neutral_pct     NUMERIC(5,2),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sentiment_politician_date ON sentiment_snapshots(politician_id, date DESC);

-- ============================================================
-- Events and timeline
-- ============================================================
CREATE TABLE events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       TEXT NOT NULL,
    description TEXT,
    event_type  TEXT NOT NULL CHECK (event_type IN ('rally','debate','iebc_deadline','court_hearing','parliamentary','inauguration','other')),
    location    TEXT,
    latitude    DOUBLE PRECISION,
    longitude   DOUBLE PRECISION,
    start_time  TIMESTAMPTZ,
    end_time    TIMESTAMPTZ,
    source_url  TEXT,
    source_id   UUID REFERENCES sources(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_start ON events(start_time);

CREATE TABLE event_participants (
    event_id        UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    politician_id   UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    role            TEXT,
    PRIMARY KEY (event_id, politician_id)
);

CREATE TABLE election_timeline (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id     UUID NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,
    milestone_type  TEXT NOT NULL CHECK (milestone_type IN (
        'registration_open','registration_close','nomination_start','nomination_end',
        'campaign_start','silence_period','election_day','results_announcement',
        'inauguration','other'
    )),
    date            DATE NOT NULL,
    status          TEXT NOT NULL DEFAULT 'upcoming' CHECK (status IN ('upcoming','current','passed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_timeline_election ON election_timeline(election_id);
CREATE INDEX idx_timeline_date ON election_timeline(date);

-- trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_politicians_updated BEFORE UPDATE ON politicians FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_parties_updated BEFORE UPDATE ON political_parties FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_coalitions_updated BEFORE UPDATE ON coalitions FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_elections_updated BEFORE UPDATE ON elections FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_candidacies_updated BEFORE UPDATE ON candidacies FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_manifestos_updated BEFORE UPDATE ON manifestos FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_promises_updated BEFORE UPDATE ON promises FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_court_cases_updated BEFORE UPDATE ON court_cases FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_integrity_flags_updated BEFORE UPDATE ON integrity_flags FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_sources_updated BEFORE UPDATE ON sources FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_news_sources_updated BEFORE UPDATE ON news_sources FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_events_updated BEFORE UPDATE ON events FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_timeline_updated BEFORE UPDATE ON election_timeline FOR EACH ROW EXECUTE FUNCTION update_updated_at();
