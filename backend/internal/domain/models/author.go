package models

import (
	"time"

	"github.com/google/uuid"
)

// Author represents an author entity
type Author struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`

	// Populated from joins
	Localizations map[string]*AuthorLocalization `json:"localizations,omitempty"`
	Name          string                          `json:"name,omitempty"` // Current locale name
	Bio           string                          `json:"bio,omitempty"`  // Current locale bio
}

// AuthorLocalization represents localized author information
type AuthorLocalization struct {
	AuthorID  uuid.UUID `json:"authorId" db:"author_id"`
	Lang      string    `json:"lang" db:"lang"`
	Name      string    `json:"name" db:"name"`
	Bio       string    `json:"bio,omitempty" db:"bio"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// NovelAuthor represents the many-to-many relationship between novels and authors
type NovelAuthor struct {
	NovelID   uuid.UUID `json:"novelId" db:"novel_id"`
	AuthorID  uuid.UUID `json:"authorId" db:"author_id"`
	IsPrimary bool      `json:"isPrimary" db:"is_primary"`
	SortOrder int       `json:"sortOrder" db:"sort_order"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	// Populated from joins
	Author *Author `json:"author,omitempty"`
}

// CreateAuthorRequest represents the request to create an author
type CreateAuthorRequest struct {
	Slug          string                            `json:"slug" validate:"required,min=1,max=255"`
	Localizations []CreateAuthorLocalizationRequest `json:"localizations" validate:"required,min=1,dive"`
}

// CreateAuthorLocalizationRequest represents localized data for creating an author
type CreateAuthorLocalizationRequest struct {
	Lang string `json:"lang" validate:"required,len=2"`
	Name string `json:"name" validate:"required,min=1,max=255"`
	Bio  string `json:"bio,omitempty" validate:"max=10000"`
}

// UpdateAuthorRequest represents the request to update an author
type UpdateAuthorRequest struct {
	Slug          string                            `json:"slug" validate:"omitempty,min=1,max=255"`
	Localizations []UpdateAuthorLocalizationRequest `json:"localizations,omitempty" validate:"omitempty,dive"`
}

// UpdateAuthorLocalizationRequest represents localized data for updating an author
type UpdateAuthorLocalizationRequest struct {
	Lang string `json:"lang" validate:"required,len=2"`
	Name string `json:"name" validate:"required,min=1,max=255"`
	Bio  string `json:"bio,omitempty" validate:"max=10000"`
}

// AuthorsFilter represents filters for listing authors
type AuthorsFilter struct {
	Query  string `json:"query,omitempty"`
	Lang   string `json:"lang,omitempty"`
	Sort   string `json:"sort,omitempty"`   // name, created, novels_count
	Order  string `json:"order,omitempty"`  // asc, desc
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// AuthorsResponse represents a paginated list of authors
type AuthorsResponse struct {
	Authors    []Author `json:"authors"`
	TotalCount int      `json:"totalCount"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
}

// UpdateNovelAuthorsRequest represents the request to update novel authors
type UpdateNovelAuthorsRequest struct {
	Authors []NovelAuthorInput `json:"authors" validate:"required,dive"`
}

// NovelAuthorInput represents author input for a novel
type NovelAuthorInput struct {
	AuthorID  string `json:"authorId" validate:"required,uuid"`
	IsPrimary bool   `json:"isPrimary"`
	SortOrder int    `json:"sortOrder"`
}
