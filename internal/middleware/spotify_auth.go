package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
)

const (
	TokenRefreshBuffer = 15 * time.Minute
)

type SpotifyAuthMiddleware struct {
	spotifyIntegrationService services.SpotifyIntegrationServicer
	spotifyClient             spotifyclient.SpotifyAPI
	logger                    *slog.Logger
}

func NewSpotifyAuthMiddleware(
	spotifyIntegrationService services.SpotifyIntegrationServicer,
	spotifyClient spotifyclient.SpotifyAPI,
	logger *slog.Logger,
) *SpotifyAuthMiddleware {
	return &SpotifyAuthMiddleware{
		spotifyIntegrationService: spotifyIntegrationService,
		spotifyClient:             spotifyClient,
		logger:                    logger.With("component", "SpotifyAuthMiddleware"),
	}
}

func (m *SpotifyAuthMiddleware) RequireSpotifyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := requestcontext.GetUserFromContext(ctx)
		if !ok {
			m.logger.WarnContext(ctx, "user not available in context for spotify auth")
			http.Error(w, "user not available in context", http.StatusUnauthorized)
			return
		}

		spotifyIntegration, err := m.spotifyIntegrationService.GetIntegrationByUserID(ctx, user.ID)
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to get spotify integration", "user_id", user.ID, "error", err)
			http.Error(w, "no spotify integration available for user", http.StatusUnauthorized)
			return
		}

		// Check if token needs refreshing (expires within buffer time)
		if spotifyIntegration.ExpiresAt.Before(time.Now().Add(TokenRefreshBuffer)) {
			m.logger.InfoContext(ctx, "refreshing spotify tokens",
				"user_id", user.ID,
				"expires_at", spotifyIntegration.ExpiresAt,
			)

			refreshedIntegration, err := m.refreshTokens(ctx, spotifyIntegration)
			if err != nil {
				m.logger.ErrorContext(ctx, "failed to refresh spotify tokens",
					"user_id", user.ID,
					"integration_id", spotifyIntegration.ID,
					"error", err,
				)
				http.Error(w, "failed to refresh spotify tokens", http.StatusUnauthorized)
				return
			}

			spotifyIntegration = refreshedIntegration
			m.logger.InfoContext(ctx, "successfully refreshed spotify tokens",
				"user_id", user.ID,
				"new_expires_at", spotifyIntegration.ExpiresAt,
			)
		}

		// Add spotify integration to request context
		ctxWithAuth := requestcontext.ContextWithSpotifyAuth(ctx, spotifyIntegration)
		next.ServeHTTP(w, r.WithContext(ctxWithAuth))
	})
}

// refreshTokens handles the token refresh process and database update
func (m *SpotifyAuthMiddleware) refreshTokens(ctx context.Context, integration *models.SpotifyIntegration) (*models.SpotifyIntegration, error) {
	tokenResponse, err := m.spotifyClient.RefreshTokens(ctx, integration.RefreshToken)
	if err != nil {
		return nil, err
	}

	tokenUpdate := &models.SpotifyIntegrationTokenRefresh{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresIn:    tokenResponse.ExpiresIn,
	}

	// If Spotify didn't return a new refresh token, keep the current one
	if tokenUpdate.RefreshToken == "" {
		tokenUpdate.RefreshToken = integration.RefreshToken
	}

	err = m.spotifyIntegrationService.UpdateTokens(ctx, integration.ID, tokenUpdate)
	if err != nil {
		return nil, err
	}

	updatedIntegration := *integration
	updatedIntegration.AccessToken = tokenUpdate.AccessToken
	updatedIntegration.RefreshToken = tokenUpdate.RefreshToken
	updatedIntegration.ExpiresAt = time.Now().Add(time.Duration(tokenUpdate.ExpiresIn) * time.Second)

	return &updatedIntegration, nil
}
