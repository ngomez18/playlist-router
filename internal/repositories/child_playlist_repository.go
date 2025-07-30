package repositories

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=child_playlist_repository.go -destination=mocks/mock_child_playlist_repository.go -package=mocks

type ChildPlaylistRepository interface {
	Create(ctx context.Context, userID, basePlaylistID, name, description, spotifyPlaylistID string, filterRules *models.AudioFeatureFilters) (*models.ChildPlaylist, error)
	Delete(ctx context.Context, id, userID string) error
	GetByID(ctx context.Context, id, userID string) (*models.ChildPlaylist, error)
	GetByBasePlaylistID(ctx context.Context, basePlaylistID, userID string) ([]*models.ChildPlaylist, error)
	Update(ctx context.Context, id, userID string, req *models.UpdateChildPlaylistRequest) (*models.ChildPlaylist, error)
}
