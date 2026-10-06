CREATE TYPE link_status AS ENUM ('active', 'disabled', 'expired');

ALTER TABLE links ADD COLUMN status link_status NOT NULL DEFAULT 'active';

ALTER TABLE links ADD COLUMN max_clicks BIGINT CHECK (max_clicks IS NULL OR max_clicks > 0);

CREATE INDEX IF NOT EXISTS idx_links_status ON links (status);