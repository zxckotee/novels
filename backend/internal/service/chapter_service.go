package service

import (
	"context"
	"errors"
	"fmt"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrChapterNotFound = errors.New("chapter not found")
	ErrChapterExists   = errors.New("chapter with this number already exists")
)

// ChapterService сервис для работы с главами
type ChapterService struct {
	chapterRepo  *repository.ChapterRepository
	novelRepo    *repository.NovelRepository
	progressRepo *repository.ProgressRepository
}

// NewChapterService создает новый ChapterService
func NewChapterService(
	chapterRepo *repository.ChapterRepository,
	novelRepo *repository.NovelRepository,
	progressRepo *repository.ProgressRepository,
) *ChapterService {
	return &ChapterService{
		chapterRepo:  chapterRepo,
		novelRepo:    novelRepo,
		progressRepo: progressRepo,
	}
}

// ListByNovel получает список глав новеллы
func (s *ChapterService) ListByNovel(ctx context.Context, novelSlug string, params models.ChapterListParams) (*models.ChaptersListResponse, error) {
	// Устанавливаем значения по умолчанию
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Sort == "" {
		params.Sort = "number"
	}
	if params.Order == "" {
		params.Order = "asc"
	}

	chapters, novel, total, err := s.chapterRepo.ListByNovel(ctx, novelSlug, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list chapters: %w", err)
	}

	if novel == nil {
		return nil, ErrNovelNotFound
	}

	totalPages := 0
	if params.Limit > 0 {
		totalPages = (total + params.Limit - 1) / params.Limit
	}
	return &models.ChaptersListResponse{
		Novel:    novel,
		Chapters: chapters,
		Pagination: models.Pagination{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetByID получает главу по ID
func (s *ChapterService) GetByID(ctx context.Context, id uuid.UUID, lang string) (*models.ChapterWithContent, error) {
	if lang == "" {
		lang = "ru"
	}

	chapter, err := s.chapterRepo.GetByID(ctx, id, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}
	if chapter == nil {
		return nil, ErrChapterNotFound
	}

	return chapter, nil
}

// GetForReader получает главу для чтения (с увеличением просмотров)
func (s *ChapterService) GetForReader(ctx context.Context, id uuid.UUID, lang string, userID *uuid.UUID) (*models.ChapterWithContent, error) {
	chapter, err := s.GetByID(ctx, id, lang)
	if err != nil {
		return nil, err
	}

	// Увеличиваем счетчик просмотров
	_ = s.chapterRepo.IncrementViews(ctx, id)

	// Сохраняем прогресс если пользователь авторизован
	if userID != nil {
		progress := &models.ReadingProgress{
			UserID:    *userID,
			NovelID:   chapter.NovelID,
			ChapterID: chapter.ID,
			Position:  0,
		}
		_ = s.progressRepo.Save(ctx, progress)
	}

	return chapter, nil
}

// Create создает новую главу (админ)
func (s *ChapterService) Create(ctx context.Context, req *models.CreateChapterRequest) (*models.Chapter, error) {
	// Проверяем существование новеллы
	novel, err := s.novelRepo.GetByID(ctx, req.NovelID, "ru")
	if err != nil {
		return nil, fmt.Errorf("failed to get novel: %w", err)
	}
	if novel == nil {
		return nil, ErrNovelNotFound
	}

	chapter, err := s.chapterRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chapter: %w", err)
	}

	return chapter, nil
}

// Update обновляет главу (админ)
func (s *ChapterService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateChapterRequest) error {
	// Проверяем существование главы
	existing, err := s.chapterRepo.GetByID(ctx, id, "ru")
	if err != nil {
		return fmt.Errorf("failed to get chapter: %w", err)
	}
	if existing == nil {
		return ErrChapterNotFound
	}

	if err := s.chapterRepo.Update(ctx, id, req); err != nil {
		return fmt.Errorf("failed to update chapter: %w", err)
	}

	return nil
}

// Delete удаляет главу (админ)
func (s *ChapterService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.chapterRepo.GetByID(ctx, id, "ru")
	if err != nil {
		return fmt.Errorf("failed to get chapter: %w", err)
	}
	if existing == nil {
		return ErrChapterNotFound
	}

	if err := s.chapterRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}

	return nil
}

// SaveProgress сохраняет позицию чтения
func (s *ChapterService) SaveProgress(ctx context.Context, userID, chapterID uuid.UUID, position int) error {
	// Получаем novel_id для главы
	novelID, err := s.chapterRepo.GetNovelIDByChapter(ctx, chapterID)
	if err != nil {
		return fmt.Errorf("failed to get novel id: %w", err)
	}

	progress := &models.ReadingProgress{
		UserID:    userID,
		NovelID:   novelID,
		ChapterID: chapterID,
		Position:  position,
	}

	if err := s.progressRepo.Save(ctx, progress); err != nil {
		return fmt.Errorf("failed to save progress: %w", err)
	}

	return nil
}

// GetProgress получает прогресс чтения для новеллы
func (s *ChapterService) GetProgress(ctx context.Context, userID uuid.UUID, novelSlug string) (*models.ReadingProgressWithChapter, error) {
	// Получаем ID новеллы по slug
	novel, err := s.novelRepo.GetBySlug(ctx, novelSlug, "ru")
	if err != nil {
		return nil, fmt.Errorf("failed to get novel: %w", err)
	}
	if novel == nil {
		return nil, ErrNovelNotFound
	}

	progress, err := s.progressRepo.Get(ctx, userID, novel.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}

	return progress, nil
}
