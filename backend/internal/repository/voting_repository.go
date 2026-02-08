package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"novels-backend/internal/domain/models"
)

type VotingRepository struct {
	db *sqlx.DB
}

func NewVotingRepository(db *sqlx.DB) *VotingRepository {
	return &VotingRepository{db: db}
}

// ============================================
// PROPOSALS
// ============================================

// CreateProposal creates a new novel proposal
func (r *VotingRepository) CreateProposal(ctx context.Context, proposal *models.NovelProposal) error {
	query := `
		INSERT INTO novel_proposals (
			id, user_id, original_link, status,
			title, alt_titles, author, description, cover_url,
			genres, tags, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`
	
	if proposal.ID == uuid.Nil {
		proposal.ID = uuid.New()
	}
	now := time.Now()
	proposal.CreatedAt = now
	proposal.UpdatedAt = now
	
	_, err := r.db.ExecContext(ctx, query,
		proposal.ID, proposal.UserID, proposal.OriginalLink, proposal.Status,
		proposal.Title, pq.Array(proposal.AltTitles), proposal.Author,
		proposal.Description, proposal.CoverURL,
		pq.Array(proposal.Genres), pq.Array(proposal.Tags),
		proposal.CreatedAt, proposal.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create proposal: %w", err)
	}
	
	return nil
}

// GetProposalByShortID finds a proposal by short ID (first 8 chars of UUID)
func (r *VotingRepository) GetProposalByShortID(ctx context.Context, shortID string) (*uuid.UUID, error) {
	if len(shortID) != 8 {
		return nil, fmt.Errorf("short ID must be exactly 8 characters")
	}
	query := `SELECT id FROM novel_proposals WHERE id::text LIKE $1 || '%' LIMIT 1`
	var foundID uuid.UUID
	err := r.db.GetContext(ctx, &foundID, query, shortID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get proposal by short id: %w", err)
	}
	return &foundID, nil
}

// GetProposalByID returns a proposal by ID
func (r *VotingRepository) GetProposalByID(ctx context.Context, id uuid.UUID) (*models.NovelProposal, error) {
	query := `
		SELECT 
			np.id, np.user_id, np.original_link, np.status,
			np.title, np.alt_titles, np.author, np.description, np.cover_url,
			np.genres, np.tags, np.vote_score, np.votes_count, np.translation_tickets_invested,
			np.moderator_id, np.reject_reason,
			np.created_at, np.updated_at
		FROM novel_proposals np
		WHERE np.id = $1
	`
	
	var proposal models.NovelProposal
	var altTitles, genres, tags pq.StringArray
	
	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&proposal.ID, &proposal.UserID, &proposal.OriginalLink, &proposal.Status,
		&proposal.Title, &altTitles, &proposal.Author, &proposal.Description, &proposal.CoverURL,
		&genres, &tags, &proposal.VoteScore, &proposal.VotesCount, &proposal.TranslationTicketsInvested,
		&proposal.ModeratorID, &proposal.RejectReason,
		&proposal.CreatedAt, &proposal.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get proposal by id: %w", err)
	}
	
	proposal.AltTitles = altTitles
	proposal.Genres = genres
	proposal.Tags = tags
	
	return &proposal, nil
}

