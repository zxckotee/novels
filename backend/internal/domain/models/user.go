package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole представляет роль пользователя
type UserRole string

const (
	RoleGuest     UserRole = "guest"
	RoleUser      UserRole = "user"
	RolePremium   UserRole = "premium"
	RoleModerator UserRole = "moderator"
	RoleAdmin     UserRole = "admin"
)

// User представляет пользователя системы
type User struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	IsBanned     bool       `db:"is_banned" json:"is_banned"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// UserProfile представляет профиль пользователя
type UserProfile struct {
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	DisplayName string    `db:"display_name" json:"display_name"`
	AvatarKey   *string   `db:"avatar_key" json:"avatar_key,omitempty"`
	Bio         *string   `db:"bio" json:"bio,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// UserWithProfile объединяет пользователя и его профиль
type UserWithProfile struct {
	User
	Profile UserProfile `json:"profile"`
	Roles   []UserRole  `json:"roles"`
}

// RefreshToken представляет refresh токен
type RefreshToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email,max=255"`
	Password    string `json:"password" validate:"required,min=8,max=72"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
}

// LoginRequest представляет запрос на вход
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse представляет ответ на успешную аутентификацию
type AuthResponse struct {
	User        *UserWithProfile `json:"user"`
	AccessToken string           `json:"access_token"`
	ExpiresIn   int              `json:"expires_in"` // в секундах
}

// UserResponse представляет публичную информацию о пользователе
type UserResponse struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Roles       []UserRole `json:"roles"`
	CreatedAt   time.Time  `json:"created_at"`
}
