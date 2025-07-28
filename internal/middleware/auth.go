package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
	"github.com/pocketbase/pocketbase/apis"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

type AuthMiddleware struct {
	userService services.UserServicer
}

func NewAuthMiddleware(userService services.UserServicer) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header is required", http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "token is required", http.StatusUnauthorized)
			return
		}

		// Validate token using user service
		user, err := m.userService.ValidateAuthToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// If no auth header, continue without authentication
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// If auth header exists, try to validate it
		if token, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			if token != "" {
				// Try to validate token using user service
				user, err := m.userService.ValidateAuthToken(r.Context(), token)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					
					return
				}
			}
		}

		// Invalid token format or validation failed, continue without auth
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) ExtractUserID(r *http.Request) (string, error) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		return "", apis.NewUnauthorizedError("user not found in context", nil)
	}
	return user.ID, nil
}

func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}
