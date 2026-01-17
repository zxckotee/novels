package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AdminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetSettings получает все настройки
func (r *AdminRepository) GetSettings(ctx context.Context) ([]models.AppSetting, error) {
	query := `SELECT key, value, description, updated_by, updated_at FROM app_settings ORDER BY key`
	
	var settings []models.AppSetting
	err := r.db.SelectContext(ctx, &settings, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	
	return settings, nil
}

// GetSetting получает одну настройку по ключу
func (r *AdminRepository) GetSetting(ctx context.Context, key string) (*models.AppSetting, error) {
	query := `SELECT key, value, description, updated_by, updated_at FROM app_settings WHERE key = $1`
	
	var setting models.AppSetting
	err := r.db.GetContext(ctx, &setting, query, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}
	
	return &setting, nil
}

// UpdateSetting обновляет настройку
func (r *AdminRepository) UpdateSetting(ctx context.Context, key string, value json.RawMessage, updatedBy uuid.UUID) error {
	query := `UPDATE app_settings SET value = $1, updated_by = $2, updated_at = NOW() WHERE key = $3`
	
	result, err := r.db.ExecContext(ctx, query, value, updatedBy, key)
	if err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("setting '%s' not found", key)
	}
	
	return nil
}

// GetAuditLogs получает логи с фильтрацией
func (r *AdminRepository) GetAuditLogs(ctx context.Context, filter models.AdminLogsFilter) ([]models.AdminAuditLog, int, error) {
	baseQuery := `FROM admin_audit_log al`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.ActorUserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("al.actor_user_id = $%d", argIndex))
		args = append(args, *filter.ActorUserID)
		argIndex++
	}

	if filter.Action != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("al.action = $%d", argIndex))
		args = append(args, filter.Action)
		argIndex++
	}

	if filter.EntityType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("al.entity_type = $%d", argIndex))
		args = append(args, filter.EntityType)
		argIndex++
	}

	if filter.EntityID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("al.entity_id = $%d", argIndex))
		args = append(args, *filter.EntityID)
		argIndex++
	}

	if filter.StartDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("al.created_at >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("al.created_at <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count
	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) "+baseQuery+" "+whereClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Select
	selectQuery := fmt.Sprintf(`
		SELECT al.id, al.actor_user_id, al.action, al.entity_type, al.entity_id, 
		       al.details, al.ip_address, al.user_agent, al.created_at
		%s %s
		ORDER BY al.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, argIndex, argIndex+1)

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var logs []models.AdminAuditLog
	err = r.db.SelectContext(ctx, &logs, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get logs: %w", err)
	}

	return logs, total, nil
}

// LogAction создает запись в логе
func (r *AdminRepository) LogAction(ctx context.Context, actorUserID uuid.UUID, action, entityType string, entityID *uuid.UUID, details json.RawMessage, ipAddress, userAgent string) error {
	query := `
		INSERT INTO admin_audit_log (actor_user_id, action, entity_type, entity_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := r.db.ExecContext(ctx, query, actorUserID, action, entityType, entityID, details, ipAddress, userAgent)
	if err != nil {
		return fmt.Errorf("failed to log action: %w", err)
	}
	
	return nil
}

// GetStats получает статистику платформы
func (r *AdminRepository) GetStats(ctx context.Context) (*models.AdminStatsOverview, error) {
	stats := &models.AdminStatsOverview{}
	
	// Total counts
	row := r.db.QueryRowContext(ctx, `
		SELECT 
			(SELECT COUNT(*) FROM novels) as total_novels,
			(SELECT COUNT(*) FROM chapters) as total_chapters,
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM comments WHERE NOT is_deleted) as total_comments,
			(SELECT COUNT(*) FROM comment_reports WHERE status = 'pending') as pending_reports
	`)
	
	err := row.Scan(&stats.TotalNovels, &stats.TotalChapters, &stats.TotalUsers, &stats.TotalComments, &stats.PendingReports)
	if err != nil {
		return nil, fmt.Errorf("failed to get total stats: %w", err)
	}

	// Weekly stats
	weekAgo := time.Now().AddDate(0, 0, -7)
	row = r.db.QueryRowContext(ctx, `
		SELECT 
			(SELECT COUNT(*) FROM users WHERE created_at >= $1) as new_users,
			(SELECT COUNT(*) FROM novels WHERE created_at >= $1) as new_novels,
			(SELECT COUNT(*) FROM chapters WHERE created_at >= $1) as new_chapters
	`, weekAgo)
	
	err = row.Scan(&stats.NewUsersThisWeek, &stats.NewNovelsThisWeek, &stats.NewChaptersThisWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly stats: %w", err)
	}

	// Averages (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	row = r.db.QueryRowContext(ctx, `
		SELECT 
			COALESCE(COUNT(*)::float / 30, 0) as avg_chapters,
			(SELECT COALESCE(COUNT(*)::float / 30, 0) FROM comments WHERE created_at >= $1 AND NOT is_deleted) as avg_comments
		FROM chapters
		WHERE created_at >= $1
	`, thirtyDaysAgo)
	
	err = row.Scan(&stats.AvgChaptersPerDay, &stats.AvgCommentsPerDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get average stats: %w", err)
	}

	return stats, nil
}
