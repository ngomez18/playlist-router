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

type ChildPlaylistRepositoryPocketbase struct {
	collection Collection
	app        *pocketbase.PocketBase
	log        *slog.Logger
}

func NewChildPlaylistRepositoryPocketbase(pb *pocketbase.PocketBase) *ChildPlaylistRepositoryPocketbase {
	return &ChildPlaylistRepositoryPocketbase{
		collection: CollectionChildPlaylist,
		app:        pb,
		log:        pb.Logger().With("component", "ChildPlaylistRepositoryPocketbase"),
	}
}

func (cpRepo *ChildPlaylistRepositoryPocketbase) Create(
	ctx context.Context,
	userID,
	basePlaylistID,
	name,
	description,
	spotifyPlaylistID string,
	filterRules *models.AudioFeatureFilters,
) (*models.ChildPlaylist, error) {
	collection, err := cpRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	childPlaylist := core.NewRecord(collection)
	childPlaylist.Set("user_id", userID)
	childPlaylist.Set("base_playlist_id", basePlaylistID)
	childPlaylist.Set("name", name)
	childPlaylist.Set("description", description)
	childPlaylist.Set("spotify_playlist_id", spotifyPlaylistID)
	childPlaylist.Set("is_active", true)

	// Serialize filter rules to JSON
	if filterRules != nil {
		filterRulesJSON, err := json.Marshal(filterRules)
		if err != nil {
			cpRepo.log.ErrorContext(ctx, "unable to serialize filter rules", "filter_rules", filterRules, "error", err)
			return nil, fmt.Errorf(`%w: failed to serialize filter rules: %s`, repositories.ErrDatabaseOperation, err.Error())
		}

		childPlaylist.Set("filter_rules", string(filterRulesJSON))
	}

	err = cpRepo.app.Save(childPlaylist)
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to store child_playlist record", "record", childPlaylist, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	cpRepo.log.InfoContext(ctx, "child_playlist stored successfully", "record", childPlaylist)

	return recordToChildPlaylist(childPlaylist), nil
}

func (cpRepo *ChildPlaylistRepositoryPocketbase) Delete(ctx context.Context, id, userID string) error {
	collection, err := cpRepo.getCollection(ctx)
	if err != nil {
		return err
	}

	record, err := cpRepo.app.FindRecordById(collection, id)
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to find child_playlist record", "id", id, "error", err)
		return fmt.Errorf(`%w: %s`, repositories.ErrChildPlaylistNotFound, err.Error())
	}

	// Check ownership (belongs to the specified user)
	if record.GetString("user_id") != userID {
		cpRepo.log.ErrorContext(ctx, "unauthorized delete attempt",
			"id", id,
			"user_id", userID,
			"actual_user_id", record.GetString("user_id"),
		)
		return repositories.ErrUnauthorized
	}

	err = cpRepo.app.Delete(record)
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to delete child_playlist record", "id", id, "error", err)
		return fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	cpRepo.log.InfoContext(ctx, "child_playlist deleted successfully", "id", id, "user_id", userID)
	return nil
}

func (cpRepo *ChildPlaylistRepositoryPocketbase) GetByID(ctx context.Context, id, userID string) (*models.ChildPlaylist, error) {
	collection, err := cpRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record, err := cpRepo.app.FindRecordById(collection, id)
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to find child_playlist record", "id", id, "error", err)
		return nil, repositories.ErrChildPlaylistNotFound
	}

	// Check ownership (belongs to the specified user)
	if record.GetString("user_id") != userID {
		cpRepo.log.ErrorContext(ctx, "unauthorized access attempt",
			"id", id,
			"user_id", userID,
			"actual_user_id", record.GetString("user_id"),
		)
		return nil, repositories.ErrUnauthorized
	}

	cpRepo.log.InfoContext(ctx, "child_playlist retrieved successfully", "child_playlist", record)
	return recordToChildPlaylist(record), nil
}

