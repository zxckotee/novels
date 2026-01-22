package service

import (
	"context"
	"fmt"
	"time"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

// NewsService handles news business logic
type NewsService struct {
	newsRepo *repository.NewsRepository
	userRepo *repository.UserRepository
}

// NewNewsService creates a new news service
func NewNewsService(
	newsRepo *repository.NewsRepository,
	userRepo *repository.UserRepository,
) *NewsService {
	return &NewsService{
		newsRepo: newsRepo,
		userRepo: userRepo,
	}
}

// Create creates a new news post (admin/moderator only)
func (s *NewsService) Create(ctx context.Context, authorID uuid.UUID, req *models.CreateNewsRequest) (*models.NewsPost, error) {
	// Generate unique slug
	baseSlug := slug.Make(req.Title)
	if baseSlug == "" {
		baseSlug = fmt.Sprintf("news-%d", time.Now().Unix())
	}

	newsSlug := baseSlug
	counter := 1
	for {
		existing, err := s.newsRepo.GetBySlug(ctx, newsSlug)
		if err != nil {
			return nil, fmt.Errorf("checking slug: %w", err)
		}
		if existing == nil {
			break
		}
		newsSlug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	news := &models.NewsPost{
		Slug:     newsSlug,
		Title:    req.Title,
		Summary:  req.Summary,
		Content:  req.Content,
		CoverURL: req.CoverURL,
		Category: req.Category,
		AuthorID: authorID,
		IsPinned: req.IsPinned,
	}

	if err := s.newsRepo.Create(ctx, news); err != nil {
		return nil, fmt.Errorf("creating news: %w", err)
	}

	// Publish immediately if requested
	if req.Publish {
		if err := s.newsRepo.Publish(ctx, news.ID); err != nil {
			return nil, fmt.Errorf("publishing news: %w", err)
		}
		news.IsPublished = true
		now := time.Now()
		news.PublishedAt = &now
	}

	return news, nil
}

// GetByID gets a news post by ID
func (s *NewsService) GetByID(ctx context.Context, id uuid.UUID) (*models.NewsPost, error) {
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if news == nil {
		return nil, nil
	}

	// Load author
	if err := s.enrichNews(ctx, news); err != nil {
		return nil, err
	}

	return news, nil
}

// GetBySlug gets a news post by slug with localization
func (s *NewsService) GetBySlug(ctx context.Context, newsSlug string, lang string) (*models.NewsPost, error) {
	news, err := s.newsRepo.GetLocalizedNews(ctx, newsSlug, lang)
	if err != nil {
		return nil, err
	}
	if news == nil {
		return nil, nil
	}

	// Only return if published
	if !news.IsPublished {
		return nil, nil
	}

	// Load author
	if err := s.enrichNews(ctx, news); err != nil {
		return nil, err
	}

	// Increment views
	_ = s.newsRepo.IncrementViews(ctx, news.ID)

	return news, nil
}

// GetBySlugAdmin gets a news post by slug (admin view, includes unpublished)
func (s *NewsService) GetBySlugAdmin(ctx context.Context, newsSlug string) (*models.NewsPost, error) {
	news, err := s.newsRepo.GetBySlug(ctx, newsSlug)
	if err != nil {
		return nil, err
	}
	if news == nil {
		return nil, nil
	}

	// Load author and localizations
	if err := s.enrichNews(ctx, news); err != nil {
		return nil, err
	}

	locs, err := s.newsRepo.GetLocalizations(ctx, news.ID)
	if err != nil {
		return nil, err
	}
	news.Localizations = locs

	return news, nil
}

func (s *NewsService) enrichNews(ctx context.Context, news *models.NewsPost) error {
	user, err := s.userRepo.GetByID(ctx, news.AuthorID)
	if err != nil {
		return err
	}
	if user != nil {
		var avatarURL *string
		if user.Profile.AvatarKey != nil {
			url := "/uploads/" + *user.Profile.AvatarKey
			avatarURL = &url
		}
		news.Author = &models.UserPublic{
			ID:          user.ID,
			DisplayName: user.Profile.DisplayName,
			AvatarURL:   avatarURL,
		}
	}
	return nil
}

// Update updates a news post
func (s *NewsService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateNewsRequest) (*models.NewsPost, error) {
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if news == nil {
		return nil, fmt.Errorf("news not found")
	}

	if req.Title != nil {
		news.Title = *req.Title
	}
	if req.Summary != nil {
		news.Summary = *req.Summary
	}
	if req.Content != nil {
		news.Content = *req.Content
	}
	if req.CoverURL != nil {
		news.CoverURL = *req.CoverURL
	}
	if req.Category != nil {
		news.Category = *req.Category
	}
	if req.IsPinned != nil {
		news.IsPinned = *req.IsPinned
	}

	if err := s.newsRepo.Update(ctx, news); err != nil {
		return nil, fmt.Errorf("updating news: %w", err)
	}

	return news, nil
}

// Delete deletes a news post
func (s *NewsService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.newsRepo.Delete(ctx, id)
}

// Publish publishes a news post
func (s *NewsService) Publish(ctx context.Context, id uuid.UUID) error {
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if news == nil {
		return fmt.Errorf("news not found")
	}
	if news.IsPublished {
		return nil // Already published
	}

	return s.newsRepo.Publish(ctx, id)
}

// Unpublish unpublishes a news post
func (s *NewsService) Unpublish(ctx context.Context, id uuid.UUID) error {
	return s.newsRepo.Unpublish(ctx, id)
}

// List lists news posts
func (s *NewsService) List(ctx context.Context, params models.NewsListParams) (*models.NewsListResponse, error) {
	// Default to published only for public
	if params.IsPublished == nil {
		published := true
		params.IsPublished = &published
	}

	news, total, err := s.newsRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// Load authors
	for i := range news {
		user, _ := s.userRepo.GetByID(ctx, news[i].AuthorID) // TODO: optimize with batch load
		if user != nil {
			var avatarURL *string
			if user.Profile.AvatarKey != nil {
				url := "/uploads/" + *user.Profile.AvatarKey
				avatarURL = &url
			}
			news[i].Author = &models.UserPublic{
				ID:          user.ID,
				DisplayName: user.Profile.DisplayName,
				AvatarURL:   avatarURL,
			}
		}
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &models.NewsListResponse{
		News:       news,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetLatest gets latest news for homepage
func (s *NewsService) GetLatest(ctx context.Context, limit int) ([]models.NewsCard, error) {
	news, err := s.newsRepo.GetLatest(ctx, limit)
	if err != nil {
		return nil, err
	}

	for i := range news {
		user, _ := s.userRepo.GetByID(ctx, news[i].AuthorID)
		if user != nil {
			var avatarURL *string
			if user.Profile.AvatarKey != nil {
				url := "/uploads/" + *user.Profile.AvatarKey
				avatarURL = &url
			}
			news[i].Author = &models.UserPublic{
				ID:          user.ID,
				DisplayName: user.Profile.DisplayName,
				AvatarURL:   avatarURL,
			}
		}
	}

	return news, nil
}

// SetLocalization adds or updates a news localization
func (s *NewsService) SetLocalization(ctx context.Context, newsID uuid.UUID, req *models.NewsLocalizationRequest) error {
	news, err := s.newsRepo.GetByID(ctx, newsID)
	if err != nil {
		return err
	}
	if news == nil {
		return fmt.Errorf("news not found")
	}

	loc := &models.NewsLocalization{
		NewsID:  newsID,
		Lang:    req.Lang,
		Title:   req.Title,
		Summary: req.Summary,
		Content: req.Content,
	}

	return s.newsRepo.CreateLocalization(ctx, loc)
}

// DeleteLocalization deletes a news localization
func (s *NewsService) DeleteLocalization(ctx context.Context, newsID uuid.UUID, lang string) error {
	return s.newsRepo.DeleteLocalization(ctx, newsID, lang)
}

// SetPinned sets or unsets pinned status
func (s *NewsService) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	news, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if news == nil {
		return fmt.Errorf("news not found")
	}

	return s.newsRepo.SetPinned(ctx, id, pinned)
}

// GetPinned gets pinned news
func (s *NewsService) GetPinned(ctx context.Context) ([]models.NewsCard, error) {
	news, err := s.newsRepo.GetPinnedNews(ctx)
	if err != nil {
		return nil, err
	}

	for i := range news {
		user, _ := s.userRepo.GetByID(ctx, news[i].AuthorID)
		if user != nil {
			var avatarURL *string
			if user.Profile.AvatarKey != nil {
				url := "/uploads/" + *user.Profile.AvatarKey
				avatarURL = &url
			}
			news[i].Author = &models.UserPublic{
				ID:          user.ID,
				DisplayName: user.Profile.DisplayName,
				AvatarURL:   avatarURL,
			}
		}
	}

	return news, nil
}
