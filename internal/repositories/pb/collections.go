package pb

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Collection string

var (
	CollectionUsers              Collection = "users"
	CollectionBasePlaylist       Collection = "base_playlists"
	CollectionChildPlaylist      Collection = "child_playlists"
	CollectionSpotifyIntegration Collection = "spotify_integrations"
	CollectionSyncEvent          Collection = "sync_events"
)

func GetCollection(ctx context.Context, app *pocketbase.PocketBase, collectionName Collection) (*core.Collection, error) {
	collection, err := app.FindCollectionByNameOrId(string(collectionName))
	if err != nil {
		app.Logger().ErrorContext(ctx, "unable to find collection", "collection", collectionName, "error", err)
		return nil, repositories.ErrCollectionNotFound
	}

	return collection, nil
}
