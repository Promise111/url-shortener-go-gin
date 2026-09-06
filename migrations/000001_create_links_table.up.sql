CREATE TABLE IF NOT EXISTS links (
    id BIGSERIAL PRIMARY KEY ,
    long_url TEXT NOT NULL CHECK (long_url <> '' AND char_length(long_url) <= 2048),
    short_code VARCHAR(64) UNIQUE NOT NULL CHECK (short_code <> ''),
    expires_at TIMESTAMPTZ,
    clicks BIGINT NOT NULL DEFAULT 0 CHECK (clicks >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);