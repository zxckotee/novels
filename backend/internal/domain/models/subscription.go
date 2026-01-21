package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCanceled  SubscriptionStatus = "canceled"
	SubscriptionStatusPastDue   SubscriptionStatus = "past_due"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
)

// SubscriptionPlanCode represents subscription plan identifiers
type SubscriptionPlanCode string

const (
	PlanBasic   SubscriptionPlanCode = "basic"
	PlanPremium SubscriptionPlanCode = "premium"
	// PlanVIP matches UI + DB plan code "vip"
	PlanVIP SubscriptionPlanCode = "vip"
	// PlanUltimate kept for backwards-compatibility (legacy code)
	PlanUltimate SubscriptionPlanCode = "ultimate"
)

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan struct {
	ID          uuid.UUID            `json:"id" db:"id"`
	Code        SubscriptionPlanCode `json:"code" db:"code"`
	Title       string               `json:"title" db:"title"`
	Description string               `json:"description" db:"description"`
	Price       int                  `json:"price" db:"price"` // in cents/kopecks
	Currency    string               `json:"currency" db:"currency"`
	Period      string               `json:"period" db:"period"` // monthly, yearly
	IsActive    bool                 `json:"isActive" db:"is_active"`
	
	// Features
	Features    PlanFeatures         `json:"features" db:"features"`
	
	CreatedAt   time.Time            `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time            `json:"updatedAt" db:"updated_at"`
}

// PlanFeatures represents the features of a subscription plan
type PlanFeatures struct {
	DailyVoteMultiplier      int  `json:"dailyVoteMultiplier"`      // multiplier for daily votes
	// NOTE: These fields are still named "Monthly*" for compatibility with stored JSON,
	// but they are interpreted as WEEKLY grant amounts (Wed 00:00 UTC) in the current economy.
	MonthlyNovelRequests      int `json:"monthlyNovelRequests"`      // novel request tickets per week
	MonthlyTranslationTickets int `json:"monthlyTranslationTickets"` // translation tickets per week
	AdFree                   bool `json:"adFree"`                   // no ads
	CanEditDescriptions      bool `json:"canEditDescriptions"`      // can edit novel descriptions
	CanRequestRetranslation  bool `json:"canRequestRetranslation"`  // can request chapter retranslation
	PrioritySupport          bool `json:"prioritySupport"`          // priority support
	ExclusiveBadge           bool `json:"exclusiveBadge"`           // exclusive profile badge
}

// Subscription represents a user's subscription
type Subscription struct {
	ID        uuid.UUID          `json:"id" db:"id"`
	UserID    uuid.UUID          `json:"userId" db:"user_id"`
	PlanID    uuid.UUID          `json:"planId" db:"plan_id"`
	Status    SubscriptionStatus `json:"status" db:"status"`
	StartsAt  time.Time          `json:"startsAt" db:"starts_at"`
	EndsAt    time.Time          `json:"endsAt" db:"ends_at"`
	
	// Payment info
	ExternalID *string `json:"externalId,omitempty" db:"external_id"` // Stripe/YooKassa ID
	
	// Auto-renewal
	AutoRenew bool `json:"autoRenew" db:"auto_renew"`
	
	CanceledAt *time.Time `json:"canceledAt,omitempty" db:"canceled_at"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time  `json:"updatedAt" db:"updated_at"`
	
	// Populated
	Plan *SubscriptionPlan `json:"plan,omitempty"`
}

// SubscriptionGrant represents a ticket grant from subscription
type SubscriptionGrant struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SubscriptionID uuid.UUID  `json:"subscriptionId" db:"subscription_id"`
	UserID         uuid.UUID  `json:"userId" db:"user_id"`
	Type           TicketType `json:"type" db:"type"`
	Amount         int        `json:"amount" db:"amount"`
	GrantedAt      time.Time  `json:"grantedAt" db:"granted_at"`
	ForMonth       string     `json:"forMonth" db:"for_month"` // YYYY-MM format for idempotency
}

// UserSubscriptionInfo represents user's subscription info with features
type UserSubscriptionInfo struct {
	HasActiveSubscription bool              `json:"hasActiveSubscription"`
	Subscription          *Subscription     `json:"subscription,omitempty"`
	Plan                  *SubscriptionPlan `json:"plan,omitempty"`
	Features              *PlanFeatures     `json:"features,omitempty"`
	DaysRemaining         int               `json:"daysRemaining,omitempty"`
}

// CreateSubscriptionRequest represents a request to create/purchase subscription
type CreateSubscriptionRequest struct {
	PlanID    string `json:"planId" validate:"required,uuid"`
	AutoRenew bool   `json:"autoRenew"`
}

// CancelSubscriptionRequest represents a request to cancel subscription
type CancelSubscriptionRequest struct {
	Reason string `json:"reason,omitempty"`
}

// SubscriptionPlansResponse represents available plans
type SubscriptionPlansResponse struct {
	Plans []SubscriptionPlan `json:"plans"`
}

// DefaultPlanFeatures returns default features for each plan
func DefaultPlanFeatures() map[SubscriptionPlanCode]PlanFeatures {
	return map[SubscriptionPlanCode]PlanFeatures{
		PlanBasic: {
			DailyVoteMultiplier:       1,
			MonthlyNovelRequests:      1,
			MonthlyTranslationTickets: 3,
			AdFree:                    true,
			CanEditDescriptions:       false,
			CanRequestRetranslation:   false,
			PrioritySupport:           false,
			ExclusiveBadge:            false,
		},
		PlanPremium: {
			DailyVoteMultiplier:       2,
			MonthlyNovelRequests:      2,
			MonthlyTranslationTickets: 5,
			AdFree:                    true,
			CanEditDescriptions:       true,
			CanRequestRetranslation:   true,
			PrioritySupport:           false,
			ExclusiveBadge:            true,
		},
		PlanVIP: {
			DailyVoteMultiplier:       5,
			MonthlyNovelRequests:      5,
			MonthlyTranslationTickets: 15,
			AdFree:                    true,
			CanEditDescriptions:       true,
			CanRequestRetranslation:   true,
			PrioritySupport:           true,
			ExclusiveBadge:            true,
		},
		// Legacy alias
		PlanUltimate: {
			DailyVoteMultiplier:       5,
			MonthlyNovelRequests:      5,
			MonthlyTranslationTickets: 15,
			AdFree:                    true,
			CanEditDescriptions:       true,
			CanRequestRetranslation:   true,
			PrioritySupport:           true,
			ExclusiveBadge:            true,
		},
	}
}

// IsActive checks if subscription is currently active
func (s *Subscription) IsActive() bool {
	if s.Status != SubscriptionStatusActive {
		return false
	}
	now := time.Now()
	return now.After(s.StartsAt) && now.Before(s.EndsAt)
}

// DaysRemaining returns the number of days remaining in subscription
func (s *Subscription) DaysRemaining() int {
	if !s.IsActive() {
		return 0
	}
	remaining := time.Until(s.EndsAt)
	return int(remaining.Hours() / 24)
}

// GetDailyVoteMultiplier returns the daily vote multiplier for the subscription
func (s *Subscription) GetDailyVoteMultiplier() int {
	if s.Plan == nil || !s.IsActive() {
		return 1
	}
	return s.Plan.Features.DailyVoteMultiplier
}
