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
	collection string
	app        *pocketbase.PocketBase
	log        *slog.Logger
}

func NewBasePlaylistRepositoryPocketbase(pb *pocketbase.PocketBase) *BasePlaylistRepositoryPocketbase {
	return &BasePlaylistRepositoryPocketbase{
		collection: "base_playlists",
		app:        pb,
		log:        pb.Logger().With("component", "BasePlaylistRepositoryPocketbase"),
	}
}

func (bpRepo *BasePlaylistRepositoryPocketbase) Create(ctx context.Context, userId, name, spotifyPlaylistId string) (*models.BasePlaylist, error) {
	collection, err := bpRepo.app.FindCollectionByNameOrId(bpRepo.collection)
	if err != nil {
		bpRepo.log.ErrorContext(ctx, "unable to find collection", "collection", bpRepo.collection, "error", err)
		return nil, repositories.ErrCollectionNotFound
	}

	basePlaylist := core.NewRecord(collection)
	basePlaylist.Set("user_id", userId)
	basePlaylist.Set("name", name)
	basePlaylist.Set("spotify_playlist_id", spotifyPlaylistId)
	basePlaylist.Set("is_active", true)
	basePlaylist.Set("last_synced", nil)

	err = bpRepo.app.Save(basePlaylist)
	if err != nil {
		bpRepo.log.ErrorContext(ctx, "unable to store base_playlist record", "record", basePlaylist, "error", err)
		return nil, fmt.Errorf(`%w: %s`, repositories.ErrInvalidBasePlaylist, err.Error())
	}

	bpRepo.log.InfoContext(ctx, "base_playlist stored successfully", "record", basePlaylist)

	return recordToBasePlaylist(basePlaylist), nil
}

//
// GetBasePlaylist(id string) (*models.BasePlaylist, error)
// DeleteBasePlaylist(id string) error

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
