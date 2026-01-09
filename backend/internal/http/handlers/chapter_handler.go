package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ChapterHandler обработчик эндпоинтов глав
type ChapterHandler struct {
	chapterService *service.ChapterService
}

// NewChapterHandler создает новый ChapterHandler
func NewChapterHandler(chapterService *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{
		chapterService: chapterService,
	}
}

// ListByNovel получает список глав новеллы
// GET /api/v1/novels/{slug}/chapters
func (h *ChapterHandler) ListByNovel(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	params := h.parseListParams(r)

	result, err := h.chapterService.ListByNovel(r.Context(), slug, params)
	if err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		response.InternalError(w, "failed to get chapters")
		return
	}

	response.OK(w, result)
}

// GetByID получает главу по ID
// GET /api/v1/chapters/{id}
func (h *ChapterHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid chapter id")
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ru"
	}

	// Получаем пользователя если авторизован
	var userID *uuid.UUID
	user := models.GetUserFromContext(r.Context())
	if user != nil {
		userID = &user.ID
	}

	chapter, err := h.chapterService.GetForReader(r.Context(), id, lang, userID)
	if err != nil {
		if errors.Is(err, service.ErrChapterNotFound) {
			response.NotFound(w, "chapter not found")
			return
		}
		response.InternalError(w, "failed to get chapter")
		return
	}

	response.OK(w, chapter)
}

// SaveProgress сохраняет прогресс чтения
// POST /api/v1/chapters/{id}/progress
func (h *ChapterHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(r.Context())
	if user == nil {
		response.Unauthorized(w, "not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	chapterID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid chapter id")
		return
	}

	var req struct {
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.chapterService.SaveProgress(r.Context(), user.ID, chapterID, req.Position); err != nil {
		response.InternalError(w, "failed to save progress")
		return
	}

	response.OK(w, map[string]string{"message": "progress saved"})
}

// GetProgress получает прогресс чтения для новеллы
// GET /api/v1/novels/{slug}/progress
func (h *ChapterHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(r.Context())
	if user == nil {
		response.Unauthorized(w, "not authenticated")
		return
	}

	slug := chi.URLParam(r, "slug")

	progress, err := h.chapterService.GetProgress(r.Context(), user.ID, slug)
	if err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		response.InternalError(w, "failed to get progress")
		return
	}

	if progress == nil {
		response.OK(w, map[string]interface{}{"progress": nil})
		return
	}

	response.OK(w, map[string]interface{}{"progress": progress})
}

// parseListParams парсит параметры списка
func (h *ChapterHandler) parseListParams(r *http.Request) models.ChapterListParams {
	params := models.ChapterListParams{
		Lang:  r.URL.Query().Get("lang"),
		Sort:  r.URL.Query().Get("sort"),
		Order: r.URL.Query().Get("order"),
	}

	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
		params.Page = page
	}
	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		params.Limit = limit
	}

	return params
}
