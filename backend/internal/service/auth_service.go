package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"novels-backend/internal/config"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUserBanned         = errors.New("user is banned")
)

// AuthService сервис аутентификации
type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

// NewAuthService создает новый AuthService
func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	now := time.Now()
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Создаем профиль
	profile := &models.UserProfile{
		UserID:      user.ID,
		DisplayName: req.DisplayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.userRepo.Create(ctx, user, profile); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Получаем созданного пользователя с профилем
	userWithProfile, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	// Генерируем токены
	return s.generateTokens(ctx, userWithProfile)
}

// Login аутентифицирует пользователя
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.IsBanned {
		return nil, ErrUserBanned
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Обновляем время последнего входа
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Генерируем токены
	return s.generateTokens(ctx, user)
}

// RefreshToken обновляет access token по refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	// Хешируем токен для поиска в базе
	tokenHash := hashToken(refreshToken)

	// Проверяем токен в базе
	storedToken, err := s.userRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil || storedToken == nil {
		return nil, ErrInvalidToken
	}

	if storedToken.RevokedAt != nil {
		return nil, ErrInvalidToken
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// Парсим refresh token для проверки структуры
	claims, err := s.parseToken(refreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, err
	}

	// Проверяем, что это refresh token
	if claims.TokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	// Проверяем соответствие UserID
	if claims.UserID != storedToken.UserID {
		return nil, ErrInvalidToken
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if user.IsBanned {
		return nil, ErrUserBanned
	}

	// Отзываем старый refresh token
	_ = s.userRepo.RevokeRefreshToken(ctx, tokenHash)

	// Генерируем новые токены
	return s.generateTokens(ctx, user)
}

// Logout отзывает токены пользователя
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.userRepo.RevokeRefreshToken(ctx, tokenHash)
}

// LogoutAll отзывает все токены пользователя
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.RevokeAllUserTokens(ctx, userID)
}

// ValidateToken проверяет access token и возвращает claims
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	claims, err := s.parseToken(tokenString, s.cfg.JWT.Secret)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ChangePassword изменяет пароль пользователя
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	// Проверяем старый пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Отзываем все токены
	_ = s.userRepo.RevokeAllUserTokens(ctx, userID)

	return nil
}

// generateTokens генерирует пару access и refresh токенов
func (s *AuthService) generateTokens(ctx context.Context, user *models.UserWithProfile) (*models.AuthResponse, error) {
	now := time.Now()

	// Access token
	accessClaims := &models.JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Roles:     user.Roles,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWT.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := &models.JWTClaims{
		UserID:    user.ID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWT.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.JWT.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Хешируем токен для сохранения в базе
	tokenHash := hashToken(refreshTokenString)

	// Сохраняем refresh token в базе
	tokenRecord := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(s.cfg.JWT.RefreshTokenTTL),
		CreatedAt: now,
	}

	if err := s.userRepo.SaveRefreshToken(ctx, tokenRecord); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.cfg.JWT.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// hashToken хеширует токен для безопасного хранения
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// parseToken парсит JWT токен
func (s *AuthService) parseToken(tokenString, secret string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
