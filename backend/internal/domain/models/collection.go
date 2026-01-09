package models

import (
	"time"

	"github.com/google/uuid"
)

// Collection represents a user-created collection of novels
type Collection struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"userId" db:"user_id"`
	Slug        string    `json:"slug" db:"slug"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description,omitempty" db:"description"`
	CoverURL    string    `json:"coverUrl,omitempty" db:"cover_url"`
	IsPublic    bool      `json:"isPublic" db:"is_public"`
	IsFeatured  bool      `json:"isFeatured" db:"is_featured"`
	ViewsCount  int       `json:"viewsCount" db:"views_count"`
	VotesCount  int       `json:"votesCount" db:"votes_count"`
	ItemsCount  int       `json:"itemsCount" db:"items_count"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`

	// Relations (not stored in DB)
	User     *UserPublic       `json:"user,omitempty"`
	Items    []CollectionItem  `json:"items,omitempty"`
	UserVote *int              `json:"userVote,omitempty"` // Current user's vote
}

// CollectionItem represents a novel in a collection
type CollectionItem struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CollectionID uuid.UUID `json:"collectionId" db:"collection_id"`
	NovelID      uuid.UUID `json:"novelId" db:"novel_id"`
	Position     int       `json:"position" db:"position"`
	Note         string    `json:"note,omitempty" db:"note"`
	AddedAt      time.Time `json:"addedAt" db:"added_at"`

	// Relations
	Novel *NovelCard `json:"novel,omitempty"`
}

// CollectionVote represents a user's vote on a collection
type CollectionVote struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CollectionID uuid.UUID `json:"collectionId" db:"collection_id"`
	UserID       uuid.UUID `json:"userId" db:"user_id"`
	Value        int       `json:"value" db:"value"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// CollectionCard is a lightweight collection for lists
type CollectionCard struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"userId" db:"user_id"`
	Slug        string      `json:"slug" db:"slug"`
	Title       string      `json:"title" db:"title"`
	Description string      `json:"description,omitempty" db:"description"`
	CoverURL    string      `json:"coverUrl,omitempty" db:"cover_url"`
	VotesCount  int         `json:"votesCount" db:"votes_count"`
	ItemsCount  int         `json:"itemsCount" db:"items_count"`
	CreatedAt   time.Time   `json:"createdAt" db:"created_at"`
	User        *UserPublic `json:"user,omitempty"`
	PreviewCovers []string  `json:"previewCovers,omitempty"` // First N novel covers
}

// CreateCollectionRequest for creating a new collection
type CreateCollectionRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"max=2000"`
	CoverURL    string `json:"coverUrl" validate:"omitempty,url"`
	IsPublic    bool   `json:"isPublic"`
}

// UpdateCollectionRequest for updating a collection
type UpdateCollectionRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	CoverURL    *string `json:"coverUrl" validate:"omitempty,url"`
	IsPublic    *bool   `json:"isPublic"`
}

// AddToCollectionRequest for adding a novel to a collection
type AddToCollectionRequest struct {
	NovelID  uuid.UUID `json:"novelId" validate:"required"`
	Position *int      `json:"position"`
	Note     string    `json:"note" validate:"max=500"`
}

// UpdateCollectionItemRequest for updating an item in a collection
type UpdateCollectionItemRequest struct {
	Position *int    `json:"position"`
	Note     *string `json:"note" validate:"omitempty,max=500"`
}

// ReorderCollectionItemsRequest for reordering items
type ReorderCollectionItemsRequest struct {
	Items []struct {
		NovelID  uuid.UUID `json:"novelId"`
		Position int       `json:"position"`
	} `json:"items" validate:"required"`
}

// CollectionListParams for filtering collections
type CollectionListParams struct {
	UserID     *uuid.UUID `json:"userId"`
	IsFeatured *bool      `json:"isFeatured"`
	Sort       string     `json:"sort"` // popular, recent, votes
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

// CollectionListResponse paginated response
type CollectionListResponse struct {
	Collections []CollectionCard `json:"collections"`
	Total       int              `json:"total"`
	Page        int              `json:"page"`
	Limit       int              `json:"limit"`
	TotalPages  int              `json:"totalPages"`
}
