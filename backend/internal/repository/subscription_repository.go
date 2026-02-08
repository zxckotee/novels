package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels-backend/internal/domain/models"
)

type SubscriptionRepository struct {
	db *sqlx.DB
}

func NewSubscriptionRepository(db *sqlx.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// ============================================
// SUBSCRIPTION PLANS
// ============================================

// GetAllPlans returns all active subscription plans
func (r *SubscriptionRepository) GetAllPlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	plans := []models.SubscriptionPlan{}
	
	query := `
		SELECT id, code, title, description, price, currency, period, is_active, features, created_at, updated_at
		FROM subscription_plans
		WHERE is_active = true
		ORDER BY price ASC
	`
	
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all plans: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var plan models.SubscriptionPlan
		var featuresJSON []byte
		
		err := rows.Scan(
			&plan.ID, &plan.Code, &plan.Title, &plan.Description,
			&plan.Price, &plan.Currency, &plan.Period, &plan.IsActive,
			&featuresJSON, &plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		
		// Parse features JSON
		if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
			return nil, fmt.Errorf("parse features: %w", err)
		}
		
		plans = append(plans, plan)
	}
	
	return plans, nil
}

// GetPlanByID returns a subscription plan by ID
func (r *SubscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	query := `
		SELECT id, code, title, description, price, currency, period, is_active, features, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1
	`
	
	var plan models.SubscriptionPlan
	var featuresJSON []byte
	
	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&plan.ID, &plan.Code, &plan.Title, &plan.Description,
		&plan.Price, &plan.Currency, &plan.Period, &plan.IsActive,
		&featuresJSON, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get plan by id: %w", err)
	}
	
	if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
		return nil, fmt.Errorf("parse features: %w", err)
	}
	
	return &plan, nil
}

// GetPlanByCode returns a subscription plan by code
func (r *SubscriptionRepository) GetPlanByCode(ctx context.Context, code models.SubscriptionPlanCode) (*models.SubscriptionPlan, error) {
	query := `
		SELECT id, code, title, description, price, currency, period, is_active, features, created_at, updated_at
		FROM subscription_plans
		WHERE code = $1
	`
	
	var plan models.SubscriptionPlan
	var featuresJSON []byte
	
	err := r.db.QueryRowxContext(ctx, query, code).Scan(
		&plan.ID, &plan.Code, &plan.Title, &plan.Description,
		&plan.Price, &plan.Currency, &plan.Period, &plan.IsActive,
		&featuresJSON, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get plan by code: %w", err)
	}
	
	if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
		return nil, fmt.Errorf("parse features: %w", err)
	}
	
	return &plan, nil
}

// ============================================
// SUBSCRIPTIONS
// ============================================

// CreateSubscription creates a new subscription
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, plan_id, status, starts_at, ends_at,
			external_id, auto_renew, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now
	
	_, err := r.db.ExecContext(ctx, query,
		sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.StartsAt, sub.EndsAt,
		sub.ExternalID, sub.AutoRenew, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	
	return nil
}

