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

// NewsHandler handles news-related requests
type NewsHandler struct {
	newsService *service.NewsService
}

// NewNewsHandler creates a new news handler
func NewNewsHandler(newsService *service.NewsService) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
	}
}

// List returns a list of news posts
// GET /news
func (h *NewsHandler) List(w http.ResponseWriter, r *http.Request) {
	params := models.NewsListParams{
		Lang:  r.URL.Query().Get("lang"),
		Page:  1,
		Limit: 20,
	}

	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		params.Page = page
	}
	if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit > 0 && limit <= 50 {
		params.Limit = limit
	}

	if category := r.URL.Query().Get("category"); category != "" {
		cat := models.NewsCategory(category)
		params.Category = &cat
	}

	result, err := h.newsService.List(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list news", err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetBySlug returns a news post by slug
// GET /news/{slug}
func (h *NewsHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ru"
	}

	news, err := h.newsService.GetBySlug(r.Context(), slug, lang)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get news", err)
		return
	}
	if news == nil {
		response.Error(w, http.StatusNotFound, "News not found", nil)
		return
	}

	response.JSON(w, http.StatusOK, news)
}

// GetLatest returns latest news for homepage
// GET /news/latest
func (h *NewsHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	limit := 5
	if l, _ := strconv.Atoi(r.URL.Query().Get("limit")); l > 0 && l <= 10 {
		limit = l
	}

	news, err := h.newsService.GetLatest(r.Context(), limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get latest news", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"news": news,
	})
}

// GetPinned returns pinned news
// GET /news/pinned
func (h *NewsHandler) GetPinned(w http.ResponseWriter, r *http.Request) {
	news, err := h.newsService.GetPinned(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get pinned news", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"news": news,
	})
}

// Create creates a new news post (admin/moderator only)
// POST /admin/news
func (h *NewsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.CreateNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	news, err := h.newsService.Create(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create news", err)
		return
	}

	response.JSON(w, http.StatusCreated, news)
}

// Update updates a news post
// PUT /admin/news/{id}
func (h *NewsHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid news ID", nil)
		return
	}

	var req models.UpdateNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	news, err := h.newsService.Update(r.Context(), id, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update news", err)
		return
	}

	response.JSON(w, http.StatusOK, news)
}

// Delete deletes a news post
// DELETE /admin/news/{id}
func (h *NewsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid news ID", nil)
		return
	}

	if err := h.newsService.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete news", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Publish publishes a news post
// POST /admin/news/{id}/publish
func (h *NewsHandler) Publish(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid news ID", nil)
		return
	}

	if err := h.newsService.Publish(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to publish news", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "published"})
}

// Unpublish unpublishes a news post
// POST /admin/news/{id}/unpublish
func (h *NewsHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid news ID", nil)
		return
	}

	if err := h.newsService.Unpublish(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to unpublish news", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "unpublished"})
}

// SetPinned sets pinned status
// POST /admin/news/{id}/pin
func (h *NewsHandler) SetPinned(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid news ID", nil)
		return
	}

	var req struct {
		Pinned bool `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.newsService.SetPinned(r.Context(), id, req.Pinned); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to set pinned status", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"pinned": req.Pinned})
}

// SetLocalization adds or updates a news localization
// PUT /admin/news/{id}/localizations/{lang}
func (h *NewsHandler) SetLocalization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid news ID", nil)
		return
	}

	lang := chi.URLParam(r, "lang")

	var req models.NewsLocalizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	req.Lang = lang

	if err := h.newsService.SetLocalization(r.Context(), id, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to set localization", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// DeleteLocalization deletes a news localization
// DELETE /admin/news/{id}/localizations/{lang}
func (h *NewsHandler) DeleteLocalization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid news ID", nil)
		return
	}

	lang := chi.URLParam(r, "lang")

	if err := h.newsService.DeleteLocalization(r.Context(), id, lang); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete localization", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetAdminBySlug returns a news post for admin (includes unpublished)
// GET /admin/news/{slug}
func (h *NewsHandler) GetAdminBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	news, err := h.newsService.GetBySlugAdmin(r.Context(), slug)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get news", err)
		return
	}
	if news == nil {
		response.Error(w, http.StatusNotFound, "News not found", nil)
		return
	}

	response.JSON(w, http.StatusOK, news)
}

// ListAdmin returns a list of news posts for admin (includes unpublished)
// GET /admin/news
func (h *NewsHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	params := models.NewsListParams{
		Page:  1,
		Limit: 20,
	}

	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		params.Page = page
	}
	if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit > 0 && limit <= 50 {
		params.Limit = limit
	}

	if category := r.URL.Query().Get("category"); category != "" {
		cat := models.NewsCategory(category)
		params.Category = &cat
	}

	// Allow viewing all statuses
	if published := r.URL.Query().Get("published"); published != "" {
		isPublished := published == "true"
		params.IsPublished = &isPublished
	} else {
		params.IsPublished = nil // Show all
	}

	result, err := h.newsService.List(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list news", err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}
