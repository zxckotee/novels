package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"novels/internal/domain/models"
	"novels/internal/repository"
)

var (
	ErrInvalidListCode = errors.New("invalid bookmark list code")
	ErrNovelNotFound   = errors.New("novel not found")
)

type BookmarkService struct {
	bookmarkRepo *repository.BookmarkRepository
	novelRepo    *repository.NovelRepository
	xpService    *XPService
}

func NewBookmarkService(
	bookmarkRepo *repository.BookmarkRepository,
	novelRepo *repository.NovelRepository,
	xpService *XPService,
) *BookmarkService {
	return &BookmarkService{
		bookmarkRepo: bookmarkRepo,
		novelRepo:    novelRepo,
		xpService:    xpService,
	}
}

// GetLists retrieves user's bookmark lists
func (s *BookmarkService) GetLists(ctx context.Context, userID uuid.UUID) ([]models.BookmarkList, error) {
	return s.bookmarkRepo.GetOrCreateLists(ctx, userID)
}

// List retrieves bookmarks with filters
func (s *BookmarkService) List(ctx context.Context, userID uuid.UUID, filter models.BookmarksFilter, lang string) (*models.BookmarksResponse, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		filter.Limit = 20
	}
	if filter.Sort == "" {
		filter.Sort = "latest_update"
	}
	
	filter.UserID = userID
	
	return s.bookmarkRepo.List(ctx, filter, lang)
}

// AddBookmark adds a novel to a bookmark list
func (s *BookmarkService) AddBookmark(ctx context.Context, userID uuid.UUID, novelIDStr string, listCode models.BookmarkListCode) (*models.Bookmark, error) {
	// Validate list code
	if !isValidListCode(listCode) {
		return nil, ErrInvalidListCode
	}
	
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		return nil, ErrNovelNotFound
	}
	
	// Check if novel exists
	novel, err := s.novelRepo.GetByID(ctx, novelID, "ru")
	if err != nil || novel == nil {
		return nil, ErrNovelNotFound
	}
	
	// Get or create lists
	lists, err := s.bookmarkRepo.GetOrCreateLists(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Find the target list
	var targetList *models.BookmarkList
	for _, list := range lists {
		if list.Code == listCode {
			targetList = &list
			break
		}
	}
	
	if targetList == nil {
		return nil, ErrInvalidListCode
	}
	
	// Check if bookmark already exists
	existing, err := s.bookmarkRepo.GetBookmark(ctx, userID, novelID)
	if err != nil {
		return nil, err
	}
	
	if existing != nil {
		// Update existing bookmark
		err = s.bookmarkRepo.Update(ctx, userID, novelID, targetList.ID)
		if err != nil {
			return nil, err
		}
		existing.ListID = targetList.ID
		return existing, nil
	}
	
	// Create new bookmark
	bookmark := &models.Bookmark{
		ID:      uuid.New(),
		UserID:  userID,
		NovelID: novelID,
		ListID:  targetList.ID,
	}
	
	err = s.bookmarkRepo.Create(ctx, bookmark)
	if err != nil {
		return nil, err
	}
	
	// Award XP for first bookmark
	if s.xpService != nil {
		_ = s.xpService.AwardXP(ctx, userID, models.XPEventBookmark, 3, "novel", novelID)
	}
	
	return bookmark, nil
}

// UpdateBookmark moves a bookmark to another list
func (s *BookmarkService) UpdateBookmark(ctx context.Context, userID uuid.UUID, novelIDStr string, listCode models.BookmarkListCode) error {
	if !isValidListCode(listCode) {
		return ErrInvalidListCode
	}
	
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		return ErrNovelNotFound
	}
	
	// Get the target list
	list, err := s.bookmarkRepo.GetListByCode(ctx, userID, listCode)
	if err != nil {
		return err
	}
	if list == nil {
		return ErrInvalidListCode
	}
	
	return s.bookmarkRepo.Update(ctx, userID, novelID, list.ID)
}

// RemoveBookmark removes a novel from bookmarks
func (s *BookmarkService) RemoveBookmark(ctx context.Context, userID uuid.UUID, novelIDStr string) error {
	novelID, err := uuid.Parse(novelIDStr)
	if err != nil {
		return ErrNovelNotFound
	}
	
	return s.bookmarkRepo.Delete(ctx, userID, novelID)
}

// GetBookmarkStatus gets bookmark status for a novel
func (s *BookmarkService) GetBookmarkStatus(ctx context.Context, userID, novelID uuid.UUID) (*models.BookmarkListCode, error) {
	return s.bookmarkRepo.GetNovelBookmarkStatus(ctx, userID, novelID)
}

// GetStats gets bookmark statistics for a user
func (s *BookmarkService) GetStats(ctx context.Context, userID uuid.UUID) ([]models.BookmarkListStats, error) {
	return s.bookmarkRepo.GetStats(ctx, userID)
}

// isValidListCode checks if the list code is valid
func isValidListCode(code models.BookmarkListCode) bool {
	for _, validCode := range models.SystemBookmarkLists {
		if code == validCode {
			return true
		}
	}
	return false
}
