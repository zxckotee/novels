package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"novels-backend/internal/domain/models"
	"novels-backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CommentAdminRepository интерфейс для админ работы с комментариями
type CommentAdminRepository interface {
	AdminListComments(ctx context.Context, filter models.AdminCommentsFilter) ([]models.Comment, int, error)
	SoftDeleteComment(ctx context.Context, commentID uuid.UUID) error
	HardDeleteComment(ctx context.Context, commentID uuid.UUID) error
	GetReports(ctx context.Context, filter models.ReportsFilter) ([]models.CommentReport, int, error)
	ResolveReport(ctx context.Context, reportID uuid.UUID, action, reason string) error
}

// CommentAdminHandler обработчик админских эндпоинтов для комментариев
type CommentAdminHandler struct {
	commentRepo CommentAdminRepository
}

func NewCommentAdminHandler(commentRepo CommentAdminRepository) *CommentAdminHandler {
	return &CommentAdminHandler{
		commentRepo: commentRepo,
	}
}

// ListComments получает список комментариев для админа
// GET /api/v1/admin/comments
func (h *CommentAdminHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	filter := models.AdminCommentsFilter{
		TargetType: models.TargetType(r.URL.Query().Get("targetType")),
		Sort:       r.URL.Query().Get("sort"),
		Page:       parseIntQuery(r, "page", 1),
		Limit:      parseIntQuery(r, "limit", 20),
	}

	if targetID := r.URL.Query().Get("targetId"); targetID != "" {
		if id, err := uuid.Parse(targetID); err == nil {
			filter.TargetID = &id
		}
	}

	if userID := r.URL.Query().Get("userId"); userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			filter.UserID = &id
		}
	}

	if isDeleted := r.URL.Query().Get("isDeleted"); isDeleted != "" {
		b := isDeleted == "true"
		filter.IsDeleted = &b
	}

	comments, total, err := h.commentRepo.AdminListComments(r.Context(), filter)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, models.CommentsResponse{
		Comments:   comments,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	})
}

// SoftDeleteComment помечает комментарий как удаленный
// DELETE /api/v1/admin/comments/{id}
func (h *CommentAdminHandler) SoftDeleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	commentID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid comment id")
		return
	}

	err = h.commentRepo.SoftDeleteComment(r.Context(), commentID)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "comment soft deleted"})
}

// HardDeleteComment полностью удаляет комментарий
// DELETE /api/v1/admin/comments/{id}/hard
func (h *CommentAdminHandler) HardDeleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	commentID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid comment id")
		return
	}

	err = h.commentRepo.HardDeleteComment(r.Context(), commentID)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "comment permanently deleted"})
}

// ListReports получает список жалоб
// GET /api/v1/admin/reports
func (h *CommentAdminHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	filter := models.ReportsFilter{
		Status: r.URL.Query().Get("status"),
		Sort:   r.URL.Query().Get("sort"),
		Page:   parseIntQuery(r, "page", 1),
		Limit:  parseIntQuery(r, "limit", 20),
	}

	if commentID := r.URL.Query().Get("commentId"); commentID != "" {
		if id, err := uuid.Parse(commentID); err == nil {
			filter.CommentID = &id
		}
	}

	if reporterID := r.URL.Query().Get("reporterId"); reporterID != "" {
		if id, err := uuid.Parse(reporterID); err == nil {
			filter.ReporterID = &id
		}
	}

	reports, total, err := h.commentRepo.GetReports(r.Context(), filter)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, models.ReportsResponse{
		Reports:    reports,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	})
}

// ResolveReport обрабатывает жалобу
// POST /api/v1/admin/reports/{id}/resolve
func (h *CommentAdminHandler) ResolveReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	reportID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid report id")
		return
	}

	var req models.ResolveReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Action == "" {
		response.BadRequest(w, "action is required")
		return
	}

	err = h.commentRepo.ResolveReport(r.Context(), reportID, req.Action, req.Reason)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, map[string]string{"message": "report processed"})
}
