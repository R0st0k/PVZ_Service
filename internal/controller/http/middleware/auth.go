package middleware

import (
	"context"
	"net/http"
	api "pvz-service/api/generated"
	"pvz-service/internal/service"
	"strings"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
)

// AuthMiddleware создает middleware для проверки JWT токена
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(AuthorizationHeader)
			if authHeader == "" {
				http.Error(w, "authorization header is required", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, BearerPrefix) {
				http.Error(w, "authorization header must start with 'Bearer '", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
			email, role, err := authService.GetUserFromToken(tokenString)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_email", email)
			ctx = context.WithValue(ctx, "user_role", api.UserRole(role))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleMiddleware создает middleware для проверки роли пользователя
func RoleMiddlewareMulti(allowedRoles ...api.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("user_role").(api.UserRole)
			if !ok {
				http.Error(w, "insufficient permissionsss", http.StatusForbidden)
				return
			}

			// Проверяем, есть ли роль пользователя в списке разрешенных
			allowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
