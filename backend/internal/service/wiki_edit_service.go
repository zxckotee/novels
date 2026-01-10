package service

import (
	"context"
	"fmt"

	"novels/internal/domain/models"
	"novels/internal/repository"

	"github.com/google/uuid"
)

// WikiEditService handles wiki edit business logic
type WikiEditService struct {
	wikiRepo        *repository.WikiEditRepository
	novelRepo       *repository.NovelRepository
	userRepo        *repository.UserRepository
	subscriptionSvc *SubscriptionService
}

// NewWikiEditService creates a new wiki edit service
func NewWikiEditService(
	wikiRepo *repository.WikiEditRepository,
	novelRepo *repository.NovelRepository,
	userRepo *repository.UserRepository,
	subscriptionSvc *SubscriptionService,
) *WikiEditService {
	return &WikiEditService{
		wikiRepo:        wikiRepo,
		novelRepo:       novelRepo,
		userRepo:        userRepo,
		subscriptionSvc: subscriptionSvc,
	}
}

// CreateEditRequest creates a new edit request (Premium users only)
func (s *WikiEditService) CreateEditRequest(ctx context.Context, userID, novelID uuid.UUID, req *models.CreateEditRequestRequest) (*models.NovelEditRequest, error) {
	// Check if user can edit descriptions (Premium feature)
	canEdit, err := s.subscriptionSvc.HasFeature(ctx, userID, "can_edit_descriptions")
	if err != nil {
		return nil, fmt.Errorf("checking feature: %w", err)
	}
	if !canEdit {
		return nil, fmt.Errorf("editing descriptions requires Premium subscription")
	}

	// Check if novel exists
	novel, err := s.novelRepo.GetByID(ctx, novelID)
	if err != nil {
		return nil, err
	}
	if novel == nil {
		return nil, fmt.Errorf("novel not found")
	}

	// Check if user already has a pending request for this novel
	hasPending, err := s.wikiRepo.HasPendingEditRequest(ctx, userID, novelID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, fmt.Errorf("you already have a pending edit request for this novel")
	}

	// Create the request
	editRequest := &models.NovelEditRequest{
		NovelID:    novelID,
		UserID:     userID,
		EditReason: req.EditReason,
	}

	if err := s.wikiRepo.CreateEditRequest(ctx, editRequest); err != nil {
		return nil, fmt.Errorf("creating edit request: %w", err)
	}

	// Add changes
	for _, change := range req.Changes {
		// Get old value for the field
		oldValue, err := s.getFieldValue(ctx, novelID, change.FieldType, change.Lang)
		if err != nil {
			return nil, fmt.Errorf("getting old value: %w", err)
		}

		editChange := &models.NovelEditRequestChange{
			RequestID: editRequest.ID,
			FieldType: change.FieldType,
			Lang:      change.Lang,
			OldValue:  oldValue,
			NewValue:  change.NewValue,
		}

		if err := s.wikiRepo.AddEditChange(ctx, editChange); err != nil {
			return nil, fmt.Errorf("adding change: %w", err)
		}
		editRequest.Changes = append(editRequest.Changes, *editChange)
	}

	return editRequest, nil
}

func (s *WikiEditService) getFieldValue(ctx context.Context, novelID uuid.UUID, fieldType models.EditFieldType, lang *string) (string, error) {
	novel, err := s.novelRepo.GetByID(ctx, novelID)
	if err != nil {
		return "", err
	}
	if novel == nil {
		return "", fmt.Errorf("novel not found")
	}

	switch fieldType {
	case models.EditFieldCoverURL:
		return novel.CoverURL, nil
	case models.EditFieldReleaseYear:
		return fmt.Sprintf("%d", novel.ReleaseYear), nil
	case models.EditFieldOriginalChaptersCount:
		return fmt.Sprintf("%d", novel.OriginalChaptersCount), nil
	case models.EditFieldTranslationStatus:
		return string(novel.TranslationStatus), nil
	case models.EditFieldTitle, models.EditFieldDescription, models.EditFieldAltTitles:
		if lang != nil {
			loc, err := s.novelRepo.GetLocalization(ctx, novelID, *lang)
			if err != nil {
				return "", err
			}
			if loc != nil {
				switch fieldType {
				case models.EditFieldTitle:
					return loc.Title, nil
				case models.EditFieldDescription:
					return loc.Description, nil
				case models.EditFieldAltTitles:
					return loc.AltTitles, nil
				}
			}
		}
	}

	return "", nil
}

