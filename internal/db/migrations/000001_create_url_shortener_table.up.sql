CREATE TABLE IF NOT EXISTS url (
    id BIGSERIAL PRIMARY KEY,
    original_url TEXT UNIQUE NOT NULL,
    short_code VARCHAR(16) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_short_code ON url (short_code);
CREATE INDEX IF NOT EXISTS idx_url_expires_at ON url (expires_at);