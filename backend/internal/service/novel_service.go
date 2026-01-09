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
	ErrNovelNotFound = errors.New("novel not found")
	ErrNovelSlugExists = errors.New("novel with this slug already exists")
)

// NovelService сервис для работы с новеллами
type NovelService struct {
	novelRepo *repository.NovelRepository
}

// NewNovelService создает новый NovelService
func NewNovelService(novelRepo *repository.NovelRepository) *NovelService {
	return &NovelService{
		novelRepo: novelRepo,
	}
}

// List получает список новелл с фильтрацией и пагинацией
func (s *NovelService) List(ctx context.Context, params models.NovelListParams) (*models.NovelListResponse, error) {
	// Устанавливаем значения по умолчанию
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Lang == "" {
		params.Lang = "ru"
	}
	if params.Sort == "" {
		params.Sort = "created_at"
	}
	if params.Order == "" {
		params.Order = "desc"
	}

	novels, total, err := s.novelRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list novels: %w", err)
	}

	totalPages := total / params.Limit
	if total%params.Limit > 0 {
		totalPages++
	}

	return &models.NovelListResponse{
		Novels: novels,
		Pagination: models.Pagination{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetBySlug получает новеллу по slug
func (s *NovelService) GetBySlug(ctx context.Context, slug, lang string) (*models.NovelDetail, error) {
	if lang == "" {
		lang = "ru"
	}

	novel, err := s.novelRepo.GetBySlug(ctx, slug, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel: %w", err)
	}
	if novel == nil {
		return nil, ErrNovelNotFound
	}

	return novel, nil
}

// GetByID получает новеллу по ID
func (s *NovelService) GetByID(ctx context.Context, id uuid.UUID, lang string) (*models.NovelDetail, error) {
	if lang == "" {
		lang = "ru"
	}

	novel, err := s.novelRepo.GetByID(ctx, id, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel: %w", err)
	}
	if novel == nil {
		return nil, ErrNovelNotFound
	}

	return novel, nil
}

// Search поиск новелл по ключевым словам
func (s *NovelService) Search(ctx context.Context, query string, params models.NovelListParams) (*models.NovelListResponse, error) {
	// Устанавливаем значения по умолчанию
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Lang == "" {
		params.Lang = "ru"
	}

	novels, total, err := s.novelRepo.Search(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search novels: %w", err)
	}

	totalPages := total / params.Limit
	if total%params.Limit > 0 {
		totalPages++
	}

	return &models.NovelListResponse{
		Novels: novels,
		Pagination: models.Pagination{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// Create создает новую новеллу (админ)
func (s *NovelService) Create(ctx context.Context, req *models.CreateNovelRequest) (*models.Novel, error) {
	// Проверяем уникальность slug
	existing, err := s.novelRepo.GetBySlug(ctx, req.Slug, "ru")
	if err != nil {
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}
	if existing != nil {
		return nil, ErrNovelSlugExists
	}

	novel, err := s.novelRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create novel: %w", err)
	}

	return novel, nil
}

// Update обновляет новеллу (админ)
func (s *NovelService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateNovelRequest) error {
	// Проверяем существование
	existing, err := s.novelRepo.GetByID(ctx, id, "ru")
	if err != nil {
		return fmt.Errorf("failed to get novel: %w", err)
	}
	if existing == nil {
		return ErrNovelNotFound
	}

	// Проверяем уникальность slug, если он меняется
	if req.Slug != nil && *req.Slug != existing.Slug {
		slugExists, err := s.novelRepo.GetBySlug(ctx, *req.Slug, "ru")
		if err != nil {
			return fmt.Errorf("failed to check slug: %w", err)
		}
		if slugExists != nil {
			return ErrNovelSlugExists
		}
	}

	if err := s.novelRepo.Update(ctx, id, req); err != nil {
		return fmt.Errorf("failed to update novel: %w", err)
	}

	return nil
}

// Delete удаляет новеллу (админ)
func (s *NovelService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.novelRepo.GetByID(ctx, id, "ru")
	if err != nil {
		return fmt.Errorf("failed to get novel: %w", err)
	}
	if existing == nil {
		return ErrNovelNotFound
	}

	if err := s.novelRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete novel: %w", err)
	}

	return nil
}

// GetPopular получает популярные новеллы
func (s *NovelService) GetPopular(ctx context.Context, lang string, limit int) ([]models.NovelCard, error) {
	if limit <= 0 {
		limit = 10
	}
	if lang == "" {
		lang = "ru"
	}

	params := models.NovelListParams{
		Lang:  lang,
		Limit: limit,
		Page:  1,
		Sort:  "views_daily",
		Order: "desc",
	}

	novels, _, err := s.novelRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular novels: %w", err)
	}

	return novels, nil
}

// GetLatestUpdates получает последние обновления
func (s *NovelService) GetLatestUpdates(ctx context.Context, lang string, limit int) ([]models.NovelCard, error) {
	if limit <= 0 {
		limit = 20
	}
	if lang == "" {
		lang = "ru"
	}

	params := models.NovelListParams{
		Lang:  lang,
		Limit: limit,
		Page:  1,
		Sort:  "updated_at",
		Order: "desc",
	}

	novels, _, err := s.novelRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest updates: %w", err)
	}

	return novels, nil
}

// GetNewReleases получает новинки
func (s *NovelService) GetNewReleases(ctx context.Context, lang string, limit int) ([]models.NovelCard, error) {
	if limit <= 0 {
		limit = 10
	}
	if lang == "" {
		lang = "ru"
	}

	params := models.NovelListParams{
		Lang:  lang,
		Limit: limit,
		Page:  1,
		Sort:  "created_at",
		Order: "desc",
	}

	novels, _, err := s.novelRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get new releases: %w", err)
	}

	return novels, nil
}

