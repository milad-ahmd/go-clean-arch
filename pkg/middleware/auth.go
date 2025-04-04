package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/milad-ahmd/go-clean-arch/internal/domain"
	"github.com/milad-ahmd/go-clean-arch/pkg/logger"
	"go.uber.org/zap"
)

// contextKey is a custom type for context keys
type contextKey string

// UserKey is the key for the user in the context
const UserKey contextKey = "user"

// Auth middleware for authentication
func Auth(userUseCase domain.UserUseCase, logger logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Check if the header has the Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Extract the token
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token and get user
			user, err := userUseCase.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Error("Failed to validate token", zap.Error(err))
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware for role-based authorization
func RequireRole(roles ...domain.Role) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := r.Context().Value(UserKey).(*domain.User)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext gets the user from the context
func GetUserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(UserKey).(*domain.User)
	return user, ok
}
