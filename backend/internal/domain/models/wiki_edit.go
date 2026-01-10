package models

import (
	"time"

	"github.com/google/uuid"
)

// EditRequestStatus represents the status of an edit request
type EditRequestStatus string

const (
	EditRequestStatusPending   EditRequestStatus = "pending"
	EditRequestStatusApproved  EditRequestStatus = "approved"
	EditRequestStatusRejected  EditRequestStatus = "rejected"
	EditRequestStatusWithdrawn EditRequestStatus = "withdrawn"
)

// EditFieldType represents the type of field being edited
type EditFieldType string

const (
	EditFieldTitle                 EditFieldType = "title"
	EditFieldAltTitles             EditFieldType = "alt_titles"
	EditFieldDescription           EditFieldType = "description"
	EditFieldAuthor                EditFieldType = "author"
	EditFieldCoverURL              EditFieldType = "cover_url"
	EditFieldReleaseYear           EditFieldType = "release_year"
	EditFieldOriginalChaptersCount EditFieldType = "original_chapters_count"
	EditFieldGenres                EditFieldType = "genres"
	EditFieldTags                  EditFieldType = "tags"
	EditFieldTranslationStatus     EditFieldType = "translation_status"
)

// NovelEditRequest represents a request to edit a novel
type NovelEditRequest struct {
	ID               uuid.UUID         `json:"id" db:"id"`
	NovelID          uuid.UUID         `json:"novelId" db:"novel_id"`
	UserID           uuid.UUID         `json:"userId" db:"user_id"`
	Status           EditRequestStatus `json:"status" db:"status"`
	EditReason       string            `json:"editReason,omitempty" db:"edit_reason"`
	ModeratorID      *uuid.UUID        `json:"moderatorId,omitempty" db:"moderator_id"`
	ModeratorComment string            `json:"moderatorComment,omitempty" db:"moderator_comment"`
	ReviewedAt       *time.Time        `json:"reviewedAt,omitempty" db:"reviewed_at"`
	CreatedAt        time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time         `json:"updatedAt" db:"updated_at"`

	// Relations
	User      *UserPublic              `json:"user,omitempty"`
	Moderator *UserPublic              `json:"moderator,omitempty"`
	Novel     *NovelCard               `json:"novel,omitempty"`
	Changes   []NovelEditRequestChange `json:"changes,omitempty"`
}

// NovelEditRequestChange represents a single field change in an edit request
type NovelEditRequestChange struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	RequestID uuid.UUID     `json:"requestId" db:"request_id"`
	FieldType EditFieldType `json:"fieldType" db:"field_type"`
	Lang      *string       `json:"lang,omitempty" db:"lang"`
	OldValue  string        `json:"oldValue,omitempty" db:"old_value"`
	NewValue  string        `json:"newValue" db:"new_value"`
	CreatedAt time.Time     `json:"createdAt" db:"created_at"`
}

// NovelEditHistory represents a record of applied edit
type NovelEditHistory struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	NovelID   uuid.UUID     `json:"novelId" db:"novel_id"`
	RequestID *uuid.UUID    `json:"requestId,omitempty" db:"request_id"`
	UserID    uuid.UUID     `json:"userId" db:"user_id"`
	FieldType EditFieldType `json:"fieldType" db:"field_type"`
	Lang      *string       `json:"lang,omitempty" db:"lang"`
	OldValue  string        `json:"oldValue,omitempty" db:"old_value"`
	NewValue  string        `json:"newValue" db:"new_value"`
	CreatedAt time.Time     `json:"createdAt" db:"created_at"`

	// Relations
	User *UserPublic `json:"user,omitempty"`
}

// CreateEditRequestRequest for creating an edit request
type CreateEditRequestRequest struct {
	EditReason string                  `json:"editReason" validate:"max=500"`
	Changes    []EditChangeRequest     `json:"changes" validate:"required,min=1,dive"`
}

// EditChangeRequest single field change request
type EditChangeRequest struct {
	FieldType EditFieldType `json:"fieldType" validate:"required"`
	Lang      *string       `json:"lang" validate:"omitempty,len=2"`
	NewValue  string        `json:"newValue" validate:"required"`
}

// ReviewEditRequestRequest for reviewing (approve/reject) an edit request
type ReviewEditRequestRequest struct {
	Action  string `json:"action" validate:"required,oneof=approve reject"`
	Comment string `json:"comment" validate:"max=500"`
}

// EditRequestListParams for filtering edit requests
type EditRequestListParams struct {
	NovelID *uuid.UUID         `json:"novelId"`
	UserID  *uuid.UUID         `json:"userId"`
	Status  *EditRequestStatus `json:"status"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
}

// EditRequestListResponse paginated response
type EditRequestListResponse struct {
	Requests   []NovelEditRequest `json:"requests"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"totalPages"`
}

// EditHistoryListParams for filtering edit history
type EditHistoryListParams struct {
	NovelID *uuid.UUID `json:"novelId"`
	UserID  *uuid.UUID `json:"userId"`
	Page    int        `json:"page"`
	Limit   int        `json:"limit"`
}

// EditHistoryListResponse paginated response
type EditHistoryListResponse struct {
	History    []NovelEditHistory `json:"history"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"totalPages"`
}

// PlatformStats global platform statistics
type PlatformStats struct {
	TotalNovels         int       `json:"totalNovels" db:"total_novels"`
	TotalChapters       int       `json:"totalChapters" db:"total_chapters"`
	TotalUsers          int       `json:"totalUsers" db:"total_users"`
	TotalComments       int       `json:"totalComments" db:"total_comments"`
	TotalCollections    int       `json:"totalCollections" db:"total_collections"`
	TotalVotesCast      int       `json:"totalVotesCast" db:"total_votes_cast"`
	TotalTicketsSpent   int64     `json:"totalTicketsSpent" db:"total_tickets_spent"`
	ProposalsTranslated int       `json:"proposalsTranslated" db:"proposals_translated"`
	UpdatedAt           time.Time `json:"updatedAt" db:"updated_at"`
}