// GetTrending получает трендовые новеллы (по росту просмотров)
func (s *NovelService) GetTrending(ctx context.Context, lang string, limit int) ([]models.NovelCard, error) {
	if limit <= 0 {
		limit = 10
	}
	if lang == "" {
		lang = "ru"
	}

	// Для трендов используем views_daily как основную метрику
	// В будущем можно добавить более сложную логику
	params := models.NovelListParams{
		Lang:  lang,
		Limit: limit,
		Page:  1,
		Sort:  "views_daily",
		Order: "desc",
	}

	novels, _, err := s.novelRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending novels: %w", err)
	}

	return novels, nil
}

// GetTopRated получает новеллы с лучшим рейтингом
func (s *NovelService) GetTopRated(ctx context.Context, lang string, limit int) ([]models.NovelCard, error) {
	if limit <= 0 {
		limit = 10
	}
	if lang == "" {
		lang = "ru"
	}

	params := models.NovelListParams{
		Lang:  lang,
		Limit: limit,
		Page:  1,
		Sort:  "rating",
		Order: "desc",
	}

	novels, _, err := s.novelRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get top rated novels: %w", err)
	}

	return novels, nil
}

// Rate добавляет оценку новелле
func (s *NovelService) Rate(ctx context.Context, userID, novelID uuid.UUID, value int) error {
	if value < 1 || value > 5 {
		return errors.New("rating value must be between 1 and 5")
	}

	// Проверяем существование новеллы
	existing, err := s.novelRepo.GetByID(ctx, novelID, "ru")
	if err != nil {
		return fmt.Errorf("failed to get novel: %w", err)
	}
	if existing == nil {
		return ErrNovelNotFound
	}

	if err := s.novelRepo.AddRating(ctx, userID, novelID, value); err != nil {
		return fmt.Errorf("failed to add rating: %w", err)
	}

	return nil
}

// GetUserRating получает оценку пользователя для новеллы
func (s *NovelService) GetUserRating(ctx context.Context, userID, novelID uuid.UUID) (int, error) {
	rating, err := s.novelRepo.GetUserRating(ctx, userID, novelID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user rating: %w", err)
	}
	return rating, nil
}

// IncrementViews увеличивает счетчик просмотров
func (s *NovelService) IncrementViews(ctx context.Context, novelID uuid.UUID) error {
	if err := s.novelRepo.IncrementViews(ctx, novelID); err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}
	return nil
}

// GetAllGenres получает все жанры
func (s *NovelService) GetAllGenres(ctx context.Context, lang string) ([]models.Genre, error) {
	if lang == "" {
		lang = "ru"
	}
	return s.novelRepo.GetAllGenres(ctx, lang)
}

// GetAllTags получает все теги
func (s *NovelService) GetAllTags(ctx context.Context, lang string) ([]models.Tag, error) {
	if lang == "" {
		lang = "ru"
	}
	return s.novelRepo.GetAllTags(ctx, lang)
}
