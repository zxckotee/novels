package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/novels/backend/internal/domain/models"
	"github.com/novels/backend/internal/http/middleware"
	"github.com/novels/backend/internal/service"
	"github.com/novels/backend/pkg/response"
	"github.com/rs/zerolog"
)

type VotingHandler struct {
	votingService *service.VotingService
	logger        zerolog.Logger
}

func NewVotingHandler(votingService *service.VotingService, logger zerolog.Logger) *VotingHandler {
	return &VotingHandler{
		votingService: votingService,
		logger:        logger,
	}
}

// ============================================
// PROPOSALS
// ============================================

// CreateProposal creates a new novel proposal
// POST /api/v1/proposals
func (h *VotingHandler) CreateProposal(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	var req models.CreateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	
	// Validate
	if req.OriginalLink == "" {
		response.Error(w, http.StatusBadRequest, "original_link is required", nil)
		return
	}
	if req.Title == "" {
		response.Error(w, http.StatusBadRequest, "title is required", nil)
		return
	}
	if req.Description == "" || len(req.Description) < 100 {
		response.Error(w, http.StatusBadRequest, "description must be at least 100 characters", nil)
		return
	}
	if len(req.Genres) == 0 {
		response.Error(w, http.StatusBadRequest, "at least one genre is required", nil)
		return
	}
	
	proposal, err := h.votingService.CreateProposal(r.Context(), userID, req)
	if err != nil {
		if err == service.ErrInsufficientTickets {
			response.Error(w, http.StatusPaymentRequired, "insufficient novel request tickets", nil)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to create proposal")
		response.Error(w, http.StatusInternalServerError, "failed to create proposal", nil)
		return
	}
	
	response.JSON(w, http.StatusCreated, proposal)
}

// GetProposal returns a proposal by ID
// GET /api/v1/proposals/{id}
func (h *VotingHandler) GetProposal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid proposal ID", nil)
		return
	}
	
	var currentUserID *uuid.UUID
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		currentUserID = &uid
	}
	
	proposal, err := h.votingService.GetProposal(r.Context(), id, currentUserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get proposal")
		response.Error(w, http.StatusInternalServerError, "failed to get proposal", nil)
		return
	}
	if proposal == nil {
		response.Error(w, http.StatusNotFound, "proposal not found", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, proposal)
}

// ListProposals returns proposals with filters
// GET /api/v1/proposals
func (h *VotingHandler) ListProposals(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	
	filter := models.ProposalFilter{
		Sort:  r.URL.Query().Get("sort"),
		Page:  page,
		Limit: limit,
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		s := models.ProposalStatus(status)
		filter.Status = &s
	}
	
	proposals, err := h.votingService.ListProposals(r.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list proposals")
		response.Error(w, http.StatusInternalServerError, "failed to list proposals", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, proposals)
}

// GetMyProposals returns the current user's proposals
// GET /api/v1/proposals/my
func (h *VotingHandler) GetMyProposals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	
	proposals, err := h.votingService.GetMyProposals(r.Context(), userID, page, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get my proposals")
		response.Error(w, http.StatusInternalServerError, "failed to get proposals", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, proposals)
}

// UpdateProposal updates a proposal
// PUT /api/v1/proposals/{id}
func (h *VotingHandler) UpdateProposal(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid proposal ID", nil)
		return
	}
	
	var req models.UpdateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	
	proposal, err := h.votingService.UpdateProposal(r.Context(), id, userID, req)
	if err != nil {
		if err == service.ErrProposalNotFound {
			response.Error(w, http.StatusNotFound, "proposal not found", nil)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update proposal")
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	response.JSON(w, http.StatusOK, proposal)
}

// SubmitProposal submits a proposal for moderation
// POST /api/v1/proposals/{id}/submit
func (h *VotingHandler) SubmitProposal(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid proposal ID", nil)
		return
	}
	
	err = h.votingService.SubmitProposal(r.Context(), id, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to submit proposal")
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "proposal submitted for moderation"})
}

