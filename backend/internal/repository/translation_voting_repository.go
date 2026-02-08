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

type TranslationVotingRepository struct {
	db *sqlx.DB
}

func NewTranslationVotingRepository(db *sqlx.DB) *TranslationVotingRepository {
	return &TranslationVotingRepository{db: db}
}

func (r *TranslationVotingRepository) GetTargetByID(ctx context.Context, id uuid.UUID) (*models.TranslationVoteTarget, error) {
	const q = `
		SELECT id, novel_id, proposal_id, status, translation_tickets_invested, created_at, updated_at
		FROM translation_vote_targets
		WHERE id = $1
	`
	var t models.TranslationVoteTarget
	err := r.db.QueryRowxContext(ctx, q, id).Scan(
		&t.ID, &t.NovelID, &t.ProposalID, &t.Status, &t.TranslationTicketsInvested, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get target: %w", err)
	}
	return &t, nil
}

func (r *TranslationVotingRepository) GetTargetByNovelID(ctx context.Context, novelID uuid.UUID) (*models.TranslationVoteTarget, error) {
	const q = `
		SELECT id, novel_id, proposal_id, status, translation_tickets_invested, created_at, updated_at
		FROM translation_vote_targets
		WHERE novel_id = $1
	`
	var t models.TranslationVoteTarget
	err := r.db.QueryRowxContext(ctx, q, novelID).Scan(
		&t.ID, &t.NovelID, &t.ProposalID, &t.Status, &t.TranslationTicketsInvested, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get target by novel: %w", err)
	}
	return &t, nil
}

func (r *TranslationVotingRepository) GetTargetByProposalID(ctx context.Context, proposalID uuid.UUID) (*models.TranslationVoteTarget, error) {
	const q = `
		SELECT id, novel_id, proposal_id, status, translation_tickets_invested, created_at, updated_at
		FROM translation_vote_targets
		WHERE proposal_id = $1
	`
	var t models.TranslationVoteTarget
	err := r.db.QueryRowxContext(ctx, q, proposalID).Scan(
		&t.ID, &t.NovelID, &t.ProposalID, &t.Status, &t.TranslationTicketsInvested, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get target by proposal: %w", err)
	}
	return &t, nil
}

