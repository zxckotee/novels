package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/service"
	"github.com/rs/zerolog"
)

// DailyVoteGrantJob handles daily vote granting to all users
type DailyVoteGrantJob struct {
	db            *sqlx.DB
	ticketService *service.TicketService
	logger        zerolog.Logger
}

func NewDailyVoteGrantJob(
	db *sqlx.DB,
	ticketService *service.TicketService,
	logger zerolog.Logger,
) *DailyVoteGrantJob {
	return &DailyVoteGrantJob{
		db:            db,
		ticketService: ticketService,
		logger:        logger.With().Str("job", "daily_vote_grant").Logger(),
	}
}

// Run executes the daily vote grant job
// Should be triggered at 00:00 UTC (3:00 MSK)
func (j *DailyVoteGrantJob) Run(ctx context.Context) error {
	j.logger.Info().Msg("Starting daily vote grant job")
	
	startTime := time.Now()
	today := time.Now().UTC().Format("2006-01-02")
	
	// Create grant log entry
	grantLogID, err := j.createGrantLog(ctx, today)
	if err != nil {
		return fmt.Errorf("create grant log: %w", err)
	}
	
	// Get all active users
	usersProcessed := 0
	totalVotesGranted := 0
	
	// Process users in batches
	batchSize := 1000
	offset := 0
	
	for {
		users, err := j.getUserBatch(ctx, batchSize, offset)
		if err != nil {
			j.updateGrantLogError(ctx, grantLogID, err.Error())
			return fmt.Errorf("get user batch: %w", err)
		}
		
		if len(users) == 0 {
			break
		}
		
		for _, userID := range users {
			err := j.ticketService.GrantDailyVotes(ctx, userID)
			if err != nil {
				j.logger.Error().Err(err).
					Str("user_id", userID.String()).
					Msg("Failed to grant daily votes to user")
				continue
			}
			
			usersProcessed++
			totalVotesGranted += models.DefaultDailyVoteAmount
		}
		
		offset += batchSize
		
		// Small delay between batches to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}
	
	// Update grant log
	duration := time.Since(startTime)
	err = j.completeGrantLog(ctx, grantLogID, usersProcessed, totalVotesGranted)
	if err != nil {
		j.logger.Error().Err(err).Msg("Failed to complete grant log")
	}
	
	j.logger.Info().
		Int("users_processed", usersProcessed).
		Int("total_votes_granted", totalVotesGranted).
		Dur("duration", duration).
		Msg("Daily vote grant job completed")
	
	return nil
}

// getUserBatch returns a batch of active user IDs
func (j *DailyVoteGrantJob) getUserBatch(ctx context.Context, limit, offset int) ([]uuid.UUID, error) {
	query := `
		SELECT id FROM users 
		WHERE is_banned = false 
		ORDER BY id 
		LIMIT $1 OFFSET $2
	`
	
	rows, err := j.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		users = append(users, id)
	}
	
	return users, nil
}

// createGrantLog creates a log entry for the daily vote grant
func (j *DailyVoteGrantJob) createGrantLog(ctx context.Context, grantDate string) (interface{}, error) {
	query := `
		INSERT INTO daily_vote_grants (grant_date, status, started_at)
		VALUES ($1::date, 'running', NOW())
		ON CONFLICT (grant_date) DO UPDATE SET
			status = 'running',
			started_at = NOW(),
			users_processed = 0,
			total_votes_granted = 0,
			error_message = NULL
		RETURNING id
	`
	
	var id interface{}
	err := j.db.QueryRowContext(ctx, query, grantDate).Scan(&id)
	return id, err
}

// completeGrantLog marks the grant log as completed
func (j *DailyVoteGrantJob) completeGrantLog(ctx context.Context, id interface{}, usersProcessed, totalVotes int) error {
	query := `
		UPDATE daily_vote_grants
		SET status = 'completed',
			completed_at = NOW(),
			users_processed = $2,
			total_votes_granted = $3
		WHERE id = $1
	`
	
	_, err := j.db.ExecContext(ctx, query, id, usersProcessed, totalVotes)
	return err
}

// updateGrantLogError marks the grant log as failed
func (j *DailyVoteGrantJob) updateGrantLogError(ctx context.Context, id interface{}, errorMsg string) {
	query := `
		UPDATE daily_vote_grants
		SET status = 'failed',
			completed_at = NOW(),
			error_message = $2
		WHERE id = $1
	`
	
	j.db.ExecContext(ctx, query, id, errorMsg)
}

// GetLastGrantStatus returns the status of the last daily vote grant
func (j *DailyVoteGrantJob) GetLastGrantStatus(ctx context.Context) (*DailyGrantStatus, error) {
	query := `
		SELECT 
			grant_date, status, users_processed, total_votes_granted,
			started_at, completed_at, error_message
		FROM daily_vote_grants
		ORDER BY grant_date DESC
		LIMIT 1
	`
	
	var status DailyGrantStatus
	var completedAt sql.NullTime
	var errorMessage sql.NullString
	
	err := j.db.QueryRowContext(ctx, query).Scan(
		&status.GrantDate, &status.Status, &status.UsersProcessed,
		&status.TotalVotesGranted, &status.StartedAt, &completedAt, &errorMessage,
	)
	if err != nil {
		return nil, err
	}
	
	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
	}
	if errorMessage.Valid {
		status.ErrorMessage = &errorMessage.String
	}
	
	return &status, nil
}

// DailyGrantStatus represents the status of a daily vote grant
type DailyGrantStatus struct {
	GrantDate         string     `json:"grantDate"`
	Status            string     `json:"status"`
	UsersProcessed    int        `json:"usersProcessed"`
	TotalVotesGranted int        `json:"totalVotesGranted"`
	StartedAt         time.Time  `json:"startedAt"`
	CompletedAt       *time.Time `json:"completedAt,omitempty"`
	ErrorMessage      *string    `json:"errorMessage,omitempty"`
}
