-- Migration: 010_translation_ticket_investments
-- Description: Track translation ticket investments per proposal and exclude them from vote_score
-- Created: 2026-01-21

-- Add column to store invested translation tickets (sum of translation_ticket votes)
ALTER TABLE novel_proposals
    ADD COLUMN IF NOT EXISTS translation_tickets_invested INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_novel_proposals_translation_invested
    ON novel_proposals(translation_tickets_invested DESC);

-- Backfill proposal aggregates from votes table
UPDATE novel_proposals np
SET
    vote_score = COALESCE((
        SELECT SUM(v.amount)::int
        FROM votes v
        WHERE v.proposal_id = np.id AND v.ticket_type = 'daily_vote'
    ), 0),
    votes_count = COALESCE((
        SELECT COUNT(*)::int
        FROM votes v
        WHERE v.proposal_id = np.id AND v.ticket_type = 'daily_vote'
    ), 0),
    translation_tickets_invested = COALESCE((
        SELECT SUM(v.amount)::int
        FROM votes v
        WHERE v.proposal_id = np.id AND v.ticket_type = 'translation_ticket'
    ), 0);

-- Replace trigger so only daily_vote affects vote_score/votes_count, and translation_ticket affects translation_tickets_invested.
CREATE OR REPLACE FUNCTION update_proposal_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.ticket_type = 'daily_vote' THEN
            UPDATE novel_proposals
            SET vote_score = vote_score + NEW.amount,
                votes_count = votes_count + 1,
                updated_at = NOW()
            WHERE id = NEW.proposal_id;
        ELSIF NEW.ticket_type = 'translation_ticket' THEN
            UPDATE novel_proposals
            SET translation_tickets_invested = translation_tickets_invested + NEW.amount,
                updated_at = NOW()
            WHERE id = NEW.proposal_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.ticket_type = 'daily_vote' THEN
            UPDATE novel_proposals
            SET vote_score = vote_score - OLD.amount,
                votes_count = votes_count - 1,
                updated_at = NOW()
            WHERE id = OLD.proposal_id;
        ELSIF OLD.ticket_type = 'translation_ticket' THEN
            UPDATE novel_proposals
            SET translation_tickets_invested = translation_tickets_invested - OLD.amount,
                updated_at = NOW()
            WHERE id = OLD.proposal_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_proposal_vote_score ON votes;
DROP TRIGGER IF EXISTS trg_update_proposal_stats ON votes;
CREATE TRIGGER trg_update_proposal_stats
AFTER INSERT OR DELETE ON votes
FOR EACH ROW
EXECUTE FUNCTION update_proposal_stats();

