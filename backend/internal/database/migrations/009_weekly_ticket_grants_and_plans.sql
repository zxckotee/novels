-- Migration: 009_weekly_ticket_grants_and_plans
-- Description: Weekly ticket grant logs + align subscription plans with UI (vip code) and weekly ticket amounts
-- Created: 2026-01-21

-- ============================================
-- WEEKLY TICKET GRANTS LOG
-- ============================================

CREATE TABLE IF NOT EXISTS weekly_ticket_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grant_date DATE NOT NULL,
    users_processed INTEGER NOT NULL DEFAULT 0,
    novel_requests_granted INTEGER NOT NULL DEFAULT 0,
    translation_tickets_granted INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'running', -- running, completed, failed
    error_message TEXT,
    UNIQUE (grant_date)
);

CREATE INDEX IF NOT EXISTS idx_weekly_ticket_grants_date ON weekly_ticket_grants(grant_date);

COMMENT ON TABLE weekly_ticket_grants IS 'Logs of weekly ticket grant executions (Wed 00:00 UTC)';

-- ============================================
-- ALIGN SUBSCRIPTION PLANS WITH UI + WEEKLY AMOUNTS
-- ============================================
-- UI expects plan code "vip". Previously DB used "ultimate".
-- Also we interpret monthlyNovelRequests/monthlyTranslationTickets as WEEKLY amounts now.

-- Rename ultimate -> vip (keep id)
UPDATE subscription_plans
SET code = 'vip',
    title = 'VIP',
    updated_at = NOW()
WHERE code = 'ultimate';

-- Premium: 2 Novel Request / week, 5 Translation Ticket / week, daily vote multiplier stays 2
UPDATE subscription_plans
SET features = jsonb_set(
        jsonb_set(
            features,
            '{monthlyNovelRequests}',
            '2'::jsonb,
            true
        ),
        '{monthlyTranslationTickets}',
        '5'::jsonb,
        true
    ),
    updated_at = NOW()
WHERE code = 'premium';

-- VIP: 5 Novel Request / week, 15 Translation Ticket / week, daily vote multiplier = 5
UPDATE subscription_plans
SET features = jsonb_set(
        jsonb_set(
            jsonb_set(
                features,
                '{dailyVoteMultiplier}',
                '5'::jsonb,
                true
            ),
            '{monthlyNovelRequests}',
            '5'::jsonb,
            true
        ),
        '{monthlyTranslationTickets}',
        '15'::jsonb,
        true
    ),
    updated_at = NOW()
WHERE code = 'vip';