// GetActiveSubscription returns the active subscription for a user
func (r *SubscriptionRepository) GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.plan_id, s.status, s.starts_at, s.ends_at,
			s.external_id, s.auto_renew, s.canceled_at, s.created_at, s.updated_at,
			sp.code, sp.title, sp.description, sp.price, sp.currency, sp.period, sp.features
		FROM subscriptions s
		JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.user_id = $1 AND s.status = 'active' AND s.ends_at > NOW()
		-- If multiple actives exist due to a bug/race, prefer the highest tier plan.
		ORDER BY COALESCE((sp.features->>'dailyVoteMultiplier')::int, 1) DESC, sp.price DESC, s.ends_at DESC, s.created_at DESC
		LIMIT 1
	`
	
	var sub models.Subscription
	var plan models.SubscriptionPlan
	var featuresJSON []byte
	
	err := r.db.QueryRowxContext(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.PlanID, &sub.Status, &sub.StartsAt, &sub.EndsAt,
		&sub.ExternalID, &sub.AutoRenew, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
		&plan.Code, &plan.Title, &plan.Description, &plan.Price, &plan.Currency, &plan.Period, &featuresJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active subscription: %w", err)
	}
	
	if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
		return nil, fmt.Errorf("parse features: %w", err)
	}
	
	plan.ID = sub.PlanID
	sub.Plan = &plan
	
	return &sub, nil
}

// ConsolidateActiveSubscriptions ensures the user has at most one active subscription.
// It keeps the current best active subscription (by plan tier) and extends its ends_at
// to the maximum ends_at among actives, then cancels the rest.
func (r *SubscriptionRepository) ConsolidateActiveSubscriptions(ctx context.Context, userID uuid.UUID) error {
	// Find the best active subscription id
	var keepID uuid.UUID
	err := r.db.GetContext(ctx, &keepID, `
		SELECT s.id
		FROM subscriptions s
		JOIN subscription_plans sp ON sp.id = s.plan_id
		WHERE s.user_id = $1 AND s.status = 'active' AND s.ends_at > NOW()
		ORDER BY COALESCE((sp.features->>'dailyVoteMultiplier')::int, 1) DESC, sp.price DESC, s.ends_at DESC, s.created_at DESC
		LIMIT 1
	`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("select best active subscription: %w", err)
	}

	// Extend keep ends_at to max ends_at among actives (preserve remaining time)
	_, err = r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET ends_at = (
			SELECT MAX(ends_at) FROM subscriptions
			WHERE user_id = $1 AND status = 'active' AND ends_at > NOW()
		),
		updated_at = NOW()
		WHERE id = $2
	`, userID, keepID)
	if err != nil {
		return fmt.Errorf("extend keep subscription: %w", err)
	}

	// Cancel others
	_, _ = r.CancelOtherActiveSubscriptions(ctx, userID, keepID)
	return nil
}

