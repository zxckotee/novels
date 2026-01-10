-- Migration: 002_mvp2_comments_bookmarks_xp
-- Description: Add tables for comments, bookmarks, and XP system (MVP 2)

-- ============================================
-- COMMENTS SYSTEM
-- ============================================

-- Add missing columns to comments table
ALTER TABLE comments ADD COLUMN IF NOT EXISTS is_spoiler BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS likes_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS dislikes_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS replies_count INTEGER NOT NULL DEFAULT 0;

-- Comment votes (like/dislike)
CREATE TABLE IF NOT EXISTS comment_votes (
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value INTEGER NOT NULL CHECK (value IN (-1, 1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (comment_id, user_id)
);

-- Comment reports
CREATE TABLE IF NOT EXISTS comment_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'resolved', 'dismissed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (comment_id, user_id)
);

-- ============================================
-- BOOKMARKS SYSTEM
-- ============================================

-- Bookmark lists (system and custom)
CREATE TABLE IF NOT EXISTS bookmark_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, code)
);

-- Bookmarks
CREATE TABLE IF NOT EXISTS bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    list_id UUID NOT NULL REFERENCES bookmark_lists(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, novel_id)
);

-- ============================================
-- XP AND LEVELS SYSTEM
-- ============================================

-- User XP
CREATE TABLE IF NOT EXISTS user_xp (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    xp_total BIGINT NOT NULL DEFAULT 0,
    level INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- XP events (audit log)
CREATE TABLE IF NOT EXISTS xp_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    delta BIGINT NOT NULL,
    ref_type VARCHAR(50),
    ref_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Achievements
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    icon_key VARCHAR(100),
    condition JSONB,
    xp_reward BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User achievements
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, achievement_id)
);

-- ============================================
-- INDEXES
-- ============================================

-- Comment indexes
CREATE INDEX IF NOT EXISTS idx_comment_votes_user ON comment_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_comment_reports_status ON comment_reports(status);
CREATE INDEX IF NOT EXISTS idx_comment_reports_comment ON comment_reports(comment_id);

-- Bookmark indexes
CREATE INDEX IF NOT EXISTS idx_bookmark_lists_user ON bookmark_lists(user_id);
CREATE INDEX IF NOT EXISTS idx_bookmarks_user ON bookmarks(user_id);
CREATE INDEX IF NOT EXISTS idx_bookmarks_novel ON bookmarks(novel_id);
CREATE INDEX IF NOT EXISTS idx_bookmarks_list ON bookmarks(list_id);

-- XP indexes
CREATE INDEX IF NOT EXISTS idx_xp_events_user ON xp_events(user_id);
CREATE INDEX IF NOT EXISTS idx_xp_events_type ON xp_events(type);
CREATE INDEX IF NOT EXISTS idx_xp_events_created ON xp_events(created_at);
CREATE INDEX IF NOT EXISTS idx_user_xp_level ON user_xp(level);
CREATE INDEX IF NOT EXISTS idx_user_xp_total ON user_xp(xp_total DESC);

-- User achievements index
CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);

-- ============================================
-- DEFAULT ACHIEVEMENTS
-- ============================================

INSERT INTO achievements (code, title, description, icon_key, xp_reward) VALUES
    ('first_chapter', 'Первая глава', 'Прочитать первую главу', 'book-open', 50),
    ('bookworm_10', 'Книжный червь', 'Прочитать 10 глав', 'book', 100),
    ('bookworm_100', 'Заядлый читатель', 'Прочитать 100 глав', 'books', 500),
    ('bookworm_1000', 'Мастер чтения', 'Прочитать 1000 глав', 'crown', 2000),
    ('first_comment', 'Первый комментарий', 'Оставить первый комментарий', 'message', 25),
    ('commentator_10', 'Комментатор', 'Оставить 10 комментариев', 'messages', 100),
    ('commentator_100', 'Активный участник', 'Оставить 100 комментариев', 'chat', 500),
    ('first_bookmark', 'Первая закладка', 'Добавить первую книгу в закладки', 'bookmark', 15),
    ('collector_10', 'Коллекционер', 'Добавить 10 книг в закладки', 'bookmarks', 75),
    ('collector_100', 'Библиотекарь', 'Добавить 100 книг в закладки', 'library', 300)
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- FUNCTIONS AND TRIGGERS
-- ============================================

-- Function to update comment counts
CREATE OR REPLACE FUNCTION update_comment_vote_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.value = 1 THEN
            UPDATE comments SET likes_count = likes_count + 1 WHERE id = NEW.comment_id;
        ELSE
            UPDATE comments SET dislikes_count = dislikes_count + 1 WHERE id = NEW.comment_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.value = 1 THEN
            UPDATE comments SET likes_count = GREATEST(0, likes_count - 1) WHERE id = OLD.comment_id;
        ELSE
            UPDATE comments SET dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = OLD.comment_id;
        END IF;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.value != NEW.value THEN
            IF NEW.value = 1 THEN
                UPDATE comments SET likes_count = likes_count + 1, dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = NEW.comment_id;
            ELSE
                UPDATE comments SET likes_count = GREATEST(0, likes_count - 1), dislikes_count = dislikes_count + 1 WHERE id = NEW.comment_id;
            END IF;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger for comment vote counts (drop if exists first)
DROP TRIGGER IF EXISTS trigger_comment_vote_counts ON comment_votes;
CREATE TRIGGER trigger_comment_vote_counts
    AFTER INSERT OR UPDATE OR DELETE ON comment_votes
    FOR EACH ROW EXECUTE FUNCTION update_comment_vote_counts();

-- Function to update replies count
CREATE OR REPLACE FUNCTION update_comment_replies_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.parent_id IS NOT NULL THEN
        UPDATE comments SET replies_count = replies_count + 1 WHERE id = NEW.parent_id;
    ELSIF TG_OP = 'DELETE' AND OLD.parent_id IS NOT NULL THEN
        UPDATE comments SET replies_count = GREATEST(0, replies_count - 1) WHERE id = OLD.parent_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger for replies count
DROP TRIGGER IF EXISTS trigger_comment_replies_count ON comments;
CREATE TRIGGER trigger_comment_replies_count
    AFTER INSERT OR DELETE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_comment_replies_count();
