package service

import (
	"context"
	"encoding/json"
	"fmt"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
)

// AdminRepository интерфейс для работы с админ-функциями
type AdminRepository interface {
	GetSettings(ctx context.Context) ([]models.AppSetting, error)
	GetSetting(ctx context.Context, key string) (*models.AppSetting, error)
	UpdateSetting(ctx context.Context, key string, value json.RawMessage, updatedBy uuid.UUID) error
	GetAuditLogs(ctx context.Context, filter models.AdminLogsFilter) ([]models.AdminAuditLog, int, error)
	LogAction(ctx context.Context, actorUserID uuid.UUID, action, entityType string, entityID *uuid.UUID, details json.RawMessage, ipAddress, userAgent string) error
	GetStats(ctx context.Context) (*models.AdminStatsOverview, error)
}

// AdminService сервис для системных админ-функций
type AdminService struct {
	adminRepo AdminRepository
}

func NewAdminService(adminRepo AdminRepository) *AdminService {
	return &AdminService{adminRepo: adminRepo}
}

// GetSettings получает все настройки
func (s *AdminService) GetSettings(ctx context.Context) ([]models.AppSetting, error) {
	return s.adminRepo.GetSettings(ctx)
}

// GetSetting получает настройку по ключу
func (s *AdminService) GetSetting(ctx context.Context, key string) (*models.AppSetting, error) {
	setting, err := s.adminRepo.GetSetting(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}
	if setting == nil {
		return nil, ErrNotFound
	}
	return setting, nil
}

// UpdateSetting обновляет настройку
func (s *AdminService) UpdateSetting(ctx context.Context, key string, value json.RawMessage, updatedBy uuid.UUID) error {
	// Проверяем, что настройка существует
	_, err := s.adminRepo.GetSetting(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check setting: %w", err)
	}

	return s.adminRepo.UpdateSetting(ctx, key, value, updatedBy)
}

// GetAuditLogs получает логи с фильтрацией
func (s *AdminService) GetAuditLogs(ctx context.Context, filter models.AdminLogsFilter) (*models.AdminLogsResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}

	logs, total, err := s.adminRepo.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	if logs == nil {
		logs = []models.AdminAuditLog{}
	}

	return &models.AdminLogsResponse{
		Logs:       logs,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

// LogAction сохраняет действие в логе
func (s *AdminService) LogAction(ctx context.Context, actorUserID uuid.UUID, action, entityType string, entityID *uuid.UUID, details json.RawMessage, ipAddress, userAgent string) error {
	return s.adminRepo.LogAction(ctx, actorUserID, action, entityType, entityID, details, ipAddress, userAgent)
}

// GetStats получает статистику платформы
func (s *AdminService) GetStats(ctx context.Context) (*models.AdminStatsOverview, error) {
	return s.adminRepo.GetStats(ctx)
}
