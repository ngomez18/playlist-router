package pb

import (
	"fmt"

	"github.com/ngomez18/playlist-router/internal/config"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// InitCollections creates all required collections if they don't exist
func InitCollections(app *pocketbase.PocketBase, cfg *config.Config) error {
	if err := createAdminUser(app, cfg); err != nil {
		return err
	}

	if err := createBasePlaylistCollection(app); err != nil {
		return err
	}

	if err := createSpotifyIntegrationsCollection(app); err != nil {
		return err
	}

	if err := createChildPlaylistCollection(app); err != nil {
		return err
	}

	if err := createSyncEventCollection(app); err != nil {
		return err
	}

	return nil
}

func createAdminUser(app *pocketbase.PocketBase, cfg *config.Config) error {
	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	user, err := app.FindRecordsByFilter(
		superusers,
		"email = {:email}",
		"created",
		1,
		0,
		dbx.Params{"email": cfg.AdminEmail},
	)
	if err != nil || len(user) > 0 {
		return err
	}

	record := core.NewRecord(superusers)
	record.Set("email", cfg.AdminEmail)
	record.Set("password", cfg.AdminPassword)

	return app.Save(record)
}

// createBasePlaylistCollection creates the base_playlists collection
func createBasePlaylistCollection(app *pocketbase.PocketBase) error {
	// Check if base_playlists collection exists
	_, err := app.FindCollectionByNameOrId(string(CollectionBasePlaylist))
	if err == nil {
		// Collection already exists
		return nil
	}

	// Create base_playlists collection
	collection := core.NewBaseCollection(string(CollectionBasePlaylist))

	// Add fields - user_id as relation to users collection
	collection.Fields.Add(&core.RelationField{
		Name:          "user_id",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  "_pb_users_auth_",
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Max:      100,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "spotify_playlist_id",
		Required: true,
	})

	collection.Fields.Add(&core.BoolField{
		Name:     "is_active",
		Required: false,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	// Create unique index on user_id + spotify_playlist_id to prevent duplicate imports
	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_base_playlists_user_spotify ON base_playlists (user_id, spotify_playlist_id)",
	}

	return app.Save(collection)
}

// createSpotifyIntegrationsCollection creates the spotify_integrations collection
func createSpotifyIntegrationsCollection(app *pocketbase.PocketBase) error {
	// Check if spotify_integrations collection exists
	_, err := app.FindCollectionByNameOrId(string(CollectionSpotifyIntegration))
	if err == nil {
		// Collection already exists
		return nil
	}

	// Create spotify_integrations collection
	collection := core.NewBaseCollection(string(CollectionSpotifyIntegration))

	// Foreign key to users collection (PocketBase relation field)
	collection.Fields.Add(&core.RelationField{
		Name:          "user",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  "_pb_users_auth_",
		CascadeDelete: true,
	})

	// Spotify user ID (unique identifier from Spotify)
	collection.Fields.Add(&core.TextField{
		Name:     "spotify_id",
		Required: true,
	})

	// Access token (encrypted by PocketBase automatically for security)
	collection.Fields.Add(&core.TextField{
		Name:     "access_token",
		Required: true,
	})

	// Refresh token
	collection.Fields.Add(&core.TextField{
		Name:     "refresh_token",
		Required: true,
	})

	// Token type (usually "Bearer")
	collection.Fields.Add(&core.TextField{
		Name:     "token_type",
		Required: false,
	})

	// Token expiration timestamp
	collection.Fields.Add(&core.DateField{
		Name:     "expires_at",
		Required: true,
	})

	// Scopes granted by user
	collection.Fields.Add(&core.TextField{
		Name:     "scope",
		Required: false,
	})

	// Spotify display name
	collection.Fields.Add(&core.TextField{
		Name:     "display_name",
		Required: false,
		Max:      200,
	})

	// Standard timestamp fields
	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	// Create unique index on user_id to ensure one integration per user
	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_spotify_integrations_user ON spotify_integrations (user)",
	}

	return app.Save(collection)
}

// createChildPlaylistCollection creates the child_playlists collection
func createChildPlaylistCollection(app *pocketbase.PocketBase) error {
	// Check if child_playlists collection exists
	_, err := app.FindCollectionByNameOrId(string(CollectionChildPlaylist))
	if err == nil {
		// Collection already exists
		return nil
	}

	// Get the base_playlists collection to reference it properly
	basePlaylistCollection, err := app.FindCollectionByNameOrId(string(CollectionBasePlaylist))
	if err != nil {
		return fmt.Errorf("base_playlists collection must exist before creating child_playlists: %w", err)
	}

	// Create child_playlists collection
	collection := core.NewBaseCollection(string(CollectionChildPlaylist))

	// Add fields - user_id as relation to users collection
	collection.Fields.Add(&core.RelationField{
		Name:          "user_id",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  "_pb_users_auth_",
		CascadeDelete: true,
	})

	// Foreign key to base_playlists collection
	collection.Fields.Add(&core.RelationField{
		Name:          "base_playlist_id",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  basePlaylistCollection.Id,
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Max:      100,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "description",
		Required: false,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "spotify_playlist_id",
		Required: true,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "filter_rules",
		Required: false,
	})

	collection.Fields.Add(&core.BoolField{
		Name:     "is_active",
		Required: false,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	// Create unique index on base_playlist_id + spotify_playlist_id to prevent duplicate child playlists
	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_child_playlists_base_spotify ON child_playlists (base_playlist_id, spotify_playlist_id)",
	}

	return app.Save(collection)
}

// createSyncEventCollection creates the sync_events collection
func createSyncEventCollection(app *pocketbase.PocketBase) error {
	// Check if sync_events collection exists
	_, err := app.FindCollectionByNameOrId(string(CollectionSyncEvent))
	if err == nil {
		// Collection already exists
		return nil
	}

	basePlaylistCollection, err := app.FindCollectionByNameOrId(string(CollectionBasePlaylist))
	if err != nil {
		return fmt.Errorf("base_playlists collection must exist before creating sync_events: %w", err)
	}

	// Create sync_events collection
	collection := core.NewBaseCollection(string(CollectionSyncEvent))

	// Add fields
	collection.Fields.Add(&core.RelationField{
		Name:          "user_id",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  "_pb_users_auth_",
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.RelationField{
		Name:          "base_playlist_id",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  basePlaylistCollection.Id,
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.TextField{
		Name: "child_playlist_ids",
	})

	collection.Fields.Add(&core.TextField{
		Name:     "status",
		Required: true,
	})

	collection.Fields.Add(&core.DateField{
		Name:     "started_at",
		Required: true,
	})

	collection.Fields.Add(&core.DateField{
		Name: "completed_at",
	})

	collection.Fields.Add(&core.TextField{
		Name: "error_message",
	})

	collection.Fields.Add(&core.NumberField{
		Name: "tracks_processed",
	})

	collection.Fields.Add(&core.NumberField{
		Name: "total_api_requests",
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return app.Save(collection)
}
