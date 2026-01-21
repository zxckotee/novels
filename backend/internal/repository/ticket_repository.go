package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels-backend/internal/domain/models"
)

type TicketRepository struct {
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// GetBalance returns the balance for a specific ticket type
func (r *TicketRepository) GetBalance(ctx context.Context, userID uuid.UUID, ticketType models.TicketType) (int, error) {
	var balance int
	query := `SELECT COALESCE(balance, 0) FROM ticket_balances WHERE user_id = $1 AND type = $2`
	
	err := r.db.GetContext(ctx, &balance, query, userID, ticketType)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get ticket balance: %w", err)
	}
	
	return balance, nil
}

// GetAllBalances returns all ticket balances for a user
func (r *TicketRepository) GetAllBalances(ctx context.Context, userID uuid.UUID) ([]models.TicketBalance, error) {
	balances := []models.TicketBalance{}
	query := `SELECT user_id, type, balance, updated_at FROM ticket_balances WHERE user_id = $1`
	
	err := r.db.SelectContext(ctx, &balances, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get all balances: %w", err)
	}
	
	return balances, nil
}

// GetWalletInfo returns complete wallet info for a user
func (r *TicketRepository) GetWalletInfo(ctx context.Context, userID uuid.UUID) (*models.WalletInfo, error) {
	query := `
		SELECT 
			$1::uuid as user_id,
			COALESCE((SELECT balance FROM ticket_balances WHERE user_id = $1 AND type = 'daily_vote'), 0) as daily_votes,
			COALESCE((SELECT balance FROM ticket_balances WHERE user_id = $1 AND type = 'novel_request'), 0) as novel_requests,
			COALESCE((SELECT balance FROM ticket_balances WHERE user_id = $1 AND type = 'translation_ticket'), 0) as translation_tickets
	`
	
	wallet := &models.WalletInfo{}
	err := r.db.QueryRowxContext(ctx, query, userID).Scan(
		&wallet.UserID,
		&wallet.DailyVotes,
		&wallet.NovelRequests,
		&wallet.TranslationTickets,
	)
	if err != nil {
		return nil, fmt.Errorf("get wallet info: %w", err)
	}
	
	// Calculate next daily reset (3:00 MSK = 0:00 UTC)
	now := time.Now().UTC()
	nextReset := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if now.After(nextReset) {
		nextReset = nextReset.Add(24 * time.Hour)
	}
	wallet.NextDailyReset = nextReset
	
	return wallet, nil
}

// CreateTransaction creates a new ticket transaction
func (r *TicketRepository) CreateTransaction(ctx context.Context, tx models.TicketTransaction) error {
	query := `
		INSERT INTO ticket_transactions (id, user_id, type, delta, reason, ref_type, ref_id, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}

	// IMPORTANT: empty string is not NULL in Postgres and will collide under the UNIQUE constraint.
	// Treat empty idempotency keys as NULL so the key is truly "optional".
	var idempotencyKey any
	if tx.IdempotencyKey != "" {
		idempotencyKey = tx.IdempotencyKey
	} else {
		idempotencyKey = nil
	}
	
	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.UserID, tx.Type, tx.Delta, tx.Reason,
		tx.RefType, tx.RefID, idempotencyKey, tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	
	return nil
}

// CreateTransactionTx creates a transaction within a database transaction
func (r *TicketRepository) CreateTransactionTx(ctx context.Context, dbTx *sqlx.Tx, tx models.TicketTransaction) error {
	query := `
		INSERT INTO ticket_transactions (id, user_id, type, delta, reason, ref_type, ref_id, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}

	// IMPORTANT: empty string is not NULL in Postgres and will collide under the UNIQUE constraint.
	// Treat empty idempotency keys as NULL so the key is truly "optional".
	var idempotencyKey any
	if tx.IdempotencyKey != "" {
		idempotencyKey = tx.IdempotencyKey
	} else {
		idempotencyKey = nil
	}
	
	_, err := dbTx.ExecContext(ctx, query,
		tx.ID, tx.UserID, tx.Type, tx.Delta, tx.Reason,
		tx.RefType, tx.RefID, idempotencyKey, tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create transaction tx: %w", err)
	}
	
	return nil
}

// SpendTickets spends tickets atomically (checks balance, creates transaction, updates balance)
func (r *TicketRepository) SpendTickets(ctx context.Context, userID uuid.UUID, ticketType models.TicketType, amount int, reason, refType string, refID *uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Check current balance
	var balance int
	query := `SELECT COALESCE(balance, 0) FROM ticket_balances WHERE user_id = $1 AND type = $2 FOR UPDATE`
	err = tx.GetContext(ctx, &balance, query, userID, ticketType)
	if errors.Is(err, sql.ErrNoRows) {
		balance = 0
	} else if err != nil {
		return fmt.Errorf("check balance: %w", err)
	}
	
	if balance < amount {
		return fmt.Errorf("insufficient balance: have %d, need %d", balance, amount)
	}
	
	// Create transaction
	transaction := models.TicketTransaction{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      ticketType,
		Delta:     -amount,
		Reason:    reason,
		RefType:   refType,
		RefID:     refID,
		CreatedAt: time.Now(),
	}
	
	err = r.CreateTransactionTx(ctx, tx, transaction)
	if err != nil {
		return fmt.Errorf("create spend transaction: %w", err)
	}
	
	return tx.Commit()
}

