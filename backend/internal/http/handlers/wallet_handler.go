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

type WalletHandler struct {
	ticketService *service.TicketService
	logger        zerolog.Logger
}

func NewWalletHandler(ticketService *service.TicketService, logger zerolog.Logger) *WalletHandler {
	return &WalletHandler{
		ticketService: ticketService,
		logger:        logger,
	}
}

// GetWallet returns the current user's wallet
// GET /api/v1/wallet
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	wallet, err := h.ticketService.GetWallet(r.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get wallet")
		response.Error(w, http.StatusInternalServerError, "failed to get wallet", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, wallet)
}

// GetTransactions returns the current user's ticket transactions
// GET /api/v1/wallet/transactions
func (h *WalletHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	// Parse query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	var ticketType *models.TicketType
	if t := r.URL.Query().Get("type"); t != "" {
		tt := models.TicketType(t)
		ticketType = &tt
	}
	
	transactions, err := h.ticketService.GetTransactions(r.Context(), userID, ticketType, page, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get transactions")
		response.Error(w, http.StatusInternalServerError, "failed to get transactions", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, transactions)
}

// GetStats returns the current user's ticket statistics
// GET /api/v1/wallet/stats
func (h *WalletHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	stats, err := h.ticketService.GetUserStats(r.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get stats")
		response.Error(w, http.StatusInternalServerError, "failed to get stats", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, stats)
}

// GetLeaderboard returns the ticket spending leaderboard
// GET /api/v1/leaderboard
func (h *WalletHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	entries, err := h.ticketService.GetLeaderboard(r.Context(), period, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get leaderboard")
		response.Error(w, http.StatusInternalServerError, "failed to get leaderboard", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"period":  period,
		"entries": entries,
	})
}

// GrantTickets grants tickets to a user (admin only)
// POST /api/v1/admin/tickets/grant
func (h *WalletHandler) GrantTickets(w http.ResponseWriter, r *http.Request) {
	var req models.GrantTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	
	// Validate
	if req.UserID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "user_id is required", nil)
		return
	}
	if req.Amount < 1 {
		response.Error(w, http.StatusBadRequest, "amount must be positive", nil)
		return
	}
	if req.Reason == "" {
		req.Reason = models.ReasonAdminAdjustment
	}
	
	err := h.ticketService.GrantTickets(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to grant tickets")
		response.Error(w, http.StatusInternalServerError, "failed to grant tickets", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "tickets granted successfully"})
}

// GetUserWallet returns a specific user's wallet (admin only)
// GET /api/v1/admin/users/{userId}/wallet
func (h *WalletHandler) GetUserWallet(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID", nil)
		return
	}
	
	wallet, err := h.ticketService.GetWallet(r.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get user wallet")
		response.Error(w, http.StatusInternalServerError, "failed to get wallet", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, wallet)
}
