CREATE TABLE IF NOT EXISTS email_verification_tokens(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ADD INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);