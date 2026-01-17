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

// AuthorAdminHandler обработчик админских эндпоинтов для авторов
type AuthorAdminHandler struct {
	authorService *service.AuthorService
}

// NewAuthorAdminHandler создает новый AuthorAdminHandler
func NewAuthorAdminHandler(authorService *service.AuthorService) *AuthorAdminHandler {
	return &AuthorAdminHandler{
		authorService: authorService,
	}
}

// ListAuthors получает список авторов
// GET /api/v1/admin/authors
func (h *AuthorAdminHandler) ListAuthors(w http.ResponseWriter, r *http.Request) {
	filter := models.AuthorsFilter{
		Query: r.URL.Query().Get("query"),
		Lang:  r.URL.Query().Get("lang"),
		Sort:  r.URL.Query().Get("sort"),
		Order: r.URL.Query().Get("order"),
		Page:  parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 20),
	}

	result, err := h.authorService.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, result)
}

// GetAuthor получает автора по ID
// GET /api/v1/admin/authors/{id}
func (h *AuthorAdminHandler) GetAuthor(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid author id")
		return
	}

	author, err := h.authorService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "author not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, author)
}

// CreateAuthor создает нового автора
// POST /api/v1/admin/authors
func (h *AuthorAdminHandler) CreateAuthor(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Валидация
	if req.Slug == "" {
		response.BadRequest(w, "slug is required")
		return
	}
	if len(req.Localizations) == 0 {
		response.BadRequest(w, "at least one localization is required")
		return
	}

	author, err := h.authorService.Create(r.Context(), &req)
	if err != nil {
		if err.Error() == "author with slug '"+req.Slug+"' already exists" {
			response.Conflict(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	response.Created(w, author)
}

// UpdateAuthor обновляет автора
// PUT /api/v1/admin/authors/{id}
func (h *AuthorAdminHandler) UpdateAuthor(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid author id")
		return
	}

	var req models.UpdateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	author, err := h.authorService.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "author not found")
			return
		}
		if err.Error() == "author with slug '"+req.Slug+"' already exists" {
			response.Conflict(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, author)
}

// DeleteAuthor удаляет автора
// DELETE /api/v1/admin/authors/{id}
func (h *AuthorAdminHandler) DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid author id")
		return
	}

	if err := h.authorService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "author not found")
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, map[string]string{"message": "author deleted"})
}

// GetNovelAuthors получает авторов новеллы
// GET /api/v1/admin/novels/{id}/authors
func (h *AuthorAdminHandler) GetNovelAuthors(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	novelID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid novel id")
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ru"
	}

	authors, err := h.authorService.GetNovelAuthors(r.Context(), novelID, lang)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, authors)
}

// UpdateNovelAuthors обновляет авторов новеллы
// PUT /api/v1/admin/novels/{id}/authors
func (h *AuthorAdminHandler) UpdateNovelAuthors(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	novelID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid novel id")
		return
	}

	var req models.UpdateNovelAuthorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.authorService.UpdateNovelAuthors(r.Context(), novelID, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, map[string]string{"message": "novel authors updated"})
}
