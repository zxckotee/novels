package middleware

import (
	"context"
	"net/http"
	"strings"

	"novels-backend/internal/config"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/service"
	"novels-backend/pkg/response"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserRoleKey  contextKey = "user_role"
	UserRolesKey contextKey = "user_roles"
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
		
		// Добавляем все роли в контекст
		ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
		
		// Для обратной совместимости также добавляем первую роль
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
			// Получаем все роли пользователя
			userRolesVal := r.Context().Value(UserRolesKey)
			
			// Проверяем тип ролей
			var userRoleStrings []string
			switch v := userRolesVal.(type) {
			case []models.UserRole:
				for _, role := range v {
					userRoleStrings = append(userRoleStrings, string(role))
				}
			case []string:
				userRoleStrings = v
			default:
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Access denied - no roles found")
				return
			}

			// Проверяем, есть ли хотя бы одна требуемая роль у пользователя
			hasRole := false
			for _, userRole := range userRoleStrings {
				// Admin всегда имеет доступ
				if userRole == "admin" {
					hasRole = true
					break
				}
				
				// Проверяем каждую требуемую роль
				for _, requiredRole := range roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				
				if hasRole {
					break
				}
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
					ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
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
