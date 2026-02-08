package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/events"
	"novels-backend/internal/repository"
)

type TranslationVotingService struct {
	repo       *repository.TranslationVotingRepository
	votingRepo *repository.VotingRepository
	ticketRepo *repository.TicketRepository
	events     *events.Bus
	logger     zerolog.Logger
}

func NewTranslationVotingService(
	repo *repository.TranslationVotingRepository,
	votingRepo *repository.VotingRepository,
	ticketRepo *repository.TicketRepository,
	eventBus *events.Bus,
	logger zerolog.Logger,
) *TranslationVotingService {
	return &TranslationVotingService{
		repo:       repo,
		votingRepo: votingRepo,
		ticketRepo: ticketRepo,
		events:     eventBus,
		logger:     logger.With().Str("service", "translation_voting").Logger(),
	}
}

func (s *TranslationVotingService) GetTranslationLeaderboard(ctx context.Context, limit int) (*models.TranslationLeaderboard, error) {
	entries, err := s.repo.ListLeaderboard(ctx, limit)
	if err != nil {
		return nil, err
	}
	return &models.TranslationLeaderboard{Entries: entries}, nil
}

func (s *TranslationVotingService) CastTranslationVote(ctx context.Context, userID uuid.UUID, req models.CastTranslationVoteRequest) error {
	if req.Amount < 1 {
		return fmt.Errorf("amount must be positive")
	}

	var target *models.TranslationVoteTarget
	var err error

	// Resolve target by explicit targetId, or by novelId/proposalId.
	switch {
	case req.TargetID != nil && *req.TargetID != "":
		tid, e := uuid.Parse(*req.TargetID)
		if e != nil {
			return fmt.Errorf("invalid targetId")
		}
		target, err = s.repo.GetTargetByID(ctx, tid)
	case req.NovelID != nil && *req.NovelID != "":
		nid, e := uuid.Parse(*req.NovelID)
		if e != nil {
			return fmt.Errorf("invalid novelId")
		}
		target, err = s.repo.EnsureTargetForNovel(ctx, nid)
	case req.ProposalID != nil && *req.ProposalID != "":
		pid, e := uuid.Parse(*req.ProposalID)
		if e != nil {
			return fmt.Errorf("invalid proposalId")
		}

		// Prevent voting for your own proposal (same rule as daily voting).
		p, e := s.votingRepo.GetProposalByID(ctx, pid)
		if e != nil {
			return e
		}
		if p == nil {
			return ErrProposalNotFound
		}
		if p.UserID == userID {
			return ErrCannotVoteOwnProposal
		}

		target, err = s.repo.EnsureTargetForProposal(ctx, pid)
	default:
		return fmt.Errorf("targetId or novelId or proposalId is required")
	}
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("target not found")
	}
	if target.Status != models.TranslationTargetStatusVoting {
		return fmt.Errorf("target is not in voting status")
	}

	// Spend translation tickets.
	if err := s.ticketRepo.SpendTickets(ctx, userID, models.TicketTypeTranslationTicket, req.Amount, "translation_vote", "translation_vote_target", &target.ID); err != nil {
		return err
	}

	// Record vote.
	if err := s.repo.CastTranslationVote(ctx, userID, target.ID, req.Amount); err != nil {
		return err
	}

	return nil
}

// ProcessTranslationWinner picks the winner and updates translation_vote_targets.status.
// Does NOT reset other targets' tickets.
func (s *TranslationVotingService) ProcessTranslationWinner(ctx context.Context) error {
	return s.processTranslationWinner(ctx, false)
}

// ProcessTranslationWinnerForce is an admin/testing helper: it can select a winner even if tickets == 0.
func (s *TranslationVotingService) ProcessTranslationWinnerForce(ctx context.Context) error {
	return s.processTranslationWinner(ctx, true)
}

func (s *TranslationVotingService) processTranslationWinner(ctx context.Context, force bool) error {
	top, err := s.repo.GetTopTarget(ctx)
	if err != nil {
		return err
	}
	if top == nil {
		s.logger.Info().Msg("No translation targets to process")
		return nil
	}
	if !force && top.TranslationTicketsInvested < 1 {
		s.logger.Info().Msg("No translation targets with tickets to process")
		return nil
	}

	// Winner status depends on whether it's already released (novel_id) or still announced (proposal_id).
	nextStatus := models.TranslationTargetStatusTranslating
	if top.ProposalID != nil && top.NovelID == nil {
		nextStatus = models.TranslationTargetStatusWaitingRelease
	}

	if err := s.repo.UpdateTargetStatus(ctx, top.ID, nextStatus); err != nil {
		return err
	}

	s.logger.Info().
		Str("target_id", top.ID.String()).
		Str("next_status", string(nextStatus)).
		Int("tickets", top.TranslationTicketsInvested).
		Msg("Translation vote winner selected")

	if s.events != nil {
		_ = s.events.Publish(ctx, events.TranslationVoteWinnerSelected{
			TargetID:   top.ID,
			NovelID:    top.NovelID,
			ProposalID: top.ProposalID,
		})
	}

	return nil
}

// OnProposalReleased binds proposal->novel for translation target and moves waiting_release -> translating.
func (s *TranslationVotingService) OnProposalReleased(ctx context.Context, proposalID uuid.UUID, novelID uuid.UUID) error {
	updated, err := s.repo.BindProposalToNovel(ctx, proposalID, novelID)
	if err != nil {
		return err
	}
	if updated == nil {
		return nil
	}

	s.logger.Info().
		Str("proposal_id", proposalID.String()).
		Str("novel_id", novelID.String()).
		Str("target_id", updated.ID.String()).
		Str("status", string(updated.Status)).
		Msg("Translation target bound to released novel")

	return nil
}

