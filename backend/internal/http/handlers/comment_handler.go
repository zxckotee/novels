package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/http/middleware"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// List godoc
// @Summary List comments
// @Description Get comments for a target (novel, chapter, news)
// @Tags comments
// @Accept json
// @Produce json
// @Param target_type query string true "Target type (novel, chapter, news)"
// @Param target_id query string true "Target ID (UUID)"
// @Param parent_id query string false "Parent comment ID for nested replies"
// @Param sort query string false "Sort order (newest, oldest, top)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{data=models.CommentsResponse}
// @Router /comments [get]
func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	targetType := r.URL.Query().Get("target_type")
	targetIDStr := r.URL.Query().Get("target_id")
	parentIDStr := r.URL.Query().Get("parent_id")
	anchor := r.URL.Query().Get("anchor")
	sort := r.URL.Query().Get("sort")
	
	if targetType == "" || targetIDStr == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "target_type and target_id are required")
		return
	}
	
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid target_id")
		return
	}
	
	filter := models.CommentsFilter{
		TargetType: models.TargetType(targetType),
		TargetID:   targetID,
		Sort:       sort,
		Page:       parseIntQuery(r, "page", 1),
		Limit:      parseIntQuery(r, "limit", 20),
	}

	if anchor != "" {
		filter.Anchor = &anchor
	}
	
	if parentIDStr != "" {
		parentID, err := uuid.Parse(parentIDStr)
		if err == nil {
			filter.ParentID = &parentID
		}
	}
	
	// Get viewer ID if logged in
	var viewerID *uuid.UUID
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			viewerID = &userID
		}
	}
	
	result, err := h.commentService.List(r.Context(), filter, viewerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get comments")
		return
	}
	
	response.JSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary Get comment by ID
// @Description Get a single comment by ID
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} response.Response{data=models.Comment}
// @Router /comments/{id} [get]
func (h *CommentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}
	
	var viewerID *uuid.UUID
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			viewerID = &userID
		}
	}
	
	comment, err := h.commentService.GetByID(r.Context(), id, viewerID)
	if err != nil {
		if err == service.ErrCommentNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "comment not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get comment")
		return
	}
	
	response.JSON(w, http.StatusOK, comment)
}

// Create godoc
// @Summary Create comment
// @Description Create a new comment (requires authentication)
// @Tags comments
// @Accept json
// @Produce json
// @Param body body models.CreateCommentRequest true "Comment data"
// @Success 201 {object} response.Response{data=models.Comment}
// @Router /comments [post]
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	
	// Validate
	if req.Body == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "body is required")
		return
	}
	if req.TargetType == "" || req.TargetID == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "target_type and target_id are required")
		return
	}
	
	comment, err := h.commentService.Create(r.Context(), req, userID)
	if err != nil {
		switch err {
		case service.ErrInvalidParent:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid parent comment")
		case service.ErrCommentDeleted:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "cannot reply to deleted comment")
		case service.ErrMaxDepthExceeded:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "maximum reply depth exceeded")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create comment")
		}
		return
	}
	
	response.JSON(w, http.StatusCreated, comment)
}

// Update godoc
// @Summary Update comment
// @Description Update own comment (requires authentication)
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param body body models.UpdateCommentRequest true "Comment data"
// @Success 200 {object} response.Response{data=models.Comment}
// @Router /comments/{id} [put]
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}
	
	var req models.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	
	comment, err := h.commentService.Update(r.Context(), id, req, userID)
	if err != nil {
		switch err {
		case service.ErrCommentNotFound:
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "comment not found")
		case service.ErrCommentNotOwned:
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "you can only edit your own comments")
		case service.ErrCommentDeleted:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "cannot edit deleted comment")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update comment")
		}
		return
	}
	
	response.JSON(w, http.StatusOK, comment)
}

// Delete godoc
// @Summary Delete comment
// @Description Delete own comment (requires authentication)
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Success 204 "No Content"
// @Router /comments/{id} [delete]
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}
	
	// Check if user is admin
	role := middleware.GetUserRole(r.Context())
	isAdmin := role == string(models.RoleAdmin) || role == string(models.RoleModerator)
	
	err = h.commentService.Delete(r.Context(), id, userID, isAdmin)
	if err != nil {
		switch err {
		case service.ErrCommentNotFound:
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "comment not found")
		case service.ErrCommentNotOwned:
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "you can only delete your own comments")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete comment")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Vote godoc
// @Summary Vote on comment
// @Description Like or dislike a comment (requires authentication)
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param body body models.VoteCommentRequest true "Vote data"
// @Success 204 "No Content"
// @Router /comments/{id}/vote [post]
func (h *CommentHandler) Vote(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}
	
	var req models.VoteCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	
	if req.Value != 1 && req.Value != -1 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "value must be 1 or -1")
		return
	}
	
	err = h.commentService.Vote(r.Context(), id, userID, req.Value)
	if err != nil {
		switch err {
		case service.ErrCommentNotFound:
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "comment not found")
		case service.ErrCommentDeleted:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "cannot vote on deleted comment")
		case service.ErrCannotVoteOwnComment:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "cannot vote on your own comment")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to vote")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Report godoc
// @Summary Report comment
// @Description Report a comment for moderation (requires authentication)
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param body body models.ReportCommentRequest true "Report data"
// @Success 204 "No Content"
// @Router /comments/{id}/report [post]
func (h *CommentHandler) Report(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id")
		return
	}
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}
	
	var req models.ReportCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	
	if len(req.Reason) < 10 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "reason must be at least 10 characters")
		return
	}
	
	err = h.commentService.Report(r.Context(), id, userID, req.Reason)
	if err != nil {
		if err == service.ErrCommentNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "comment not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to report comment")
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetReplies godoc
// @Summary Get comment replies
// @Description Get replies to a specific comment
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param limit query int false "Number of replies" default(10)
// @Success 200 {object} response.Response{data=[]models.Comment}
// @Router /comments/{id}/replies [get]
func (h *CommentHandler) GetReplies(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}
	
	limit := parseIntQuery(r, "limit", 10)
	
	var viewerID *uuid.UUID
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			viewerID = &userID
		}
	}
	
	replies, err := h.commentService.GetReplies(r.Context(), id, limit, viewerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get replies")
		return
	}
	
	response.JSON(w, http.StatusOK, replies)
}
