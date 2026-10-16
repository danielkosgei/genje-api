-- Feeds carry their section (e.g. Standard "politics.php") but the items often do not,
-- so each source records the category its articles default to.
ALTER TABLE news_sources ADD COLUMN default_category TEXT;

-- The publisher is what readers see ("The Standard"), not the feed name ("Standard - Politics").
ALTER TABLE news_sources ADD COLUMN publisher TEXT;

UPDATE news_sources SET publisher = CASE outlet
    WHEN 'nation'  THEN 'Nation Africa'
    WHEN 'standard' THEN 'The Standard'
    WHEN 'star'    THEN 'The Star'
    WHEN 'citizen' THEN 'Citizen Digital'
    WHEN 'ktn'     THEN 'KTN News'
    ELSE name
END;

-- Lets the seeder upsert sources by name instead of inserting once and never updating.
CREATE UNIQUE INDEX idx_news_sources_name ON news_sources(name);
