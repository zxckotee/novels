package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

// CollectionService handles collection business logic
type CollectionService struct {
	collectionRepo *repository.CollectionRepository
	novelRepo      *repository.NovelRepository
	userRepo       *repository.UserRepository
}

// NewCollectionService creates a new collection service
func NewCollectionService(
	collectionRepo *repository.CollectionRepository,
	novelRepo *repository.NovelRepository,
	userRepo *repository.UserRepository,
) *CollectionService {
	return &CollectionService{
		collectionRepo: collectionRepo,
		novelRepo:      novelRepo,
		userRepo:       userRepo,
	}
}

// Create creates a new collection
func (s *CollectionService) Create(ctx context.Context, userID uuid.UUID, req *models.CreateCollectionRequest) (*models.Collection, error) {
	// Generate slug
	baseSlug := slug.Make(req.Title)
	if baseSlug == "" {
		baseSlug = fmt.Sprintf("collection-%d", time.Now().Unix())
	}

	// Check slug uniqueness for this user
	collectionSlug := baseSlug
	counter := 1
	for {
		existing, err := s.collectionRepo.GetBySlug(ctx, userID, collectionSlug)
		if err != nil {
			return nil, fmt.Errorf("checking slug: %w", err)
		}
		if existing == nil {
			break
		}
		collectionSlug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	collection := &models.Collection{
		UserID:      userID,
		Slug:        collectionSlug,
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		IsPublic:    req.IsPublic,
	}

	if err := s.collectionRepo.Create(ctx, collection); err != nil {
		return nil, fmt.Errorf("creating collection: %w", err)
	}

	return collection, nil
}

// GetByID gets a collection by ID
func (s *CollectionService) GetByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*models.Collection, error) {
	collection, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, nil
	}

	// Check visibility
	if !collection.IsPublic && (viewerID == nil || *viewerID != collection.UserID) {
		return nil, nil
	}

	// Load user
	if err := s.enrichCollection(ctx, collection, viewerID); err != nil {
		return nil, err
	}

	return collection, nil
}

// GetBySlug gets a collection by user ID and slug
func (s *CollectionService) GetBySlug(ctx context.Context, userID uuid.UUID, colSlug string, viewerID *uuid.UUID) (*models.Collection, error) {
	collection, err := s.collectionRepo.GetBySlug(ctx, userID, colSlug)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, nil
	}

	// Check visibility
	if !collection.IsPublic && (viewerID == nil || *viewerID != collection.UserID) {
		return nil, nil
	}

	// Load items and user
	if err := s.enrichCollection(ctx, collection, viewerID); err != nil {
		return nil, err
	}

	// Increment views
	_ = s.collectionRepo.IncrementViews(ctx, collection.ID)

	return collection, nil
}

func (s *CollectionService) enrichCollection(ctx context.Context, collection *models.Collection, viewerID *uuid.UUID) error {
	// Load user
	user, err := s.userRepo.GetByID(ctx, collection.UserID)
	if err != nil {
		return err
	}
	if user != nil {
		var avatarURL *string
		if user.Profile.AvatarKey != nil {
			url := "/uploads/" + *user.Profile.AvatarKey
			avatarURL = &url
		}
		collection.User = &models.UserPublic{
			ID:          user.ID,
			DisplayName: user.Profile.DisplayName,
			AvatarURL:   avatarURL,
		}
	}

	// Load items
	items, err := s.collectionRepo.GetItems(ctx, collection.ID)
	if err != nil {
		return err
	}
	collection.Items = items

	// Load viewer's vote
	if viewerID != nil {
		vote, err := s.collectionRepo.GetUserVote(ctx, collection.ID, *viewerID)
		if err != nil {
			return err
		}
		if vote != nil {
			collection.UserVote = &vote.Value
		}
	}

	return nil
}

// Update updates a collection
func (s *CollectionService) Update(ctx context.Context, id, userID uuid.UUID, req *models.UpdateCollectionRequest) (*models.Collection, error) {
	collection, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, fmt.Errorf("collection not found")
	}
	if collection.UserID != userID {
		return nil, fmt.Errorf("not authorized")
	}

	if req.Title != nil {
		collection.Title = *req.Title
	}
	if req.Description != nil {
		collection.Description = *req.Description
	}
	if req.CoverURL != nil {
		collection.CoverURL = *req.CoverURL
	}
	if req.IsPublic != nil {
		collection.IsPublic = *req.IsPublic
	}

	if err := s.collectionRepo.Update(ctx, collection); err != nil {
		return nil, fmt.Errorf("updating collection: %w", err)
	}

	return collection, nil
}

// Delete deletes a collection
func (s *CollectionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	collection, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}
	if collection.UserID != userID {
		return fmt.Errorf("not authorized")
	}

	return s.collectionRepo.Delete(ctx, id)
}

