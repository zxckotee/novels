package service

import (
	"context"
	"fmt"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
)

// AuthorRepository интерфейс для работы с авторами
type AuthorRepository interface {
	List(ctx context.Context, filter models.AuthorsFilter) ([]models.Author, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Author, error)
	GetBySlug(ctx context.Context, slug string) (*models.Author, error)
	Create(ctx context.Context, req *models.CreateAuthorRequest) (*models.Author, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateAuthorRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetNovelAuthors(ctx context.Context, novelID uuid.UUID, lang string) ([]models.NovelAuthor, error)
	UpdateNovelAuthors(ctx context.Context, novelID uuid.UUID, authors []models.NovelAuthorInput) error
}

// AuthorService сервис для работы с авторами
type AuthorService struct {
	authorRepo AuthorRepository
}

// NewAuthorService создает новый AuthorService
func NewAuthorService(authorRepo AuthorRepository) *AuthorService {
	return &AuthorService{
		authorRepo: authorRepo,
	}
}

// List получает список авторов
func (s *AuthorService) List(ctx context.Context, filter models.AuthorsFilter) (*models.AuthorsResponse, error) {
	// Устанавливаем дефолтные значения
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Lang == "" {
		filter.Lang = "ru"
	}

	authors, total, err := s.authorRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list authors: %w", err)
	}

	return &models.AuthorsResponse{
		Authors:    authors,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

// GetByID получает автора по ID
func (s *AuthorService) GetByID(ctx context.Context, id uuid.UUID) (*models.Author, error) {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	if author == nil {
		return nil, ErrNotFound
	}
	return author, nil
}

// GetBySlug получает автора по slug
func (s *AuthorService) GetBySlug(ctx context.Context, slug string) (*models.Author, error) {
	author, err := s.authorRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	if author == nil {
		return nil, ErrNotFound
	}
	return author, nil
}

// Create создает нового автора
func (s *AuthorService) Create(ctx context.Context, req *models.CreateAuthorRequest) (*models.Author, error) {
	// Проверяем, что slug уникален
	existing, err := s.authorRepo.GetBySlug(ctx, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("author with slug '%s' already exists", req.Slug)
	}

	// Создаем автора
	author, err := s.authorRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create author: %w", err)
	}

	return author, nil
}

// Update обновляет автора
func (s *AuthorService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateAuthorRequest) (*models.Author, error) {
	// Проверяем, что автор существует
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	if author == nil {
		return nil, ErrNotFound
	}

	// Если изменяется slug, проверяем уникальность
	if req.Slug != "" && req.Slug != author.Slug {
		existing, err := s.authorRepo.GetBySlug(ctx, req.Slug)
		if err != nil {
			return nil, fmt.Errorf("failed to check slug: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("author with slug '%s' already exists", req.Slug)
		}
	}

	// Обновляем автора
	err = s.authorRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update author: %w", err)
	}

	// Возвращаем обновленного автора
	return s.authorRepo.GetByID(ctx, id)
}

// Delete удаляет автора
func (s *AuthorService) Delete(ctx context.Context, id uuid.UUID) error {
	// Проверяем, что автор существует
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get author: %w", err)
	}
	if author == nil {
		return ErrNotFound
	}

	// Удаляем автора
	err = s.authorRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete author: %w", err)
	}

	return nil
}

// GetNovelAuthors получает авторов новеллы
func (s *AuthorService) GetNovelAuthors(ctx context.Context, novelID uuid.UUID, lang string) ([]models.NovelAuthor, error) {
	if lang == "" {
		lang = "ru"
	}

	authors, err := s.authorRepo.GetNovelAuthors(ctx, novelID, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel authors: %w", err)
	}

	return authors, nil
}

// UpdateNovelAuthors обновляет авторов новеллы
func (s *AuthorService) UpdateNovelAuthors(ctx context.Context, novelID uuid.UUID, req *models.UpdateNovelAuthorsRequest) error {
	// Проверяем, что все авторы существуют
	for _, author := range req.Authors {
		authorID, err := uuid.Parse(author.AuthorID)
		if err != nil {
			return fmt.Errorf("invalid author ID '%s': %w", author.AuthorID, err)
		}

		existing, err := s.authorRepo.GetByID(ctx, authorID)
		if err != nil {
			return fmt.Errorf("failed to check author: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("author with ID '%s' not found", author.AuthorID)
		}
	}

	// Обновляем авторов
	err := s.authorRepo.UpdateNovelAuthors(ctx, novelID, req.Authors)
	if err != nil {
		return fmt.Errorf("failed to update novel authors: %w", err)
	}

	return nil
}
