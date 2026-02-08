package models

import (
	"time"

	"github.com/google/uuid"
)

type TranslationVoteTargetStatus string

const (
	TranslationTargetStatusVoting         TranslationVoteTargetStatus = "voting"
	TranslationTargetStatusWaitingRelease TranslationVoteTargetStatus = "waiting_release"
	TranslationTargetStatusTranslating    TranslationVoteTargetStatus = "translating"
	TranslationTargetStatusCompleted      TranslationVoteTargetStatus = "completed"
	TranslationTargetStatusCancelled      TranslationVoteTargetStatus = "cancelled"
)

type TranslationVoteTarget struct {
	ID                        uuid.UUID                 `json:"id" db:"id"`
	NovelID                   *uuid.UUID                `json:"novelId,omitempty" db:"novel_id"`
	ProposalID                *uuid.UUID                `json:"proposalId,omitempty" db:"proposal_id"`
	Status                    TranslationVoteTargetStatus `json:"status" db:"status"`
	TranslationTicketsInvested int                      `json:"translationTicketsInvested" db:"translation_tickets_invested"`
	CreatedAt                 time.Time                 `json:"createdAt" db:"created_at"`
	UpdatedAt                 time.Time                 `json:"updatedAt" db:"updated_at"`
}

type TranslationLeaderboardEntry struct {
	TargetID uuid.UUID `json:"targetId"`
	Status   TranslationVoteTargetStatus `json:"status"`
	Score    int       `json:"score"`

	NovelID    *uuid.UUID `json:"novelId,omitempty"`
	ProposalID *uuid.UUID `json:"proposalId,omitempty"`

	Title    string  `json:"title"`
	CoverURL *string `json:"coverUrl,omitempty"`
}

type TranslationLeaderboard struct {
	Entries []TranslationLeaderboardEntry `json:"entries"`
}

type CastTranslationVoteRequest struct {
	TargetID   *string `json:"targetId,omitempty"`   // optional if voting by ref
	NovelID    *string `json:"novelId,omitempty"`    // one of novelId/proposalId/targetId
	ProposalID *string `json:"proposalId,omitempty"` // one of novelId/proposalId/targetId
	Amount     int     `json:"amount" validate:"required,min=1"`
}

