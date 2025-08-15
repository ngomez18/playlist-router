package services

import (
	"context"
	"fmt"
	"log/slog"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

//go:generate mockgen -source=child_playlist_service.go -destination=mocks/mock_child_playlist_service.go -package=mocks

type ChildPlaylistServicer interface {
	CreateChildPlaylist(ctx context.Context, userID, basePlaylistID string, input *models.CreateChildPlaylistRequest) (*models.ChildPlaylist, error)
	DeleteChildPlaylist(ctx context.Context, id, userID string) error
	GetChildPlaylist(ctx context.Context, id, userID string) (*models.ChildPlaylist, error)
	GetChildPlaylistsByBasePlaylistID(ctx context.Context, basePlaylistID, userID string) ([]*models.ChildPlaylist, error)
	UpdateChildPlaylist(ctx context.Context, id, userID string, input *models.UpdateChildPlaylistRequest) (*models.ChildPlaylist, error)
	UpdateChildPlaylistSpotifyID(ctx context.Context, id, userID, spotifyID string) (*models.ChildPlaylist, error)
}

type ChildPlaylistService struct {
	childPlaylistRepo      repositories.ChildPlaylistRepository
	basePlaylistRepo       repositories.BasePlaylistRepository
	spotifyIntegrationRepo repositories.SpotifyIntegrationRepository
	spotifyClient          spotifyclient.SpotifyAPI
	logger                 *slog.Logger
}

func NewChildPlaylistService(
	childPlaylistRepo repositories.ChildPlaylistRepository,
	basePlaylistRepo repositories.BasePlaylistRepository,
	spotifyIntegrationRepo repositories.SpotifyIntegrationRepository,
	spotifyClient spotifyclient.SpotifyAPI,
	logger *slog.Logger,
) *ChildPlaylistService {
	return &ChildPlaylistService{
		childPlaylistRepo:      childPlaylistRepo,
		basePlaylistRepo:       basePlaylistRepo,
		spotifyIntegrationRepo: spotifyIntegrationRepo,
		spotifyClient:          spotifyClient,
		logger:                 logger.With("component", "ChildPlaylistService"),
	}
}

func (cpService *ChildPlaylistService) CreateChildPlaylist(ctx context.Context, userID, basePlaylistID string, input *models.CreateChildPlaylistRequest) (*models.ChildPlaylist, error) {
	cpService.logger.InfoContext(ctx, "creating child playlist", "user_id", userID, "base_playlist_id", basePlaylistID, "input", input)

	basePlaylist, err := cpService.basePlaylistRepo.GetByID(ctx, basePlaylistID, userID)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to get base playlist", "base_playlist_id", basePlaylistID, "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to get base playlist: %w", err)
	}

	// Create playlist in Spotify with naming format: [Base Name] > Child Name
	spotifyPlaylistName := models.BuildChildPlaylistName(basePlaylist.Name, input.Name)
	cpService.logger.InfoContext(ctx, "creating spotify playlist", "spotify_name", spotifyPlaylistName)

	spotifyPlaylist, err := cpService.spotifyClient.CreatePlaylist(
		ctx,
		spotifyPlaylistName,
		models.BuildChildPlaylistDescription(input.Description),
		false, // private by default
	)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to create playlist in spotify", "error", err.Error())
		return nil, fmt.Errorf("failed to create spotify playlist: %w", err)
	}

	cpService.logger.InfoContext(ctx, "successfully created spotify playlist", "spotify_playlist_id", spotifyPlaylist.ID, "name", spotifyPlaylist.Name)

	// Create the child playlist record in our database
	fields := repositories.CreateChildPlaylistFields{
		UserID:            userID,
		BasePlaylistID:    basePlaylistID,
		Name:              input.Name,
		Description:       input.Description,
		SpotifyPlaylistID: spotifyPlaylist.ID,
		FilterRules:       input.FilterRules,
		IsActive:          true,
	}
	childPlaylist, err := cpService.childPlaylistRepo.Create(ctx, fields)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to create child playlist", "error", err.Error())
		return nil, fmt.Errorf("failed to create child playlist: %w", err)
	}

	cpService.logger.InfoContext(ctx, "child playlist created successfully", "child_playlist", childPlaylist)
	return childPlaylist, nil
}

func (cpService *ChildPlaylistService) DeleteChildPlaylist(ctx context.Context, id, userID string) error {
	cpService.logger.InfoContext(ctx, "deleting child playlist", "id", id, "user_id", userID)

	// Get the child playlist to retrieve the Spotify playlist ID
	childPlaylist, err := cpService.childPlaylistRepo.GetByID(ctx, id, userID)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to get child playlist for deletion", "id", id, "user_id", userID, "error", err.Error())
		return fmt.Errorf("failed to get child playlist: %w", err)
	}

	// Delete from Spotify first
	err = cpService.spotifyClient.DeletePlaylist(ctx, childPlaylist.SpotifyPlaylistID)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to delete playlist from spotify", "spotify_playlist_id", childPlaylist.SpotifyPlaylistID, "error", err.Error())
		return fmt.Errorf("failed to delete spotify playlist: %w", err)
	}

	cpService.logger.InfoContext(ctx, "successfully deleted spotify playlist", "spotify_playlist_id", childPlaylist.SpotifyPlaylistID)

	// Delete from database
	err = cpService.childPlaylistRepo.Delete(ctx, id, userID)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to delete child playlist from database", "id", id, "error", err.Error())
		return fmt.Errorf("failed to delete child playlist: %w", err)
	}

	cpService.logger.InfoContext(ctx, "child playlist deleted successfully", "id", id, "user_id", userID)
	return nil
}

