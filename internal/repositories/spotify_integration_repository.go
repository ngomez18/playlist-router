package repositories

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=spotify_integration_repository.go -destination=mocks/mock_spotify_integration_repository.go -package=mocks

type SpotifyIntegrationRepository interface {
	CreateOrUpdate(ctx context.Context, userID string, integration *models.SpotifyIntegration) (*models.SpotifyIntegration, error)
	GetByUserID(ctx context.Context, userID string) (*models.SpotifyIntegration, error)
	GetBySpotifyID(ctx context.Context, spotifyID string) (*models.SpotifyIntegration, error)
	UpdateTokens(ctx context.Context, integrationID string, tokens *models.SpotifyTokenResponse) error
	Delete(ctx context.Context, userID string) error
}