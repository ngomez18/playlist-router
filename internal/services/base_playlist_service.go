package services

import (
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

type BasePlaylistServicer interface {
	CreateBasePlaylist(input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error)
}

type BasePlaylistService struct {
	basePlaylistRepo *repositories.BasePlaylistRepository
}

func NewBasePlaylistService(basePlaylistRepo *repositories.BasePlaylistRepository) *BasePlaylistService {
	return &BasePlaylistService{
		basePlaylistRepo: basePlaylistRepo,
	}
}

func (bpService *BasePlaylistService) CreateBasePlaylist(input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error) {
	return nil, nil
}