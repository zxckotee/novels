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

// CollectionHandler handles collection-related requests
type CollectionHandler struct {
	collectionService *service.CollectionService
}

// NewCollectionHandler creates a new collection handler
func NewCollectionHandler(collectionService *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
	}
}

// List returns a list of collections
// GET /collections
func (h *CollectionHandler) List(w http.ResponseWriter, r *http.Request) {
	params := models.CollectionListParams{
		Sort:  r.URL.Query().Get("sort"),
		Page:  1,
		Limit: 20,
	}

	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		params.Page = page
	}
	if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit > 0 && limit <= 50 {
		params.Limit = limit
	}

	if userIDStr := r.URL.Query().Get("userId"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			params.UserID = &userID
		}
	}

	if r.URL.Query().Get("featured") == "true" {
		featured := true
		params.IsFeatured = &featured
	}

	result, err := h.collectionService.List(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list collections")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetByID returns a collection by ID
// GET /collections/{id}
func (h *CollectionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	var viewerID *uuid.UUID
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			viewerID = &uid
		}
	}

	collection, err := h.collectionService.GetByID(r.Context(), id, viewerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get collection")
		return
	}
	if collection == nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Collection not found")
		return
	}

	response.JSON(w, http.StatusOK, collection)
}

// GetPopular returns popular collections
// GET /collections/popular
func (h *CollectionHandler) GetPopular(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l, _ := strconv.Atoi(r.URL.Query().Get("limit")); l > 0 && l <= 20 {
		limit = l
	}

	collections, err := h.collectionService.GetPopular(r.Context(), limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get popular collections")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"collections": collections,
	})
}

// GetFeatured returns featured collections
// GET /collections/featured
func (h *CollectionHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l, _ := strconv.Atoi(r.URL.Query().Get("limit")); l > 0 && l <= 20 {
		limit = l
	}

	collections, err := h.collectionService.GetFeatured(r.Context(), limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get featured collections")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"collections": collections,
	})
}

// Create creates a new collection
// POST /collections
func (h *CollectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
		return
	}

	var req models.CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	collection, err := h.collectionService.Create(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create collection")
		return
	}

	response.JSON(w, http.StatusCreated, collection)
}

// Update updates a collection
// PUT /collections/{id}
func (h *CollectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	var req models.UpdateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	collection, err := h.collectionService.Update(r.Context(), id, userID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update collection")
		return
	}

	response.JSON(w, http.StatusOK, collection)
}

// Delete deletes a collection
// DELETE /collections/{id}
func (h *CollectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	if err := h.collectionService.Delete(r.Context(), id, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete collection")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// AddItem adds a novel to a collection
// POST /collections/{id}/items
func (h *CollectionHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
		return
	}

	idStr := chi.URLParam(r, "id")
	collectionID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	var req models.AddToCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if err := h.collectionService.AddItem(r.Context(), collectionID, userID, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add item")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"status": "added"})
}

// RemoveItem removes a novel from a collection
// DELETE /collections/{id}/items/{novelId}
func (h *CollectionHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
		return
	}

	collectionIDStr := chi.URLParam(r, "id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	novelIDStr := chi.URLParam(r, "novelId")
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid novel ID")
		return
	}

	if err := h.collectionService.RemoveItem(r.Context(), collectionID, novelID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove item")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ReorderItems reorders items in a collection
// PUT /collections/{id}/items/reorder
func (h *CollectionHandler) ReorderItems(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
		return
	}

	idStr := chi.URLParam(r, "id")
	collectionID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	var req models.ReorderCollectionItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if err := h.collectionService.ReorderItems(r.Context(), collectionID, userID, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reorder items")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}

// Vote votes on a collection
// POST /collections/{id}/vote
func (h *CollectionHandler) Vote(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
		return
	}

	idStr := chi.URLParam(r, "id")
	collectionID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	if err := h.collectionService.Vote(r.Context(), collectionID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to vote")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "voted"})
}

// GetUserCollections returns collections for a user
// GET /users/{id}/collections
func (h *CollectionHandler) GetUserCollections(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	profileUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID")
		return
	}

	var viewerID *uuid.UUID
	viewerIDStr := middleware.GetUserID(r.Context())
	if viewerIDStr != "" {
		if uid, err := uuid.Parse(viewerIDStr); err == nil {
			viewerID = &uid
		}
	}

	collections, err := h.collectionService.GetUserCollections(r.Context(), profileUserID, viewerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user collections")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"collections": collections,
	})
}

// SetFeatured sets featured status (admin only)
// POST /admin/collections/{id}/featured
func (h *CollectionHandler) SetFeatured(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid collection ID")
		return
	}

	var req struct {
		Featured bool `json:"featured"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if err := h.collectionService.SetFeatured(r.Context(), id, req.Featured); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set featured status")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"featured": req.Featured})
}
