package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"
	"github.com/rs/zerolog"
)

var (
	ErrInsufficientTickets = errors.New("insufficient tickets")
	ErrProposalNotFound    = errors.New("proposal not found")
	ErrProposalNotVoting   = errors.New("proposal is not in voting status")
	ErrCannotVoteOwnProposal = errors.New("cannot vote for own proposal")
)

type VotingService struct {
	votingRepo *repository.VotingRepository
	ticketRepo *repository.TicketRepository
	logger     zerolog.Logger
}

func NewVotingService(
	votingRepo *repository.VotingRepository,
	ticketRepo *repository.TicketRepository,
	logger zerolog.Logger,
) *VotingService {
	return &VotingService{
		votingRepo: votingRepo,
		ticketRepo: ticketRepo,
		logger:     logger,
	}
}

// ============================================
// PROPOSALS
// ============================================

// CreateProposal creates a new novel proposal
func (s *VotingService) CreateProposal(ctx context.Context, userID uuid.UUID, req models.CreateProposalRequest) (*models.NovelProposal, error) {
	// Check if user has novel request ticket
	balance, err := s.ticketRepo.GetBalance(ctx, userID, models.TicketTypeNovelRequest)
	if err != nil {
		return nil, fmt.Errorf("check balance: %w", err)
	}
	if balance < 1 {
		return nil, ErrInsufficientTickets
	}
	
	// Create proposal
	proposal := &models.NovelProposal{
		ID:           uuid.New(),
		UserID:       userID,
		OriginalLink: req.OriginalLink,
		// UI submits a completed proposal; it should go straight to moderation.
		Status:       models.ProposalStatusModeration,
		Title:        req.Title,
		AltTitles:    req.AltTitles,
		Author:       req.Author,
		Description:  req.Description,
		CoverURL:     req.CoverURL,
		Genres:       req.Genres,
		Tags:         req.Tags,
	}
	
	// Start transaction
	tx, err := s.votingRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	
	// Create proposal
	err = s.votingRepo.CreateProposal(ctx, proposal)
	if err != nil {
		return nil, fmt.Errorf("create proposal: %w", err)
	}
	
	// Spend novel request ticket
	err = s.ticketRepo.SpendTickets(ctx, userID, models.TicketTypeNovelRequest, 1,
		models.ReasonProposalCreated, "proposal", &proposal.ID)
	if err != nil {
		return nil, fmt.Errorf("spend ticket: %w", err)
	}
	
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	
	s.logger.Info().
		Str("proposal_id", proposal.ID.String()).
		Str("user_id", userID.String()).
		Str("title", proposal.Title).
		Msg("Proposal created")
	
	return proposal, nil
}

// GetProposal returns a proposal by ID
func (s *VotingService) GetProposal(ctx context.Context, id uuid.UUID, currentUserID *uuid.UUID) (*models.NovelProposal, error) {
	return s.votingRepo.GetProposalWithUser(ctx, id, currentUserID)
}

