package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/novels/backend/internal/domain/models"
	"github.com/novels/backend/internal/repository"
	"github.com/rs/zerolog"
)

var (
	ErrPlanNotFound         = errors.New("subscription plan not found")
	ErrAlreadySubscribed    = errors.New("user already has an active subscription")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrNotAuthorized        = errors.New("not authorized")
)

type SubscriptionService struct {
	subRepo    *repository.SubscriptionRepository
	ticketRepo *repository.TicketRepository
	logger     zerolog.Logger
}

func NewSubscriptionService(
	subRepo *repository.SubscriptionRepository,
	ticketRepo *repository.TicketRepository,
	logger zerolog.Logger,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:    subRepo,
		ticketRepo: ticketRepo,
		logger:     logger,
	}
}

// GetPlans returns all available subscription plans
func (s *SubscriptionService) GetPlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	return s.subRepo.GetAllPlans(ctx)
}

// GetPlan returns a subscription plan by ID
func (s *SubscriptionService) GetPlan(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	return s.subRepo.GetPlanByID(ctx, id)
}

// Subscribe creates a new subscription for a user
func (s *SubscriptionService) Subscribe(ctx context.Context, userID uuid.UUID, req models.CreateSubscriptionRequest) (*models.Subscription, error) {
	// Check if user already has active subscription
	existingSub, err := s.subRepo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check existing subscription: %w", err)
	}
	if existingSub != nil {
		return nil, ErrAlreadySubscribed
	}
	
	// Get plan
	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		return nil, errors.New("invalid plan ID")
	}
	
	plan, err := s.subRepo.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if plan == nil || !plan.IsActive {
		return nil, ErrPlanNotFound
	}
	
	// Calculate subscription period
	now := time.Now().UTC()
	var endsAt time.Time
	switch plan.Period {
	case "yearly":
		endsAt = now.AddDate(1, 0, 0)
	default: // monthly
		endsAt = now.AddDate(0, 1, 0)
	}
	
	// Create subscription
	subscription := &models.Subscription{
		ID:        uuid.New(),
		UserID:    userID,
		PlanID:    planID,
		Status:    models.SubscriptionStatusActive,
		StartsAt:  now,
		EndsAt:    endsAt,
		AutoRenew: req.AutoRenew,
	}
	
	err = s.subRepo.CreateSubscription(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}
	
	subscription.Plan = plan
	
	// Grant initial tickets
	err = s.grantMonthlyTickets(ctx, subscription)
	if err != nil {
		s.logger.Error().Err(err).
			Str("subscription_id", subscription.ID.String()).
			Msg("Failed to grant initial tickets")
	}
	
	s.logger.Info().
		Str("user_id", userID.String()).
		Str("plan", string(plan.Code)).
		Time("ends_at", endsAt).
		Msg("Subscription created")
	
	return subscription, nil
}

// GetUserSubscription returns the user's active subscription
func (s *SubscriptionService) GetUserSubscription(ctx context.Context, userID uuid.UUID) (*models.UserSubscriptionInfo, error) {
	return s.subRepo.GetUserSubscriptionInfo(ctx, userID)
}

// GetUserSubscriptionHistory returns all user subscriptions
func (s *SubscriptionService) GetUserSubscriptionHistory(ctx context.Context, userID uuid.UUID) ([]models.Subscription, error) {
	return s.subRepo.GetUserSubscriptions(ctx, userID)
}

// CancelSubscription cancels a subscription
func (s *SubscriptionService) CancelSubscription(ctx context.Context, userID uuid.UUID, subscriptionID uuid.UUID) error {
	sub, err := s.subRepo.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}
	
	if sub.UserID != userID {
		return ErrNotAuthorized
	}
	
	if sub.Status != models.SubscriptionStatusActive {
		return errors.New("subscription is not active")
	}
	
	err = s.subRepo.CancelSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}
	
	s.logger.Info().
		Str("user_id", userID.String()).
		Str("subscription_id", subscriptionID.String()).
		Msg("Subscription canceled")
	
	return nil
}

// HasFeature checks if user has a specific premium feature
func (s *SubscriptionService) HasFeature(ctx context.Context, userID uuid.UUID, feature string) (bool, error) {
	info, err := s.subRepo.GetUserSubscriptionInfo(ctx, userID)
	if err != nil {
		return false, err
	}
	
	if !info.HasActiveSubscription || info.Features == nil {
		return false, nil
	}
	
	switch feature {
	case "ad_free":
		return info.Features.AdFree, nil
	case "edit_descriptions":
		return info.Features.CanEditDescriptions, nil
	case "request_retranslation":
		return info.Features.CanRequestRetranslation, nil
	case "priority_support":
		return info.Features.PrioritySupport, nil
	case "exclusive_badge":
		return info.Features.ExclusiveBadge, nil
	default:
		return false, nil
	}
}

