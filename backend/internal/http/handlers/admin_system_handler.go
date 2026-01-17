package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AdminSystemHandler обработчик системных админ-эндпоинтов
type AdminSystemHandler struct {
	adminService *service.AdminService
}

func NewAdminSystemHandler(adminService *service.AdminService) *AdminSystemHandler {
	return &AdminSystemHandler{
		adminService: adminService,
	}
}

// GetSettings получает все настройки приложения
// GET /api/v1/admin/settings
func (h *AdminSystemHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.adminService.GetSettings(r.Context())
	if err != nil {
		response.InternalError(w)
		return
	}
	
	response.OK(w, settings)
}

// GetSetting получает одну настройку по ключу
// GET /api/v1/admin/settings/{key}
func (h *AdminSystemHandler) GetSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		response.BadRequest(w, "key is required")
		return
	}

	setting, err := h.adminService.GetSetting(r.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "setting not found")
			return
		}
		response.InternalError(w)
		return
	}
	
	response.OK(w, setting)
}

// UpdateSetting обновляет настройку
// PUT /api/v1/admin/settings/{key}
func (h *AdminSystemHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		response.BadRequest(w, "key is required")
		return
	}

	var req models.UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Get user ID from context (assuming middleware sets it)
	userIDValue := r.Context().Value("userID")
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		userID = uuid.Nil // Fallback
	}

	err := h.adminService.UpdateSetting(r.Context(), key, req.Value, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.NotFound(w, "setting not found")
			return
		}
		response.InternalError(w)
		return
	}
	
	response.OK(w, map[string]string{"message": "setting updated"})
}

// GetLogs получает логи действий администраторов
// GET /api/v1/admin/logs
func (h *AdminSystemHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	filter := models.AdminLogsFilter{
		Action:     r.URL.Query().Get("action"),
		EntityType: r.URL.Query().Get("entityType"),
		Page:       parseIntQuery(r, "page", 1),
		Limit:      parseIntQuery(r, "limit", 50),
	}

	// Parse actor user ID
	if actorID := r.URL.Query().Get("actorUserId"); actorID != "" {
		if id, err := uuid.Parse(actorID); err == nil {
			filter.ActorUserID = &id
		}
	}

	// Parse entity ID
	if entityID := r.URL.Query().Get("entityId"); entityID != "" {
		if id, err := uuid.Parse(entityID); err == nil {
			filter.EntityID = &id
		}
	}

	// Parse dates
	if startDateStr := r.URL.Query().Get("startDate"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endDateStr := r.URL.Query().Get("endDate"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &t
		}
	}

	result, err := h.adminService.GetAuditLogs(r.Context(), filter)
	if err != nil {
		response.InternalError(w)
		return
	}
	
	response.OK(w, result)
}

// GetStats получает статистику платформы
// GET /api/v1/admin/stats
func (h *AdminSystemHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetStats(r.Context())
	if err != nil {
		response.InternalError(w)
		return
	}
	
	response.OK(w, stats)
}
