package repositories

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=sync_event_repository.go -destination=mocks/mock_sync_event_repository.go -package=mocks

type SyncEventRepository interface {
	Create(ctx context.Context, syncEvent *models.SyncEvent) (*models.SyncEvent, error)
	Update(ctx context.Context, id string, syncEvent *models.SyncEvent) (*models.SyncEvent, error)
	GetByID(ctx context.Context, id string) (*models.SyncEvent, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.SyncEvent, error)
	GetByBasePlaylistID(ctx context.Context, basePlaylistID string) ([]*models.SyncEvent, error)
}
