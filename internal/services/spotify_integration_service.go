package services

import (
	"context"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

//go:generate mockgen -source=spotify_integration_service.go -destination=mocks/mock_spotify_integration_service.go -package=mocks

type SpotifyIntegrationServicer interface {
	CreateOrUpdateIntegration(ctx context.Context, userID string, integration *models.SpotifyIntegration) (*models.SpotifyIntegration, error)
	GetIntegrationByUserID(ctx context.Context, userID string) (*models.SpotifyIntegration, error)
	GetIntegrationBySpotifyID(ctx context.Context, spotifyID string) (*models.SpotifyIntegration, error)
	UpdateTokens(ctx context.Context, integrationID string, tokens *models.SpotifyTokenResponse) error
	DeleteIntegration(ctx context.Context, userID string) error
}

type SpotifyIntegrationService struct {
	integrationRepo repositories.SpotifyIntegrationRepository
	logger          *slog.Logger
}

func NewSpotifyIntegrationService(integrationRepo repositories.SpotifyIntegrationRepository, logger *slog.Logger) *SpotifyIntegrationService {
	return &SpotifyIntegrationService{
		integrationRepo: integrationRepo,
		logger:          logger.With("component", "SpotifyIntegrationService"),
	}
}

func (sis *SpotifyIntegrationService) CreateOrUpdateIntegration(ctx context.Context, userID string, integration *models.SpotifyIntegration) (*models.SpotifyIntegration, error) {
	sis.logger.InfoContext(ctx, "creating or updating spotify integration", "user_id", userID, "spotify_id", integration.SpotifyID)

	result, err := sis.integrationRepo.CreateOrUpdate(ctx, userID, integration)
	if err != nil {
		sis.logger.ErrorContext(ctx, "unable to upsert spotify integration", "integration", integration, "error", err.Error())
		return nil, err
	}

	sis.logger.InfoContext(ctx, "spotify integration upserted successfully", "integration", result)
	return result, nil
}

func (sis *SpotifyIntegrationService) GetIntegrationByUserID(ctx context.Context, userID string) (*models.SpotifyIntegration, error) {
	sis.logger.InfoContext(ctx, "retrieving spotify integration by user ID", "user_id", userID)

	integration, err := sis.integrationRepo.GetByUserID(ctx, userID)
	if err != nil {
		sis.logger.ErrorContext(ctx, "unable to fetch spotify integration by user ID", "user_id", userID, "error", err.Error())
		return nil, err
	}

	sis.logger.InfoContext(ctx, "spotify integration retrieved successfully by user ID", "integration_id", integration.ID, "user_id", userID, "spotify_id", integration.SpotifyID)
	return integration, nil
}

func (sis *SpotifyIntegrationService) GetIntegrationBySpotifyID(ctx context.Context, spotifyID string) (*models.SpotifyIntegration, error) {
	sis.logger.InfoContext(ctx, "retrieving spotify integration by spotify ID", "spotify_id", spotifyID)

	integration, err := sis.integrationRepo.GetBySpotifyID(ctx, spotifyID)
	if err != nil {
		sis.logger.ErrorContext(ctx, "unable to fetch spotify integration by spotify ID", "spotify_id", spotifyID, "error", err.Error())
		return nil, err
	}

	sis.logger.InfoContext(ctx, "spotify integration retrieved successfully by spotify ID", "integration_id", integration.ID, "spotify_id", spotifyID, "user_id", integration.UserID)
	return integration, nil
}

func (sis *SpotifyIntegrationService) UpdateTokens(ctx context.Context, integrationID string, tokens *models.SpotifyTokenResponse) error {
	sis.logger.InfoContext(ctx, "updating spotify integration tokens", "integration_id", integrationID)

	err := sis.integrationRepo.UpdateTokens(ctx, integrationID, tokens)
	if err != nil {
		sis.logger.ErrorContext(ctx, "unable to update spotify integration tokens", "integration_id", integrationID, "error", err.Error())
		return err
	}

	sis.logger.InfoContext(ctx, "spotify integration tokens updated successfully", "integration_id", integrationID)
	return nil
}

func (sis *SpotifyIntegrationService) DeleteIntegration(ctx context.Context, userID string) error {
	sis.logger.InfoContext(ctx, "deleting spotify integration", "user_id", userID)

	err := sis.integrationRepo.Delete(ctx, userID)
	if err != nil {
		sis.logger.ErrorContext(ctx, "failed to delete spotify integration", "user_id", userID, "error", err.Error())
		return err
	}

	sis.logger.InfoContext(ctx, "spotify integration deleted successfully", "user_id", userID)
	return nil
}