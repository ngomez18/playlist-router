package repositories

import "github.com/ngomez18/playlist-router/internal/models"

type BasePlaylistRepository interface {
	Create(userId, name, spotifyPlaylistId string) (*models.BasePlaylist, error)
	GetByID(id string) (*models.BasePlaylist, error)
	Delete(id string) error
}
