package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"
)

// AuthHandler обработчик эндпоинтов аутентификации
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler создает новый AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register обрабатывает регистрацию пользователя
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Валидация
	if req.Email == "" {
		response.BadRequest(w, "email is required")
		return
	}
	if req.Password == "" {
		response.BadRequest(w, "password is required")
		return
	}
	if len(req.Password) < 8 {
		response.BadRequest(w, "password must be at least 8 characters")
		return
	}
	if req.DisplayName == "" {
		// Используем часть email как имя по умолчанию
		parts := strings.Split(req.Email, "@")
		req.DisplayName = parts[0]
	}

	authResp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			response.Conflict(w, "email already exists")
			return
		}
		response.InternalError(w, "failed to register user")
		return
	}

	// Устанавливаем refresh token в httpOnly cookie
	h.setRefreshTokenCookie(w, authResp.RefreshToken)

	response.Created(w, authResp)
}

// Login обрабатывает вход пользователя
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "email and password are required")
		return
	}

	authResp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(w, "invalid email or password")
			return
		}
		if errors.Is(err, service.ErrUserBanned) {
			response.Forbidden(w, "user is banned")
			return
		}
		response.InternalError(w, "failed to login")
		return
	}

	h.setRefreshTokenCookie(w, authResp.RefreshToken)

	response.OK(w, authResp)
}

// Logout обрабатывает выход пользователя
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Получаем refresh token из cookie
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		_ = h.authService.Logout(r.Context(), cookie.Value)
	}

	// Удаляем cookie
	h.clearRefreshTokenCookie(w)

	response.OK(w, map[string]string{"message": "logged out successfully"})
}

// Refresh обновляет access token
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Получаем refresh token из cookie или body
	var refreshToken string

	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	} else {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		response.Unauthorized(w, "refresh token is required")
		return
	}

	authResp, err := h.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrTokenExpired) {
			h.clearRefreshTokenCookie(w)
			response.Unauthorized(w, "invalid or expired refresh token")
			return
		}
		if errors.Is(err, service.ErrUserBanned) {
			h.clearRefreshTokenCookie(w)
			response.Forbidden(w, "user is banned")
			return
		}
		response.InternalError(w, "failed to refresh token")
		return
	}

	h.setRefreshTokenCookie(w, authResp.RefreshToken)

	response.OK(w, authResp)
}

// Me возвращает информацию о текущем пользователе
// GET /api/v1/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(r.Context())
	if user == nil {
		response.Unauthorized(w, "not authenticated")
		return
	}

	response.OK(w, user)
}

// ChangePassword обрабатывает смену пароля
// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(r.Context())
	if user == nil {
		response.Unauthorized(w, "not authenticated")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		response.BadRequest(w, "old_password and new_password are required")
		return
	}
	if len(req.NewPassword) < 8 {
		response.BadRequest(w, "new password must be at least 8 characters")
		return
	}

	err := h.authService.ChangePassword(r.Context(), user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.BadRequest(w, "old password is incorrect")
			return
		}
		response.InternalError(w, "failed to change password")
		return
	}

	// Удаляем cookie, чтобы пользователь перелогинился
	h.clearRefreshTokenCookie(w)

	response.OK(w, map[string]string{"message": "password changed successfully"})
}

// setRefreshTokenCookie устанавливает refresh token в httpOnly cookie
func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(30 * 24 * time.Hour / time.Second), // 30 дней
	})
}

// clearRefreshTokenCookie удаляет cookie refresh token
func (h *AuthHandler) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