// IsPremium checks if user has any active subscription
func (s *SubscriptionService) IsPremium(ctx context.Context, userID uuid.UUID) (bool, error) {
	info, err := s.subRepo.GetUserSubscriptionInfo(ctx, userID)
	if err != nil {
		return false, err
	}
	return info.HasActiveSubscription, nil
}

// ============================================
// MONTHLY TICKET GRANTS
// ============================================

// grantMonthlyTickets grants monthly tickets based on subscription plan
func (s *SubscriptionService) grantMonthlyTickets(ctx context.Context, sub *models.Subscription) error {
	if sub.Plan == nil {
		plan, err := s.subRepo.GetPlanByID(ctx, sub.PlanID)
		if err != nil || plan == nil {
			return fmt.Errorf("get plan: %w", err)
		}
		sub.Plan = plan
	}
	
	forMonth := time.Now().UTC().Format("2006-01")
	
	// Grant Novel Request tickets
	if sub.Plan.Features.MonthlyNovelRequests > 0 {
		exists, _ := s.subRepo.HasGrantForMonth(ctx, sub.ID, models.TicketTypeNovelRequest, forMonth)
		if !exists {
			err := s.ticketRepo.GrantTickets(ctx, sub.UserID, models.TicketTypeNovelRequest,
				sub.Plan.Features.MonthlyNovelRequests, models.ReasonSubscriptionGrant,
				"subscription", &sub.ID,
				fmt.Sprintf("sub_grant:%s:%s:novel_request", sub.ID.String(), forMonth))
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to grant novel request tickets")
			} else {
				s.subRepo.CreateGrant(ctx, &models.SubscriptionGrant{
					SubscriptionID: sub.ID,
					UserID:         sub.UserID,
					Type:           models.TicketTypeNovelRequest,
					Amount:         sub.Plan.Features.MonthlyNovelRequests,
					ForMonth:       forMonth,
				})
			}
		}
	}
	
	// Grant Translation Tickets
	if sub.Plan.Features.MonthlyTranslationTickets > 0 {
		exists, _ := s.subRepo.HasGrantForMonth(ctx, sub.ID, models.TicketTypeTranslationTicket, forMonth)
		if !exists {
			err := s.ticketRepo.GrantTickets(ctx, sub.UserID, models.TicketTypeTranslationTicket,
				sub.Plan.Features.MonthlyTranslationTickets, models.ReasonSubscriptionGrant,
				"subscription", &sub.ID,
				fmt.Sprintf("sub_grant:%s:%s:translation", sub.ID.String(), forMonth))
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to grant translation tickets")
			} else {
				s.subRepo.CreateGrant(ctx, &models.SubscriptionGrant{
					SubscriptionID: sub.ID,
					UserID:         sub.UserID,
					Type:           models.TicketTypeTranslationTicket,
					Amount:         sub.Plan.Features.MonthlyTranslationTickets,
					ForMonth:       forMonth,
				})
			}
		}
	}
	
	return nil
}

// ProcessMonthlyGrants processes monthly ticket grants for all active subscriptions
func (s *SubscriptionService) ProcessMonthlyGrants(ctx context.Context) error {
	// This would typically be called by a cron job at the start of each month
	
	// Get all active subscriptions
	// Note: This is a simplified version. In production, you'd want to paginate
	// and process in batches
	
	s.logger.Info().Msg("Processing monthly subscription grants")
	
	// For now, we rely on grants being made when subscription is created
	// and when user accesses their subscription (lazy grant)
	
	return nil
}

// ============================================
// CRON JOBS
// ============================================

// ExpireSubscriptions marks expired subscriptions as expired
func (s *SubscriptionService) ExpireSubscriptions(ctx context.Context) error {
	count, err := s.subRepo.ExpireSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("expire subscriptions: %w", err)
	}
	
	if count > 0 {
		s.logger.Info().
			Int64("count", count).
			Msg("Subscriptions expired")
	}
	
	return nil
}

// GetExpiringSubscriptions returns subscriptions expiring soon (for notifications)
func (s *SubscriptionService) GetExpiringSubscriptions(ctx context.Context, withinDays int) ([]models.Subscription, error) {
	duration := time.Duration(withinDays) * 24 * time.Hour
	return s.subRepo.GetExpiringSubscriptions(ctx, duration)
}

// GetStats returns subscription statistics
func (s *SubscriptionService) GetStats(ctx context.Context) (map[string]int, error) {
	return s.subRepo.GetSubscriptionStats(ctx)
}
