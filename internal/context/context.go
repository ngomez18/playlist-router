package requestcontext

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

type contextKey string

const (
	UserContextKey        contextKey = "user"
	SpotifyAuthContextKey contextKey = "spotify_integration"
)

func ContextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

func ContextWithSpotifyAuth(ctx context.Context, spotifyAuth *models.SpotifyIntegration) context.Context {
	return context.WithValue(ctx, SpotifyAuthContextKey, spotifyAuth)
}

func GetSpotifyAuthFromContext(ctx context.Context) (*models.SpotifyIntegration, bool) {
	s, ok := ctx.Value(SpotifyAuthContextKey).(*models.SpotifyIntegration)
	return s, ok
}

func GetUserAndSpotifyAuthFromContext(ctx context.Context) (*models.User, *models.SpotifyIntegration, bool) {
	user, userOk := ctx.Value(UserContextKey).(*models.User)
	spotifyIntegration, spotifyIntegrationOk := ctx.Value(SpotifyAuthContextKey).(*models.SpotifyIntegration)

	return user, spotifyIntegration, userOk && spotifyIntegrationOk
}
