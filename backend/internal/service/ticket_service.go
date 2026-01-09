package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/novels/backend/internal/domain/models"
	"github.com/novels/backend/internal/repository"
	"github.com/rs/zerolog"
)

type TicketService struct {
	ticketRepo *repository.TicketRepository
	subRepo    *repository.SubscriptionRepository
	logger     zerolog.Logger
}

func NewTicketService(
	ticketRepo *repository.TicketRepository,
	subRepo *repository.SubscriptionRepository,
	logger zerolog.Logger,
) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
		subRepo:    subRepo,
		logger:     logger,
	}
}

// GetWallet returns the wallet info for a user
func (s *TicketService) GetWallet(ctx context.Context, userID uuid.UUID) (*models.WalletInfo, error) {
	return s.ticketRepo.GetWalletInfo(ctx, userID)
}

// GetBalance returns the balance for a specific ticket type
func (s *TicketService) GetBalance(ctx context.Context, userID uuid.UUID, ticketType models.TicketType) (int, error) {
	return s.ticketRepo.GetBalance(ctx, userID, ticketType)
}

// SpendTickets spends tickets for a specific purpose
func (s *TicketService) SpendTickets(ctx context.Context, userID uuid.UUID, req models.SpendTicketRequest) error {
	var refID *uuid.UUID
	if req.RefID != "" {
		parsed, err := uuid.Parse(req.RefID)
		if err == nil {
			refID = &parsed
		}
	}
	
	reason := string(req.Type) + "_spent"
	if req.RefType == "vote" {
		reason = models.ReasonVoteCast
	} else if req.RefType == "proposal" {
		reason = models.ReasonProposalCreated
	} else if req.RefType == "translation" {
		reason = models.ReasonTranslationRequest
	}
	
	err := s.ticketRepo.SpendTickets(ctx, userID, req.Type, req.Amount, reason, req.RefType, refID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", userID.String()).
			Str("ticket_type", string(req.Type)).
			Int("amount", req.Amount).
			Msg("Failed to spend tickets")
		return err
	}
	
	s.logger.Info().
		Str("user_id", userID.String()).
		Str("ticket_type", string(req.Type)).
		Int("amount", req.Amount).
		Msg("Tickets spent")
	
	return nil
}

// GrantTickets grants tickets to a user (for admin/system use)
func (s *TicketService) GrantTickets(ctx context.Context, req models.GrantTicketRequest) error {
	idempotencyKey := fmt.Sprintf("grant:%s:%s:%s:%d",
		req.UserID.String(), req.Type, time.Now().Format("2006-01-02"), req.Amount)
	
	err := s.ticketRepo.GrantTickets(ctx, req.UserID, req.Type, req.Amount, req.Reason, "admin", nil, idempotencyKey)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", req.UserID.String()).
			Str("ticket_type", string(req.Type)).
			Int("amount", req.Amount).
			Msg("Failed to grant tickets")
		return err
	}
	
	s.logger.Info().
		Str("user_id", req.UserID.String()).
		Str("ticket_type", string(req.Type)).
		Int("amount", req.Amount).
		Str("reason", req.Reason).
		Msg("Tickets granted")
	
	return nil
}

// GrantDailyVotes grants daily votes to a user
func (s *TicketService) GrantDailyVotes(ctx context.Context, userID uuid.UUID) error {
	// Get multiplier from subscription
	multiplier, err := s.subRepo.GetDailyVoteMultiplier(ctx, userID)
	if err != nil {
		multiplier = 1
	}
	
	amount := models.DefaultDailyVoteAmount * multiplier
	
	// Create idempotency key based on date
	today := time.Now().UTC().Format("2006-01-02")
	idempotencyKey := fmt.Sprintf("daily_vote:%s:%s", today, userID.String())
	
	// Check if already granted today
	exists, err := s.ticketRepo.CheckIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return fmt.Errorf("check idempotency: %w", err)
	}
	if exists {
		s.logger.Debug().
			Str("user_id", userID.String()).
			Msg("Daily votes already granted today")
		return nil
	}
	
	// Set daily votes (replaces old balance)
	err = s.ticketRepo.SetDailyVotes(ctx, userID, amount, idempotencyKey)
	if err != nil {
		s.logger.Error().Err(err).
			Str("user_id", userID.String()).
			Int("amount", amount).
			Msg("Failed to grant daily votes")
		return err
	}
	
	s.logger.Info().
		Str("user_id", userID.String()).
		Int("amount", amount).
		Int("multiplier", multiplier).
		Msg("Daily votes granted")
	
	return nil
}

// GetTransactions returns paginated transactions for a user
func (s *TicketService) GetTransactions(ctx context.Context, userID uuid.UUID, ticketType *models.TicketType, page, limit int) (*models.TransactionsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	filter := models.TransactionFilter{
		UserID: &userID,
		Type:   ticketType,
		Page:   page,
		Limit:  limit,
	}
	
	transactions, total, err := s.ticketRepo.GetTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return &models.TransactionsResponse{
		Transactions: transactions,
		TotalCount:   total,
		Page:         page,
		Limit:        limit,
	}, nil
}

// GetLeaderboard returns the leaderboard of top spenders
func (s *TicketService) GetLeaderboard(ctx context.Context, period string, limit int) ([]LeaderboardEntry, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	results, err := s.ticketRepo.GetLeaderboard(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	
	entries := make([]LeaderboardEntry, len(results))
	for i, r := range results {
		entries[i] = LeaderboardEntry{
			Rank:         i + 1,
			UserID:       r.UserID,
			DisplayName:  r.DisplayName,
			AvatarURL:    r.AvatarURL,
			Level:        r.Level,
			TicketsSpent: r.TicketsSpent,
		}
	}
	
	return entries, nil
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	UserID       uuid.UUID `json:"userId"`
	DisplayName  string    `json:"displayName"`
	AvatarURL    *string   `json:"avatarUrl,omitempty"`
	Level        int       `json:"level"`
	TicketsSpent int       `json:"ticketsSpent"`
}

// GetUserStats returns ticket statistics for a user
func (s *TicketService) GetUserStats(ctx context.Context, userID uuid.UUID) (*UserTicketStats, error) {
	stats := &UserTicketStats{}
	
	// Get wallet
	wallet, err := s.ticketRepo.GetWalletInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats.Wallet = wallet
	
	// Get total spent
	now := time.Now().UTC()
	
	stats.SpentToday, _ = s.ticketRepo.GetTotalSpentInPeriod(ctx, userID, nil, now.Add(-24*time.Hour))
	stats.SpentThisWeek, _ = s.ticketRepo.GetTotalSpentInPeriod(ctx, userID, nil, now.Add(-7*24*time.Hour))
	stats.SpentThisMonth, _ = s.ticketRepo.GetTotalSpentInPeriod(ctx, userID, nil, now.Add(-30*24*time.Hour))
	
	return stats, nil
}

// UserTicketStats represents user's ticket statistics
type UserTicketStats struct {
	Wallet         *models.WalletInfo `json:"wallet"`
	SpentToday     int                `json:"spentToday"`
	SpentThisWeek  int                `json:"spentThisWeek"`
	SpentThisMonth int                `json:"spentThisMonth"`
}
