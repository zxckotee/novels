package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/http/middleware"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"
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
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}
	
	var req models.CreateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	
	// Validate
	if req.OriginalLink == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "original_link is required")
		return
	}
	if req.Title == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "title is required")
		return
	}
	if req.Description == "" || len(req.Description) < 100 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "description must be at least 100 characters")
		return
	}
	if len(req.Genres) == 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "at least one genre is required")
		return
	}
	
	proposal, err := h.votingService.CreateProposal(r.Context(), userID, req)
	if err != nil {
		if err == service.ErrInsufficientTickets {
			response.Error(w, http.StatusPaymentRequired, "PAYMENT_REQUIRED", "Insufficient novel request tickets")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to create proposal")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create proposal")
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
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid proposal ID")
		return
	}
	
	var currentUserID *uuid.UUID
	if uidStr := middleware.GetUserID(r.Context()); uidStr != "" {
		if uid, err := uuid.Parse(uidStr); err == nil {
			currentUserID = &uid
		}
	}
	
	proposal, err := h.votingService.GetProposal(r.Context(), id, currentUserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get proposal")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get proposal")
		return
	}
	if proposal == nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Proposal not found")
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
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list proposals")
		return
	}
	
	response.JSON(w, http.StatusOK, proposals)
}

// GetMyProposals returns the current user's proposals
// GET /api/v1/proposals/my
func (h *VotingHandler) GetMyProposals(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
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
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get proposals")
		return
	}
	
	response.JSON(w, http.StatusOK, proposals)
}

// UpdateProposal updates a proposal
// PUT /api/v1/proposals/{id}
func (h *VotingHandler) UpdateProposal(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid proposal ID")
		return
	}
	
	var req models.UpdateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	
	proposal, err := h.votingService.UpdateProposal(r.Context(), id, userID, req)
	if err != nil {
		if err == service.ErrProposalNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Proposal not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update proposal")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	
	response.JSON(w, http.StatusOK, proposal)
}

// SubmitProposal submits a proposal for moderation
// POST /api/v1/proposals/{id}/submit
func (h *VotingHandler) SubmitProposal(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid proposal ID")
		return
	}
	
	err = h.votingService.SubmitProposal(r.Context(), id, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to submit proposal")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "proposal submitted for moderation"})
}

// DeleteProposal deletes a proposal
// DELETE /api/v1/proposals/{id}
func (h *VotingHandler) DeleteProposal(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid proposal ID")
		return
	}
	
	err = h.votingService.DeleteProposal(r.Context(), id, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete proposal")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
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
	moderatorIDStr := middleware.GetUserID(r.Context())
	if moderatorIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	moderatorID, err := uuid.Parse(moderatorIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid proposal ID")
		return
	}
	
	var req models.ModerateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	
	if req.Action != "approve" && req.Action != "reject" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "action must be 'approve' or 'reject'")
		return
	}
	
	err = h.votingService.ModerateProposal(r.Context(), id, moderatorID, req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to moderate proposal")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
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
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get proposals")
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
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}
	
	var req models.CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	
	// Validate
	if req.ProposalID == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "proposal_id is required")
		return
	}
	if req.Amount < 1 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "amount must be positive")
		return
	}
	if req.TicketType != models.TicketTypeDailyVote && req.TicketType != models.TicketTypeTranslationTicket {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticket_type must be 'daily_vote' or 'translation_ticket'")
		return
	}
	
	err = h.votingService.CastVote(r.Context(), userID, req)
	if err != nil {
		if err == service.ErrInsufficientTickets {
			response.Error(w, http.StatusPaymentRequired, "PAYMENT_REQUIRED", "Insufficient tickets")
			return
		}
		if err == service.ErrProposalNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Proposal not found")
			return
		}
		if err == service.ErrProposalNotVoting {
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Proposal is not in voting status")
			return
		}
		if err == service.ErrCannotVoteOwnProposal {
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Cannot vote for your own proposal")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to cast vote")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cast vote")
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
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get leaderboard")
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
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get stats")
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
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get proposals")
		return
	}
	
	response.JSON(w, http.StatusOK, proposals)
}
