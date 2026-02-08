-- Import runs: track parsing/import orchestration attempts for proposals

CREATE TABLE IF NOT EXISTS import_runs (
    id UUID PRIMARY KEY,
    proposal_id UUID NOT NULL REFERENCES novel_proposals(id) ON DELETE CASCADE,
    novel_id UUID NULL REFERENCES novels(id) ON DELETE SET NULL,
    importer TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_import_runs_proposal_id ON import_runs (proposal_id);
CREATE INDEX IF NOT EXISTS idx_import_runs_status ON import_runs (status);
CREATE INDEX IF NOT EXISTS idx_import_runs_started_at ON import_runs (started_at DESC);