func (cpService *ChildPlaylistService) GetChildPlaylist(ctx context.Context, id, userID string) (*models.ChildPlaylist, error) {
	cpService.logger.InfoContext(ctx, "retrieving child playlist", "id", id, "user_id", userID)

	childPlaylist, err := cpService.childPlaylistRepo.GetByID(ctx, id, userID)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to retrieve child playlist", "id", id, "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to retrieve child playlist: %w", err)
	}

	cpService.logger.InfoContext(ctx, "child playlist retrieved successfully", "child_playlist", childPlaylist)
	return childPlaylist, nil
}

func (cpService *ChildPlaylistService) GetChildPlaylistsByBasePlaylistID(ctx context.Context, basePlaylistID, userID string) ([]*models.ChildPlaylist, error) {
	cpService.logger.InfoContext(ctx, "retrieving child playlists for base playlist", "base_playlist_id", basePlaylistID, "user_id", userID)

	childPlaylists, err := cpService.childPlaylistRepo.GetByBasePlaylistID(ctx, basePlaylistID, userID)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to retrieve child playlists for base playlist", "base_playlist_id", basePlaylistID, "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to retrieve child playlists: %w", err)
	}

	cpService.logger.InfoContext(ctx, "child playlists retrieved successfully", "base_playlist_id", basePlaylistID, "user_id", userID, "count", len(childPlaylists))
	return childPlaylists, nil
}

func (cpService *ChildPlaylistService) UpdateChildPlaylist(ctx context.Context, id, userID string, input *models.UpdateChildPlaylistRequest) (*models.ChildPlaylist, error) {
	cpService.logger.InfoContext(ctx, "updating child playlist", "id", id, "user_id", userID, "input", input)

	// Update the child playlist in our database first
	updateFields := repositories.UpdateChildPlaylistFields{
		Name:        input.Name,
		Description: input.Description,
		IsActive:    input.IsActive,
		FilterRules: input.FilterRules,
	}
	updatedChildPlaylist, err := cpService.childPlaylistRepo.Update(ctx, id, userID, updateFields)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to update child playlist", "id", id, "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to update child playlist: %w", err)
	}

	spotifyUpdate := struct {
		name         string
		description  string
		shouldUpdate bool
	}{}
	if input.Name != nil {
		spotifyUpdate.shouldUpdate = true
		basePlaylist, err := cpService.basePlaylistRepo.GetByID(ctx, updatedChildPlaylist.BasePlaylistID, userID)
		if err != nil {
			cpService.logger.ErrorContext(ctx, "failed to get base playlist for name update", "base_playlist_id", updatedChildPlaylist.BasePlaylistID, "error", err.Error())
			return nil, fmt.Errorf("failed to get base playlist: %w", err)
		}

		spotifyUpdate.name = models.BuildChildPlaylistName(basePlaylist.Name, *input.Name)
	}

	if input.Description != nil {
		spotifyUpdate.shouldUpdate = true
		spotifyUpdate.description = models.BuildChildPlaylistDescription(*input.Description)
	}

	if spotifyUpdate.shouldUpdate {
		// Update Spotify playlist metadata
		err = cpService.spotifyClient.UpdatePlaylist(
			ctx,
			updatedChildPlaylist.SpotifyPlaylistID,
			spotifyUpdate.name,
			spotifyUpdate.description,
		)
		if err != nil {
			cpService.logger.ErrorContext(ctx, "failed to update spotify playlist", "spotify_playlist_id", updatedChildPlaylist.SpotifyPlaylistID, "error", err.Error())
			return nil, fmt.Errorf("failed to update spotify playlist: %w", err)
		}

		cpService.logger.InfoContext(ctx, "successfully updated spotify playlist",
			"spotify_playlist_id", updatedChildPlaylist.SpotifyPlaylistID,
			"name", spotifyUpdate.name,
			"description", spotifyUpdate.description,
		)
	}

	cpService.logger.InfoContext(ctx, "child playlist updated successfully", "child_playlist", updatedChildPlaylist)
	return updatedChildPlaylist, nil
}

func (cpService *ChildPlaylistService) UpdateChildPlaylistSpotifyID(ctx context.Context, id, userID, spotifyID string) (*models.ChildPlaylist, error) {
	cpService.logger.InfoContext(ctx, "updating child playlist spotify id", "id", id, "user_id", userID, "spotify_id", spotifyID)

	updateFields := repositories.UpdateChildPlaylistFields{SpotifyPlaylistID: &spotifyID}

	updatedChildPlaylist, err := cpService.childPlaylistRepo.Update(ctx, id, userID, updateFields)
	if err != nil {
		cpService.logger.ErrorContext(ctx, "failed to update child playlist", "id", id, "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to update child playlist: %w", err)
	}

	cpService.logger.InfoContext(ctx, "child playlist updated successfully", "child_playlist", updatedChildPlaylist)
	return updatedChildPlaylist, nil
}
