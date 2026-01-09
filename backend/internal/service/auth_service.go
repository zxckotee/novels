package service

import (
	"context"
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
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Roles:        []string{models.RoleUser},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Создаем профиль
	profile := &models.UserProfile{
		UserID:      user.ID,
		DisplayName: req.DisplayName,
	}

	if err := s.userRepo.CreateProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Генерируем токены
	return s.generateTokens(ctx, user)
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
	// Парсим refresh token
	claims, err := s.parseToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, err
	}

	// Проверяем, что это refresh token
	if claims.TokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	// Проверяем токен в базе
	storedToken, err := s.userRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil || storedToken == nil {
		return nil, ErrInvalidToken
	}

	if storedToken.RevokedAt != nil {
		return nil, ErrInvalidToken
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
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
	_ = s.userRepo.RevokeRefreshToken(ctx, refreshToken)

	// Генерируем новые токены
	return s.generateTokens(ctx, user)
}

// Logout отзывает токены пользователя
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.userRepo.RevokeRefreshToken(ctx, refreshToken)
}

// LogoutAll отзывает все токены пользователя
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.RevokeAllRefreshTokens(ctx, userID)
}

// ValidateToken проверяет access token и возвращает claims
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	claims, err := s.parseToken(tokenString, s.cfg.JWTSecret)
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
	_ = s.userRepo.RevokeAllRefreshTokens(ctx, userID)

	return nil
}

// generateTokens генерирует пару access и refresh токенов
func (s *AuthService) generateTokens(ctx context.Context, user *models.User) (*models.AuthResponse, error) {
	now := time.Now()

	// Access token
	accessClaims := &models.JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Roles:     user.Roles,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := &models.JWTClaims{
		UserID:    user.ID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTRefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.JWTRefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Сохраняем refresh token в базе
	tokenRecord := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: now.Add(s.cfg.JWTRefreshExpiration),
	}

	if err := s.userRepo.SaveRefreshToken(ctx, tokenRecord); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.cfg.JWTExpiration.Seconds()),
		TokenType:    "Bearer",
		User: &models.UserPublic{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Roles:       user.Roles,
		},
	}, nil
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