func (r *TranslationVotingRepository) EnsureTargetForNovel(ctx context.Context, novelID uuid.UUID) (*models.TranslationVoteTarget, error) {
	existing, err := r.GetTargetByNovelID(ctx, novelID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	const q = `
		INSERT INTO translation_vote_targets (novel_id, status)
		VALUES ($1, 'voting')
		ON CONFLICT (novel_id) WHERE novel_id IS NOT NULL DO UPDATE SET updated_at = NOW()
		RETURNING id, novel_id, proposal_id, status, translation_tickets_invested, created_at, updated_at
	`
	var t models.TranslationVoteTarget
	err = r.db.QueryRowxContext(ctx, q, novelID).Scan(
		&t.ID, &t.NovelID, &t.ProposalID, &t.Status, &t.TranslationTicketsInvested, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ensure target for novel: %w", err)
	}
	return &t, nil
}

func (r *TranslationVotingRepository) EnsureTargetForProposal(ctx context.Context, proposalID uuid.UUID) (*models.TranslationVoteTarget, error) {
	existing, err := r.GetTargetByProposalID(ctx, proposalID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	const q = `
		INSERT INTO translation_vote_targets (proposal_id, status)
		VALUES ($1, 'voting')
		ON CONFLICT (proposal_id) WHERE proposal_id IS NOT NULL DO UPDATE SET updated_at = NOW()
		RETURNING id, novel_id, proposal_id, status, translation_tickets_invested, created_at, updated_at
	`
	var t models.TranslationVoteTarget
	err = r.db.QueryRowxContext(ctx, q, proposalID).Scan(
		&t.ID, &t.NovelID, &t.ProposalID, &t.Status, &t.TranslationTicketsInvested, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ensure target for proposal: %w", err)
	}
	return &t, nil
}

func (r *TranslationVotingRepository) CastTranslationVote(ctx context.Context, userID uuid.UUID, targetID uuid.UUID, amount int) error {
	const q = `
		INSERT INTO translation_votes (user_id, target_id, amount)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, q, userID, targetID, amount)
	if err != nil {
		return fmt.Errorf("insert translation vote: %w", err)
	}
	return nil
}

func (r *TranslationVotingRepository) ListLeaderboard(ctx context.Context, limit int) ([]models.TranslationLeaderboardEntry, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Best-effort: title comes from either proposal.title or novel_localizations.title (ru).
	const q = `
		SELECT
			t.id AS target_id,
			t.status,
			t.translation_tickets_invested AS score,
			t.novel_id,
			t.proposal_id,
			COALESCE(p.title, nl.title, 'Untitled') AS title,
			COALESCE(p.cover_url, NULL) AS cover_url
		FROM translation_vote_targets t
		LEFT JOIN novel_proposals p ON p.id = t.proposal_id
		LEFT JOIN novel_localizations nl ON nl.novel_id = t.novel_id AND nl.lang = 'ru'
		WHERE t.status = 'voting'
		ORDER BY t.translation_tickets_invested DESC, t.created_at ASC
		LIMIT $1
	`
	rows, err := r.db.QueryxContext(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list translation leaderboard: %w", err)
	}
	defer rows.Close()

	out := make([]models.TranslationLeaderboardEntry, 0, limit)
	for rows.Next() {
		var e models.TranslationLeaderboardEntry
		if err := rows.Scan(&e.TargetID, &e.Status, &e.Score, &e.NovelID, &e.ProposalID, &e.Title, &e.CoverURL); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, e)
	}
	return out, nil
}

// OpsTranslationTargetEntry is a richer view used by admin/ops screens.
type OpsTranslationTargetEntry struct {
	TargetID uuid.UUID `json:"targetId" db:"target_id"`
	Status   models.TranslationVoteTargetStatus `json:"status" db:"status"`
	Score    int       `json:"score" db:"score"`

	NovelID    *uuid.UUID `json:"novelId,omitempty" db:"novel_id"`
	ProposalID *uuid.UUID `json:"proposalId,omitempty" db:"proposal_id"`

	Title    string  `json:"title" db:"title"`
	CoverURL *string `json:"coverUrl,omitempty" db:"cover_url"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// ListTargetsForOps returns targets of any status, with best-effort title/cover for admin ops pages.
func (r *TranslationVotingRepository) ListTargetsForOps(ctx context.Context, limit int) ([]OpsTranslationTargetEntry, error) {
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	const q = `
		SELECT
			t.id AS target_id,
			t.status,
			t.translation_tickets_invested AS score,
			t.novel_id,
			t.proposal_id,
			COALESCE(p.title, nl.title, 'Untitled') AS title,
			COALESCE(p.cover_url, NULL) AS cover_url,
			t.updated_at
		FROM translation_vote_targets t
		LEFT JOIN novel_proposals p ON p.id = t.proposal_id
		LEFT JOIN novel_localizations nl ON nl.novel_id = t.novel_id AND nl.lang = 'ru'
		ORDER BY t.updated_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryxContext(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list targets for ops: %w", err)
	}
	defer rows.Close()
	out := make([]OpsTranslationTargetEntry, 0, limit)
	for rows.Next() {
		var e OpsTranslationTargetEntry
		if err := rows.StructScan(&e); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *TranslationVotingRepository) GetTopTarget(ctx context.Context) (*models.TranslationVoteTarget, error) {
	const q = `
		SELECT id, novel_id, proposal_id, status, translation_tickets_invested, created_at, updated_at
		FROM translation_vote_targets
		WHERE status = 'voting'
		ORDER BY translation_tickets_invested DESC, created_at ASC
		LIMIT 1
	`
	var t models.TranslationVoteTarget
	err := r.db.QueryRowxContext(ctx, q).Scan(
		&t.ID, &t.NovelID, &t.ProposalID, &t.Status, &t.TranslationTicketsInvested, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get top target: %w", err)
	}
	return &t, nil
}

func (r *TranslationVotingRepository) UpdateTargetStatus(ctx context.Context, targetID uuid.UUID, status models.TranslationVoteTargetStatus) error {
	const q = `
		UPDATE translation_vote_targets
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, q, targetID, status)
	if err != nil {
		return fmt.Errorf("update target status: %w", err)
	}
	return nil
}

// BindProposalToNovel converts a proposal-based target into a novel-based target (preserving votes),
// and optionally moves it from waiting_release -> translating.
func (r *TranslationVotingRepository) BindProposalToNovel(ctx context.Context, proposalID uuid.UUID, novelID uuid.UUID) (*models.TranslationVoteTarget, error) {
	// Single statement keeps row identity & accumulated score.
	const q = `
		UPDATE translation_vote_targets
		SET novel_id = $2,
		    proposal_id = NULL,
		    status = CASE WHEN status = 'waiting_release' THEN 'translating' ELSE status END,
		    updated_at = NOW()
		WHERE proposal_id = $1
		RETURNING id, novel_id, proposal_id, status, translation_tickets_invested, created_at, updated_at
	`
	var t models.TranslationVoteTarget
	err := r.db.QueryRowxContext(ctx, q, proposalID, novelID).Scan(
		&t.ID, &t.NovelID, &t.ProposalID, &t.Status, &t.TranslationTicketsInvested, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("bind proposal to novel: %w", err)
	}
	return &t, nil
}

// TouchUpdatedAt is used by ON CONFLICT upserts and tests.
func (r *TranslationVotingRepository) TouchUpdatedAt(ctx context.Context, targetID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE translation_vote_targets SET updated_at = $2 WHERE id = $1`, targetID, time.Now().UTC())
	return err
}

