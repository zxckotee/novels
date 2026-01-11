package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// UserRepository репозиторий для работы с пользователями
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository создает новый UserRepository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *models.User, profile *models.UserProfile) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаем пользователя
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("user with this email already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Создаем профиль
	profileQuery := `
		INSERT INTO user_profiles (user_id, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.ExecContext(ctx, profileQuery, profile.UserID, profile.DisplayName, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user profile: %w", err)
	}

	// Добавляем роль user по умолчанию
	roleQuery := `INSERT INTO user_roles (user_id, role) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, roleQuery, user.ID, models.RoleUser)
	if err != nil {
		return fmt.Errorf("failed to add user role: %w", err)
	}

	return tx.Commit()
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UserWithProfile, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.is_banned, u.last_login_at, u.created_at, u.updated_at,
		       p.display_name, p.avatar_key, p.bio
		FROM users u
		JOIN user_profiles p ON u.id = p.user_id
		WHERE u.id = $1
	`
	
	var user models.UserWithProfile
	var profile struct {
		DisplayName string  `db:"display_name"`
		AvatarKey   *string `db:"avatar_key"`
		Bio         *string `db:"bio"`
	}

	row := r.db.QueryRowxContext(ctx, query, id)
	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.IsBanned, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&profile.DisplayName, &profile.AvatarKey, &profile.Bio,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Profile = models.UserProfile{
		UserID:      user.ID,
		DisplayName: profile.DisplayName,
		AvatarKey:   profile.AvatarKey,
		Bio:         profile.Bio,
	}

	// Получаем роли пользователя
	roles, err := r.GetRoles(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.UserWithProfile, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.is_banned, u.last_login_at, u.created_at, u.updated_at,
		       p.display_name, p.avatar_key, p.bio
		FROM users u
		JOIN user_profiles p ON u.id = p.user_id
		WHERE u.email = $1
	`
	
	var user models.UserWithProfile
	var profile struct {
		DisplayName string  `db:"display_name"`
		AvatarKey   *string `db:"avatar_key"`
		Bio         *string `db:"bio"`
	}

	row := r.db.QueryRowxContext(ctx, query, email)
	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.IsBanned, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&profile.DisplayName, &profile.AvatarKey, &profile.Bio,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Profile = models.UserProfile{
		UserID:      user.ID,
		DisplayName: profile.DisplayName,
		AvatarKey:   profile.AvatarKey,
		Bio:         profile.Bio,
	}

	// Получаем роли пользователя
	roles, err := r.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return &user, nil
}

// GetRoles получает роли пользователя
func (r *UserRepository) GetRoles(ctx context.Context, userID uuid.UUID) ([]models.UserRole, error) {
	query := `SELECT role FROM user_roles WHERE user_id = $1`
	
	var roles []models.UserRole
	err := r.db.SelectContext(ctx, &roles, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// UpdateLastLogin обновляет время последнего входа
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// SaveRefreshToken сохраняет refresh токен
func (r *UserRepository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

// GetRefreshToken получает refresh токен по хешу
func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`
	
	var token models.RefreshToken
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

// RevokeRefreshToken отзывает refresh токен
func (r *UserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE token_hash = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllUserTokens отзывает все токены пользователя
func (r *UserRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}
	return nil
}

// CleanupExpiredTokens удаляет истекшие токены
func (r *UserRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked_at IS NOT NULL`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}

// UpdatePassword обновляет пароль пользователя
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
