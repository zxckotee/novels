-- Import run cookies: store cookie_header for Cloudflare recovery

CREATE TABLE IF NOT EXISTS import_run_cookies (
    run_id UUID PRIMARY KEY REFERENCES import_runs(id) ON DELETE CASCADE,
    cookie_header TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add cloudflare_blocked flag to import_runs
ALTER TABLE import_runs ADD COLUMN IF NOT EXISTS cloudflare_blocked BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_import_runs_cloudflare_blocked ON import_runs (cloudflare_blocked) WHERE cloudflare_blocked = TRUE;
