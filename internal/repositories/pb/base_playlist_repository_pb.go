package pb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type BasePlaylistRepositoryPocketbase struct {
	collection Collection
	app        *pocketbase.PocketBase
	log        *slog.Logger
}

func NewBasePlaylistRepositoryPocketbase(pb *pocketbase.PocketBase) *BasePlaylistRepositoryPocketbase {
	return &BasePlaylistRepositoryPocketbase{
		collection: CollectionBasePlaylist,
		app:        pb,
		log:        pb.Logger().With("component", "BasePlaylistRepositoryPocketbase"),
	}
}

func (bpRepo *BasePlaylistRepositoryPocketbase) Create(ctx context.Context, userId, name, spotifyPlaylistId string) (*models.BasePlaylist, error) {
	collection, err := bpRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	basePlaylist := core.NewRecord(collection)
	basePlaylist.Set("user_id", userId)
	basePlaylist.Set("name", name)
	basePlaylist.Set("spotify_playlist_id", spotifyPlaylistId)
	basePlaylist.Set("is_active", true)

	err = bpRepo.app.Save(basePlaylist)
	if err != nil {
		bpRepo.log.ErrorContext(ctx, "unable to store base_playlist record", "record", basePlaylist, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	bpRepo.log.InfoContext(ctx, "base_playlist stored successfully", "record", basePlaylist)

	return recordToBasePlaylist(basePlaylist), nil
}

func (bpRepo *BasePlaylistRepositoryPocketbase) Delete(ctx context.Context, id, userId string) error {
	collection, err := bpRepo.getCollection(ctx)
	if err != nil {
		return err
	}

	record, err := bpRepo.app.FindRecordById(collection, id)
	if err != nil {
		bpRepo.log.ErrorContext(ctx, "unable to find base_playlist record", "id", id, "error", err)
		return repositories.ErrBasePlaylistNotFound
	}

	// Check ownership
	if record.GetString("user_id") != userId {
		bpRepo.log.ErrorContext(ctx, "unauthorized delete attempt",
			"id", id,
			"requested_by", userId,
		)
		return repositories.ErrUnauthorized
	}

	err = bpRepo.app.Delete(record)
	if err != nil {
		bpRepo.log.ErrorContext(ctx, "unable to delete base_playlist record", "id", id, "error", err)
		return fmt.Errorf(`%w: %s`, repositories.ErrDatabaseOperation, err.Error())
	}

	bpRepo.log.InfoContext(ctx, "base_playlist deleted successfully", "id", id, "user_id", userId)
	return nil
}

func (bpRepo *BasePlaylistRepositoryPocketbase) GetByID(ctx context.Context, id, userId string) (*models.BasePlaylist, error) {
	collection, err := bpRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record, err := bpRepo.app.FindRecordById(collection, id)
	if err != nil {
		bpRepo.log.ErrorContext(ctx, "unable to find base_playlist record", "id", id, "error", err)
		return nil, repositories.ErrBasePlaylistNotFound
	}

	// Check ownership
	if record.GetString("user_id") != userId {
		bpRepo.log.ErrorContext(ctx, "unauthorized access attempt",
			"id", id,
			"requested_by", userId,
		)
		return nil, repositories.ErrUnauthorized
	}

	bpRepo.log.InfoContext(ctx, "base_playlist retrieved successfully", "base_playlist", record)
	return recordToBasePlaylist(record), nil
}

func (bpRepo *BasePlaylistRepositoryPocketbase) getCollection(ctx context.Context) (*core.Collection, error) {
	collection, err := bpRepo.app.FindCollectionByNameOrId(string(bpRepo.collection))
	if err != nil {
		bpRepo.log.ErrorContext(ctx, "unable to find collection", "collection", bpRepo.collection, "error", err)
		return nil, repositories.ErrCollectionNotFound
	}

	return collection, nil
}

func recordToBasePlaylist(record *core.Record) *models.BasePlaylist {
	return &models.BasePlaylist{
		ID:                record.Id,
		UserID:            record.GetString("user_id"),
		Name:              record.GetString("name"),
		SpotifyPlaylistID: record.GetString("spotify_playlist_id"),
		IsActive:          record.GetBool("is_active"),
		Created:           record.GetDateTime("created").Time(),
		Updated:           record.GetDateTime("updated").Time(),
	}
}
