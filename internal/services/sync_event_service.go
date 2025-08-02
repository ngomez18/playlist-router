package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

//go:generate mockgen -source=sync_event_service.go -destination=mocks/mock_sync_event_service.go -package=mocks

type SyncEventServicer interface {
	CreateSyncEvent(ctx context.Context, syncEvent *models.SyncEvent) (*models.SyncEvent, error)
	UpdateSyncEvent(ctx context.Context, id string, syncEvent *models.SyncEvent) (*models.SyncEvent, error)
	GetSyncEvent(ctx context.Context, id string) (*models.SyncEvent, error)
	HasActiveSyncForBasePlaylist(ctx context.Context, userID, basePlaylistID string) (bool, error)
	HasActiveSyncForUser(ctx context.Context, userID string) (bool, error)
}

type SyncEventService struct {
	syncEventRepo repositories.SyncEventRepository
	logger        *slog.Logger
}

func NewSyncEventService(
	syncEventRepo repositories.SyncEventRepository,
	logger *slog.Logger,
) *SyncEventService {
	return &SyncEventService{
		syncEventRepo: syncEventRepo,
		logger:        logger.With("component", "SyncEventService"),
	}
}

func (seService *SyncEventService) CreateSyncEvent(ctx context.Context, syncEvent *models.SyncEvent) (*models.SyncEvent, error) {
	seService.logger.InfoContext(ctx, "creating sync event", "user_id", syncEvent.UserID, "base_playlist_id", syncEvent.BasePlaylistID)

	createdSyncEvent, err := seService.syncEventRepo.Create(ctx, syncEvent)
	if err != nil {
		seService.logger.ErrorContext(ctx, "failed to create sync event", "user_id", syncEvent.UserID, "base_playlist_id", syncEvent.BasePlaylistID, "error", err.Error())
		return nil, fmt.Errorf("failed to create sync event: %w", err)
	}

	seService.logger.InfoContext(ctx, "sync event created successfully", "sync_event_id", createdSyncEvent.ID, "user_id", syncEvent.UserID)
	return createdSyncEvent, nil
}

func (seService *SyncEventService) UpdateSyncEvent(ctx context.Context, id string, syncEvent *models.SyncEvent) (*models.SyncEvent, error) {
	seService.logger.InfoContext(ctx, "updating sync event", "sync_event_id", id, "status", syncEvent.Status)

	updatedSyncEvent, err := seService.syncEventRepo.Update(ctx, id, syncEvent)
	if err != nil {
		seService.logger.ErrorContext(ctx, "failed to update sync event", "sync_event_id", id, "error", err.Error())
		return nil, fmt.Errorf("failed to update sync event: %w", err)
	}

	seService.logger.InfoContext(ctx, "sync event updated successfully", "sync_event_id", id, "status", updatedSyncEvent.Status)
	return updatedSyncEvent, nil
}

func (seService *SyncEventService) GetSyncEvent(ctx context.Context, id string) (*models.SyncEvent, error) {
	seService.logger.InfoContext(ctx, "retrieving sync event", "sync_event_id", id)

	syncEvent, err := seService.syncEventRepo.GetByID(ctx, id)
	if err != nil {
		seService.logger.ErrorContext(ctx, "failed to retrieve sync event", "sync_event_id", id, "error", err.Error())
		return nil, fmt.Errorf("failed to retrieve sync event: %w", err)
	}

	seService.logger.InfoContext(ctx, "sync event retrieved successfully", "sync_event_id", id, "status", syncEvent.Status)
	return syncEvent, nil
}

func (seService *SyncEventService) HasActiveSyncForBasePlaylist(ctx context.Context, userID, basePlaylistID string) (bool, error) {
	seService.logger.InfoContext(ctx, "checking for active sync", "user_id", userID, "base_playlist_id", basePlaylistID)

	syncEvents, err := seService.syncEventRepo.GetByBasePlaylistID(ctx, basePlaylistID)
	if err != nil {
		seService.logger.ErrorContext(ctx, "failed to get sync events for active sync check", "user_id", userID, "base_playlist_id", basePlaylistID, "error", err.Error())
		return false, fmt.Errorf("failed to check for active sync: %w", err)
	}

	// Check if any sync events for this base playlist and user are in progress
	for _, syncEvent := range syncEvents {
		if syncEvent.UserID == userID && syncEvent.Status == models.SyncStatusInProgress {
			seService.logger.InfoContext(ctx, "active sync found", "user_id", userID, "base_playlist_id", basePlaylistID, "sync_event_id", syncEvent.ID)
			return true, nil
		}
	}

	seService.logger.InfoContext(ctx, "no active sync found", "user_id", userID, "base_playlist_id", basePlaylistID)
	return false, nil
}

func (seService *SyncEventService) HasActiveSyncForUser(ctx context.Context, userID string) (bool, error) {
	seService.logger.InfoContext(ctx, "checking for active sync", "user_id", userID)

	syncEvents, err := seService.syncEventRepo.GetByUserID(ctx, userID)
	if err != nil {
		seService.logger.ErrorContext(ctx, "failed to get sync events for active sync check", "user_id", userID, "error", err.Error())
		return false, fmt.Errorf("failed to check for active sync: %w", err)
	}

for _, syncEvent := range syncEvents {
		if syncEvent.Status == models.SyncStatusInProgress {
			seService.logger.InfoContext(ctx, "active sync found", "user_id", userID, "sync_event_id", syncEvent.ID)
			return true, nil
		}
	}

	seService.logger.InfoContext(ctx, "no active sync found", "user_id", userID)
	return false, nil
}