package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"
	"github.com/rs/zerolog"
)

// WeeklyTicketGrantJob grants weekly tickets:
// - Subscription-based: Premium/VIP get novel_request + translation_ticket weekly
// - Level reward: users with level > 10 get +1 novel_request weekly
//
// Scheduled for Wednesday 00:00 UTC (03:00 MSK).
type WeeklyTicketGrantJob struct {
	db      *sqlx.DB
	subRepo *repository.SubscriptionRepository
	tickets *repository.TicketRepository
	logger  zerolog.Logger
}

func NewWeeklyTicketGrantJob(
	db *sqlx.DB,
	subRepo *repository.SubscriptionRepository,
	ticketRepo *repository.TicketRepository,
	logger zerolog.Logger,
) *WeeklyTicketGrantJob {
	return &WeeklyTicketGrantJob{
		db:      db,
		subRepo: subRepo,
		tickets: ticketRepo,
		logger:  logger.With().Str("job", "weekly_ticket_grant").Logger(),
	}
}

// Run executes the weekly ticket grant.
func (j *WeeklyTicketGrantJob) Run(ctx context.Context) error {
	j.logger.Info().Msg("Starting weekly ticket grant job")

	startTime := time.Now()
	grantDate := time.Now().UTC().Format("2006-01-02") // Wednesday date (UTC)

	grantLogID, err := j.createGrantLog(ctx, grantDate)
	if err != nil {
		return fmt.Errorf("create grant log: %w", err)
	}

	usersProcessed := 0
	totalNovelRequests := 0
	totalTranslationTickets := 0

	batchSize := 1000
	offset := 0

	for {
		rows, err := j.getUserBatch(ctx, batchSize, offset)
		if err != nil {
			j.updateGrantLogError(ctx, grantLogID, err.Error())
			return fmt.Errorf("get user batch: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			nrSub, ttSub := subscriptionWeeklyAmounts(row.PlanCode)
			nrLevel := 0
			if row.Level > 10 {
				nrLevel = 1
			}

			// Subscription-based grants (idempotent per user+date+type)
			if nrSub > 0 {
				key := fmt.Sprintf("weekly_sub:%s:%s:novel_request", grantDate, row.UserID.String())
				if err := j.tickets.GrantTickets(ctx, row.UserID, models.TicketTypeNovelRequest, nrSub,
					models.ReasonSubscriptionGrant, "weekly_grant", nil, key); err != nil {
					j.logger.Error().Err(err).Str("user_id", row.UserID.String()).Msg("Failed to grant weekly novel request tickets (subscription)")
				} else {
					totalNovelRequests += nrSub
				}
			}
			if ttSub > 0 {
				key := fmt.Sprintf("weekly_sub:%s:%s:translation_ticket", grantDate, row.UserID.String())
				if err := j.tickets.GrantTickets(ctx, row.UserID, models.TicketTypeTranslationTicket, ttSub,
					models.ReasonSubscriptionGrant, "weekly_grant", nil, key); err != nil {
					j.logger.Error().Err(err).Str("user_id", row.UserID.String()).Msg("Failed to grant weekly translation tickets (subscription)")
				} else {
					totalTranslationTickets += ttSub
				}
			}

			// Level-based bonus (also idempotent)
			if nrLevel > 0 {
				key := fmt.Sprintf("weekly_level:%s:%s:novel_request", grantDate, row.UserID.String())
				if err := j.tickets.GrantTickets(ctx, row.UserID, models.TicketTypeNovelRequest, nrLevel,
					models.ReasonLevelReward, "weekly_grant", nil, key); err != nil {
					j.logger.Error().Err(err).Str("user_id", row.UserID.String()).Msg("Failed to grant weekly novel request ticket (level reward)")
				} else {
					totalNovelRequests += nrLevel
				}
			}

			usersProcessed++
		}

		offset += batchSize
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)
	if err := j.completeGrantLog(ctx, grantLogID, usersProcessed, totalNovelRequests, totalTranslationTickets); err != nil {
		j.logger.Error().Err(err).Msg("Failed to complete weekly grant log")
	}

	j.logger.Info().
		Int("users_processed", usersProcessed).
		Int("novel_requests_granted", totalNovelRequests).
		Int("translation_tickets_granted", totalTranslationTickets).
		Dur("duration", duration).
		Msg("Weekly ticket grant job completed")

	return nil
}

type weeklyUserRow struct {
	UserID   uuid.UUID
	Level    int
	PlanCode *string
}

func (j *WeeklyTicketGrantJob) getUserBatch(ctx context.Context, limit, offset int) ([]weeklyUserRow, error) {
	// Use active_subscriptions view for plan_code; include level from user_xp.
	query := `
		SELECT
			u.id,
			COALESCE(ux.level, 1) as level,
			asub.plan_code
		FROM users u
		LEFT JOIN user_xp ux ON ux.user_id = u.id
		LEFT JOIN active_subscriptions asub ON asub.user_id = u.id
		WHERE u.is_banned = false
		ORDER BY u.id
		LIMIT $1 OFFSET $2
	`

	rows, err := j.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []weeklyUserRow
	for rows.Next() {
		var r weeklyUserRow
		if err := rows.Scan(&r.UserID, &r.Level, &r.PlanCode); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func subscriptionWeeklyAmounts(planCode *string) (novelRequests int, translationTickets int) {
	if planCode == nil {
		return 0, 0
	}
	switch *planCode {
	case "premium":
		return 2, 5
	case "vip":
		return 5, 15
	default:
		return 0, 0
	}
}

// ===== logging table: weekly_ticket_grants =====

func (j *WeeklyTicketGrantJob) createGrantLog(ctx context.Context, grantDate string) (uuid.UUID, error) {
	query := `
		INSERT INTO weekly_ticket_grants (grant_date, status, started_at)
		VALUES ($1::date, 'running', NOW())
		ON CONFLICT (grant_date) DO UPDATE SET
			status = 'running',
			started_at = NOW(),
			users_processed = 0,
			novel_requests_granted = 0,
			translation_tickets_granted = 0,
			error_message = NULL
		RETURNING id
	`

	var id uuid.UUID
	err := j.db.QueryRowContext(ctx, query, grantDate).Scan(&id)
	return id, err
}

func (j *WeeklyTicketGrantJob) completeGrantLog(ctx context.Context, id uuid.UUID, usersProcessed, novelRequests, translationTickets int) error {
	query := `
		UPDATE weekly_ticket_grants
		SET status = 'completed',
			completed_at = NOW(),
			users_processed = $2,
			novel_requests_granted = $3,
			translation_tickets_granted = $4
		WHERE id = $1
	`
	_, err := j.db.ExecContext(ctx, query, id, usersProcessed, novelRequests, translationTickets)
	return err
}

func (j *WeeklyTicketGrantJob) updateGrantLogError(ctx context.Context, id uuid.UUID, errorMsg string) {
	query := `
		UPDATE weekly_ticket_grants
		SET status = 'failed',
			completed_at = NOW(),
			error_message = $2
		WHERE id = $1
	`
	j.db.ExecContext(ctx, query, id, errorMsg)
}

// WeeklyGrantStatus represents the status of a weekly ticket grant.
type WeeklyGrantStatus struct {
	GrantDate               string     `json:"grantDate"`
	Status                  string     `json:"status"`
	UsersProcessed          int        `json:"usersProcessed"`
	NovelRequestsGranted    int        `json:"novelRequestsGranted"`
	TranslationTicketsGranted int      `json:"translationTicketsGranted"`
	StartedAt               time.Time  `json:"startedAt"`
	CompletedAt             *time.Time `json:"completedAt,omitempty"`
	ErrorMessage            *string    `json:"errorMessage,omitempty"`
}

func (j *WeeklyTicketGrantJob) GetLastGrantStatus(ctx context.Context) (*WeeklyGrantStatus, error) {
	query := `
		SELECT
			grant_date, status, users_processed, novel_requests_granted, translation_tickets_granted,
			started_at, completed_at, error_message
		FROM weekly_ticket_grants
		ORDER BY grant_date DESC
		LIMIT 1
	`

	var status WeeklyGrantStatus
	var completedAt sql.NullTime
	var errorMessage sql.NullString

	err := j.db.QueryRowContext(ctx, query).Scan(
		&status.GrantDate, &status.Status, &status.UsersProcessed,
		&status.NovelRequestsGranted, &status.TranslationTicketsGranted,
		&status.StartedAt, &completedAt, &errorMessage,
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

