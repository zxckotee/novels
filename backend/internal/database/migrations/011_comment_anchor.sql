-- Migration 011: Comment anchors (paragraph-level threads)
-- Adds optional `anchor` to scope comments within the same target (e.g. chapter paragraph "p:12")

ALTER TABLE comments
    ADD COLUMN IF NOT EXISTS anchor TEXT;

-- Helpful index for listing root comments by target + anchor
CREATE INDEX IF NOT EXISTS idx_comments_target_anchor ON comments(target_type, target_id, anchor);