// GetEditRequest gets an edit request by ID
func (s *WikiEditService) GetEditRequest(ctx context.Context, id uuid.UUID) (*models.NovelEditRequest, error) {
	request, err := s.wikiRepo.GetEditRequestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, nil
	}

	// Load changes
	changes, err := s.wikiRepo.GetEditRequestChanges(ctx, id)
	if err != nil {
		return nil, err
	}
	request.Changes = changes

	// Load user
	user, err := s.userRepo.GetByID(ctx, request.UserID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		request.User = &models.UserPublic{
			ID:          user.ID,
			DisplayName: user.Profile.DisplayName,
			AvatarURL:   user.Profile.AvatarURL,
		}
	}

	// Load moderator if reviewed
	if request.ModeratorID != nil {
		moderator, err := s.userRepo.GetByID(ctx, *request.ModeratorID)
		if err != nil {
			return nil, err
		}
		if moderator != nil {
			request.Moderator = &models.UserPublic{
				ID:          moderator.ID,
				DisplayName: moderator.Profile.DisplayName,
				AvatarURL:   moderator.Profile.AvatarURL,
			}
		}
	}

	return request, nil
}

// ListEditRequests lists edit requests
func (s *WikiEditService) ListEditRequests(ctx context.Context, params models.EditRequestListParams) (*models.EditRequestListResponse, error) {
	requests, total, err := s.wikiRepo.ListEditRequests(ctx, params)
	if err != nil {
		return nil, err
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &models.EditRequestListResponse{
		Requests:   requests,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetPendingRequests gets pending requests for moderation
func (s *WikiEditService) GetPendingRequests(ctx context.Context, page, limit int) (*models.EditRequestListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	requests, total, err := s.wikiRepo.GetPendingRequests(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Enrich with user data
	for i := range requests {
		user, _ := s.userRepo.GetByID(ctx, requests[i].UserID)
		if user != nil {
			requests[i].User = &models.UserPublic{
				ID:          user.ID,
				DisplayName: user.Profile.DisplayName,
				AvatarURL:   user.Profile.AvatarURL,
			}
		}
	}

	totalPages := (total + limit - 1) / limit

	return &models.EditRequestListResponse{
		Requests:   requests,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ReviewEditRequest reviews (approves/rejects) an edit request
func (s *WikiEditService) ReviewEditRequest(ctx context.Context, requestID, moderatorID uuid.UUID, req *models.ReviewEditRequestRequest) error {
	// Check request exists and is pending
	request, err := s.wikiRepo.GetEditRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if request == nil {
		return fmt.Errorf("edit request not found")
	}
	if request.Status != models.EditRequestStatusPending {
		return fmt.Errorf("edit request is not pending")
	}

	switch req.Action {
	case "approve":
		return s.wikiRepo.ApproveEditRequest(ctx, requestID, moderatorID, req.Comment)
	case "reject":
		return s.wikiRepo.RejectEditRequest(ctx, requestID, moderatorID, req.Comment)
	default:
		return fmt.Errorf("invalid action: %s", req.Action)
	}
}

// WithdrawEditRequest withdraws user's own edit request
func (s *WikiEditService) WithdrawEditRequest(ctx context.Context, requestID, userID uuid.UUID) error {
	return s.wikiRepo.WithdrawEditRequest(ctx, requestID, userID)
}

// GetEditHistory gets edit history for a novel
func (s *WikiEditService) GetEditHistory(ctx context.Context, params models.EditHistoryListParams) (*models.EditHistoryListResponse, error) {
	history, total, err := s.wikiRepo.GetEditHistory(ctx, params)
	if err != nil {
		return nil, err
	}

	// Enrich with user data
	for i := range history {
		user, _ := s.userRepo.GetByID(ctx, history[i].UserID)
		if user != nil {
			history[i].User = &models.UserPublic{
				ID:          user.ID,
				DisplayName: user.Profile.DisplayName,
				AvatarURL:   user.Profile.AvatarURL,
			}
		}
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &models.EditHistoryListResponse{
		History:    history,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetUserEditRequests gets all edit requests by a user
func (s *WikiEditService) GetUserEditRequests(ctx context.Context, userID uuid.UUID) ([]models.NovelEditRequest, error) {
	return s.wikiRepo.GetUserEditRequests(ctx, userID)
}

// GetPlatformStats gets global platform statistics
func (s *WikiEditService) GetPlatformStats(ctx context.Context) (*models.PlatformStats, error) {
	return s.wikiRepo.GetPlatformStats(ctx)
}

// RefreshPlatformStats refreshes the cached statistics
func (s *WikiEditService) RefreshPlatformStats(ctx context.Context) error {
	return s.wikiRepo.RefreshPlatformStats(ctx)
}

// CountPendingRequests counts pending edit requests (for moderation badge)
func (s *WikiEditService) CountPendingRequests(ctx context.Context) (int, error) {
	return s.wikiRepo.CountPendingEditRequests(ctx)
}