// GetProposalWithUser returns a proposal with user info
func (r *VotingRepository) GetProposalWithUser(ctx context.Context, id uuid.UUID, currentUserID *uuid.UUID) (*models.NovelProposal, error) {
	proposal, err := r.GetProposalByID(ctx, id)
	if err != nil || proposal == nil {
		return proposal, err
	}
	
	// Get user info
	userQuery := `
		SELECT 
			u.id,
			COALESCE(up.display_name, u.email) as display_name,
			up.avatar_key as avatar_url,
			COALESCE(ux.level, 1) as level
		FROM users u
		LEFT JOIN user_profiles up ON u.id = up.user_id
		LEFT JOIN user_xp ux ON u.id = ux.user_id
		WHERE u.id = $1
	`
	
	user := &models.ProposalUser{}
	err = r.db.QueryRowxContext(ctx, userQuery, proposal.UserID).Scan(
		&user.ID, &user.DisplayName, &user.AvatarURL, &user.Level,
	)
	if err == nil {
		proposal.User = user
	}
	
	// Get current user's vote if provided
	if currentUserID != nil {
		var userVote int
		voteQuery := `SELECT COALESCE(SUM(amount), 0) FROM votes WHERE proposal_id = $1 AND user_id = $2 AND ticket_type = 'daily_vote'`
		r.db.GetContext(ctx, &userVote, voteQuery, id, *currentUserID)
		if userVote > 0 {
			proposal.UserVote = &userVote
		}
	}
	
	return proposal, nil
}

// ListProposals returns proposals with filters
func (r *VotingRepository) ListProposals(ctx context.Context, filter models.ProposalFilter) ([]models.NovelProposal, int, error) {
	proposals := []models.NovelProposal{}
	
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argNum := 1
	
	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("np.status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}
	
	if filter.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("np.user_id = $%d", argNum))
		args = append(args, *filter.UserID)
		argNum++
	}
	
	whereClause := "WHERE " + whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}
	
	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM novel_proposals np %s", whereClause)
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count proposals: %w", err)
	}
	
	// Sort
	orderBy := "np.created_at DESC"
	switch filter.Sort {
	case "votes":
		orderBy = "np.vote_score DESC, np.created_at ASC"
	case "oldest":
		orderBy = "np.created_at ASC"
	case "newest":
		orderBy = "np.created_at DESC"
	}
	
	// Get proposals
	selectQuery := fmt.Sprintf(`
		SELECT
			np.id, np.user_id, np.original_link, np.status,
			np.title, np.alt_titles, np.author, np.description, np.cover_url,
			np.genres, np.tags, np.vote_score, np.votes_count, np.translation_tickets_invested,
			np.moderator_id, np.reject_reason,
			np.created_at, np.updated_at,
			COALESCE(up.display_name, u.email) as user_display_name,
			up.avatar_key as user_avatar,
			COALESCE(ux.level, 1) as user_level
		FROM novel_proposals np
		LEFT JOIN users u ON np.user_id = u.id
		LEFT JOIN user_profiles up ON np.user_id = up.user_id
		LEFT JOIN user_xp ux ON np.user_id = ux.user_id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argNum, argNum+1)
	
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	
	rows, err := r.db.QueryxContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list proposals: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var proposal models.NovelProposal
		var altTitles, genres, tags pq.StringArray
		var userDisplayName string
		var userAvatarURL *string
		var userLevel int
		
		err := rows.Scan(
			&proposal.ID, &proposal.UserID, &proposal.OriginalLink, &proposal.Status,
			&proposal.Title, &altTitles, &proposal.Author, &proposal.Description, &proposal.CoverURL,
			&genres, &tags, &proposal.VoteScore, &proposal.VotesCount, &proposal.TranslationTicketsInvested,
			&proposal.ModeratorID, &proposal.RejectReason,
			&proposal.CreatedAt, &proposal.UpdatedAt,
			&userDisplayName, &userAvatarURL, &userLevel,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan proposal: %w", err)
		}
		
		proposal.AltTitles = altTitles
		proposal.Genres = genres
		proposal.Tags = tags
		proposal.User = &models.ProposalUser{
			ID:          proposal.UserID,
			DisplayName: userDisplayName,
			AvatarURL:   userAvatarURL,
			Level:       userLevel,
		}
		
		proposals = append(proposals, proposal)
	}
	
	return proposals, total, nil
}

// UpdateProposal updates a proposal
func (r *VotingRepository) UpdateProposal(ctx context.Context, proposal *models.NovelProposal) error {
	query := `
		UPDATE novel_proposals SET
			original_link = $2,
			title = $3,
			alt_titles = $4,
			author = $5,
			description = $6,
			cover_url = $7,
			genres = $8,
			tags = $9,
			updated_at = NOW()
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query,
		proposal.ID, proposal.OriginalLink,
		proposal.Title, pq.Array(proposal.AltTitles), proposal.Author,
		proposal.Description, proposal.CoverURL,
		pq.Array(proposal.Genres), pq.Array(proposal.Tags),
	)
	if err != nil {
		return fmt.Errorf("update proposal: %w", err)
	}
	
	return nil
}

