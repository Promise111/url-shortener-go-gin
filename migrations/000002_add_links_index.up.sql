CREATE INDEX IF NOT EXISTS idx_links_expires_at ON links (expires_at)
WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_links_clicks ON links (clicks);
CREATE INDEX IF NOT EXISTS idx_links_created_at ON links (created_at DESC);
