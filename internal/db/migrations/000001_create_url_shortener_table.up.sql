CREATE TABLE IF NOT EXISTS url (
    id BIGSERIAL PRIMARY KEY,
    original_url TEXT UNIQUE NOT NULL,
    short_code VARCHAR(16) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_short_code ON url (short_code);
CREATE INDEX IF NOT EXISTS idx_url_expires_at ON url (expires_at);

-- Function to add a new URL and return the short code if it already exists
CREATE OR REPLACE FUNCTION add_url(
  p_original_url text,
  p_short_code   text
) RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
  v_short_code text;
BEGIN
  -- Try to insert; if the original_url already exists, do nothing but try to get the existing short_code
  INSERT INTO url (original_url, short_code)
  VALUES (p_original_url, p_short_code)
  ON CONFLICT (original_url) DO NOTHING
  RETURNING short_code INTO v_short_code;

  IF v_short_code IS NOT NULL THEN
    RETURN v_short_code; -- inserted successfully
  END IF;

  -- Row already existed; return the existing short_code
  SELECT short_code
    INTO v_short_code
    FROM url
   WHERE original_url = p_original_url;

  RETURN v_short_code;
END;
$$;