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

// NovelHandler обработчик эндпоинтов новелл
type NovelHandler struct {
	novelService *service.NovelService
}

// NewNovelHandler создает новый NovelHandler
func NewNovelHandler(novelService *service.NovelService) *NovelHandler {
	return &NovelHandler{
		novelService: novelService,
	}
}

// List получает список новелл
// GET /api/v1/novels
func (h *NovelHandler) List(w http.ResponseWriter, r *http.Request) {
	params := h.parseListParams(r)

	result, err := h.novelService.List(r.Context(), params)
	if err != nil {
		response.InternalError(w, "failed to get novels")
		return
	}

	response.OK(w, result)
}

// Search поиск новелл
// GET /api/v1/novels/search
func (h *NovelHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		response.BadRequest(w, "search query is required")
		return
	}

	params := h.parseListParams(r)

	result, err := h.novelService.Search(r.Context(), query, params)
	if err != nil {
		response.InternalError(w, "failed to search novels")
		return
	}

	response.OK(w, result)
}

// GetBySlug получает новеллу по slug
// GET /api/v1/novels/{slug}
func (h *NovelHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ru"
	}

	novel, err := h.novelService.GetBySlug(r.Context(), slug, lang)
	if err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		response.InternalError(w, "failed to get novel")
		return
	}

	// Увеличиваем счетчик просмотров
	_ = h.novelService.IncrementViews(r.Context(), novel.ID)

	response.OK(w, novel)
}

// GetPopular получает популярные новеллы
// GET /api/v1/novels/popular
func (h *NovelHandler) GetPopular(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	limit := h.parseLimit(r, 10)

	novels, err := h.novelService.GetPopular(r.Context(), lang, limit)
	if err != nil {
		response.InternalError(w, "failed to get popular novels")
		return
	}

	response.OK(w, novels)
}

// GetLatestUpdates получает последние обновления
// GET /api/v1/novels/latest
func (h *NovelHandler) GetLatestUpdates(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	limit := h.parseLimit(r, 20)

	novels, err := h.novelService.GetLatestUpdates(r.Context(), lang, limit)
	if err != nil {
		response.InternalError(w, "failed to get latest updates")
		return
	}

	response.OK(w, novels)
}

// GetNewReleases получает новинки
// GET /api/v1/novels/new
func (h *NovelHandler) GetNewReleases(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	limit := h.parseLimit(r, 10)

	novels, err := h.novelService.GetNewReleases(r.Context(), lang, limit)
	if err != nil {
		response.InternalError(w, "failed to get new releases")
		return
	}

	response.OK(w, novels)
}

// GetTrending получает трендовые новеллы
// GET /api/v1/novels/trending
func (h *NovelHandler) GetTrending(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	limit := h.parseLimit(r, 10)

	novels, err := h.novelService.GetTrending(r.Context(), lang, limit)
	if err != nil {
		response.InternalError(w, "failed to get trending novels")
		return
	}

	response.OK(w, novels)
}

// GetTopRated получает новеллы с лучшим рейтингом
// GET /api/v1/novels/top-rated
func (h *NovelHandler) GetTopRated(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	limit := h.parseLimit(r, 10)

	novels, err := h.novelService.GetTopRated(r.Context(), lang, limit)
	if err != nil {
		response.InternalError(w, "failed to get top rated novels")
		return
	}

	response.OK(w, novels)
}

// Rate добавляет оценку новелле
// POST /api/v1/novels/{slug}/rate
func (h *NovelHandler) Rate(w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(r.Context())
	if user == nil {
		response.Unauthorized(w, "not authenticated")
		return
	}

	slug := chi.URLParam(r, "slug")

	var req struct {
		Value int `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Value < 1 || req.Value > 5 {
		response.BadRequest(w, "rating value must be between 1 and 5")
		return
	}

	// Получаем ID новеллы
	novel, err := h.novelService.GetBySlug(r.Context(), slug, "ru")
	if err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		response.InternalError(w, "failed to get novel")
		return
	}

	if err := h.novelService.Rate(r.Context(), user.ID, novel.ID, req.Value); err != nil {
		response.InternalError(w, "failed to rate novel")
		return
	}

	response.OK(w, map[string]string{"message": "rated successfully"})
}

// GetUserRating получает оценку пользователя
// GET /api/v1/novels/{slug}/my-rating
func (h *NovelHandler) GetUserRating(w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(r.Context())
	if user == nil {
		response.Unauthorized(w, "not authenticated")
		return
	}

	slug := chi.URLParam(r, "slug")

	novel, err := h.novelService.GetBySlug(r.Context(), slug, "ru")
	if err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		response.InternalError(w, "failed to get novel")
		return
	}

	rating, err := h.novelService.GetUserRating(r.Context(), user.ID, novel.ID)
	if err != nil {
		response.InternalError(w, "failed to get rating")
		return
	}

	response.OK(w, map[string]int{"rating": rating})
}

// GetGenres получает все жанры
// GET /api/v1/genres
func (h *NovelHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")

	genres, err := h.novelService.GetAllGenres(r.Context(), lang)
	if err != nil {
		response.InternalError(w, "failed to get genres")
		return
	}

	response.OK(w, genres)
}

// GetTags получает все теги
// GET /api/v1/tags
func (h *NovelHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")

	tags, err := h.novelService.GetAllTags(r.Context(), lang)
	if err != nil {
		response.InternalError(w, "failed to get tags")
		return
	}

	response.OK(w, tags)
}

// parseListParams парсит параметры списка из запроса
func (h *NovelHandler) parseListParams(r *http.Request) models.NovelListParams {
	params := models.NovelListParams{
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

	// Парсинг жанров
	if genres := r.URL.Query().Get("genres"); genres != "" {
		for _, g := range splitAndTrim(genres) {
			if id, err := uuid.Parse(g); err == nil {
				params.GenreIDs = append(params.GenreIDs, id)
			}
		}
	}

	// Парсинг тегов
	if tags := r.URL.Query().Get("tags"); tags != "" {
		for _, t := range splitAndTrim(tags) {
			if id, err := uuid.Parse(t); err == nil {
				params.TagIDs = append(params.TagIDs, id)
			}
		}
	}

	// Статус перевода
	if status := r.URL.Query().Get("status"); status != "" {
		params.TranslationStatus = &status
	}

	return params
}

// parseLimit парсит параметр limit
func (h *NovelHandler) parseLimit(r *http.Request, defaultLimit int) int {
	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		return limit
	}
	return defaultLimit
}

// splitAndTrim разделяет строку и удаляет пробелы
func splitAndTrim(s string) []string {
	var result []string
	for _, part := range splitString(s, ",") {
		if trimmed := trimString(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
