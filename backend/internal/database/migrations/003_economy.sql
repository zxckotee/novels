-- Migration: 003_economy
-- Description: Economy system (tickets, voting, subscriptions)
-- Created: 2026-01-09

-- ============================================
-- TICKET TYPES
-- ============================================

DO $$ BEGIN
    CREATE TYPE ticket_type AS ENUM ('daily_vote', 'novel_request', 'translation_ticket');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ============================================
-- TICKET BALANCES
-- ============================================

CREATE TABLE IF NOT EXISTS ticket_balances (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type ticket_type NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, type)
);

CREATE INDEX IF NOT EXISTS idx_ticket_balances_user ON ticket_balances(user_id);

-- ============================================
-- TICKET TRANSACTIONS
-- ============================================

CREATE TABLE IF NOT EXISTS ticket_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type ticket_type NOT NULL,
    delta INTEGER NOT NULL, -- positive = credit, negative = debit
    reason VARCHAR(100) NOT NULL,
    ref_type VARCHAR(50), -- proposal, vote, subscription, admin, etc.
    ref_id UUID,
    idempotency_key VARCHAR(255) UNIQUE, -- for preventing duplicate grants
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ticket_transactions_user ON ticket_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_ticket_transactions_user_type ON ticket_transactions(user_id, type);
CREATE INDEX IF NOT EXISTS idx_ticket_transactions_created ON ticket_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_ticket_transactions_idempotency ON ticket_transactions(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- ============================================
-- PROPOSAL STATUS
-- ============================================

DO $$ BEGIN
    CREATE TYPE proposal_status AS ENUM ('draft', 'moderation', 'voting', 'accepted', 'rejected', 'translating');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ============================================
-- NOVEL PROPOSALS
-- ============================================

CREATE TABLE IF NOT EXISTS novel_proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_link TEXT NOT NULL,
    status proposal_status NOT NULL DEFAULT 'draft',
    
    -- Novel metadata (draft)
    title VARCHAR(500) NOT NULL,
    alt_titles TEXT[] DEFAULT ARRAY[]::TEXT[],
    author VARCHAR(255),
    description TEXT,
    cover_url TEXT,
    genres TEXT[] DEFAULT ARRAY[]::TEXT[],
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    
    -- Voting stats (denormalized for performance)
    vote_score INTEGER NOT NULL DEFAULT 0,
    votes_count INTEGER NOT NULL DEFAULT 0,
    
    -- Moderation
    moderator_id UUID REFERENCES users(id),
    reject_reason TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_novel_proposals_user ON novel_proposals(user_id);
CREATE INDEX IF NOT EXISTS idx_novel_proposals_status ON novel_proposals(status);
CREATE INDEX IF NOT EXISTS idx_novel_proposals_voting ON novel_proposals(status, vote_score DESC) WHERE status = 'voting';
CREATE INDEX IF NOT EXISTS idx_novel_proposals_created ON novel_proposals(created_at);

-- ============================================
-- VOTING POLLS
-- ============================================

CREATE TABLE IF NOT EXISTS voting_polls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, closed
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_voting_polls_status ON voting_polls(status);
CREATE INDEX IF NOT EXISTS idx_voting_polls_active ON voting_polls(ends_at) WHERE status = 'active';

-- ============================================
-- VOTES
-- ============================================

CREATE TABLE IF NOT EXISTS votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID REFERENCES voting_polls(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    proposal_id UUID NOT NULL REFERENCES novel_proposals(id) ON DELETE CASCADE,
    ticket_type ticket_type NOT NULL,
    amount INTEGER NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_votes_poll ON votes(poll_id);
CREATE INDEX IF NOT EXISTS idx_votes_user ON votes(user_id);
CREATE INDEX IF NOT EXISTS idx_votes_proposal ON votes(proposal_id);
CREATE INDEX IF NOT EXISTS idx_votes_created ON votes(created_at);

-- ============================================
-- SUBSCRIPTION PLANS
-- ============================================

CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL, -- in kopecks/cents
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    period VARCHAR(20) NOT NULL DEFAULT 'monthly', -- monthly, yearly
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Features as JSONB for flexibility
    features JSONB NOT NULL DEFAULT '{}'::JSONB,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default subscription plans
INSERT INTO subscription_plans (code, title, description, price, currency, features) VALUES
('basic', 'Basic', 'Базовая подписка с отключением рекламы', 19900, 'RUB', '{
    "dailyVoteMultiplier": 1,
    "monthlyNovelRequests": 1,
    "monthlyTranslationTickets": 3,
    "adFree": true,
    "canEditDescriptions": false,
    "canRequestRetranslation": false,
    "prioritySupport": false,
    "exclusiveBadge": false
}'::JSONB),
('premium', 'Premium', 'Премиум подписка с расширенными возможностями', 49900, 'RUB', '{
    "dailyVoteMultiplier": 2,
    "monthlyNovelRequests": 3,
    "monthlyTranslationTickets": 10,
    "adFree": true,
    "canEditDescriptions": true,
    "canRequestRetranslation": true,
    "prioritySupport": false,
    "exclusiveBadge": true
}'::JSONB),
('ultimate', 'Ultimate', 'Максимальная подписка со всеми привилегиями', 99900, 'RUB', '{
    "dailyVoteMultiplier": 3,
    "monthlyNovelRequests": 5,
    "monthlyTranslationTickets": 25,
    "adFree": true,
    "canEditDescriptions": true,
    "canRequestRetranslation": true,
    "prioritySupport": true,
    "exclusiveBadge": true
}'::JSONB)
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- SUBSCRIPTION STATUS
-- ============================================

DO $$ BEGIN
    CREATE TYPE subscription_status AS ENUM ('active', 'canceled', 'past_due', 'expired');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ============================================
-- SUBSCRIPTIONS
-- ============================================

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    status subscription_status NOT NULL DEFAULT 'active',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ NOT NULL,
    
    -- External payment reference
    external_id VARCHAR(255),
    
    -- Auto-renewal
    auto_renew BOOLEAN NOT NULL DEFAULT true,
    
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(user_id, ends_at) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_subscriptions_expires ON subscriptions(ends_at) WHERE status = 'active';

-- ============================================
-- SUBSCRIPTION GRANTS (for tracking monthly grants)
-- ============================================

CREATE TABLE IF NOT EXISTS subscription_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type ticket_type NOT NULL,
    amount INTEGER NOT NULL,
    for_month VARCHAR(7) NOT NULL, -- YYYY-MM format
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (subscription_id, type, for_month)
);

CREATE INDEX IF NOT EXISTS idx_subscription_grants_user ON subscription_grants(user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_grants_month ON subscription_grants(for_month);

-- ============================================
-- DAILY VOTE GRANTS LOG (for tracking daily grant execution)
-- ============================================

CREATE TABLE IF NOT EXISTS daily_vote_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grant_date DATE NOT NULL,
    users_processed INTEGER NOT NULL DEFAULT 0,
    total_votes_granted INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'running', -- running, completed, failed
    error_message TEXT,
    UNIQUE (grant_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_vote_grants_date ON daily_vote_grants(grant_date);

-- ============================================
-- LEADERBOARD CACHE (for performance)
-- ============================================

CREATE TABLE IF NOT EXISTS leaderboard_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period VARCHAR(10) NOT NULL, -- day, week, month
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tickets_spent INTEGER NOT NULL DEFAULT 0,
    rank INTEGER NOT NULL DEFAULT 0,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (period, user_id)
);

CREATE INDEX IF NOT EXISTS idx_leaderboard_cache_period ON leaderboard_cache(period, rank);
CREATE INDEX IF NOT EXISTS idx_leaderboard_cache_user ON leaderboard_cache(user_id);

-- ============================================
-- TRIGGERS
-- ============================================

-- Update vote_score on novel_proposals when votes change
CREATE OR REPLACE FUNCTION update_proposal_vote_score()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE novel_proposals
        SET vote_score = vote_score + NEW.amount,
            votes_count = votes_count + 1,
            updated_at = NOW()
        WHERE id = NEW.proposal_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE novel_proposals
        SET vote_score = vote_score - OLD.amount,
            votes_count = votes_count - 1,
            updated_at = NOW()
        WHERE id = OLD.proposal_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_proposal_vote_score ON votes;
CREATE TRIGGER trg_update_proposal_vote_score
AFTER INSERT OR DELETE ON votes
FOR EACH ROW
EXECUTE FUNCTION update_proposal_vote_score();

-- Update ticket_balances on transaction
CREATE OR REPLACE FUNCTION update_ticket_balance()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO ticket_balances (user_id, type, balance, updated_at)
    VALUES (NEW.user_id, NEW.type, GREATEST(0, NEW.delta), NOW())
    ON CONFLICT (user_id, type)
    DO UPDATE SET
        balance = GREATEST(0, ticket_balances.balance + NEW.delta),
        updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_ticket_balance ON ticket_transactions;
CREATE TRIGGER trg_update_ticket_balance
AFTER INSERT ON ticket_transactions
FOR EACH ROW
EXECUTE FUNCTION update_ticket_balance();

-- Update subscriptions.updated_at
CREATE OR REPLACE FUNCTION update_subscription_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_subscription_timestamp ON subscriptions;
CREATE TRIGGER trg_update_subscription_timestamp
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION update_subscription_timestamp();

-- Update novel_proposals.updated_at
DROP TRIGGER IF EXISTS trg_update_proposal_timestamp ON novel_proposals;
CREATE TRIGGER trg_update_proposal_timestamp
BEFORE UPDATE ON novel_proposals
FOR EACH ROW
EXECUTE FUNCTION update_subscription_timestamp();

-- ============================================
-- HELPER VIEWS
-- ============================================

-- Active subscriptions view
CREATE OR REPLACE VIEW active_subscriptions AS
SELECT 
    s.*,
    sp.code as plan_code,
    sp.title as plan_title,
    sp.features
FROM subscriptions s
JOIN subscription_plans sp ON s.plan_id = sp.id
WHERE s.status = 'active'
AND s.ends_at > NOW();

-- Voting leaderboard view (proposals sorted by votes)
CREATE OR REPLACE VIEW voting_leaderboard AS
SELECT 
    np.*,
    u.email as user_email,
    up.display_name as user_display_name,
    up.avatar_key as user_avatar,
    ux.level as user_level
FROM novel_proposals np
JOIN users u ON np.user_id = u.id
LEFT JOIN user_profiles up ON np.user_id = up.user_id
LEFT JOIN user_xp ux ON np.user_id = ux.user_id
WHERE np.status = 'voting'
ORDER BY np.vote_score DESC, np.created_at ASC;

-- User wallet view
CREATE OR REPLACE VIEW user_wallets AS
SELECT 
    u.id as user_id,
    COALESCE(dv.balance, 0) as daily_votes,
    COALESCE(nr.balance, 0) as novel_requests,
    COALESCE(tt.balance, 0) as translation_tickets
FROM users u
LEFT JOIN ticket_balances dv ON u.id = dv.user_id AND dv.type = 'daily_vote'
LEFT JOIN ticket_balances nr ON u.id = nr.user_id AND nr.type = 'novel_request'
LEFT JOIN ticket_balances tt ON u.id = tt.user_id AND tt.type = 'translation_ticket';

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TABLE ticket_balances IS 'Current ticket balances for each user and ticket type';
COMMENT ON TABLE ticket_transactions IS 'All ticket transactions (credits and debits)';
COMMENT ON TABLE novel_proposals IS 'User-submitted proposals for novels to translate';
COMMENT ON TABLE voting_polls IS 'Voting periods for novel selection';
COMMENT ON TABLE votes IS 'Individual votes cast by users';
COMMENT ON TABLE subscription_plans IS 'Available subscription plans with features';
COMMENT ON TABLE subscriptions IS 'User subscriptions';
COMMENT ON TABLE subscription_grants IS 'Monthly ticket grants from subscriptions';
COMMENT ON TABLE daily_vote_grants IS 'Logs of daily vote grant executions';
COMMENT ON TABLE leaderboard_cache IS 'Cached leaderboard data for performance';