// List lists collections
func (s *CollectionService) List(ctx context.Context, params models.CollectionListParams) (*models.CollectionListResponse, error) {
	collections, total, err := s.collectionRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// Enrich with preview covers
	for i := range collections {
		covers, _ := s.collectionRepo.GetPreviewCovers(ctx, collections[i].ID, 4)
		collections[i].PreviewCovers = covers
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &models.CollectionListResponse{
		Collections: collections,
		Total:       total,
		Page:        params.Page,
		Limit:       params.Limit,
		TotalPages:  totalPages,
	}, nil
}

// GetUserCollections gets all collections for a user
func (s *CollectionService) GetUserCollections(ctx context.Context, userID uuid.UUID, viewerID *uuid.UUID) ([]models.CollectionCard, error) {
	includePrivate := viewerID != nil && *viewerID == userID
	collections, err := s.collectionRepo.GetUserCollections(ctx, userID, includePrivate)
	if err != nil {
		return nil, err
	}

	// Enrich with preview covers
	for i := range collections {
		covers, _ := s.collectionRepo.GetPreviewCovers(ctx, collections[i].ID, 4)
		collections[i].PreviewCovers = covers
	}

	return collections, nil
}

// AddItem adds a novel to a collection
func (s *CollectionService) AddItem(ctx context.Context, collectionID, userID uuid.UUID, req *models.AddToCollectionRequest) error {
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}
	if collection.UserID != userID {
		return fmt.Errorf("not authorized")
	}

	// Check novel exists
	novel, err := s.novelRepo.GetByID(ctx, req.NovelID, "ru")
	if err != nil {
		return err
	}
	if novel == nil {
		return fmt.Errorf("novel not found")
	}

	item := &models.CollectionItem{
		CollectionID: collectionID,
		NovelID:      req.NovelID,
		Note:         req.Note,
	}
	if req.Position != nil {
		item.Position = *req.Position
	}

	return s.collectionRepo.AddItem(ctx, item)
}

// RemoveItem removes a novel from a collection
func (s *CollectionService) RemoveItem(ctx context.Context, collectionID, novelID, userID uuid.UUID) error {
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}
	if collection.UserID != userID {
		return fmt.Errorf("not authorized")
	}

	return s.collectionRepo.RemoveItem(ctx, collectionID, novelID)
}

// UpdateItem updates a collection item
func (s *CollectionService) UpdateItem(ctx context.Context, collectionID, novelID, userID uuid.UUID, req *models.UpdateCollectionItemRequest) error {
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}
	if collection.UserID != userID {
		return fmt.Errorf("not authorized")
	}

	return s.collectionRepo.UpdateItem(ctx, collectionID, novelID, req.Position, req.Note)
}

// ReorderItems reorders items in a collection
func (s *CollectionService) ReorderItems(ctx context.Context, collectionID, userID uuid.UUID, req *models.ReorderCollectionItemsRequest) error {
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}
	if collection.UserID != userID {
		return fmt.Errorf("not authorized")
	}

	items := make([]struct {
		NovelID  uuid.UUID
		Position int
	}, len(req.Items))

	for i, item := range req.Items {
		items[i].NovelID = item.NovelID
		items[i].Position = item.Position
	}

	return s.collectionRepo.ReorderItems(ctx, collectionID, items)
}

// Vote votes on a collection
func (s *CollectionService) Vote(ctx context.Context, collectionID, userID uuid.UUID) error {
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}
	if !collection.IsPublic {
		return fmt.Errorf("cannot vote on private collection")
	}
	if collection.UserID == userID {
		return fmt.Errorf("cannot vote on own collection")
	}

	// Check if already voted
	existing, err := s.collectionRepo.GetUserVote(ctx, collectionID, userID)
	if err != nil {
		return err
	}
	if existing != nil {
		// Remove vote (toggle)
		return s.collectionRepo.RemoveVote(ctx, collectionID, userID)
	}

	vote := &models.CollectionVote{
		CollectionID: collectionID,
		UserID:       userID,
		Value:        1,
	}

	return s.collectionRepo.Vote(ctx, vote)
}

// GetFeatured gets featured collections
func (s *CollectionService) GetFeatured(ctx context.Context, limit int) ([]models.CollectionCard, error) {
	collections, err := s.collectionRepo.GetFeaturedCollections(ctx, limit)
	if err != nil {
		return nil, err
	}

	for i := range collections {
		covers, _ := s.collectionRepo.GetPreviewCovers(ctx, collections[i].ID, 4)
		collections[i].PreviewCovers = covers
	}

	return collections, nil
}

// GetPopular gets popular collections
func (s *CollectionService) GetPopular(ctx context.Context, limit int) ([]models.CollectionCard, error) {
	collections, err := s.collectionRepo.GetPopularCollections(ctx, limit)
	if err != nil {
		return nil, err
	}

	for i := range collections {
		covers, _ := s.collectionRepo.GetPreviewCovers(ctx, collections[i].ID, 4)
		collections[i].PreviewCovers = covers
	}

	return collections, nil
}

// SetFeatured sets featured status (admin only)
func (s *CollectionService) SetFeatured(ctx context.Context, id uuid.UUID, featured bool) error {
	collection, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	return s.collectionRepo.SetFeatured(ctx, id, featured)
}

// GenerateSlug generates a unique slug for a collection
func (s *CollectionService) GenerateSlug(title string) string {
	baseSlug := slug.Make(title)
	if baseSlug == "" {
		baseSlug = "collection"
	}
	return strings.ToLower(baseSlug)
}
