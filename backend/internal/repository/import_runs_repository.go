package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"novels-backend/internal/domain/models"
)

type ImportRunsRepository struct {
	db *sqlx.DB
}

func NewImportRunsRepository(db *sqlx.DB) *ImportRunsRepository {
	return &ImportRunsRepository{db: db}
}

func (r *ImportRunsRepository) Create(ctx context.Context, run *models.ImportRun) error {
	if run == nil {
		return fmt.Errorf("run is nil")
	}
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	now := time.Now().UTC()
	if run.StartedAt.IsZero() {
		run.StartedAt = now
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	if run.UpdatedAt.IsZero() {
		run.UpdatedAt = now
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO import_runs (
			id, proposal_id, novel_id, importer, status, error,
			progress_current, progress_total, checkpoint, cloudflare_blocked,
			started_at, finished_at, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,
			$11,$12,$13,$14
		)
	`, run.ID, run.ProposalID, run.NovelID, run.Importer, run.Status, run.Error,
		run.ProgressCurrent, run.ProgressTotal, mustJSON(run.Checkpoint), run.CloudflareBlocked,
		run.StartedAt, run.FinishedAt, run.CreatedAt, run.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create import run: %w", err)
	}
	return nil
}

func (r *ImportRunsRepository) SetResult(ctx context.Context, runID uuid.UUID, status models.ImportRunStatus, novelID *uuid.UUID, errMsg *string, cloudflareBlocked *bool) error {
	finished := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_runs
		SET status = $2,
		    novel_id = COALESCE($3, novel_id),
		    error = $4,
		    finished_at = $5,
		    cloudflare_blocked = COALESCE($6, cloudflare_blocked),
		    updated_at = NOW()
		WHERE id = $1
	`, runID, status, novelID, errMsg, finished, cloudflareBlocked)
	if err != nil {
		return fmt.Errorf("set import run result: %w", err)
	}
	return nil
}

func (r *ImportRunsRepository) UpdateProgress(ctx context.Context, runID uuid.UUID, current, total int, checkpoint any) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_runs
		SET progress_current = $2,
		    progress_total = $3,
		    checkpoint = $4,
		    updated_at = NOW()
		WHERE id = $1
	`, runID, current, total, mustJSON(checkpoint))
	if err != nil {
		return fmt.Errorf("update import run progress: %w", err)
	}
	return nil
}

func (r *ImportRunsRepository) SetStatus(ctx context.Context, runID uuid.UUID, status models.ImportRunStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_runs
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, runID, status)
	if err != nil {
		return fmt.Errorf("set import run status: %w", err)
	}
	return nil
}

func (r *ImportRunsRepository) SetNovelID(ctx context.Context, runID uuid.UUID, novelID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_runs
		SET novel_id = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, runID, novelID)
	if err != nil {
		return fmt.Errorf("set import run novel_id: %w", err)
	}
	return nil
}

func (r *ImportRunsRepository) List(ctx context.Context, limit int, status *models.ImportRunStatus) ([]models.ImportRun, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	q := `
		SELECT id, proposal_id, novel_id, importer, status, error,
		       progress_current, progress_total, COALESCE(checkpoint, '{}'::jsonb) AS checkpoint,
		       cloudflare_blocked,
		       started_at, finished_at, created_at, updated_at
		FROM import_runs
	`
	args := []any{}
	if status != nil {
		q += ` WHERE status = $1`
		args = append(args, *status)
		q += ` ORDER BY started_at DESC LIMIT $2`
		args = append(args, limit)
	} else {
		q += ` ORDER BY started_at DESC LIMIT $1`
		args = append(args, limit)
	}
	out := []models.ImportRun{}
	if err := r.db.SelectContext(ctx, &out, q, args...); err != nil {
		return nil, fmt.Errorf("list import runs: %w", err)
	}
	return out, nil
}

func (r *ImportRunsRepository) GetByID(ctx context.Context, runID uuid.UUID) (*models.ImportRun, error) {
	var run models.ImportRun
	err := r.db.GetContext(ctx, &run, `
		SELECT id, proposal_id, novel_id, importer, status, error,
		       progress_current, progress_total, COALESCE(checkpoint, '{}'::jsonb) AS checkpoint,
		       cloudflare_blocked,
		       started_at, finished_at, created_at, updated_at
		FROM import_runs
		WHERE id = $1
	`, runID)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func mustJSON(v any) any {
	if v == nil {
		return nil
	}
	if rm, ok := v.(json.RawMessage); ok {
		if len(rm) == 0 {
			return nil
		}
		return []byte(rm)
	}
	if b, ok := v.([]byte); ok {
		if len(b) == 0 {
			return nil
		}
		return b
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