func (cpRepo *ChildPlaylistRepositoryPocketbase) GetByBasePlaylistID(ctx context.Context, basePlaylistID, userID string) ([]*models.ChildPlaylist, error) {
	collection, err := cpRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	records, err := cpRepo.app.FindRecordsByFilter(
		collection,
		"base_playlist_id = {:basePlaylistID} && user_id = {:userID}",
		"-created", // Order by created date descending (newest first)
		0,          // limit (0 = no limit)
		0,          // offset
		dbx.Params{
			"basePlaylistID": basePlaylistID,
			"userID":         userID,
		},
	)
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to find child_playlist records for base playlist", "base_playlist_id", basePlaylistID, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	childPlaylists := make([]*models.ChildPlaylist, len(records))
	for i, record := range records {
		childPlaylists[i] = recordToChildPlaylist(record)
	}

	cpRepo.log.InfoContext(ctx, "child_playlists retrieved successfully", "base_playlist_id", basePlaylistID, "user_id", userID, "count", len(childPlaylists))
	return childPlaylists, nil
}

func (cpRepo *ChildPlaylistRepositoryPocketbase) Update(ctx context.Context, id, userID string, req *models.UpdateChildPlaylistRequest) (*models.ChildPlaylist, error) {
	collection, err := cpRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record, err := cpRepo.app.FindRecordById(collection, id)
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to find child_playlist record", "id", id, "error", err)
		return nil, repositories.ErrChildPlaylistNotFound
	}

	// Check ownership (belongs to the specified user)
	if record.GetString("user_id") != userID {
		cpRepo.log.ErrorContext(ctx, "unauthorized update attempt",
			"id", id,
			"user_id", userID,
			"actual_user_id", record.GetString("user_id"),
		)
		return nil, repositories.ErrUnauthorized
	}

	// Update fields if provided
	if req.Name != nil {
		record.Set("name", *req.Name)
	}

	if req.Description != nil {
		record.Set("description", *req.Description)
	}

	if req.IsActive != nil {
		record.Set("is_active", *req.IsActive)
	}

	if req.FilterRules != nil {
		filterRulesJSON, err := json.Marshal(req.FilterRules)
		if err != nil {
			cpRepo.log.ErrorContext(ctx, "unable to serialize filter rules", "filter_rules", req.FilterRules, "error", err)
			return nil, fmt.Errorf(`%w: failed to serialize filter rules: %s`, repositories.ErrDatabaseOperation, err.Error())
		}
		record.Set("filter_rules", string(filterRulesJSON))
	}

	err = cpRepo.app.Save(record)
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to update child_playlist record", "id", id, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	cpRepo.log.InfoContext(ctx, "child_playlist updated successfully", "id", id)
	return recordToChildPlaylist(record), nil
}

func (cpRepo *ChildPlaylistRepositoryPocketbase) getCollection(ctx context.Context) (*core.Collection, error) {
	collection, err := cpRepo.app.FindCollectionByNameOrId(string(cpRepo.collection))
	if err != nil {
		cpRepo.log.ErrorContext(ctx, "unable to find collection", "collection", cpRepo.collection, "error", err)
		return nil, repositories.ErrCollectionNotFound
	}

	return collection, nil
}

func recordToChildPlaylist(record *core.Record) *models.ChildPlaylist {
	childPlaylist := &models.ChildPlaylist{
		ID:                record.Id,
		UserID:            record.GetString("user_id"),
		BasePlaylistID:    record.GetString("base_playlist_id"),
		Name:              record.GetString("name"),
		Description:       record.GetString("description"),
		SpotifyPlaylistID: record.GetString("spotify_playlist_id"),
		IsActive:          record.GetBool("is_active"),
		Created:           record.GetDateTime("created").Time(),
		Updated:           record.GetDateTime("updated").Time(),
	}

	// Deserialize filter rules from JSON
	filterRulesJSON := record.GetString("filter_rules")
	if filterRulesJSON != "" {
		var filterRules models.AudioFeatureFilters
		if err := json.Unmarshal([]byte(filterRulesJSON), &filterRules); err == nil {
			childPlaylist.FilterRules = &filterRules
		}
	}

	return childPlaylist
}
