package models

import (
	"time"

	"github.com/google/uuid"
)

// ProposalStatus represents the status of a novel proposal
type ProposalStatus string

const (
	ProposalStatusDraft      ProposalStatus = "draft"
	ProposalStatusModeration ProposalStatus = "moderation"
	ProposalStatusVoting     ProposalStatus = "voting"
	ProposalStatusAccepted   ProposalStatus = "accepted"
	ProposalStatusRejected   ProposalStatus = "rejected"
	ProposalStatusTranslating ProposalStatus = "translating"
)

// NovelProposal represents a proposal to translate a novel
type NovelProposal struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       uuid.UUID       `json:"userId" db:"user_id"`
	OriginalLink string          `json:"originalLink" db:"original_link"`
	Status       ProposalStatus  `json:"status" db:"status"`
	
	// Novel metadata (draft)
	Title       string   `json:"title" db:"title"`
	AltTitles   []string `json:"altTitles" db:"alt_titles"`
	Author      string   `json:"author" db:"author"`
	Description string   `json:"description" db:"description"`
	CoverURL    *string  `json:"coverUrl,omitempty" db:"cover_url"`
	Genres      []string `json:"genres" db:"genres"`
	Tags        []string `json:"tags" db:"tags"`
	
	// Voting stats
	VoteScore   int       `json:"voteScore" db:"vote_score"`
	VotesCount  int       `json:"votesCount" db:"votes_count"`
	
	// Moderation
	ModeratorID   *uuid.UUID `json:"moderatorId,omitempty" db:"moderator_id"`
	RejectReason  *string    `json:"rejectReason,omitempty" db:"reject_reason"`
	
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	
	// Populated
	User     *ProposalUser `json:"user,omitempty"`
	UserVote *int          `json:"userVote,omitempty"` // User's vote amount
}

// ProposalUser represents minimal user info for proposals
type ProposalUser struct {
	ID          uuid.UUID `json:"id" db:"id"`
	DisplayName string    `json:"displayName" db:"display_name"`
	AvatarURL   *string   `json:"avatarUrl,omitempty" db:"avatar_url"`
	Level       int       `json:"level" db:"level"`
}

// VotingPoll represents a voting period
type VotingPoll struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Status    string    `json:"status" db:"status"` // active, closed
	StartsAt  time.Time `json:"startsAt" db:"starts_at"`
	EndsAt    time.Time `json:"endsAt" db:"ends_at"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// VotingEntry represents a proposal in a voting poll
type VotingEntry struct {
	PollID    uuid.UUID `json:"pollId" db:"poll_id"`
	NovelID   uuid.UUID `json:"novelId" db:"novel_id"` // Can be proposal_id
	Score     int       `json:"score" db:"score"`
	
	// Populated
	Proposal *NovelProposal `json:"proposal,omitempty"`
}

// Vote represents a single vote
type Vote struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	PollID     uuid.UUID  `json:"pollId" db:"poll_id"`
	UserID     uuid.UUID  `json:"userId" db:"user_id"`
	ProposalID uuid.UUID  `json:"proposalId" db:"proposal_id"`
	TicketType TicketType `json:"ticketType" db:"ticket_type"`
	Amount     int        `json:"amount" db:"amount"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
}

// CreateProposalRequest represents a request to create a proposal
type CreateProposalRequest struct {
	OriginalLink string   `json:"originalLink" validate:"required,url"`
	Title        string   `json:"title" validate:"required,min=2,max=200"`
	AltTitles    []string `json:"altTitles,omitempty"`
	Author       string   `json:"author" validate:"required"`
	Description  string   `json:"description" validate:"required,min=100"`
	CoverURL     *string  `json:"coverUrl,omitempty" validate:"omitempty,url"`
	Genres       []string `json:"genres" validate:"required,min=1"`
	Tags         []string `json:"tags,omitempty"`
}

// UpdateProposalRequest represents a request to update a proposal
type UpdateProposalRequest struct {
	OriginalLink *string  `json:"originalLink,omitempty" validate:"omitempty,url"`
	Title        *string  `json:"title,omitempty" validate:"omitempty,min=2,max=200"`
	AltTitles    []string `json:"altTitles,omitempty"`
	Author       *string  `json:"author,omitempty"`
	Description  *string  `json:"description,omitempty" validate:"omitempty,min=100"`
	CoverURL     *string  `json:"coverUrl,omitempty" validate:"omitempty,url"`
	Genres       []string `json:"genres,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// ModerateProposalRequest represents a moderation action
type ModerateProposalRequest struct {
	Action       string  `json:"action" validate:"required,oneof=approve reject"`
	RejectReason *string `json:"rejectReason,omitempty"`
}

// CastVoteRequest represents a request to cast a vote
type CastVoteRequest struct {
	ProposalID string     `json:"proposalId" validate:"required,uuid"`
	TicketType TicketType `json:"ticketType" validate:"required,oneof=daily_vote translation_ticket"`
	Amount     int        `json:"amount" validate:"required,min=1"`
}

// ProposalFilter represents filters for listing proposals
type ProposalFilter struct {
	Status *ProposalStatus
	UserID *uuid.UUID
	Sort   string // newest, votes, oldest
	Page   int
	Limit  int
}

// ProposalsResponse represents paginated proposals response
type ProposalsResponse struct {
	Proposals  []NovelProposal `json:"proposals"`
	TotalCount int             `json:"totalCount"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

// VotingLeaderboard represents the current voting leaderboard
type VotingLeaderboard struct {
	Poll      *VotingPoll      `json:"poll"`
	Entries   []VotingEntry    `json:"entries"`
	NextReset time.Time        `json:"nextReset"`
}

// VotingStats represents voting statistics
type VotingStats struct {
	TotalProposals      int `json:"totalProposals"`
	ActiveProposals     int `json:"activeProposals"`
	TotalVotesCast      int `json:"totalVotesCast"`
	ProposalsTranslated int `json:"proposalsTranslated"`
}
