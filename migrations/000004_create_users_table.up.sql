CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE CHECK (email <> ''),
    password_hash VARCHAR(255) NOT NULL CHECK (password_hash <> ''),
    username VARCHAR(255) NOT NULL UNIQUE CHECK (username <> '' AND CHAR_LENGTH(username) <= 255),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS id_users_username ON users (username);