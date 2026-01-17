package service

import (
	"context"
	"fmt"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
)

// GenreRepository интерфейс для работы с жанрами
type GenreRepository interface {
	List(ctx context.Context, filter models.GenresFilter) ([]models.GenreWithLocalizations, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.GenreWithLocalizations, error)
	Create(ctx context.Context, req *models.CreateGenreRequest) (*models.GenreWithLocalizations, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateGenreRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TagRepository интерфейс для работы с тегами
type TagRepository interface {
	List(ctx context.Context, filter models.TagsFilter) ([]models.TagWithLocalizations, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.TagWithLocalizations, error)
	Create(ctx context.Context, req *models.CreateTagRequest) (*models.TagWithLocalizations, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateTagRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GenreService сервис для работы с жанрами
type GenreService struct {
	genreRepo GenreRepository
}

func NewGenreService(genreRepo GenreRepository) *GenreService {
	return &GenreService{genreRepo: genreRepo}
}

func (s *GenreService) List(ctx context.Context, filter models.GenresFilter) (*models.GenresResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Lang == "" {
		filter.Lang = "ru"
	}

	genres, total, err := s.genreRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list genres: %w", err)
	}

	return &models.GenresResponse{
		Genres:     genres,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

func (s *GenreService) GetByID(ctx context.Context, id uuid.UUID) (*models.GenreWithLocalizations, error) {
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}
	if genre == nil {
		return nil, ErrNotFound
	}
	return genre, nil
}

func (s *GenreService) Create(ctx context.Context, req *models.CreateGenreRequest) (*models.GenreWithLocalizations, error) {
	genre, err := s.genreRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create genre: %w", err)
	}
	return genre, nil
}

func (s *GenreService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateGenreRequest) (*models.GenreWithLocalizations, error) {
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}
	if genre == nil {
		return nil, ErrNotFound
	}

	err = s.genreRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update genre: %w", err)
	}

	return s.genreRepo.GetByID(ctx, id)
}

func (s *GenreService) Delete(ctx context.Context, id uuid.UUID) error {
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get genre: %w", err)
	}
	if genre == nil {
		return ErrNotFound
	}

	return s.genreRepo.Delete(ctx, id)
}

// TagService сервис для работы с тегами
type TagService struct {
	tagRepo TagRepository
}

func NewTagService(tagRepo TagRepository) *TagService {
	return &TagService{tagRepo: tagRepo}
}

func (s *TagService) List(ctx context.Context, filter models.TagsFilter) (*models.TagsResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Lang == "" {
		filter.Lang = "ru"
	}

	tags, total, err := s.tagRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return &models.TagsResponse{
		Tags:       tags,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

func (s *TagService) GetByID(ctx context.Context, id uuid.UUID) (*models.TagWithLocalizations, error) {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	if tag == nil {
		return nil, ErrNotFound
	}
	return tag, nil
}

func (s *TagService) Create(ctx context.Context, req *models.CreateTagRequest) (*models.TagWithLocalizations, error) {
	tag, err := s.tagRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}
	return tag, nil
}

func (s *TagService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateTagRequest) (*models.TagWithLocalizations, error) {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	if tag == nil {
		return nil, ErrNotFound
	}

	err = s.tagRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return s.tagRepo.GetByID(ctx, id)
}

func (s *TagService) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get tag: %w", err)
	}
	if tag == nil {
		return ErrNotFound
	}

	return s.tagRepo.Delete(ctx, id)
}
