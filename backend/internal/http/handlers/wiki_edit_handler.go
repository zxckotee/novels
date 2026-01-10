package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"novels/internal/domain/models"
	"novels/internal/http/middleware"
	"novels/internal/service"
	"novels/pkg/response"

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
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	novelIDStr := chi.URLParam(r, "id")
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid novel ID", nil)
		return
	}

	var req models.CreateEditRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	req.NovelID = novelID

	editRequest, err := h.wikiEditService.CreateEditRequest(r.Context(), userID, &req)
	if err != nil {
		if err.Error() == "premium subscription required" {
			response.Error(w, http.StatusForbidden, "Premium subscription required", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to create edit request", err)
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
		response.Error(w, http.StatusBadRequest, "Invalid edit request ID", nil)
		return
	}

	editRequest, err := h.wikiEditService.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get edit request", err)
		return
	}
	if editRequest == nil {
		response.Error(w, http.StatusNotFound, "Edit request not found", nil)
		return
	}

	response.JSON(w, http.StatusOK, editRequest)
}

// GetUserEditRequests returns edit requests for the current user
// GET /me/edit-requests
func (h *WikiEditHandler) GetUserEditRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	params := models.EditRequestListParams{
		Page:  1,
		Limit: 20,
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

	result, err := h.wikiEditService.GetUserEditRequests(r.Context(), userID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get edit requests", err)
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
		response.Error(w, http.StatusBadRequest, "Invalid novel ID", nil)
		return
	}

	params := models.EditRequestListParams{
		Page:  1,
		Limit: 20,
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

	result, err := h.wikiEditService.GetNovelEditRequests(r.Context(), novelID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get edit requests", err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetPendingEditRequests returns all pending edit requests (moderator/admin)
// GET /moderation/edit-requests
func (h *WikiEditHandler) GetPendingEditRequests(w http.ResponseWriter, r *http.Request) {
	params := models.EditRequestListParams{
		Page:  1,
		Limit: 20,
	}

	// Default to pending status
	pendingStatus := models.EditRequestStatusPending
	params.Status = &pendingStatus

	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		params.Page = page
	}
	if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit > 0 && limit <= 50 {
		params.Limit = limit
	}

	// Allow overriding status
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.EditRequestStatus(statusStr)
		params.Status = &status
	}

	result, err := h.wikiEditService.GetPendingEditRequests(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get edit requests", err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// ApproveEditRequest approves an edit request (moderator/admin)
// POST /moderation/edit-requests/{id}/approve
func (h *WikiEditHandler) ApproveEditRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid edit request ID", nil)
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.wikiEditService.ApproveEditRequest(r.Context(), id, userID, req.Comment); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to approve edit request", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

// RejectEditRequest rejects an edit request (moderator/admin)
// POST /moderation/edit-requests/{id}/reject
func (h *WikiEditHandler) RejectEditRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid edit request ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Reason == "" {
		response.Error(w, http.StatusBadRequest, "Rejection reason is required", nil)
		return
	}

	if err := h.wikiEditService.RejectEditRequest(r.Context(), id, userID, req.Reason); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to reject edit request", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

// CancelEditRequest cancels an edit request (owner only)
// POST /edit-requests/{id}/cancel
func (h *WikiEditHandler) CancelEditRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid edit request ID", nil)
		return
	}

	if err := h.wikiEditService.CancelEditRequest(r.Context(), id, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to cancel edit request", err)
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
		response.Error(w, http.StatusBadRequest, "Invalid novel ID", nil)
		return
	}

	page := 1
	limit := 20
	if p, _ := strconv.Atoi(r.URL.Query().Get("page")); p > 0 {
		page = p
	}
	if l, _ := strconv.Atoi(r.URL.Query().Get("limit")); l > 0 && l <= 50 {
		limit = l
	}

	history, total, err := h.wikiEditService.GetNovelEditHistory(r.Context(), novelID, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get edit history", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetPlatformStats returns platform statistics
// GET /stats/platform
func (h *WikiEditHandler) GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.wikiEditService.GetPlatformStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get platform stats", err)
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
