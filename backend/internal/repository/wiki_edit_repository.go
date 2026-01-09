package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"novels/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// WikiEditRepository handles wiki edit database operations
type WikiEditRepository struct {
	db *sqlx.DB
}

// NewWikiEditRepository creates a new wiki edit repository
func NewWikiEditRepository(db *sqlx.DB) *WikiEditRepository {
	return &WikiEditRepository{db: db}
}

// CreateEditRequest creates a new edit request
func (r *WikiEditRepository) CreateEditRequest(ctx context.Context, request *models.NovelEditRequest) error {
	query := `
		INSERT INTO novel_edit_requests (novel_id, user_id, edit_reason)
		VALUES ($1, $2, $3)
		RETURNING id, status, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		request.NovelID,
		request.UserID,
		request.EditReason,
	).Scan(&request.ID, &request.Status, &request.CreatedAt, &request.UpdatedAt)
}

// AddEditChange adds a change to an edit request
func (r *WikiEditRepository) AddEditChange(ctx context.Context, change *models.NovelEditRequestChange) error {
	query := `
		INSERT INTO novel_edit_request_changes (request_id, field_type, lang, old_value, new_value)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.db.QueryRowxContext(ctx, query,
		change.RequestID,
		change.FieldType,
		change.Lang,
		change.OldValue,
		change.NewValue,
	).Scan(&change.ID, &change.CreatedAt)
}

// GetEditRequestByID gets an edit request by ID
func (r *WikiEditRepository) GetEditRequestByID(ctx context.Context, id uuid.UUID) (*models.NovelEditRequest, error) {
	var request models.NovelEditRequest
	query := `SELECT * FROM novel_edit_requests WHERE id = $1`
	err := r.db.GetContext(ctx, &request, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &request, err
}

// GetEditRequestChanges gets all changes for an edit request
func (r *WikiEditRepository) GetEditRequestChanges(ctx context.Context, requestID uuid.UUID) ([]models.NovelEditRequestChange, error) {
	var changes []models.NovelEditRequestChange
	query := `SELECT * FROM novel_edit_request_changes WHERE request_id = $1`
	err := r.db.SelectContext(ctx, &changes, query, requestID)
	return changes, err
}

// ListEditRequests lists edit requests with filters
func (r *WikiEditRepository) ListEditRequests(ctx context.Context, params models.EditRequestListParams) ([]models.NovelEditRequest, int, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if params.NovelID != nil {
		conditions = append(conditions, fmt.Sprintf("novel_id = $%d", argNum))
		args = append(args, *params.NovelID)
		argNum++
	}

	if params.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, *params.UserID)
		argNum++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *params.Status)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM novel_edit_requests %s`, whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Pagination
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.Limit

	query := fmt.Sprintf(`
		SELECT * FROM novel_edit_requests
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	var requests []models.NovelEditRequest
	if err := r.db.SelectContext(ctx, &requests, query, args...); err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// GetPendingRequests gets pending edit requests for moderation
func (r *WikiEditRepository) GetPendingRequests(ctx context.Context, limit, offset int) ([]models.NovelEditRequest, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM novel_edit_requests WHERE status = 'pending'`); err != nil {
		return nil, 0, err
	}

	var requests []models.NovelEditRequest
	query := `
		SELECT * FROM novel_edit_requests
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &requests, query, limit, offset)
	return requests, total, err
}

// ApproveEditRequest approves an edit request
func (r *WikiEditRepository) ApproveEditRequest(ctx context.Context, requestID, moderatorID uuid.UUID, comment string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Call the stored function that applies changes
	var success bool
	err = tx.QueryRowxContext(ctx,
		`SELECT apply_edit_request($1, $2)`, requestID, moderatorID).Scan(&success)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("failed to apply edit request")
	}

	// Add moderator comment if provided
	if comment != "" {
		_, err = tx.ExecContext(ctx,
			`UPDATE novel_edit_requests SET moderator_comment = $2 WHERE id = $1`,
			requestID, comment)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RejectEditRequest rejects an edit request
func (r *WikiEditRepository) RejectEditRequest(ctx context.Context, requestID, moderatorID uuid.UUID, comment string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE novel_edit_requests 
		SET status = 'rejected', moderator_id = $2, moderator_comment = $3, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'pending'`,
		requestID, moderatorID, comment)
	return err
}

// WithdrawEditRequest withdraws an edit request (by the original user)
func (r *WikiEditRepository) WithdrawEditRequest(ctx context.Context, requestID, userID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE novel_edit_requests 
		SET status = 'withdrawn', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'pending'`,
		requestID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("request not found or cannot be withdrawn")
	}

	return nil
}

// GetEditHistory gets edit history for a novel
func (r *WikiEditRepository) GetEditHistory(ctx context.Context, params models.EditHistoryListParams) ([]models.NovelEditHistory, int, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if params.NovelID != nil {
		conditions = append(conditions, fmt.Sprintf("novel_id = $%d", argNum))
		args = append(args, *params.NovelID)
		argNum++
	}

	if params.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, *params.UserID)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM novel_edit_history %s`, whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Pagination
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.Limit

	query := fmt.Sprintf(`
		SELECT * FROM novel_edit_history
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	var history []models.NovelEditHistory
	if err := r.db.SelectContext(ctx, &history, query, args...); err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

// GetUserEditRequests gets all edit requests by a user
func (r *WikiEditRepository) GetUserEditRequests(ctx context.Context, userID uuid.UUID) ([]models.NovelEditRequest, error) {
	var requests []models.NovelEditRequest
	query := `
		SELECT * FROM novel_edit_requests
		WHERE user_id = $1
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &requests, query, userID)
	return requests, err
}

// GetPlatformStats gets global platform statistics
func (r *WikiEditRepository) GetPlatformStats(ctx context.Context) (*models.PlatformStats, error) {
	var stats models.PlatformStats
	query := `SELECT * FROM platform_stats WHERE id = 1`
	err := r.db.GetContext(ctx, &stats, query)
	if err == sql.ErrNoRows {
		// Return empty stats if not found
		return &models.PlatformStats{}, nil
	}
	return &stats, err
}

// RefreshPlatformStats refreshes the cached platform statistics
func (r *WikiEditRepository) RefreshPlatformStats(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `SELECT refresh_platform_stats()`)
	return err
}

// CountPendingEditRequests counts pending edit requests
func (r *WikiEditRepository) CountPendingEditRequests(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM novel_edit_requests WHERE status = 'pending'`)
	return count, err
}

// HasPendingEditRequest checks if user has a pending edit request for a novel
func (r *WikiEditRepository) HasPendingEditRequest(ctx context.Context, userID, novelID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM novel_edit_requests 
			WHERE user_id = $1 AND novel_id = $2 AND status = 'pending'
		)`, userID, novelID)
	return exists, err
}
