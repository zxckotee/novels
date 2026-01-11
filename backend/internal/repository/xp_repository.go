package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels-backend/internal/domain/models"
)

type XPRepository struct {
	db *sqlx.DB
}

func NewXPRepository(db *sqlx.DB) *XPRepository {
	return &XPRepository{db: db}
}

// GetUserXP retrieves user's XP data
func (r *XPRepository) GetUserXP(ctx context.Context, userID uuid.UUID) (*models.UserXP, error) {
	var xp models.UserXP
	query := `SELECT user_id, xp_total, level, updated_at FROM user_xp WHERE user_id = $1`
	
	err := r.db.GetContext(ctx, &xp, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default values for new user
			return &models.UserXP{
				UserID:  userID,
				XPTotal: 0,
				Level:   1,
			}, nil
		}
		return nil, err
	}
	
	return &xp, nil
}

// CreateOrUpdateXP creates or updates user's XP
func (r *XPRepository) CreateOrUpdateXP(ctx context.Context, userID uuid.UUID, delta int64) (*models.UserXP, error) {
	query := `
		INSERT INTO user_xp (user_id, xp_total, level, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			xp_total = user_xp.xp_total + $2,
			level = $3,
			updated_at = NOW()
		RETURNING user_id, xp_total, level, updated_at`
	
	// First get current XP to calculate new level
	currentXP, _ := r.GetUserXP(ctx, userID)
	newXPTotal := currentXP.XPTotal + delta
	newLevel := models.CalculateLevel(newXPTotal)
	
	var xp models.UserXP
	err := r.db.QueryRowxContext(ctx, query, userID, delta, newLevel).Scan(
		&xp.UserID, &xp.XPTotal, &xp.Level, &xp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// Recalculate level based on actual total
	xp.Level = models.CalculateLevel(xp.XPTotal)
	
	// Update level if changed
	if xp.Level != newLevel {
		_, _ = r.db.ExecContext(ctx, `UPDATE user_xp SET level = $2 WHERE user_id = $1`, userID, xp.Level)
	}
	
	return &xp, nil
}

// AddXPEvent records an XP event
func (r *XPRepository) AddXPEvent(ctx context.Context, event *models.XPEvent) error {
	query := `
		INSERT INTO xp_events (id, user_id, type, delta, ref_type, ref_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`
	
	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.UserID,
		event.Type,
		event.Delta,
		event.RefType,
		event.RefID,
	)
	return err
}

// GetXPEvents retrieves XP events for a user
func (r *XPRepository) GetXPEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.XPEvent, error) {
	var events []models.XPEvent
	query := `
		SELECT id, user_id, type, delta, ref_type, ref_id, created_at
		FROM xp_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	err := r.db.SelectContext(ctx, &events, query, userID, limit, offset)
	return events, err
}

// CheckEventExists checks if an XP event already exists (for idempotency)
func (r *XPRepository) CheckEventExists(ctx context.Context, userID uuid.UUID, eventType models.XPEventType, refType string, refID uuid.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM xp_events 
			WHERE user_id = $1 AND type = $2 AND ref_type = $3 AND ref_id = $4
		)`
	
	err := r.db.GetContext(ctx, &exists, query, userID, eventType, refType, refID)
	return exists, err
}

// GetLeaderboard retrieves top users by XP
func (r *XPRepository) GetLeaderboard(ctx context.Context, limit int) ([]models.XPLeaderboardEntry, error) {
	query := `
		SELECT 
			ux.user_id,
			COALESCE(up.display_name, u.email) as display_name,
			up.avatar_url,
			ux.level,
			ux.xp_total
		FROM user_xp ux
		JOIN users u ON ux.user_id = u.id
		LEFT JOIN user_profiles up ON u.id = up.user_id
		WHERE u.is_banned = false
		ORDER BY ux.xp_total DESC
		LIMIT $1`
	
	rows, err := r.db.QueryxContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	entries := make([]models.XPLeaderboardEntry, 0)
	rank := 0
	for rows.Next() {
		rank++
		var entry models.XPLeaderboardEntry
		err := rows.Scan(
			&entry.UserID,
			&entry.DisplayName,
			&entry.AvatarURL,
			&entry.Level,
			&entry.XPTotal,
		)
		if err != nil {
			return nil, err
		}
		entry.Rank = rank
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// GetUserRank retrieves user's rank in XP leaderboard
func (r *XPRepository) GetUserRank(ctx context.Context, userID uuid.UUID) (int, error) {
	var rank int
	query := `
		SELECT COUNT(*) + 1
		FROM user_xp ux
		JOIN users u ON ux.user_id = u.id
		WHERE u.is_banned = false
		AND ux.xp_total > (SELECT COALESCE(xp_total, 0) FROM user_xp WHERE user_id = $1)`
	
	err := r.db.GetContext(ctx, &rank, query, userID)
	return rank, err
}

// GetUserStats retrieves aggregated user statistics
func (r *XPRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	stats := &models.UserStats{UserID: userID}
	
	// Count chapters read
	query := `SELECT COUNT(*) FROM reading_progress WHERE user_id = $1`
	_ = r.db.GetContext(ctx, &stats.ChaptersRead, query, userID)
	
	// Count comments
	query = `SELECT COUNT(*) FROM comments WHERE user_id = $1 AND is_deleted = false`
	_ = r.db.GetContext(ctx, &stats.CommentsCount, query, userID)
	
	// Count bookmarks
	query = `SELECT COUNT(*) FROM bookmarks WHERE user_id = $1`
	_ = r.db.GetContext(ctx, &stats.BookmarksCount, query, userID)
	
	return stats, nil
}

// GetAchievements retrieves all achievements
func (r *XPRepository) GetAchievements(ctx context.Context) ([]models.Achievement, error) {
	var achievements []models.Achievement
	query := `SELECT id, code, title, description, icon_key, condition, xp_reward, created_at FROM achievements ORDER BY created_at`
	err := r.db.SelectContext(ctx, &achievements, query)
	return achievements, err
}

// GetUserAchievements retrieves user's unlocked achievements
func (r *XPRepository) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]models.UserAchievement, error) {
	var userAchievements []models.UserAchievement
	query := `
		SELECT 
			ua.user_id, ua.achievement_id, ua.unlocked_at,
			a.id as "achievement.id",
			a.code as "achievement.code",
			a.title as "achievement.title",
			a.description as "achievement.description",
			a.icon_key as "achievement.icon_key"
		FROM user_achievements ua
		JOIN achievements a ON ua.achievement_id = a.id
		WHERE ua.user_id = $1
		ORDER BY ua.unlocked_at DESC`
	
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var ua models.UserAchievement
		var achievement models.Achievement
		err := rows.Scan(
			&ua.UserID, &ua.AchievementID, &ua.UnlockedAt,
			&achievement.ID, &achievement.Code, &achievement.Title,
			&achievement.Description, &achievement.IconKey,
		)
		if err != nil {
			return nil, err
		}
		ua.Achievement = &achievement
		userAchievements = append(userAchievements, ua)
	}
	
	return userAchievements, nil
}

// UnlockAchievement unlocks an achievement for a user
func (r *XPRepository) UnlockAchievement(ctx context.Context, userID, achievementID uuid.UUID) error {
	query := `
		INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, achievement_id) DO NOTHING`
	
	_, err := r.db.ExecContext(ctx, query, userID, achievementID)
	return err
}

// HasAchievement checks if user has an achievement
func (r *XPRepository) HasAchievement(ctx context.Context, userID, achievementID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_achievements WHERE user_id = $1 AND achievement_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, userID, achievementID)
	return exists, err
}
