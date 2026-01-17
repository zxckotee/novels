package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GenreTagAdminHandler обработчик админских эндпоинтов для жанров и тегов
type GenreTagAdminHandler struct {
	genreService *service.GenreService
	tagService   *service.TagService
}

// NewGenreTagAdminHandler создает новый GenreTagAdminHandler
func NewGenreTagAdminHandler(genreService *service.GenreService, tagService *service.TagService) *GenreTagAdminHandler {
	return &GenreTagAdminHandler{
		genreService: genreService,
		tagService:   tagService,
	}
}

// ========================
// GENRES
// ========================

// ListGenres получает список жанров
// GET /api/v1/admin/genres
func (h *GenreTagAdminHandler) ListGenres(w http.ResponseWriter, r *http.Request) {
	filter := models.GenresFilter{
		Query: r.URL.Query().Get("query"),
		Lang:  r.URL.Query().Get("lang"),
		Sort:  r.URL.Query().Get("sort"),
		Order: r.URL.Query().Get("order"),
		Page:  parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 50),
	}

	result, err := h.genreService.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, result)
}

// GetGenre получает жанр по ID
// GET /api/v1/admin/genres/{id}
func (h *GenreTagAdminHandler) GetGenre(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid genre id")
		return
	}

	genre, err := h.genreService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "genre not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, genre)
}

// CreateGenre создает новый жанр
// POST /api/v1/admin/genres
func (h *GenreTagAdminHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var req models.CreateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Slug == "" || len(req.Localizations) == 0 {
		response.BadRequest(w, "slug and localizations are required")
		return
	}

	genre, err := h.genreService.Create(r.Context(), &req)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.Created(w, genre)
}

// UpdateGenre обновляет жанр
// PUT /api/v1/admin/genres/{id}
func (h *GenreTagAdminHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid genre id")
		return
	}

	var req models.UpdateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	genre, err := h.genreService.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "genre not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, genre)
}

// DeleteGenre удаляет жанр
// DELETE /api/v1/admin/genres/{id}
func (h *GenreTagAdminHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid genre id")
		return
	}

	err = h.genreService.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "genre not found")
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, map[string]string{"message": "genre deleted"})
}

// ========================
// TAGS
// ========================

// ListTags получает список тегов
// GET /api/v1/admin/tags
func (h *GenreTagAdminHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	filter := models.TagsFilter{
		Query: r.URL.Query().Get("query"),
		Lang:  r.URL.Query().Get("lang"),
		Sort:  r.URL.Query().Get("sort"),
		Order: r.URL.Query().Get("order"),
		Page:  parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 50),
	}

	result, err := h.tagService.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, result)
}

// GetTag получает тег по ID
// GET /api/v1/admin/tags/{id}
func (h *GenreTagAdminHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid tag id")
		return
	}

	tag, err := h.tagService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "tag not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, tag)
}

// CreateTag создает новый тег
// POST /api/v1/admin/tags
func (h *GenreTagAdminHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Slug == "" || len(req.Localizations) == 0 {
		response.BadRequest(w, "slug and localizations are required")
		return
	}

	tag, err := h.tagService.Create(r.Context(), &req)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.Created(w, tag)
}

// UpdateTag обновляет тег
// PUT /api/v1/admin/tags/{id}
func (h *GenreTagAdminHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid tag id")
		return
	}

	var req models.UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	tag, err := h.tagService.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "tag not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, tag)
}

// DeleteTag удаляет тег
// DELETE /api/v1/admin/tags/{id}
func (h *GenreTagAdminHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid tag id")
		return
	}

	err = h.tagService.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "tag not found")
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, map[string]string{"message": "tag deleted"})
}
