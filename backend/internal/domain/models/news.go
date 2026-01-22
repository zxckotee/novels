package models

import (
	"time"

	"github.com/google/uuid"
)

// NewsCategory represents news category type
type NewsCategory string

const (
	NewsCategoryAnnouncement NewsCategory = "announcement"
	NewsCategoryUpdate       NewsCategory = "update"
	NewsCategoryEvent        NewsCategory = "event"
	NewsCategoryCommunity    NewsCategory = "community"
	NewsCategoryTranslation  NewsCategory = "translation"
)

// NewsPost represents a news article
type NewsPost struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	Slug          string       `json:"slug" db:"slug"`
	Title         string       `json:"title" db:"title"`
	Summary       string       `json:"summary,omitempty" db:"summary"`
	Content       string       `json:"content" db:"content"`
	CoverURL      string       `json:"coverUrl,omitempty" db:"cover_url"`
	Category      NewsCategory `json:"category" db:"category"`
	AuthorID      uuid.UUID    `json:"authorId" db:"author_id"`
	IsPublished   bool         `json:"isPublished" db:"is_published"`
	IsPinned      bool         `json:"isPinned" db:"is_pinned"`
	ViewsCount    int          `json:"viewsCount" db:"views_count"`
	CommentsCount int          `json:"commentsCount" db:"comments_count"`
	PublishedAt   *time.Time   `json:"publishedAt,omitempty" db:"published_at"`
	CreatedAt     time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time    `json:"updatedAt" db:"updated_at"`

	// Relations
	Author        *UserPublic         `json:"author,omitempty"`
	Localizations []NewsLocalization  `json:"localizations,omitempty"`
}

// NewsLocalization for multi-language news
type NewsLocalization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	NewsID    uuid.UUID `json:"newsId" db:"news_id"`
	Lang      string    `json:"lang" db:"lang"`
	Title     string    `json:"title" db:"title"`
	Summary   string    `json:"summary,omitempty" db:"summary"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// NewsCard lightweight news for lists
type NewsCard struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	Slug          string       `json:"slug" db:"slug"`
	Title         string       `json:"title" db:"title"`
	Summary       string       `json:"summary,omitempty" db:"summary"`
	CoverURL      string       `json:"coverUrl,omitempty" db:"cover_url"`
	Category      NewsCategory `json:"category" db:"category"`
	IsPinned      bool         `json:"isPinned" db:"is_pinned"`
	ViewsCount    int          `json:"viewsCount" db:"views_count"`
	CommentsCount int          `json:"commentsCount" db:"comments_count"`
	PublishedAt   *time.Time   `json:"publishedAt,omitempty" db:"published_at"`
	AuthorID      uuid.UUID    `json:"authorId" db:"author_id"`
	Author        *UserPublic  `json:"author,omitempty"`
}

// CreateNewsRequest for creating news
type CreateNewsRequest struct {
	Title     string       `json:"title" validate:"required,min=5,max=500"`
	Summary   string       `json:"summary" validate:"max=1000"`
	Content   string       `json:"content" validate:"required,min=50"`
	CoverURL  string       `json:"coverUrl" validate:"omitempty,url"`
	Category  NewsCategory `json:"category" validate:"required,oneof=announcement update event community translation"`
	IsPinned  bool         `json:"isPinned"`
	Publish   bool         `json:"publish"` // Publish immediately
}

// UpdateNewsRequest for updating news
type UpdateNewsRequest struct {
	Title     *string       `json:"title" validate:"omitempty,min=5,max=500"`
	Summary   *string       `json:"summary" validate:"omitempty,max=1000"`
	Content   *string       `json:"content" validate:"omitempty,min=50"`
	CoverURL  *string       `json:"coverUrl" validate:"omitempty,url"`
	Category  *NewsCategory `json:"category" validate:"omitempty,oneof=announcement update event community translation"`
	IsPinned  *bool         `json:"isPinned"`
}

// NewsLocalizationRequest for adding/updating localization
type NewsLocalizationRequest struct {
	Lang    string `json:"lang" validate:"required,len=2"`
	Title   string `json:"title" validate:"required,min=5,max=500"`
	Summary string `json:"summary" validate:"max=1000"`
	Content string `json:"content" validate:"required,min=50"`
}

// NewsListParams for filtering news
type NewsListParams struct {
	Category    *NewsCategory `json:"category"`
	IsPinned    *bool         `json:"isPinned"`
	IsPublished *bool         `json:"isPublished"`
	Lang        string        `json:"lang"`
	Page        int           `json:"page"`
	Limit       int           `json:"limit"`
}

// NewsListResponse paginated response
type NewsListResponse struct {
	News       []NewsCard `json:"news"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"totalPages"`
}
