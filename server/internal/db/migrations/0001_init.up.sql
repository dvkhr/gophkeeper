-- 0001_init.up.sql

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
        CREATE TYPE user_status AS ENUM ('active', 'blocked');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'data_type') THEN
        CREATE TYPE data_type AS ENUM ('loginpass', 'text', 'binary', 'card');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    login TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_data (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type data_type NOT NULL,
    encrypted_data BYTEA NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_user_data_user_id ON user_data(user_id);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);