// GetMaxActiveEndsAtTx returns MAX(ends_at) for current active subscriptions inside a transaction.
func (r *SubscriptionRepository) GetMaxActiveEndsAtTx(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (*time.Time, error) {
	var t *time.Time
	if err := tx.GetContext(ctx, &t, `
		SELECT MAX(ends_at)
		FROM subscriptions
		WHERE user_id = $1 AND status = 'active' AND ends_at > NOW()
	`, userID); err != nil {
		return nil, fmt.Errorf("get max active ends_at: %w", err)
	}
	return t, nil
}

// GetSubscriptionByID returns a subscription by ID
func (r *SubscriptionRepository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.plan_id, s.status, s.starts_at, s.ends_at,
			s.external_id, s.auto_renew, s.canceled_at, s.created_at, s.updated_at
		FROM subscriptions s
		WHERE s.id = $1
	`
	
	var sub models.Subscription
	err := r.db.GetContext(ctx, &sub, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription by id: %w", err)
	}
	
	return &sub, nil
}

// GetUserSubscriptions returns all subscriptions for a user
func (r *SubscriptionRepository) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]models.Subscription, error) {
	subscriptions := []models.Subscription{}
	
	query := `
		SELECT 
			s.id, s.user_id, s.plan_id, s.status, s.starts_at, s.ends_at,
			s.external_id, s.auto_renew, s.canceled_at, s.created_at, s.updated_at,
			sp.code, sp.title, sp.price, sp.currency
		FROM subscriptions s
		JOIN subscription_plans sp ON s.plan_id = sp.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`
	
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user subscriptions: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var sub models.Subscription
		var plan models.SubscriptionPlan
		
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.PlanID, &sub.Status, &sub.StartsAt, &sub.EndsAt,
			&sub.ExternalID, &sub.AutoRenew, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
			&plan.Code, &plan.Title, &plan.Price, &plan.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		
		plan.ID = sub.PlanID
		sub.Plan = &plan
		subscriptions = append(subscriptions, sub)
	}
	
	return subscriptions, nil
}

// UpdateSubscriptionStatus updates the status of a subscription
func (r *SubscriptionRepository) UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status models.SubscriptionStatus) error {
	query := `UPDATE subscriptions SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// CancelSubscription cancels a subscription
func (r *SubscriptionRepository) CancelSubscription(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE subscriptions 
		SET status = 'canceled', canceled_at = NOW(), auto_renew = false, updated_at = NOW() 
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ExtendSubscription extends a subscription
func (r *SubscriptionRepository) ExtendSubscription(ctx context.Context, id uuid.UUID, newEndsAt time.Time) error {
	query := `UPDATE subscriptions SET ends_at = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, newEndsAt)
	return err
}

// ============================================
// SUBSCRIPTION GRANTS
// ============================================

// CreateGrant records a subscription grant
func (r *SubscriptionRepository) CreateGrant(ctx context.Context, grant *models.SubscriptionGrant) error {
	query := `
		INSERT INTO subscription_grants (id, subscription_id, user_id, type, amount, for_month, granted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (subscription_id, type, for_month) DO NOTHING
	`
	
	if grant.ID == uuid.Nil {
		grant.ID = uuid.New()
	}
	if grant.GrantedAt.IsZero() {
		grant.GrantedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		grant.ID, grant.SubscriptionID, grant.UserID, grant.Type,
		grant.Amount, grant.ForMonth, grant.GrantedAt,
	)
	if err != nil {
		return fmt.Errorf("create grant: %w", err)
	}
	
	return nil
}

// HasGrantForMonth checks if a grant was already made for a specific month
func (r *SubscriptionRepository) HasGrantForMonth(ctx context.Context, subscriptionID uuid.UUID, ticketType models.TicketType, forMonth string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM subscription_grants 
			WHERE subscription_id = $1 AND type = $2 AND for_month = $3
		)
	`
	
	err := r.db.GetContext(ctx, &exists, query, subscriptionID, ticketType, forMonth)
	if err != nil {
		return false, fmt.Errorf("check grant exists: %w", err)
	}
	
	return exists, nil
}

// ============================================
// USER SUBSCRIPTION INFO
// ============================================

// GetUserSubscriptionInfo returns complete subscription info for a user
func (r *SubscriptionRepository) GetUserSubscriptionInfo(ctx context.Context, userID uuid.UUID) (*models.UserSubscriptionInfo, error) {
	info := &models.UserSubscriptionInfo{
		HasActiveSubscription: false,
	}
	
	sub, err := r.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	if sub != nil {
		// Safety net: if duplicates exist due to a previous race condition, consolidate.
		_ = r.ConsolidateActiveSubscriptions(ctx, userID)
		info.HasActiveSubscription = true
		info.Subscription = sub
		info.Plan = sub.Plan
		if sub.Plan != nil {
			info.Features = &sub.Plan.Features
		}
		info.DaysRemaining = sub.DaysRemaining()
	}
	
	return info, nil
}

// GetDailyVoteMultiplier returns the daily vote multiplier for a user
func (r *SubscriptionRepository) GetDailyVoteMultiplier(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(
			(SELECT (sp.features->>'dailyVoteMultiplier')::int
			 FROM subscriptions s
			 JOIN subscription_plans sp ON s.plan_id = sp.id
			 WHERE s.user_id = $1 AND s.status = 'active' AND s.ends_at > NOW()
			 LIMIT 1),
			1
		)
	`
	
	var multiplier int
	err := r.db.GetContext(ctx, &multiplier, query, userID)
	if err != nil {
		return 1, nil // Default to 1 on error
	}
	
	return multiplier, nil
}

// ============================================
// EXPIRING SUBSCRIPTIONS
// ============================================

// GetExpiringSubscriptions returns subscriptions expiring within the given duration
func (r *SubscriptionRepository) GetExpiringSubscriptions(ctx context.Context, within time.Duration) ([]models.Subscription, error) {
	subscriptions := []models.Subscription{}
	
	query := `
		SELECT id, user_id, plan_id, status, starts_at, ends_at, auto_renew, created_at
		FROM subscriptions
		WHERE status = 'active' AND ends_at <= $1 AND ends_at > NOW()
		ORDER BY ends_at ASC
	`
	
	expiryTime := time.Now().Add(within)
	err := r.db.SelectContext(ctx, &subscriptions, query, expiryTime)
	if err != nil {
		return nil, fmt.Errorf("get expiring subscriptions: %w", err)
	}
	
	return subscriptions, nil
}

// GetExpiredSubscriptions returns subscriptions that have expired
func (r *SubscriptionRepository) GetExpiredSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	subscriptions := []models.Subscription{}
	
	query := `
		SELECT id, user_id, plan_id, status, starts_at, ends_at, auto_renew, created_at
		FROM subscriptions
		WHERE status = 'active' AND ends_at <= NOW()
	`
	
	err := r.db.SelectContext(ctx, &subscriptions, query)
	if err != nil {
		return nil, fmt.Errorf("get expired subscriptions: %w", err)
	}
	
	return subscriptions, nil
}

// ExpireSubscriptions marks expired subscriptions as expired
func (r *SubscriptionRepository) ExpireSubscriptions(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET status = 'expired', updated_at = NOW() 
		 WHERE status = 'active' AND ends_at <= NOW()`,
	)
	if err != nil {
		return 0, fmt.Errorf("expire subscriptions: %w", err)
	}
	
	return result.RowsAffected()
}

// ============================================
// STATISTICS
// ============================================

// GetSubscriptionStats returns subscription statistics
func (r *SubscriptionRepository) GetSubscriptionStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)
	
	// Active subscribers count
	var activeSubscribers int
	if err := r.db.GetContext(ctx, &activeSubscribers,
		`SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'active' AND ends_at > NOW()`); err == nil {
		stats["active_subscribers"] = activeSubscribers
	}
	
	// Total subscriptions ever
	var totalSubscriptions int
	if err := r.db.GetContext(ctx, &totalSubscriptions,
		`SELECT COUNT(*) FROM subscriptions`); err == nil {
		stats["total_subscriptions"] = totalSubscriptions
	}
	
	// Subscriptions by plan
	rows, _ := r.db.QueryxContext(ctx,
		`SELECT sp.code, COUNT(s.id) 
		 FROM subscriptions s 
		 JOIN subscription_plans sp ON s.plan_id = sp.id 
		 WHERE s.status = 'active' AND s.ends_at > NOW()
		 GROUP BY sp.code`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var code string
			var count int
			rows.Scan(&code, &count)
			stats["plan_"+code] = count
		}
	}
	
	return stats, nil
}

