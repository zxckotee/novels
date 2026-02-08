-- Fix duplicate active subscriptions and prevent future duplicates.
-- Policy:
-- - If a user has multiple active subscriptions, keep the "best" plan (VIP > Premium > Basic) by dailyVoteMultiplier/price,
--   and set its ends_at to the MAX(ends_at) among actives (so we don't lose remaining time).
-- - Cancel the rest.

-- 1) Pick the winner per user and compute max ends_at across actives.
WITH active AS (
  SELECT
    s.id,
    s.user_id,
    s.ends_at,
    s.created_at,
    sp.price,
    COALESCE((sp.features->>'dailyVoteMultiplier')::int, 1) AS vote_multiplier
  FROM subscriptions s
  JOIN subscription_plans sp ON sp.id = s.plan_id
  WHERE s.status = 'active' AND s.ends_at > NOW()
),
winner AS (
  SELECT
    user_id,
    -- Keep the best plan: multiplier desc, then price desc, then newest created_at
    (ARRAY_AGG(id ORDER BY vote_multiplier DESC, price DESC, created_at DESC))[1] AS keep_id,
    MAX(ends_at) AS max_ends_at
  FROM active
  GROUP BY user_id
),
extend_keep AS (
  UPDATE subscriptions s
  SET ends_at = w.max_ends_at, updated_at = NOW()
  FROM winner w
  WHERE s.id = w.keep_id
  RETURNING s.id
)
UPDATE subscriptions s
SET status = 'canceled', canceled_at = NOW(), auto_renew = false, updated_at = NOW()
FROM winner w
WHERE s.user_id = w.user_id
  AND s.status = 'active'
  AND s.ends_at > NOW()
  AND s.id <> w.keep_id;

-- 2) If there are "active" rows that already ended, mark them expired (cleanup).
UPDATE subscriptions
SET status = 'expired', updated_at = NOW()
WHERE status = 'active' AND ends_at <= NOW();

-- 3) Enforce: only one active subscription row per user.
-- Note: status is updated to 'expired' by jobs when ends_at passes, so this stays consistent.
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_one_active_per_user
ON subscriptions (user_id)
WHERE status = 'active';