// GrantTickets grants tickets to a user (creates transaction, updates balance)
func (r *TicketRepository) GrantTickets(ctx context.Context, userID uuid.UUID, ticketType models.TicketType, amount int, reason, refType string, refID *uuid.UUID, idempotencyKey string) error {
	transaction := models.TicketTransaction{
		ID:             uuid.New(),
		UserID:         userID,
		Type:           ticketType,
		Delta:          amount,
		Reason:         reason,
		RefType:        refType,
		RefID:          refID,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}
	
	return r.CreateTransaction(ctx, transaction)
}

// SetDailyVotes sets daily votes to a specific amount (for daily grant)
func (r *TicketRepository) SetDailyVotes(ctx context.Context, userID uuid.UUID, amount int, idempotencyKey string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Get current balance
	var currentBalance int
	query := `SELECT COALESCE(balance, 0) FROM ticket_balances WHERE user_id = $1 AND type = 'daily_vote'`
	err = tx.GetContext(ctx, &currentBalance, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		currentBalance = 0
	} else if err != nil {
		return fmt.Errorf("get current balance: %w", err)
	}
	
	// Calculate delta
	delta := amount - currentBalance
	
	// Create transaction
	transaction := models.TicketTransaction{
		ID:             uuid.New(),
		UserID:         userID,
		Type:           models.TicketTypeDailyVote,
		Delta:          delta,
		Reason:         models.ReasonDailyGrant,
		RefType:        "daily_grant",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}
	
	err = r.CreateTransactionTx(ctx, tx, transaction)
	if err != nil {
		return fmt.Errorf("create daily grant transaction: %w", err)
	}
	
	return tx.Commit()
}

// GetTransactions returns paginated transactions for a user
func (r *TicketRepository) GetTransactions(ctx context.Context, filter models.TransactionFilter) ([]models.TicketTransaction, int, error) {
	transactions := []models.TicketTransaction{}
	
	baseQuery := `FROM ticket_transactions WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	
	if filter.UserID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argNum)
		args = append(args, *filter.UserID)
		argNum++
	}
	
	if filter.Type != nil {
		baseQuery += fmt.Sprintf(" AND type = $%d", argNum)
		args = append(args, *filter.Type)
		argNum++
	}
	
	// Count total
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}
	
	// Get transactions
	selectQuery := fmt.Sprintf(`
		SELECT id, user_id, type, delta, reason, ref_type, ref_id, created_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argNum, argNum+1)
	
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	
	err = r.db.SelectContext(ctx, &transactions, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get transactions: %w", err)
	}
	
	return transactions, total, nil
}

// GetTotalSpentInPeriod returns total tickets spent by a user in a period
func (r *TicketRepository) GetTotalSpentInPeriod(ctx context.Context, userID uuid.UUID, ticketType *models.TicketType, since time.Time) (int, error) {
	query := `
		SELECT COALESCE(ABS(SUM(delta)), 0)
		FROM ticket_transactions
		WHERE user_id = $1 AND delta < 0 AND created_at >= $2
	`
	args := []interface{}{userID, since}
	
	if ticketType != nil {
		query += " AND type = $3"
		args = append(args, *ticketType)
	}
	
	var total int
	err := r.db.GetContext(ctx, &total, query, args...)
	if err != nil {
		return 0, fmt.Errorf("get total spent: %w", err)
	}
	
	return total, nil
}

// GetLeaderboard returns top users by tickets spent
func (r *TicketRepository) GetLeaderboard(ctx context.Context, period string, limit int) ([]struct {
	UserID       uuid.UUID `db:"user_id"`
	TicketsSpent int       `db:"tickets_spent"`
	DisplayName  string    `db:"display_name"`
	AvatarURL    *string   `db:"avatar_url"`
	Level        int       `db:"level"`
}, error) {
	var since time.Time
	now := time.Now().UTC()
	
	switch period {
	case "day":
		since = now.Add(-24 * time.Hour)
	case "week":
		since = now.Add(-7 * 24 * time.Hour)
	case "month":
		since = now.Add(-30 * 24 * time.Hour)
	default:
		since = now.Add(-24 * time.Hour)
	}
	
	query := `
		SELECT 
			tt.user_id,
			ABS(SUM(tt.delta)) as tickets_spent,
			COALESCE(up.display_name, '') as display_name,
			up.avatar_key as avatar_url,
			COALESCE(ux.level, 1) as level
		FROM ticket_transactions tt
		LEFT JOIN user_profiles up ON tt.user_id = up.user_id
		LEFT JOIN user_xp ux ON tt.user_id = ux.user_id
		WHERE tt.delta < 0 AND tt.created_at >= $1
		GROUP BY tt.user_id, up.display_name, up.avatar_key, ux.level
		ORDER BY tickets_spent DESC
		LIMIT $2
	`
	
	var results []struct {
		UserID       uuid.UUID `db:"user_id"`
		TicketsSpent int       `db:"tickets_spent"`
		DisplayName  string    `db:"display_name"`
		AvatarURL    *string   `db:"avatar_url"`
		Level        int       `db:"level"`
	}
	
	err := r.db.SelectContext(ctx, &results, query, since, limit)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}
	
	return results, nil
}

// CheckIdempotencyKey checks if a transaction with the given key exists
func (r *TicketRepository) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticket_transactions WHERE idempotency_key = $1)`
	
	err := r.db.GetContext(ctx, &exists, query, key)
	if err != nil {
		return false, fmt.Errorf("check idempotency key: %w", err)
	}
	
	return exists, nil
}

// BeginTx starts a new database transaction
func (r *TicketRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