// UpdateProposalStatus updates proposal status
func (r *VotingRepository) UpdateProposalStatus(ctx context.Context, id uuid.UUID, status models.ProposalStatus, moderatorID *uuid.UUID, rejectReason *string) error {
	query := `
		UPDATE novel_proposals SET
			status = $2,
			moderator_id = $3,
			reject_reason = $4,
			updated_at = NOW()
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query, id, status, moderatorID, rejectReason)
	if err != nil {
		return fmt.Errorf("update proposal status: %w", err)
	}
	
	return nil
}

// SubmitProposalForModeration changes status to moderation
func (r *VotingRepository) SubmitProposalForModeration(ctx context.Context, id uuid.UUID) error {
	return r.UpdateProposalStatus(ctx, id, models.ProposalStatusModeration, nil, nil)
}

// DeleteProposal deletes a proposal
func (r *VotingRepository) DeleteProposal(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM novel_proposals WHERE id = $1", id)
	return err
}

// ============================================
// VOTES
// ============================================

// CreateVote creates a new vote
func (r *VotingRepository) CreateVote(ctx context.Context, vote *models.Vote) error {
	query := `
		INSERT INTO votes (id, poll_id, user_id, proposal_id, ticket_type, amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	if vote.ID == uuid.Nil {
		vote.ID = uuid.New()
	}
	if vote.CreatedAt.IsZero() {
		vote.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		vote.ID, vote.PollID, vote.UserID, vote.ProposalID,
		vote.TicketType, vote.Amount, vote.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create vote: %w", err)
	}
	
	return nil
}

// GetUserVotesForProposal returns user's votes for a proposal
func (r *VotingRepository) GetUserVotesForProposal(ctx context.Context, userID, proposalID uuid.UUID) (int, error) {
	var total int
	query := `SELECT COALESCE(SUM(amount), 0) FROM votes WHERE user_id = $1 AND proposal_id = $2`
	
	err := r.db.GetContext(ctx, &total, query, userID, proposalID)
	if err != nil {
		return 0, fmt.Errorf("get user votes: %w", err)
	}
	
	return total, nil
}

// ============================================
// POLLING
// ============================================

// GetActivePoll returns the current active voting poll
func (r *VotingRepository) GetActivePoll(ctx context.Context) (*models.VotingPoll, error) {
	query := `
		SELECT id, status, starts_at, ends_at, created_at
		FROM voting_polls
		WHERE status = 'active' AND ends_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	var poll models.VotingPoll
	err := r.db.GetContext(ctx, &poll, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active poll: %w", err)
	}
	
	return &poll, nil
}

// CreatePoll creates a new voting poll
func (r *VotingRepository) CreatePoll(ctx context.Context, poll *models.VotingPoll) error {
	query := `
		INSERT INTO voting_polls (id, status, starts_at, ends_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	if poll.ID == uuid.Nil {
		poll.ID = uuid.New()
	}
	if poll.CreatedAt.IsZero() {
		poll.CreatedAt = time.Now()
	}
	
	_, err := r.db.ExecContext(ctx, query,
		poll.ID, poll.Status, poll.StartsAt, poll.EndsAt, poll.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create poll: %w", err)
	}
	
	return nil
}

// ClosePoll closes a voting poll
func (r *VotingRepository) ClosePoll(ctx context.Context, pollID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE voting_polls SET status = 'closed' WHERE id = $1", pollID)
	return err
}

// GetVotingLeaderboard returns the current voting leaderboard
func (r *VotingRepository) GetVotingLeaderboard(ctx context.Context, limit int) ([]models.NovelProposal, error) {
	proposals := []models.NovelProposal{}
	
	query := `
		SELECT 
			np.id, np.user_id, np.original_link, np.status,
			np.title, np.alt_titles, np.author, np.description, np.cover_url,
			np.genres, np.tags, np.vote_score, np.votes_count, np.translation_tickets_invested,
			np.created_at, np.updated_at,
			COALESCE(up.display_name, u.email) as user_display_name,
			up.avatar_key as user_avatar,
			COALESCE(ux.level, 1) as user_level
		FROM novel_proposals np
		LEFT JOIN users u ON np.user_id = u.id
		LEFT JOIN user_profiles up ON np.user_id = up.user_id
		LEFT JOIN user_xp ux ON np.user_id = ux.user_id
		WHERE np.status = 'voting'
		ORDER BY np.vote_score DESC, np.created_at ASC
		LIMIT $1
	`
	
	rows, err := r.db.QueryxContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get voting leaderboard: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var proposal models.NovelProposal
		var altTitles, genres, tags pq.StringArray
		var userDisplayName string
		var userAvatarURL *string
		var userLevel int
		
		err := rows.Scan(
			&proposal.ID, &proposal.UserID, &proposal.OriginalLink, &proposal.Status,
			&proposal.Title, &altTitles, &proposal.Author, &proposal.Description, &proposal.CoverURL,
			&genres, &tags, &proposal.VoteScore, &proposal.VotesCount, &proposal.TranslationTicketsInvested,
			&proposal.CreatedAt, &proposal.UpdatedAt,
			&userDisplayName, &userAvatarURL, &userLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("scan proposal: %w", err)
		}
		
		proposal.AltTitles = altTitles
		proposal.Genres = genres
		proposal.Tags = tags
		proposal.User = &models.ProposalUser{
			ID:          proposal.UserID,
			DisplayName: userDisplayName,
			AvatarURL:   userAvatarURL,
			Level:       userLevel,
		}
		
		proposals = append(proposals, proposal)
	}
	
	return proposals, nil
}

// GetTopProposal returns the proposal with highest votes
func (r *VotingRepository) GetTopProposal(ctx context.Context) (*models.NovelProposal, error) {
	proposals, err := r.GetVotingLeaderboard(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(proposals) == 0 {
		return nil, nil
	}
	return &proposals[0], nil
}

// GetVotingStats returns overall voting statistics
func (r *VotingRepository) GetVotingStats(ctx context.Context) (*models.VotingStats, error) {
	stats := &models.VotingStats{}
	
	// Total proposals
	r.db.GetContext(ctx, &stats.TotalProposals, "SELECT COUNT(*) FROM novel_proposals")
	
	// Active proposals (in voting)
	r.db.GetContext(ctx, &stats.ActiveProposals, "SELECT COUNT(*) FROM novel_proposals WHERE status = 'voting'")
	
	// Total votes cast
	r.db.GetContext(ctx, &stats.TotalVotesCast, "SELECT COALESCE(SUM(amount), 0) FROM votes")
	
	// Translated proposals
	r.db.GetContext(ctx, &stats.ProposalsTranslated, "SELECT COUNT(*) FROM novel_proposals WHERE status IN ('accepted', 'translating')")
	
	return stats, nil
}

// SetProposalNovelID links a proposal to a released novel.
func (r *VotingRepository) SetProposalNovelID(ctx context.Context, proposalID uuid.UUID, novelID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE novel_proposals SET novel_id = $2, updated_at = NOW() WHERE id = $1`, proposalID, novelID)
	if err != nil {
		return fmt.Errorf("set proposal novel_id: %w", err)
	}
	return nil
}

// BeginTx starts a new database transaction
func (r *VotingRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
