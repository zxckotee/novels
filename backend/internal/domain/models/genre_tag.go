package models

import (
	"time"

	"github.com/google/uuid"
)

// GenreWithLocalizations represents a genre with all its localizations
type GenreWithLocalizations struct {
	ID            uuid.UUID                   `json:"id" db:"id"`
	Slug          string                      `json:"slug" db:"slug"`
	CreatedAt     time.Time                   `json:"createdAt" db:"created_at"`
	Localizations map[string]*GenreLocalization `json:"localizations,omitempty"`
	Name          string                      `json:"name,omitempty"` // Current locale name
}

// GenreLocalization represents localized genre information
type GenreLocalization struct {
	GenreID uuid.UUID `json:"genreId" db:"genre_id"`
	Lang    string    `json:"lang" db:"lang"`
	Name    string    `json:"name" db:"name"`
}

// TagWithLocalizations represents a tag with all its localizations
type TagWithLocalizations struct {
	ID            uuid.UUID                 `json:"id" db:"id"`
	Slug          string                    `json:"slug" db:"slug"`
	CreatedAt     time.Time                 `json:"createdAt" db:"created_at"`
	Localizations map[string]*TagLocalization `json:"localizations,omitempty"`
	Name          string                    `json:"name,omitempty"` // Current locale name
}

// TagLocalization represents localized tag information
type TagLocalization struct {
	TagID uuid.UUID `json:"tagId" db:"tag_id"`
	Lang  string    `json:"lang" db:"lang"`
	Name  string    `json:"name" db:"name"`
}

// CreateGenreRequest represents the request to create a genre
type CreateGenreRequest struct {
	Slug          string                            `json:"slug" validate:"required,min=1,max=100"`
	Localizations []CreateGenreLocalizationRequest `json:"localizations" validate:"required,min=1,dive"`
}

// CreateGenreLocalizationRequest represents localized data for creating a genre
type CreateGenreLocalizationRequest struct {
	Lang string `json:"lang" validate:"required,len=2"`
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateGenreRequest represents the request to update a genre
type UpdateGenreRequest struct {
	Slug          string                            `json:"slug,omitempty" validate:"omitempty,min=1,max=100"`
	Localizations []UpdateGenreLocalizationRequest `json:"localizations,omitempty" validate:"omitempty,dive"`
}

// UpdateGenreLocalizationRequest represents localized data for updating a genre
type UpdateGenreLocalizationRequest struct {
	Lang string `json:"lang" validate:"required,len=2"`
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CreateTagRequest represents the request to create a tag
type CreateTagRequest struct {
	Slug          string                          `json:"slug" validate:"required,min=1,max=100"`
	Localizations []CreateTagLocalizationRequest `json:"localizations" validate:"required,min=1,dive"`
}

// CreateTagLocalizationRequest represents localized data for creating a tag
type CreateTagLocalizationRequest struct {
	Lang string `json:"lang" validate:"required,len=2"`
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateTagRequest represents the request to update a tag
type UpdateTagRequest struct {
	Slug          string                          `json:"slug,omitempty" validate:"omitempty,min=1,max=100"`
	Localizations []UpdateTagLocalizationRequest `json:"localizations,omitempty" validate:"omitempty,dive"`
}

// UpdateTagLocalizationRequest represents localized data for updating a tag
type UpdateTagLocalizationRequest struct {
	Lang string `json:"lang" validate:"required,len=2"`
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// GenresFilter represents filters for listing genres
type GenresFilter struct {
	Query string `json:"query,omitempty"`
	Lang  string `json:"lang,omitempty"`
	Sort  string `json:"sort,omitempty"`  // name, created, novels_count
	Order string `json:"order,omitempty"` // asc, desc
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// GenresResponse represents a paginated list of genres
type GenresResponse struct {
	Genres     []GenreWithLocalizations `json:"genres"`
	TotalCount int                      `json:"totalCount"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
}

// TagsFilter represents filters for listing tags
type TagsFilter struct {
	Query string `json:"query,omitempty"`
	Lang  string `json:"lang,omitempty"`
	Sort  string `json:"sort,omitempty"`  // name, created, novels_count
	Order string `json:"order,omitempty"` // asc, desc
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// TagsResponse represents a paginated list of tags
type TagsResponse struct {
	Tags       []TagWithLocalizations `json:"tags"`
	TotalCount int                    `json:"totalCount"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
}
