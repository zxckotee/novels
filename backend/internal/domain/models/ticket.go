package models

import (
	"time"

	"github.com/google/uuid"
)

// TicketType represents the type of ticket
type TicketType string

const (
	TicketTypeDailyVote          TicketType = "daily_vote"
	TicketTypeNovelRequest       TicketType = "novel_request"
	TicketTypeTranslationTicket  TicketType = "translation_ticket"
)

// TicketBalance represents a user's balance for a specific ticket type
type TicketBalance struct {
	UserID    uuid.UUID  `json:"userId" db:"user_id"`
	Type      TicketType `json:"type" db:"type"`
	Balance   int        `json:"balance" db:"balance"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

// TicketTransaction represents a ticket transaction (credit/debit)
type TicketTransaction struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"userId" db:"user_id"`
	Type           TicketType `json:"type" db:"type"`
	Delta          int        `json:"delta" db:"delta"` // positive = credit, negative = debit
	Reason         string     `json:"reason" db:"reason"`
	RefType        string     `json:"refType,omitempty" db:"ref_type"`
	RefID          *uuid.UUID `json:"refId,omitempty" db:"ref_id"`
	IdempotencyKey string     `json:"-" db:"idempotency_key"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
}

// TicketTransactionReason defines standard reasons for transactions
const (
	ReasonDailyGrant         = "daily_grant"
	ReasonVoteCast           = "vote_cast"
	ReasonProposalCreated    = "proposal_created"
	ReasonTranslationRequest = "translation_request"
	ReasonSubscriptionGrant  = "subscription_grant"
	ReasonAdminAdjustment    = "admin_adjustment"
	ReasonLevelReward        = "level_reward"
)

// WalletInfo represents user's wallet with all ticket balances
type WalletInfo struct {
	UserID             uuid.UUID `json:"userId"`
	DailyVotes         int       `json:"dailyVotes"`
	NovelRequests      int       `json:"novelRequests"`
	TranslationTickets int       `json:"translationTickets"`
	NextDailyReset     time.Time `json:"nextDailyReset"`
}

// SpendTicketRequest represents a request to spend tickets
type SpendTicketRequest struct {
	Type   TicketType `json:"type" validate:"required"`
	Amount int        `json:"amount" validate:"required,min=1"`
	RefType string    `json:"refType,omitempty"`
	RefID   string    `json:"refId,omitempty"`
}

// GrantTicketRequest represents a request to grant tickets (admin)
type GrantTicketRequest struct {
	UserID uuid.UUID  `json:"userId" validate:"required"`
	Type   TicketType `json:"type" validate:"required"`
	Amount int        `json:"amount" validate:"required,min=1"`
	Reason string     `json:"reason" validate:"required"`
}

// TransactionFilter represents filters for listing transactions
type TransactionFilter struct {
	UserID *uuid.UUID
	Type   *TicketType
	Page   int
	Limit  int
}

// TransactionsResponse represents paginated transactions response
type TransactionsResponse struct {
	Transactions []TicketTransaction `json:"transactions"`
	TotalCount   int                 `json:"totalCount"`
	Page         int                 `json:"page"`
	Limit        int                 `json:"limit"`
}

// DefaultDailyVoteAmount is the default daily vote amount for regular users
const DefaultDailyVoteAmount = 1

// PremiumDailyVoteMultiplier is the multiplier for premium users
const PremiumDailyVoteMultiplier = 2
