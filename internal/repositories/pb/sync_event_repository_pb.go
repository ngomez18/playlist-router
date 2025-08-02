package pb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type SyncEventRepositoryPocketbase struct {
	collection Collection
	app        *pocketbase.PocketBase
	log        *slog.Logger
}

func NewSyncEventRepositoryPocketbase(pb *pocketbase.PocketBase) *SyncEventRepositoryPocketbase {
	return &SyncEventRepositoryPocketbase{
		collection: CollectionSyncEvent,
		app:        pb,
		log:        pb.Logger().With("component", "SyncEventRepositoryPocketbase"),
	}
}

func (seRepo *SyncEventRepositoryPocketbase) Create(ctx context.Context, syncEvent *models.SyncEvent) (*models.SyncEvent, error) {
	collection, err := seRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("user_id", syncEvent.UserID)
	record.Set("base_playlist_id", syncEvent.BasePlaylistID)
	record.Set("status", string(syncEvent.Status))
	record.Set("started_at", syncEvent.StartedAt)
	record.Set("tracks_processed", syncEvent.TracksProcessed)
	record.Set("total_api_requests", syncEvent.TotalAPIRequests)

	// Serialize child playlist IDs to JSON
	if len(syncEvent.ChildPlaylistIDs) > 0 {
		childPlaylistIDsJSON, err := json.Marshal(syncEvent.ChildPlaylistIDs)
		if err != nil {
			seRepo.log.ErrorContext(ctx, "unable to serialize child playlist IDs", "child_playlist_ids", syncEvent.ChildPlaylistIDs, "error", err)
			return nil, fmt.Errorf(`%w: failed to serialize child playlist IDs: %s`, repositories.ErrDatabaseOperation, err.Error())
		}
		record.Set("child_playlist_ids", string(childPlaylistIDsJSON))
	}

	// Set optional fields
	if syncEvent.CompletedAt != nil {
		record.Set("completed_at", *syncEvent.CompletedAt)
	}
	if syncEvent.ErrorMessage != nil {
		record.Set("error_message", *syncEvent.ErrorMessage)
	}

	err = seRepo.app.Save(record)
	if err != nil {
		seRepo.log.ErrorContext(ctx, "unable to store sync_event record", "record", record, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	seRepo.log.InfoContext(ctx, "sync_event stored successfully", "record", record)

	return recordToSyncEvent(record), nil
}

func (seRepo *SyncEventRepositoryPocketbase) Update(ctx context.Context, id string, syncEvent *models.SyncEvent) (*models.SyncEvent, error) {
	collection, err := seRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record, err := seRepo.app.FindRecordById(collection, id)
	if err != nil {
		seRepo.log.ErrorContext(ctx, "unable to find sync_event record", "id", id, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrSyncEventNotFound, err.Error())
	}

	// Update fields
	record.Set("status", string(syncEvent.Status))
	record.Set("tracks_processed", syncEvent.TracksProcessed)
	record.Set("total_api_requests", syncEvent.TotalAPIRequests)

	// Update child playlist IDs if provided (including empty slice to clear them)
	if syncEvent.ChildPlaylistIDs != nil {
		if len(syncEvent.ChildPlaylistIDs) > 0 {
			childPlaylistIDsJSON, err := json.Marshal(syncEvent.ChildPlaylistIDs)
			if err != nil {
				seRepo.log.ErrorContext(ctx, "unable to serialize child playlist IDs", "child_playlist_ids", syncEvent.ChildPlaylistIDs, "error", err)
				return nil, fmt.Errorf(`%w: failed to serialize child playlist IDs: %s`, repositories.ErrDatabaseOperation, err.Error())
			}
			record.Set("child_playlist_ids", string(childPlaylistIDsJSON))
		} else {
			// Clear child playlist IDs when empty slice is provided
			record.Set("child_playlist_ids", "")
		}
	}

	// Update optional fields
	if syncEvent.CompletedAt != nil {
		record.Set("completed_at", *syncEvent.CompletedAt)
	}
	if syncEvent.ErrorMessage != nil {
		record.Set("error_message", *syncEvent.ErrorMessage)
	}

	err = seRepo.app.Save(record)
	if err != nil {
		seRepo.log.ErrorContext(ctx, "unable to update sync_event record", "id", id, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	seRepo.log.InfoContext(ctx, "sync_event updated successfully", "id", id)
	return recordToSyncEvent(record), nil
}

func (seRepo *SyncEventRepositoryPocketbase) GetByID(ctx context.Context, id string) (*models.SyncEvent, error) {
	collection, err := seRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record, err := seRepo.app.FindRecordById(collection, id)
	if err != nil {
		seRepo.log.ErrorContext(ctx, "unable to find sync_event record", "id", id, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrSyncEventNotFound, err.Error())
	}

	seRepo.log.InfoContext(ctx, "sync_event retrieved successfully", "sync_event", record)
	return recordToSyncEvent(record), nil
}

func (seRepo *SyncEventRepositoryPocketbase) GetByUserID(ctx context.Context, userID string) ([]*models.SyncEvent, error) {
	collection, err := seRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	records, err := seRepo.app.FindRecordsByFilter(
		collection,
		"user_id = {:userID}",
		"-created", // Order by created date descending (newest first)
		0,          // limit (0 = no limit)
		0,          // offset
		dbx.Params{
			"userID": userID,
		},
	)
	if err != nil {
		seRepo.log.ErrorContext(ctx, "unable to find sync_event records for user", "user_id", userID, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	syncEvents := make([]*models.SyncEvent, len(records))
	for i, record := range records {
		syncEvents[i] = recordToSyncEvent(record)
	}

	seRepo.log.InfoContext(ctx, "sync_events retrieved successfully", "user_id", userID, "count", len(syncEvents))
	return syncEvents, nil
}

func (seRepo *SyncEventRepositoryPocketbase) GetByBasePlaylistID(ctx context.Context, basePlaylistID string) ([]*models.SyncEvent, error) {
	collection, err := seRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	records, err := seRepo.app.FindRecordsByFilter(
		collection,
		"base_playlist_id = {:basePlaylistID}",
		"-created", // Order by created date descending (newest first)
		0,          // limit (0 = no limit)
		0,          // offset
		dbx.Params{
			"basePlaylistID": basePlaylistID,
		},
	)
	if err != nil {
		seRepo.log.ErrorContext(ctx, "unable to find sync_event records for base playlist", "base_playlist_id", basePlaylistID, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	syncEvents := make([]*models.SyncEvent, len(records))
	for i, record := range records {
		syncEvents[i] = recordToSyncEvent(record)
	}

	seRepo.log.InfoContext(ctx, "sync_events retrieved successfully", "base_playlist_id", basePlaylistID, "count", len(syncEvents))
	return syncEvents, nil
}

func (seRepo *SyncEventRepositoryPocketbase) getCollection(ctx context.Context) (*core.Collection, error) {
	collection, err := seRepo.app.FindCollectionByNameOrId(string(seRepo.collection))
	if err != nil {
		seRepo.log.ErrorContext(ctx, "unable to find collection", "collection", seRepo.collection, "error", err)
		return nil, repositories.ErrCollectionNotFound
	}

	return collection, nil
}

func recordToSyncEvent(record *core.Record) *models.SyncEvent {
	syncEvent := &models.SyncEvent{
		ID:               record.Id,
		UserID:           record.GetString("user_id"),
		BasePlaylistID:   record.GetString("base_playlist_id"),
		Status:           models.SyncStatus(record.GetString("status")),
		StartedAt:        record.GetDateTime("started_at").Time(),
		TracksProcessed:  record.GetInt("tracks_processed"),
		TotalAPIRequests: record.GetInt("total_api_requests"),
		Created:          record.GetDateTime("created").Time(),
		Updated:          record.GetDateTime("updated").Time(),
	}

	// Deserialize child playlist IDs from JSON
	childPlaylistIDsJSON := record.GetString("child_playlist_ids")
	if childPlaylistIDsJSON != "" {
		var childPlaylistIDs []string
		if err := json.Unmarshal([]byte(childPlaylistIDsJSON), &childPlaylistIDs); err == nil {
			syncEvent.ChildPlaylistIDs = childPlaylistIDs
		}
	} else {
		// Ensure we have an empty slice instead of nil for consistency
		syncEvent.ChildPlaylistIDs = []string{}
	}

	// Handle optional fields
	if completedAtTime := record.GetDateTime("completed_at"); !completedAtTime.IsZero() {
		completedAt := completedAtTime.Time()
		syncEvent.CompletedAt = &completedAt
	}

	if errorMessage := record.GetString("error_message"); errorMessage != "" {
		syncEvent.ErrorMessage = &errorMessage
	}

	return syncEvent
}