// BeginTx starts a new database transaction
func (r *SubscriptionRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// CancelOtherActiveSubscriptions cancels all active subscriptions for the user except keepID.
// This is used as a safety net if duplicates already exist due to race conditions.
func (r *SubscriptionRepository) CancelOtherActiveSubscriptions(ctx context.Context, userID uuid.UUID, keepID uuid.UUID) (int64, error) {
	query := `
		UPDATE subscriptions
		SET status = 'canceled', canceled_at = NOW(), auto_renew = false, updated_at = NOW()
		WHERE user_id = $1 AND status = 'active' AND ends_at > NOW() AND id <> $2
	`
	res, err := r.db.ExecContext(ctx, query, userID, keepID)
	if err != nil {
		return 0, fmt.Errorf("cancel other active subscriptions: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// CancelAllActiveSubscriptionsTx cancels all currently active subscriptions for a user inside a tx.
func (r *SubscriptionRepository) CancelAllActiveSubscriptionsTx(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (int64, error) {
	query := `
		UPDATE subscriptions
		SET status = 'canceled', canceled_at = NOW(), auto_renew = false, updated_at = NOW()
		WHERE user_id = $1 AND status = 'active' AND ends_at > NOW()
	`
	res, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("cancel active subscriptions: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// CreateSubscriptionTx creates a new subscription within a transaction.
func (r *SubscriptionRepository) CreateSubscriptionTx(ctx context.Context, tx *sqlx.Tx, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, plan_id, status, starts_at, ends_at,
			external_id, auto_renew, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	_, err := tx.ExecContext(ctx, query,
		sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.StartsAt, sub.EndsAt,
		sub.ExternalID, sub.AutoRenew, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}
