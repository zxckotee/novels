package models

import (
	"time"

	"github.com/google/uuid"
)

// TargetType represents the type of entity a comment is attached to
type TargetType string

const (
	TargetTypeNovel   TargetType = "novel"
	TargetTypeChapter TargetType = "chapter"
	TargetTypeNews    TargetType = "news"
	TargetTypeProfile TargetType = "profile"
)

// Comment represents a nested comment in the system
type Comment struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ParentID  *uuid.UUID `json:"parentId,omitempty" db:"parent_id"`
	RootID    *uuid.UUID `json:"rootId,omitempty" db:"root_id"`
	Depth     int        `json:"depth" db:"depth"`
	
	TargetType TargetType `json:"targetType" db:"target_type"`
	TargetID   uuid.UUID  `json:"targetId" db:"target_id"`
	
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Body      string    `json:"body" db:"body"`
	IsDeleted bool      `json:"isDeleted" db:"is_deleted"`
	IsSpoiler bool      `json:"isSpoiler" db:"is_spoiler"`
	
	LikesCount    int `json:"likesCount" db:"likes_count"`
	DislikesCount int `json:"dislikesCount" db:"dislikes_count"`
	RepliesCount  int `json:"repliesCount" db:"replies_count"`
	
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	
	// Populated from joins
	User     *CommentUser `json:"user,omitempty"`
	UserVote *int         `json:"userVote,omitempty"` // -1, 0, or 1
	Replies  []Comment    `json:"replies,omitempty"`
}

// CommentUser represents minimal user info for comments
type CommentUser struct {
	ID          uuid.UUID `json:"id" db:"id"`
	DisplayName string    `json:"displayName" db:"display_name"`
	AvatarURL   *string   `json:"avatarUrl,omitempty" db:"avatar_url"`
	Level       int       `json:"level" db:"level"`
	Role        UserRole  `json:"role" db:"role"`
}

// CommentVote represents a user's vote on a comment
type CommentVote struct {
	CommentID uuid.UUID `json:"commentId" db:"comment_id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Value     int       `json:"value" db:"value"` // -1 or 1
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// CommentReport represents a report/complaint about a comment
type CommentReport struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CommentID uuid.UUID `json:"commentId" db:"comment_id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Reason    string    `json:"reason" db:"reason"`
	Status    string    `json:"status" db:"status"` // pending, resolved, dismissed
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	TargetType TargetType `json:"targetType" validate:"required,oneof=novel chapter news profile"`
	TargetID   string     `json:"targetId" validate:"required,uuid"`
	ParentID   *string    `json:"parentId,omitempty" validate:"omitempty,uuid"`
	Body       string     `json:"body" validate:"required,min=1,max=10000"`
	IsSpoiler  bool       `json:"isSpoiler"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Body      string `json:"body" validate:"required,min=1,max=10000"`
	IsSpoiler bool   `json:"isSpoiler"`
}

// VoteCommentRequest represents the request to vote on a comment
type VoteCommentRequest struct {
	Value int `json:"value" validate:"required,oneof=-1 1"`
}

// ReportCommentRequest represents the request to report a comment
type ReportCommentRequest struct {
	Reason string `json:"reason" validate:"required,min=10,max=1000"`
}

// CommentsFilter represents filters for listing comments
type CommentsFilter struct {
	TargetType TargetType `json:"targetType"`
	TargetID   uuid.UUID  `json:"targetId"`
	ParentID   *uuid.UUID `json:"parentId,omitempty"` // nil = root comments only
	UserID     *uuid.UUID `json:"userId,omitempty"`
	Sort       string     `json:"sort"` // newest, oldest, top
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

// CommentsResponse represents a paginated list of comments
type CommentsResponse struct {
	Comments   []Comment `json:"comments"`
	TotalCount int       `json:"totalCount"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
}
