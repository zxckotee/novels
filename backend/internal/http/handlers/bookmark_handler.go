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
)

type BookmarkHandler struct {
	bookmarkService *service.BookmarkService
}

func NewBookmarkHandler(bookmarkService *service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{bookmarkService: bookmarkService}
}

// List godoc
// @Summary List bookmarks
// @Description Get user's bookmarks with optional filter
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param list query string false "List code (reading, planned, dropped, completed, favorites)"
// @Param sort query string false "Sort order (latest_update, date_added, title)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{data=models.BookmarksResponse}
// @Router /bookmarks [get]
func (h *BookmarkHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	filter := models.BookmarksFilter{
		Sort:  r.URL.Query().Get("sort"),
		Page:  parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 20),
	}
	
	listCodeStr := r.URL.Query().Get("list")
	if listCodeStr != "" {
		listCode := models.BookmarkListCode(listCodeStr)
		filter.ListCode = &listCode
	}
	
	// Get language from header or default
	lang := r.Header.Get("Accept-Language")
	if lang == "" {
		lang = "ru"
	}
	
	result, err := h.bookmarkService.List(r.Context(), userID, filter, lang)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get bookmarks")
		return
	}
	
	response.JSON(w, http.StatusOK, result)
}

// GetLists godoc
// @Summary Get bookmark lists
// @Description Get user's bookmark lists with counts
// @Tags bookmarks
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.BookmarkList}
// @Router /bookmarks/lists [get]
func (h *BookmarkHandler) GetLists(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	lists, err := h.bookmarkService.GetLists(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get lists")
		return
	}
	
	response.JSON(w, http.StatusOK, lists)
}

// Add godoc
// @Summary Add bookmark
// @Description Add a novel to a bookmark list
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param body body models.CreateBookmarkRequest true "Bookmark data"
// @Success 201 {object} response.Response{data=models.Bookmark}
// @Router /bookmarks [post]
func (h *BookmarkHandler) Add(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	var req models.CreateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	
	if req.NovelID == "" || req.ListCode == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "novel_id and list_code are required")
		return
	}
	
	bookmark, err := h.bookmarkService.AddBookmark(r.Context(), userID, req.NovelID, req.ListCode)
	if err != nil {
		switch err {
		case service.ErrNovelNotFound:
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "novel not found")
		case service.ErrInvalidListCode:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid list code")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to add bookmark")
		}
		return
	}
	
	response.JSON(w, http.StatusCreated, bookmark)
}

// Update godoc
// @Summary Update bookmark
// @Description Move a bookmark to another list
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param novelId path string true "Novel ID"
// @Param body body models.UpdateBookmarkRequest true "Update data"
// @Success 204 "No Content"
// @Router /bookmarks/{novelId} [put]
func (h *BookmarkHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	novelID := chi.URLParam(r, "novelId")
	if novelID == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "novel_id is required")
		return
	}
	
	var req models.UpdateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	
	err = h.bookmarkService.UpdateBookmark(r.Context(), userID, novelID, req.ListCode)
	if err != nil {
		switch err {
		case service.ErrNovelNotFound:
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "bookmark not found")
		case service.ErrInvalidListCode:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid list code")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update bookmark")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Remove godoc
// @Summary Remove bookmark
// @Description Remove a novel from bookmarks
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param novelId path string true "Novel ID"
// @Success 204 "No Content"
// @Router /bookmarks/{novelId} [delete]
func (h *BookmarkHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	novelID := chi.URLParam(r, "novelId")
	if novelID == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "novel_id is required")
		return
	}
	
	err = h.bookmarkService.RemoveBookmark(r.Context(), userID, novelID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to remove bookmark")
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetStats godoc
// @Summary Get bookmark stats
// @Description Get bookmark counts per list
// @Tags bookmarks
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.BookmarkListStats}
// @Router /bookmarks/stats [get]
func (h *BookmarkHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	stats, err := h.bookmarkService.GetStats(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get stats")
		return
	}
	
	response.JSON(w, http.StatusOK, stats)
}

// GetNovelStatus godoc
// @Summary Get novel bookmark status
// @Description Check if novel is in any bookmark list
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param novelId path string true "Novel ID"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /bookmarks/status/{novelId} [get]
func (h *BookmarkHandler) GetNovelStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.JSON(w, http.StatusOK, map[string]interface{}{"list": nil})
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.JSON(w, http.StatusOK, map[string]interface{}{"list": nil})
		return
	}
	
	novelIDStr := chi.URLParam(r, "novelId")
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid novel_id")
		return
	}
	
	listCode, err := h.bookmarkService.GetBookmarkStatus(r.Context(), userID, novelID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get status")
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"list": listCode,
	})
}

// parseIntQuery парсит query параметр в int
func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	
	return value
}
