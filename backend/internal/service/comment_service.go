package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"novels/internal/domain/models"
	"novels/internal/repository"
)

var (
	ErrCommentNotFound     = errors.New("comment not found")
	ErrCommentDeleted      = errors.New("comment is deleted")
	ErrCommentNotOwned     = errors.New("you can only edit your own comments")
	ErrInvalidParent       = errors.New("invalid parent comment")
	ErrMaxDepthExceeded    = errors.New("maximum reply depth exceeded")
	ErrCannotVoteOwnComment = errors.New("cannot vote on your own comment")
)

const MaxCommentDepth = 5

type CommentService struct {
	commentRepo *repository.CommentRepository
	xpService   *XPService
}

func NewCommentService(commentRepo *repository.CommentRepository, xpService *XPService) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		xpService:   xpService,
	}
}

// Create creates a new comment
func (s *CommentService) Create(ctx context.Context, req models.CreateCommentRequest, userID uuid.UUID) (*models.Comment, error) {
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return nil, err
	}

	comment := &models.Comment{
		ID:         uuid.New(),
		TargetType: req.TargetType,
		TargetID:   targetID,
		UserID:     userID,
		Body:       req.Body,
		IsSpoiler:  req.IsSpoiler,
		Depth:      0,
	}

	// Handle reply
	if req.ParentID != nil {
		parentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, ErrInvalidParent
		}

		parent, err := s.commentRepo.GetByID(ctx, parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrInvalidParent
		}
		if parent.IsDeleted {
			return nil, ErrCommentDeleted
		}

		// Check max depth
		if parent.Depth >= MaxCommentDepth {
			return nil, ErrMaxDepthExceeded
		}

		comment.ParentID = &parentID
		comment.Depth = parent.Depth + 1

		// Set root ID
		if parent.RootID != nil {
			comment.RootID = parent.RootID
		} else {
			comment.RootID = &parent.ID
		}
	}

	err = s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Update parent's replies count
	if comment.ParentID != nil {
		_ = s.commentRepo.UpdateRepliesCount(ctx, *comment.ParentID)
	}

	// Award XP for commenting
	if s.xpService != nil {
		_ = s.xpService.AwardXP(ctx, userID, models.XPEventComment, 5, "comment", comment.ID)
	}

	return s.commentRepo.GetByIDWithUser(ctx, comment.ID, &userID)
}

// GetByID retrieves a comment by ID
func (s *CommentService) GetByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*models.Comment, error) {
	comment, err := s.commentRepo.GetByIDWithUser(ctx, id, viewerID)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	return comment, nil
}

// List retrieves comments with filters
func (s *CommentService) List(ctx context.Context, filter models.CommentsFilter, viewerID *uuid.UUID) (*models.CommentsResponse, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Sort == "" {
		filter.Sort = "newest"
	}

	return s.commentRepo.List(ctx, filter, viewerID)
}

// Update updates a comment
func (s *CommentService) Update(ctx context.Context, id uuid.UUID, req models.UpdateCommentRequest, userID uuid.UUID) (*models.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	if comment.IsDeleted {
		return nil, ErrCommentDeleted
	}
	if comment.UserID != userID {
		return nil, ErrCommentNotOwned
	}

	err = s.commentRepo.Update(ctx, id, req.Body, req.IsSpoiler)
	if err != nil {
		return nil, err
	}

	return s.commentRepo.GetByIDWithUser(ctx, id, &userID)
}

// Delete soft-deletes a comment
func (s *CommentService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}

	// Only owner or admin can delete
	if comment.UserID != userID && !isAdmin {
		return ErrCommentNotOwned
	}

	return s.commentRepo.Delete(ctx, id)
}

// Vote votes on a comment
func (s *CommentService) Vote(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, value int) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	if comment.IsDeleted {
		return ErrCommentDeleted
	}
	if comment.UserID == userID {
		return ErrCannotVoteOwnComment
	}

	return s.commentRepo.Vote(ctx, commentID, userID, value)
}

// Report reports a comment
func (s *CommentService) Report(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, reason string) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}

	report := &models.CommentReport{
		ID:        uuid.New(),
		CommentID: commentID,
		UserID:    userID,
		Reason:    reason,
		Status:    "pending",
	}

	return s.commentRepo.Report(ctx, report)
}

// GetReplies gets replies to a comment
func (s *CommentService) GetReplies(ctx context.Context, parentID uuid.UUID, limit int, viewerID *uuid.UUID) ([]models.Comment, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.commentRepo.GetReplies(ctx, parentID, limit, viewerID)
}
