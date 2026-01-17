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

// UserRepository интерфейс для работы с пользователями (админ методы)
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserWithProfile, error)
	ListUsers(ctx context.Context, filter models.UsersFilter) ([]models.User, int, error)
	BanUser(ctx context.Context, userID uuid.UUID, reason string) error
	UnbanUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserRoles(ctx context.Context, userID uuid.UUID, roles []string) error
}

// UserAdminHandler обработчик админских эндпоинтов для пользователей
type UserAdminHandler struct {
	userRepo UserRepository
}

func NewUserAdminHandler(userRepo UserRepository) *UserAdminHandler {
	return &UserAdminHandler{
		userRepo: userRepo,
	}
}

// ListUsers получает список пользователей
// GET /api/v1/admin/users
func (h *UserAdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	filter := models.UsersFilter{
		Query: r.URL.Query().Get("query"),
		Role:  r.URL.Query().Get("role"),
		Sort:  r.URL.Query().Get("sort"),
		Order: r.URL.Query().Get("order"),
		Page:  parseIntQuery(r, "page", 1),
		Limit: parseIntQuery(r, "limit", 20),
	}

	if banned := r.URL.Query().Get("banned"); banned != "" {
		b := banned == "true"
		filter.Banned = &b
	}

	users, total, err := h.userRepo.ListUsers(r.Context(), filter)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, models.UsersResponse{
		Users:      users,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	})
}

// GetUser получает пользователя по ID
// GET /api/v1/admin/users/{id}
func (h *UserAdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		response.InternalError(w)
		return
	}
	if user == nil {
		response.NotFound(w, "user not found")
		return
	}

	response.OK(w, user)
}

// BanUser блокирует пользователя
// POST /api/v1/admin/users/{id}/ban
func (h *UserAdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	var req models.BanUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Reason == "" {
		response.BadRequest(w, "reason is required")
		return
	}

	err = h.userRepo.BanUser(r.Context(), userID, req.Reason)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "user banned"})
}

// UnbanUser разблокирует пользователя
// POST /api/v1/admin/users/{id}/unban
func (h *UserAdminHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	err = h.userRepo.UnbanUser(r.Context(), userID)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "user unbanned"})
}

// UpdateUserRoles обновляет роли пользователя
// PUT /api/v1/admin/users/{id}/roles
func (h *UserAdminHandler) UpdateUserRoles(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	var req models.UpdateUserRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if len(req.Roles) == 0 {
		response.BadRequest(w, "roles are required")
		return
	}

	err = h.userRepo.UpdateUserRoles(r.Context(), userID, req.Roles)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, map[string]string{"message": "user roles updated"})
}
