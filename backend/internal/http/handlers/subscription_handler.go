package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/novels/backend/internal/domain/models"
	"github.com/novels/backend/internal/http/middleware"
	"github.com/novels/backend/internal/service"
	"github.com/novels/backend/pkg/response"
	"github.com/rs/zerolog"
)

type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
	logger              zerolog.Logger
}

func NewSubscriptionHandler(subscriptionService *service.SubscriptionService, logger zerolog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		logger:              logger,
	}
}

// GetPlans returns all available subscription plans
// GET /api/v1/subscriptions/plans
func (h *SubscriptionHandler) GetPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.subscriptionService.GetPlans(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get subscription plans")
		response.Error(w, http.StatusInternalServerError, "failed to get plans", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, models.SubscriptionPlansResponse{Plans: plans})
}

// GetPlan returns a subscription plan by ID
// GET /api/v1/subscriptions/plans/{id}
func (h *SubscriptionHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid plan ID", nil)
		return
	}
	
	plan, err := h.subscriptionService.GetPlan(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get plan")
		response.Error(w, http.StatusInternalServerError, "failed to get plan", nil)
		return
	}
	if plan == nil {
		response.Error(w, http.StatusNotFound, "plan not found", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, plan)
}

// GetMySubscription returns the current user's subscription info
// GET /api/v1/subscriptions/me
func (h *SubscriptionHandler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	info, err := h.subscriptionService.GetUserSubscription(r.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get subscription")
		response.Error(w, http.StatusInternalServerError, "failed to get subscription", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, info)
}

// GetMySubscriptionHistory returns the current user's subscription history
// GET /api/v1/subscriptions/history
func (h *SubscriptionHandler) GetMySubscriptionHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	subscriptions, err := h.subscriptionService.GetUserSubscriptionHistory(r.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get subscription history")
		response.Error(w, http.StatusInternalServerError, "failed to get history", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
	})
}

// Subscribe creates a new subscription
// POST /api/v1/subscriptions
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	var req models.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	
	if req.PlanID == "" {
		response.Error(w, http.StatusBadRequest, "plan_id is required", nil)
		return
	}
	
	subscription, err := h.subscriptionService.Subscribe(r.Context(), userID, req)
	if err != nil {
		if err == service.ErrAlreadySubscribed {
			response.Error(w, http.StatusConflict, "user already has an active subscription", nil)
			return
		}
		if err == service.ErrPlanNotFound {
			response.Error(w, http.StatusNotFound, "subscription plan not found", nil)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to create subscription")
		response.Error(w, http.StatusInternalServerError, "failed to create subscription", nil)
		return
	}
	
	response.JSON(w, http.StatusCreated, subscription)
}

// CancelSubscription cancels the user's subscription
// POST /api/v1/subscriptions/{id}/cancel
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	idStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid subscription ID", nil)
		return
	}
	
	err = h.subscriptionService.CancelSubscription(r.Context(), userID, subscriptionID)
	if err != nil {
		if err == service.ErrSubscriptionNotFound {
			response.Error(w, http.StatusNotFound, "subscription not found", nil)
			return
		}
		if err == service.ErrNotAuthorized {
			response.Error(w, http.StatusForbidden, "not authorized", nil)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to cancel subscription")
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "subscription canceled"})
}

// CheckFeature checks if the user has a specific premium feature
// GET /api/v1/subscriptions/features/{feature}
func (h *SubscriptionHandler) CheckFeature(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	feature := chi.URLParam(r, "feature")
	if feature == "" {
		response.Error(w, http.StatusBadRequest, "feature is required", nil)
		return
	}
	
	hasFeature, err := h.subscriptionService.HasFeature(r.Context(), userID, feature)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check feature")
		response.Error(w, http.StatusInternalServerError, "failed to check feature", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"feature":    feature,
		"hasAccess":  hasFeature,
	})
}

// IsPremium checks if the user has any premium subscription
// GET /api/v1/subscriptions/premium
func (h *SubscriptionHandler) IsPremium(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	
	isPremium, err := h.subscriptionService.IsPremium(r.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to check premium status")
		response.Error(w, http.StatusInternalServerError, "failed to check premium", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]bool{
		"isPremium": isPremium,
	})
}

// ============================================
// ADMIN
// ============================================

// GetSubscriptionStats returns subscription statistics (admin only)
// GET /api/v1/admin/subscriptions/stats
func (h *SubscriptionHandler) GetSubscriptionStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.subscriptionService.GetStats(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get subscription stats")
		response.Error(w, http.StatusInternalServerError, "failed to get stats", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, stats)
}

// GetUserSubscription returns a specific user's subscription info (admin only)
// GET /api/v1/admin/users/{userId}/subscription
func (h *SubscriptionHandler) GetUserSubscription(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID", nil)
		return
	}
	
	info, err := h.subscriptionService.GetUserSubscription(r.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get user subscription")
		response.Error(w, http.StatusInternalServerError, "failed to get subscription", nil)
		return
	}
	
	response.JSON(w, http.StatusOK, info)
}
