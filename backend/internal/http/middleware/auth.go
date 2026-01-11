package middleware

import (
	"context"
	"net/http"
	"strings"

	"novels-backend/internal/config"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

// AuthMiddleware предоставляет middleware для аутентификации
type AuthMiddleware struct {
	authService *service.AuthService
	jwtConfig   config.JWTConfig
}

// NewAuthMiddleware создает новый AuthMiddleware
func NewAuthMiddleware(authService *service.AuthService, jwtConfig config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		jwtConfig:   jwtConfig,
	}
}

// Authenticate проверяет JWT токен и добавляет информацию о пользователе в контекст
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid authorization header format")
			return
		}

		token := parts[1]

		// Валидируем токен
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			return
		}

		// Добавляем информацию о пользователе в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID.String())
		// Берем первую роль из списка (обычно одна роль)
		role := "user"
		if len(claims.Roles) > 0 {
			role = string(claims.Roles[0])
		}
		ctx = context.WithValue(ctx, UserRoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole проверяет, имеет ли пользователь требуемую роль
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
				return
			}

			// Проверяем, есть ли роль пользователя в списке разрешенных
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			// Admin имеет доступ ко всему
			if userRole == "admin" {
				hasRole = true
			}

			if !hasRole {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to access this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID извлекает ID пользователя из контекста
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserRole извлекает роль пользователя из контекста
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

// OptionalAuth проверяет JWT токен если он присутствует, но не требует его (опциональная аутентификация)
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Проверяем формат "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token := parts[1]
				
				// Валидируем токен (если есть)
				if claims, err := m.authService.ValidateToken(token); err == nil {
					// Добавляем информацию о пользователе в контекст, если токен валиден
					ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID.String())
					role := "user"
					if len(claims.Roles) > 0 {
						role = string(claims.Roles[0])
					}
					ctx = context.WithValue(ctx, UserRoleKey, role)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}
		
		// Если токена нет или он невалиден, продолжаем без пользователя в контексте
		next.ServeHTTP(w, r)
	})
}
