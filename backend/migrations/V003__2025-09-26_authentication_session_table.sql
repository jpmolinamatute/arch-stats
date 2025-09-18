-- table to store authentication session for archer
CREATE TABLE IF NOT EXISTS auth (
    auth_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archer_id UUID NOT NULL REFERENCES archer (archer_id) ON DELETE CASCADE,
    session_token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ua TEXT,
    ip_inet INET
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at
ON auth (expires_at);
