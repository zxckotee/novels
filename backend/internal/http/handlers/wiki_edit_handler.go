package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/http/middleware"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WikiEditHandler handles wiki edit request operations
type WikiEditHandler struct {
	wikiEditService *service.WikiEditService
}

// NewWikiEditHandler creates a new wiki edit handler
func NewWikiEditHandler(wikiEditService *service.WikiEditService) *WikiEditHandler {
	return &WikiEditHandler{
		wikiEditService: wikiEditService,
	}
}

// CreateEditRequest creates a new edit request (Premium only)
// POST /novels/{id}/edit-requests
func (h *WikiEditHandler) CreateEditRequest(w http.ResponseWriter, r *http.Request) {
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

	novelIDStr := chi.URLParam(r, "id")
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid novel ID")
		return
	}

	var req models.CreateEditRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	editRequest, err := h.wikiEditService.CreateEditRequest(r.Context(), userID, novelID, &req)
	if err != nil {
		if err.Error() == "editing descriptions requires Premium subscription" {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "Premium subscription required")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create edit request")
		return
	}

	response.JSON(w, http.StatusCreated, editRequest)
}

// GetEditRequest returns an edit request by ID
// GET /edit-requests/{id}
func (h *WikiEditHandler) GetEditRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid edit request ID")
		return
	}

	editRequest, err := h.wikiEditService.GetEditRequest(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get edit request")
		return
	}
	if editRequest == nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Edit request not found")
		return
	}

	response.JSON(w, http.StatusOK, editRequest)
}

// GetUserEditRequests returns edit requests for the current user
// GET /me/edit-requests
func (h *WikiEditHandler) GetUserEditRequests(w http.ResponseWriter, r *http.Request) {
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

	params := models.EditRequestListParams{
		Page:  1,
		Limit: 20,
		UserID: &userID,
	}

	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		params.Page = page
	}
	if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit > 0 && limit <= 50 {
		params.Limit = limit
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.EditRequestStatus(statusStr)
		params.Status = &status
	}

	result, err := h.wikiEditService.ListEditRequests(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get edit requests")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetNovelEditRequests returns edit requests for a specific novel
// GET /novels/{id}/edit-requests
func (h *WikiEditHandler) GetNovelEditRequests(w http.ResponseWriter, r *http.Request) {
	novelIDStr := chi.URLParam(r, "id")
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid novel ID")
		return
	}

	params := models.EditRequestListParams{
		Page:    1,
		Limit:   20,
		NovelID: &novelID,
	}

	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		params.Page = page
	}
	if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit > 0 && limit <= 50 {
		params.Limit = limit
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.EditRequestStatus(statusStr)
		params.Status = &status
	}

	result, err := h.wikiEditService.ListEditRequests(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get edit requests")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetPendingEditRequests returns all pending edit requests (moderator/admin)
// GET /moderation/edit-requests
func (h *WikiEditHandler) GetPendingEditRequests(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20

	if p, _ := strconv.Atoi(r.URL.Query().Get("page")); p > 0 {
		page = p
	}
	if l, _ := strconv.Atoi(r.URL.Query().Get("limit")); l > 0 && l <= 50 {
		limit = l
	}

	result, err := h.wikiEditService.GetPendingRequests(r.Context(), page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get edit requests")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// ApproveEditRequest approves an edit request (moderator/admin)
// POST /moderation/edit-requests/{id}/approve
func (h *WikiEditHandler) ApproveEditRequest(w http.ResponseWriter, r *http.Request) {
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
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid edit request ID")
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	reviewReq := &models.ReviewEditRequestRequest{
		Action:  "approve",
		Comment: req.Comment,
	}

	if err := h.wikiEditService.ReviewEditRequest(r.Context(), id, userID, reviewReq); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to approve edit request")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

// RejectEditRequest rejects an edit request (moderator/admin)
// POST /moderation/edit-requests/{id}/reject
func (h *WikiEditHandler) RejectEditRequest(w http.ResponseWriter, r *http.Request) {
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
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid edit request ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if req.Reason == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Rejection reason is required")
		return
	}

	reviewReq := &models.ReviewEditRequestRequest{
		Action:  "reject",
		Comment: req.Reason,
	}

	if err := h.wikiEditService.ReviewEditRequest(r.Context(), id, userID, reviewReq); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reject edit request")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

// CancelEditRequest cancels an edit request (owner only)
// POST /edit-requests/{id}/cancel
func (h *WikiEditHandler) CancelEditRequest(w http.ResponseWriter, r *http.Request) {
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
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid edit request ID")
		return
	}

	if err := h.wikiEditService.WithdrawEditRequest(r.Context(), id, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel edit request")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetNovelEditHistory returns edit history for a novel
// GET /novels/{id}/edit-history
func (h *WikiEditHandler) GetNovelEditHistory(w http.ResponseWriter, r *http.Request) {
	novelIDStr := chi.URLParam(r, "id")
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid novel ID")
		return
	}

	params := models.EditHistoryListParams{
		NovelID: &novelID,
		Page:    1,
		Limit:   20,
	}

	if p, _ := strconv.Atoi(r.URL.Query().Get("page")); p > 0 {
		params.Page = p
	}
	if l, _ := strconv.Atoi(r.URL.Query().Get("limit")); l > 0 && l <= 50 {
		params.Limit = l
	}

	result, err := h.wikiEditService.GetEditHistory(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get edit history")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetPlatformStats returns platform statistics
// GET /stats/platform
func (h *WikiEditHandler) GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.wikiEditService.GetPlatformStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get platform stats")
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
