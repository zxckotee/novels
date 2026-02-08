-- Migration: 016_import_runs_checkpoint
-- Description: Add pause/resume checkpoint/progress fields for long-running imports
-- Created: 2026-02-04

ALTER TABLE import_runs
  ADD COLUMN IF NOT EXISTS progress_current INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS progress_total INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS checkpoint JSONB NULL;

CREATE INDEX IF NOT EXISTS idx_import_runs_status_updated ON import_runs(status, updated_at DESC);

