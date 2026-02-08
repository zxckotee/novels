package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"novels-backend/internal/domain/models"
)

type ImportRunCookiesRepository struct {
	db *sqlx.DB
}

func NewImportRunCookiesRepository(db *sqlx.DB) *ImportRunCookiesRepository {
	return &ImportRunCookiesRepository{db: db}
}

// GetByRunID retrieves cookies for a specific import run
func (r *ImportRunCookiesRepository) GetByRunID(ctx context.Context, runID uuid.UUID) (*models.ImportRunCookie, error) {
	var cookie models.ImportRunCookie
	err := r.db.GetContext(ctx, &cookie, `
		SELECT run_id, cookie_header, created_at, updated_at
		FROM import_run_cookies
		WHERE run_id = $1
	`, runID)
	if err != nil {
		return nil, err
	}
	return &cookie, nil
}

// Upsert creates or updates cookies for an import run
func (r *ImportRunCookiesRepository) Upsert(ctx context.Context, runID uuid.UUID, cookieHeader string) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO import_run_cookies (run_id, cookie_header, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (run_id) DO UPDATE SET
			cookie_header = EXCLUDED.cookie_header,
			updated_at = EXCLUDED.updated_at
	`, runID, cookieHeader, now, now)
	if err != nil {
		return fmt.Errorf("upsert import run cookie: %w", err)
	}
	return nil
}

// Delete removes cookies for an import run
func (r *ImportRunCookiesRepository) Delete(ctx context.Context, runID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM import_run_cookies
		WHERE run_id = $1
	`, runID)
	if err != nil {
		return fmt.Errorf("delete import run cookie: %w", err)
	}
	return nil
}
