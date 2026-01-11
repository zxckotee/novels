package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"novels-backend/internal/service"
	"github.com/rs/zerolog"
)

// Scheduler manages background jobs
type Scheduler struct {
	db                *sqlx.DB
	ticketService     *service.TicketService
	votingService     *service.VotingService
	subscriptionService *service.SubscriptionService
	logger            zerolog.Logger
	
	dailyVoteJob      *DailyVoteGrantJob
	
	stopCh            chan struct{}
	wg                sync.WaitGroup
}

func NewScheduler(
	db *sqlx.DB,
	ticketService *service.TicketService,
	votingService *service.VotingService,
	subscriptionService *service.SubscriptionService,
	logger zerolog.Logger,
) *Scheduler {
	return &Scheduler{
		db:                  db,
		ticketService:       ticketService,
		votingService:       votingService,
		subscriptionService: subscriptionService,
		logger:              logger.With().Str("component", "scheduler").Logger(),
		stopCh:              make(chan struct{}),
	}
}

// Start starts all scheduled jobs
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info().Msg("Starting job scheduler")
	
	// Initialize jobs
	s.dailyVoteJob = NewDailyVoteGrantJob(s.db, s.ticketService, s.logger)
	
	// Start job runners
	s.wg.Add(4)
	go s.runDailyVoteJob(ctx)
	go s.runVotingWinnerJob(ctx)
	go s.runSubscriptionExpiryJob(ctx)
	go s.runCleanupJob(ctx)
}

// Stop stops all scheduled jobs
func (s *Scheduler) Stop() {
	s.logger.Info().Msg("Stopping job scheduler")
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Info().Msg("Job scheduler stopped")
}

// runDailyVoteJob runs daily at 00:00 UTC (3:00 MSK)
func (s *Scheduler) runDailyVoteJob(ctx context.Context) {
	defer s.wg.Done()
	
	// Calculate time until next 00:00 UTC
	nextRun := s.getNextUTCMidnight()
	timer := time.NewTimer(time.Until(nextRun))
	
	s.logger.Info().
		Time("next_run", nextRun).
		Msg("Daily vote job scheduled")
	
	for {
		select {
		case <-s.stopCh:
			timer.Stop()
			return
		case <-timer.C:
			s.logger.Info().Msg("Running daily vote grant job")
			
			if err := s.dailyVoteJob.Run(ctx); err != nil {
				s.logger.Error().Err(err).Msg("Daily vote grant job failed")
			}
			
			// Schedule next run
			nextRun = s.getNextUTCMidnight()
			timer.Reset(time.Until(nextRun))
			
			s.logger.Info().
				Time("next_run", nextRun).
				Msg("Daily vote job rescheduled")
		}
	}
}

// runVotingWinnerJob runs every 6 hours to pick voting winners
func (s *Scheduler) runVotingWinnerJob(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	
	s.logger.Info().Msg("Voting winner job started (every 6 hours)")
	
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.logger.Info().Msg("Running voting winner selection job")
			
			if err := s.votingService.ProcessVotingWinner(ctx); err != nil {
				s.logger.Error().Err(err).Msg("Voting winner job failed")
			}
		}
	}
}

// runSubscriptionExpiryJob runs every hour to expire subscriptions
func (s *Scheduler) runSubscriptionExpiryJob(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	s.logger.Info().Msg("Subscription expiry job started (every hour)")
	
	// Run immediately on start
	s.subscriptionService.ExpireSubscriptions(ctx)
	
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.logger.Debug().Msg("Running subscription expiry job")
			
			if err := s.subscriptionService.ExpireSubscriptions(ctx); err != nil {
				s.logger.Error().Err(err).Msg("Subscription expiry job failed")
			}
		}
	}
}

// runCleanupJob runs daily to clean up old data
func (s *Scheduler) runCleanupJob(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	s.logger.Info().Msg("Cleanup job started (every 24 hours)")
	
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.logger.Info().Msg("Running cleanup job")
			s.runCleanupTasks(ctx)
		}
	}
}

// runCleanupTasks performs various cleanup tasks
func (s *Scheduler) runCleanupTasks(ctx context.Context) {
	// Clean up old leaderboard cache
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM leaderboard_cache WHERE calculated_at < NOW() - INTERVAL '1 day'`)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to clean leaderboard cache")
	}
	
	// Clean up old daily vote grant logs (keep last 30 days)
	_, err = s.db.ExecContext(ctx,
		`DELETE FROM daily_vote_grants WHERE grant_date < CURRENT_DATE - INTERVAL '30 days'`)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to clean daily vote grant logs")
	}
	
	s.logger.Info().Msg("Cleanup tasks completed")
}

// getNextUTCMidnight returns the next 00:00 UTC
func (s *Scheduler) getNextUTCMidnight() time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	
	// If we've passed today's midnight, schedule for tomorrow
	if now.After(next) {
		next = next.Add(24 * time.Hour)
	}
	
	return next
}

// RunDailyVoteJobNow runs the daily vote job immediately (for admin/testing)
func (s *Scheduler) RunDailyVoteJobNow(ctx context.Context) error {
	return s.dailyVoteJob.Run(ctx)
}

// RunVotingWinnerJobNow runs the voting winner job immediately (for admin/testing)
func (s *Scheduler) RunVotingWinnerJobNow(ctx context.Context) error {
	return s.votingService.ProcessVotingWinner(ctx)
}

// GetDailyGrantStatus returns the status of the last daily vote grant
func (s *Scheduler) GetDailyGrantStatus(ctx context.Context) (*DailyGrantStatus, error) {
	return s.dailyVoteJob.GetLastGrantStatus(ctx)
}
