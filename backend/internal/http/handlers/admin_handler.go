package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AdminHandler обработчик админских эндпоинтов
type AdminHandler struct {
	novelService   *service.NovelService
	chapterService *service.ChapterService
	uploadDir      string
}

// NewAdminHandler создает новый AdminHandler
func NewAdminHandler(
	novelService *service.NovelService,
	chapterService *service.ChapterService,
	uploadDir string,
) *AdminHandler {
	return &AdminHandler{
		novelService:   novelService,
		chapterService: chapterService,
		uploadDir:      uploadDir,
	}
}

// ========================
// NOVELS CRUD
// ========================

// CreateNovel создает новую новеллу
// POST /api/v1/admin/novels
func (h *AdminHandler) CreateNovel(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNovelRequest
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

	novel, err := h.novelService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrNovelSlugExists) {
			response.Conflict(w, "novel with this slug already exists")
			return
		}
		response.InternalError(w)
		return
	}

	response.Created(w, novel)
}

// UpdateNovel обновляет новеллу
// PUT /api/v1/admin/novels/{id}
func (h *AdminHandler) UpdateNovel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid novel id")
		return
	}

	var req models.UpdateNovelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.novelService.Update(r.Context(), id, &req); err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		if errors.Is(err, service.ErrNovelSlugExists) {
			response.Conflict(w, "novel with this slug already exists")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "novel updated"})
}

// DeleteNovel удаляет новеллу
// DELETE /api/v1/admin/novels/{id}
func (h *AdminHandler) DeleteNovel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid novel id")
		return
	}

	if err := h.novelService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "novel deleted"})
}

// ListChapters получает список всех глав (для админа)
// GET /api/v1/admin/chapters
func (h *AdminHandler) ListChapters(w http.ResponseWriter, r *http.Request) {
	// Пока возвращаем пустой массив - для полной реализации нужно добавить метод ListAll в ChapterService
	// который будет делать SELECT с JOIN на novels для получения названий
	response.OK(w, []interface{}{})
}

// ========================
// CHAPTERS CRUD
// ========================

// CreateChapter создает новую главу
// POST /api/v1/admin/chapters
func (h *AdminHandler) CreateChapter(w http.ResponseWriter, r *http.Request) {
	var req models.CreateChapterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Валидация
	if req.NovelID == uuid.Nil {
		response.BadRequest(w, "novel_id is required")
		return
	}
	if req.Number <= 0 {
		response.BadRequest(w, "chapter number must be positive")
		return
	}
	if len(req.Contents) == 0 {
		response.BadRequest(w, "at least one content is required")
		return
	}

	chapter, err := h.chapterService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrNovelNotFound) {
			response.NotFound(w, "novel not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.Created(w, chapter)
}

// UpdateChapter обновляет главу
// PUT /api/v1/admin/chapters/{id}
func (h *AdminHandler) UpdateChapter(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid chapter id")
		return
	}

	var req models.UpdateChapterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.chapterService.Update(r.Context(), id, &req); err != nil {
		if errors.Is(err, service.ErrChapterNotFound) {
			response.NotFound(w, "chapter not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "chapter updated"})
}

// DeleteChapter удаляет главу
// DELETE /api/v1/admin/chapters/{id}
func (h *AdminHandler) DeleteChapter(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid chapter id")
		return
	}

	if err := h.chapterService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrChapterNotFound) {
			response.NotFound(w, "chapter not found")
			return
		}
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "chapter deleted"})
}

// ========================
// UPLOADS
// ========================

// UploadCover загружает обложку новеллы
// POST /api/v1/admin/novels/{id}/cover
func (h *AdminHandler) UploadCover(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid novel id")
		return
	}

	// Парсим multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		response.BadRequest(w, "file too large")
		return
	}

	file, header, err := r.FormFile("cover")
	if err != nil {
		response.BadRequest(w, "cover file is required")
		return
	}
	defer file.Close()

	// Проверяем расширение
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		response.BadRequest(w, "only jpg, jpeg, png, webp files are allowed")
		return
	}

	// Создаем директорию для обложек
	coversDir := filepath.Join(h.uploadDir, "covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		response.InternalError(w)
		return
	}

	// Генерируем имя файла
	filename := id.String() + ext
	filePath := filepath.Join(coversDir, filename)

	// Сохраняем файл
	dst, err := os.Create(filePath)
	if err != nil {
		response.InternalError(w)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.InternalError(w)
		return
	}

	// Обновляем новеллу
	coverKey := "covers/" + filename
	req := &models.UpdateNovelRequest{
		CoverImage: &coverKey,
	}
	if err := h.novelService.Update(r.Context(), id, req); err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{
		"message":   "cover uploaded",
		"cover_url": "/uploads/" + coverKey,
	})
}

// Upload handles generic file uploads
// POST /api/v1/admin/upload
func (h *AdminHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Парсим multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		response.BadRequest(w, "file too large")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "file is required")
		return
	}
	defer file.Close()

	// Проверяем расширение
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowedExts[ext] {
		response.BadRequest(w, "only jpg, jpeg, png, webp, gif files are allowed")
		return
	}

	// Создаем директорию для общих загрузок
	uploadsDir := filepath.Join(h.uploadDir, "general")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		response.InternalError(w)
		return
	}

	// Генерируем имя файла
	filename := uuid.New().String() + ext
	filePath := filepath.Join(uploadsDir, filename)

	// Сохраняем файл
	dst, err := os.Create(filePath)
	if err != nil {
		response.InternalError(w)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.InternalError(w)
		return
	}

	fileKey := "general/" + filename
	response.OK(w, map[string]string{
		"message":  "file uploaded",
		"file_url": "/uploads/" + fileKey,
		"file_key": fileKey,
	})
}

// ========================
// BULK OPERATIONS
// ========================

// BulkCreateChapters создает несколько глав за раз
// POST /api/v1/admin/novels/{id}/chapters/bulk
func (h *AdminHandler) BulkCreateChapters(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	novelID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid novel id")
		return
	}

	var req struct {
		Chapters []models.CreateChapterRequest `json:"chapters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if len(req.Chapters) == 0 {
		response.BadRequest(w, "at least one chapter is required")
		return
	}

	created := 0
	var lastError error

	for _, chapterReq := range req.Chapters {
		chapterReq.NovelID = novelID
		_, err := h.chapterService.Create(r.Context(), &chapterReq)
		if err != nil {
			lastError = err
			continue
		}
		created++
	}

	if created == 0 && lastError != nil {
		response.InternalError(w)
		return
	}

	response.Created(w, map[string]interface{}{
		"created": created,
		"total":   len(req.Chapters),
	})
}