// DeleteProposal deletes a proposal
// DELETE /api/v1/proposals/{id}
func (h *VotingHandler) DeleteProposal(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid proposal ID", nil)
		return
	}
	
	err = h.votingService.DeleteProposal(r.Context(), id, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete proposal")
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "proposal deleted"})
}

// ============================================
// MODERATION
// ============================================

// ModerateProposal approves or rejects a proposal
// POST /api/v1/moderation/proposals/{id}
func (h *VotingHandler) ModerateProposal(w http.ResponseWriter, r *http.Request) {
	moderatorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid proposal ID", nil)
		return
	}
	
	var req models.ModerateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	
	if req.Action != "approve" && req.Action != "reject" {
		response.Error(w, http.StatusBadRequest, "action must be 'approve' or 'reject'", nil)
		return
	}
	
	err = h.votingService.ModerateProposal(r.Context(), id, moderatorID, req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to moderate proposal")
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "proposal moderated"})
}

// GetPendingProposals returns proposals pending moderation
// GET /api/v1/moderation/proposals
func (h *VotingHandler) GetPendingProposals(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	status := models.ProposalStatusModeration
	filter := models.ProposalFilter{
		Status: &status,
		Sort:   "oldest",
		Page:   page,
		Limit:  20,
	}
	
	proposals, err := h.votingService.ListProposals(r.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get pending proposals")
		response.Error(w, http.StatusInternalServerError, "failed to get proposals", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, proposals)
}

// ============================================
// VOTING
// ============================================

// CastVote casts a vote for a proposal
// POST /api/v1/votes
func (h *VotingHandler) CastVote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	var req models.CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	
	// Validate
	if req.ProposalID == "" {
		response.Error(w, http.StatusBadRequest, "proposal_id is required", nil)
		return
	}
	if req.Amount < 1 {
		response.Error(w, http.StatusBadRequest, "amount must be positive", nil)
		return
	}
	if req.TicketType != models.TicketTypeDailyVote && req.TicketType != models.TicketTypeTranslationTicket {
		response.Error(w, http.StatusBadRequest, "ticket_type must be 'daily_vote' or 'translation_ticket'", nil)
		return
	}
	
	err := h.votingService.CastVote(r.Context(), userID, req)
	if err != nil {
		if err == service.ErrInsufficientTickets {
			response.Error(w, http.StatusPaymentRequired, "insufficient tickets", nil)
			return
		}
		if err == service.ErrProposalNotFound {
			response.Error(w, http.StatusNotFound, "proposal not found", nil)
			return
		}
		if err == service.ErrProposalNotVoting {
			response.Error(w, http.StatusBadRequest, "proposal is not in voting status", nil)
			return
		}
		if err == service.ErrCannotVoteOwnProposal {
			response.Error(w, http.StatusBadRequest, "cannot vote for your own proposal", nil)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to cast vote")
		response.Error(w, http.StatusInternalServerError, "failed to cast vote", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "vote cast successfully"})
}

// GetVotingLeaderboard returns the current voting leaderboard
// GET /api/v1/voting/leaderboard
func (h *VotingHandler) GetVotingLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	
	leaderboard, err := h.votingService.GetVotingLeaderboard(r.Context(), limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get voting leaderboard")
		response.Error(w, http.StatusInternalServerError, "failed to get leaderboard", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, leaderboard)
}

// GetVotingStats returns voting statistics
// GET /api/v1/voting/stats
func (h *VotingHandler) GetVotingStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.votingService.GetVotingStats(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get voting stats")
		response.Error(w, http.StatusInternalServerError, "failed to get stats", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, stats)
}

// GetVotingProposals returns proposals in voting status (convenience endpoint)
// GET /api/v1/voting/proposals
func (h *VotingHandler) GetVotingProposals(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	
	status := models.ProposalStatusVoting
	filter := models.ProposalFilter{
		Status: &status,
		Sort:   "votes",
		Page:   page,
		Limit:  limit,
	}
	
	proposals, err := h.votingService.ListProposals(r.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get voting proposals")
		response.Error(w, http.StatusInternalServerError, "failed to get proposals", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, proposals)
}
