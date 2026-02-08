-- Migration: 012_translation_voting_targets
-- Description: Translation voting as a separate leaderboard for both existing novels and announced proposals
-- Created: 2026-01-24

-- ============================================
-- Link proposal -> released novel (when daily vote releases the book)
-- ============================================

ALTER TABLE novel_proposals
    ADD COLUMN IF NOT EXISTS novel_id UUID REFERENCES novels(id);

CREATE INDEX IF NOT EXISTS idx_novel_proposals_novel_id ON novel_proposals(novel_id);

-- ============================================
-- Translation voting targets
-- ============================================

DO $$ BEGIN
    CREATE TYPE translation_vote_target_status AS ENUM ('voting', 'waiting_release', 'translating', 'completed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS translation_vote_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID REFERENCES novels(id) ON DELETE CASCADE,
    proposal_id UUID REFERENCES novel_proposals(id) ON DELETE CASCADE,
    status translation_vote_target_status NOT NULL DEFAULT 'voting',
    translation_tickets_invested INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_translation_vote_target_one_ref CHECK (
        (novel_id IS NOT NULL AND proposal_id IS NULL) OR
        (novel_id IS NULL AND proposal_id IS NOT NULL)
    )
);

-- Idempotent targets: one per novel OR one per proposal
CREATE UNIQUE INDEX IF NOT EXISTS ux_translation_vote_targets_novel_id
    ON translation_vote_targets(novel_id) WHERE novel_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_translation_vote_targets_proposal_id
    ON translation_vote_targets(proposal_id) WHERE proposal_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_translation_vote_targets_status_score
    ON translation_vote_targets(status, translation_tickets_invested DESC, created_at ASC);

-- ============================================
-- Translation votes
-- ============================================

CREATE TABLE IF NOT EXISTS translation_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES translation_vote_targets(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_translation_votes_target ON translation_votes(target_id);
CREATE INDEX IF NOT EXISTS idx_translation_votes_user ON translation_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_translation_votes_created ON translation_votes(created_at);

-- ============================================
-- Aggregate trigger for translation_vote_targets.translation_tickets_invested
-- ============================================

CREATE OR REPLACE FUNCTION update_translation_vote_target_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE translation_vote_targets
        SET translation_tickets_invested = translation_tickets_invested + NEW.amount,
            updated_at = NOW()
        WHERE id = NEW.target_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE translation_vote_targets
        SET translation_tickets_invested = translation_tickets_invested - OLD.amount,
            updated_at = NOW()
        WHERE id = OLD.target_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_translation_vote_target_stats ON translation_votes;
CREATE TRIGGER trg_update_translation_vote_target_stats
AFTER INSERT OR DELETE ON translation_votes
FOR EACH ROW
EXECUTE FUNCTION update_translation_vote_target_stats();