// ListProposals returns proposals with filters
func (s *VotingService) ListProposals(ctx context.Context, filter models.ProposalFilter) (*models.ProposalsResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		filter.Limit = 20
	}
	
	proposals, total, err := s.votingRepo.ListProposals(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return &models.ProposalsResponse{
		Proposals:  proposals,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

// UpdateProposal updates a proposal (only owner can update, only in draft status)
func (s *VotingService) UpdateProposal(ctx context.Context, id, userID uuid.UUID, req models.UpdateProposalRequest) (*models.NovelProposal, error) {
	proposal, err := s.votingRepo.GetProposalByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if proposal == nil {
		return nil, ErrProposalNotFound
	}
	
	if proposal.UserID != userID {
		return nil, errors.New("not authorized to update this proposal")
	}
	
	if proposal.Status != models.ProposalStatusDraft {
		return nil, errors.New("can only update proposals in draft status")
	}
	
	// Apply updates
	if req.OriginalLink != nil {
		proposal.OriginalLink = *req.OriginalLink
	}
	if req.Title != nil {
		proposal.Title = *req.Title
	}
	if len(req.AltTitles) > 0 {
		proposal.AltTitles = req.AltTitles
	}
	if req.Author != nil {
		proposal.Author = *req.Author
	}
	if req.Description != nil {
		proposal.Description = *req.Description
	}
	if req.CoverURL != nil {
		proposal.CoverURL = req.CoverURL
	}
	if len(req.Genres) > 0 {
		proposal.Genres = req.Genres
	}
	if len(req.Tags) > 0 {
		proposal.Tags = req.Tags
	}
	
	err = s.votingRepo.UpdateProposal(ctx, proposal)
	if err != nil {
		return nil, fmt.Errorf("update proposal: %w", err)
	}
	
	return proposal, nil
}

// SubmitProposal submits a proposal for moderation
func (s *VotingService) SubmitProposal(ctx context.Context, id, userID uuid.UUID) error {
	proposal, err := s.votingRepo.GetProposalByID(ctx, id)
	if err != nil {
		return err
	}
	if proposal == nil {
		return ErrProposalNotFound
	}
	
	if proposal.UserID != userID {
		return errors.New("not authorized")
	}
	
	if proposal.Status != models.ProposalStatusDraft {
		return errors.New("can only submit proposals in draft status")
	}
	
	return s.votingRepo.SubmitProposalForModeration(ctx, id)
}

// ModerateProposal approves or rejects a proposal (moderator only)
func (s *VotingService) ModerateProposal(ctx context.Context, id, moderatorID uuid.UUID, req models.ModerateProposalRequest) error {
	proposal, err := s.votingRepo.GetProposalByID(ctx, id)
	if err != nil {
		return err
	}
	if proposal == nil {
		return ErrProposalNotFound
	}
	
	if proposal.Status != models.ProposalStatusModeration {
		return errors.New("proposal is not in moderation status")
	}
	
	var newStatus models.ProposalStatus
	switch req.Action {
	case "approve":
		newStatus = models.ProposalStatusVoting
	case "reject":
		newStatus = models.ProposalStatusRejected
	default:
		return errors.New("invalid action")
	}
	
	err = s.votingRepo.UpdateProposalStatus(ctx, id, newStatus, &moderatorID, req.RejectReason)
	if err != nil {
		return err
	}
	
	s.logger.Info().
		Str("proposal_id", id.String()).
		Str("moderator_id", moderatorID.String()).
		Str("action", req.Action).
		Msg("Proposal moderated")
	
	return nil
}

// DeleteProposal deletes a proposal (only owner, only in draft status)
func (s *VotingService) DeleteProposal(ctx context.Context, id, userID uuid.UUID) error {
	proposal, err := s.votingRepo.GetProposalByID(ctx, id)
	if err != nil {
		return err
	}
	if proposal == nil {
		return ErrProposalNotFound
	}
	
	if proposal.UserID != userID {
		return errors.New("not authorized")
	}
	
	if proposal.Status != models.ProposalStatusDraft {
		return errors.New("can only delete proposals in draft status")
	}
	
	return s.votingRepo.DeleteProposal(ctx, id)
}

// GetMyProposals returns proposals created by the user
func (s *VotingService) GetMyProposals(ctx context.Context, userID uuid.UUID, page, limit int) (*models.ProposalsResponse, error) {
	filter := models.ProposalFilter{
		UserID: &userID,
		Sort:   "newest",
		Page:   page,
		Limit:  limit,
	}
	return s.ListProposals(ctx, filter)
}

// ============================================
// VOTING
// ============================================

// CastVote casts a vote for a proposal
func (s *VotingService) CastVote(ctx context.Context, userID uuid.UUID, req models.CastVoteRequest) error {
	proposalID, err := uuid.Parse(req.ProposalID)
	if err != nil {
		return errors.New("invalid proposal ID")
	}
	
	// Get proposal
	proposal, err := s.votingRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("get proposal: %w", err)
	}
	if proposal == nil {
		return ErrProposalNotFound
	}
	
	// Check status
	if proposal.Status != models.ProposalStatusVoting {
		return ErrProposalNotVoting
	}
	
	// Cannot vote for own proposal
	if proposal.UserID == userID {
		return ErrCannotVoteOwnProposal
	}
	
	// Check balance
	balance, err := s.ticketRepo.GetBalance(ctx, userID, req.TicketType)
	if err != nil {
		return fmt.Errorf("check balance: %w", err)
	}
	if balance < req.Amount {
		return ErrInsufficientTickets
	}
	
	// Get or create active poll
	poll, err := s.votingRepo.GetActivePoll(ctx)
	if err != nil {
		return fmt.Errorf("get active poll: %w", err)
	}
	
	if poll == nil {
		// Create new poll (24 hour cycle)
		poll = &models.VotingPoll{
			ID:       uuid.New(),
			Status:   "active",
			StartsAt: time.Now().UTC(),
			EndsAt:   time.Now().UTC().Add(24 * time.Hour),
		}
		err = s.votingRepo.CreatePoll(ctx, poll)
		if err != nil {
			return fmt.Errorf("create poll: %w", err)
		}
	}
	
	// Create vote
	vote := &models.Vote{
		ID:         uuid.New(),
		PollID:     poll.ID,
		UserID:     userID,
		ProposalID: proposalID,
		TicketType: req.TicketType,
		Amount:     req.Amount,
	}
	
	err = s.votingRepo.CreateVote(ctx, vote)
	if err != nil {
		return fmt.Errorf("create vote: %w", err)
	}
	
	// Spend tickets
	err = s.ticketRepo.SpendTickets(ctx, userID, req.TicketType, req.Amount,
		models.ReasonVoteCast, "vote", &vote.ID)
	if err != nil {
		return fmt.Errorf("spend tickets: %w", err)
	}
	
	s.logger.Info().
		Str("user_id", userID.String()).
		Str("proposal_id", proposalID.String()).
		Str("ticket_type", string(req.TicketType)).
		Int("amount", req.Amount).
		Msg("Vote cast")
	
	return nil
}

// GetVotingLeaderboard returns the current voting leaderboard
func (s *VotingService) GetVotingLeaderboard(ctx context.Context, limit int) (*models.VotingLeaderboard, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	
	poll, err := s.votingRepo.GetActivePoll(ctx)
	if err != nil {
		return nil, err
	}
	
	proposals, err := s.votingRepo.GetVotingLeaderboard(ctx, limit)
	if err != nil {
		return nil, err
	}
	
	entries := make([]models.VotingEntry, len(proposals))
	for i, p := range proposals {
		entries[i] = models.VotingEntry{
			NovelID:  p.ID,
			Score:    p.VoteScore,
			Proposal: &proposals[i],
		}
		if poll != nil {
			entries[i].PollID = poll.ID
		}
	}
	
	// Calculate next reset
	nextReset := time.Now().UTC().Add(24 * time.Hour)
	if poll != nil {
		nextReset = poll.EndsAt
	}
	
	return &models.VotingLeaderboard{
		Poll:      poll,
		Entries:   entries,
		NextReset: nextReset,
	}, nil
}

// GetVotingStats returns voting statistics
func (s *VotingService) GetVotingStats(ctx context.Context) (*models.VotingStats, error) {
	return s.votingRepo.GetVotingStats(ctx)
}

// ============================================
// CRON JOBS
// ============================================

// ProcessVotingWinner picks the winner and updates status
func (s *VotingService) ProcessVotingWinner(ctx context.Context) error {
	// Get top proposal
	topProposal, err := s.votingRepo.GetTopProposal(ctx)
	if err != nil {
		return fmt.Errorf("get top proposal: %w", err)
	}
	
	if topProposal == nil || topProposal.VoteScore < 1 {
		s.logger.Info().Msg("No proposals with votes to process")
		return nil
	}
	
	// Update status to translating
	err = s.votingRepo.UpdateProposalStatus(ctx, topProposal.ID, models.ProposalStatusTranslating, nil, nil)
	if err != nil {
		return fmt.Errorf("update proposal status: %w", err)
	}
	
	// Close current poll
	poll, err := s.votingRepo.GetActivePoll(ctx)
	if err == nil && poll != nil {
		s.votingRepo.ClosePoll(ctx, poll.ID)
	}
	
	s.logger.Info().
		Str("proposal_id", topProposal.ID.String()).
		Str("title", topProposal.Title).
		Int("vote_score", topProposal.VoteScore).
		Msg("Voting winner selected")
	
	return nil
}
