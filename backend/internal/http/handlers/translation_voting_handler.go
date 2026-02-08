package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/http/middleware"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"
)

type TranslationVotingHandler struct {
	svc    *service.TranslationVotingService
	logger zerolog.Logger
}

func NewTranslationVotingHandler(svc *service.TranslationVotingService, logger zerolog.Logger) *TranslationVotingHandler {
	return &TranslationVotingHandler{svc: svc, logger: logger.With().Str("handler", "translation_voting").Logger()}
}

// GET /api/v1/translation/leaderboard
func (h *TranslationVotingHandler) GetTranslationLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}

	lb, err := h.svc.GetTranslationLeaderboard(r.Context(), limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get translation leaderboard")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get leaderboard")
		return
	}

	response.JSON(w, http.StatusOK, lb)
}

// POST /api/v1/translation-votes
func (h *TranslationVotingHandler) CastTranslationVote(w http.ResponseWriter, r *http.Request) {
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

	var req models.CastTranslationVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	if req.Amount < 1 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "amount must be positive")
		return
	}

	err = h.svc.CastTranslationVote(r.Context(), userID, req)
	if err != nil {
		if err == service.ErrInsufficientTickets {
			response.Error(w, http.StatusPaymentRequired, "PAYMENT_REQUIRED", "Insufficient tickets")
			return
		}
		if err == service.ErrProposalNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Proposal not found")
			return
		}
		if err == service.ErrCannotVoteOwnProposal {
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Cannot vote for your own proposal")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to cast translation vote")
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "translation vote cast successfully"})
}

