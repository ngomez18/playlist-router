package repositories

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=child_playlist_repository.go -destination=mocks/mock_child_playlist_repository.go -package=mocks

type ChildPlaylistRepository interface {
	Create(ctx context.Context, fields CreateChildPlaylistFields) (*models.ChildPlaylist, error)
	Delete(ctx context.Context, id, userID string) error
	GetByID(ctx context.Context, id, userID string) (*models.ChildPlaylist, error)
	GetByBasePlaylistID(ctx context.Context, basePlaylistID, userID string) ([]*models.ChildPlaylist, error)
	Update(ctx context.Context, id, userID string, fields UpdateChildPlaylistFields) (*models.ChildPlaylist, error)
}

type CreateChildPlaylistFields struct {
	UserID            string                      `json:"user_id" validate:"required"`
	BasePlaylistID    string                      `json:"base_playlist_id" validate:"required"`
	Name              string                      `json:"name" validate:"required,min=1,max=100"`
	Description       string                      `json:"description,omitempty"`
	SpotifyPlaylistID string                      `json:"spotify_playlist_id" validate:"required"`
	FilterRules       *models.AudioFeatureFilters `json:"filter_rules,omitempty"`
	IsActive          bool                        `json:"is_active"`
}

type UpdateChildPlaylistFields struct {
	Name              *string                     `json:"name,omitempty"`
	Description       *string                     `json:"description,omitempty"`
	FilterRules       *models.AudioFeatureFilters `json:"filter_rules,omitempty"`
	IsActive          *bool                       `json:"is_active,omitempty"`
	SpotifyPlaylistID *string                     `json:"spotify_playlist_id,omitempty"`
}
