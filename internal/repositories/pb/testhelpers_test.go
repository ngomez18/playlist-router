package pb

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// NewTestApp creates a new PocketBase instance for testing
func NewTestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	tmpDir := t.TempDir()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: tmpDir,
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap test app: %v", err)
	}

	return app
}

// SetupBasePlaylistCollection creates the base_playlist collection for testing
func SetupBasePlaylistCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId(string(CollectionBasePlaylist))
	if err == nil {
		return // Collection already exists
	}

	collection := core.NewBaseCollection(string(CollectionBasePlaylist))

	collection.Fields.Add(&core.TextField{
		Name:     "user_id",
		Required: true,
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

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create base_playlist collection: %v", err)
	}
}

// CreateTestUser creates a test user and returns the user ID
func CreateTestUser(t *testing.T, app *pocketbase.PocketBase, email, name string) string {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("users collection not found: %v", err)
	}

	user := core.NewRecord(collection)
	user.Set("email", email)
	user.Set("name", name)
	user.Set("password", "test123456") // Required for auth collections
	user.Set("passwordConfirm", "test123456")

	if err := app.Save(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user.Id
}

// SetupChildPlaylistCollection creates the child_playlists collection for testing
func SetupChildPlaylistCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId(string(CollectionChildPlaylist))
	if err == nil {
		return // Collection already exists
	}

	collection := core.NewBaseCollection(string(CollectionChildPlaylist))

	collection.Fields.Add(&core.TextField{
		Name:     "user_id",
		Required: true,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "base_playlist_id",
		Required: true,
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

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create child_playlists collection: %v", err)
	}
}

// SetupSpotifyIntegrationsCollection creates the spotify_integrations collection for testing
func SetupSpotifyIntegrationsCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId(string(CollectionSpotifyIntegration))
	if err == nil {
		return // Collection already exists
	}

	// Get the users collection to reference it properly
	usersCollection, err := app.FindCollectionByNameOrId(string(CollectionUsers))
	if err != nil {
		t.Fatalf("users collection not found, make sure to call SetupUsersCollection first: %v", err)
	}

	collection := core.NewBaseCollection(string(CollectionSpotifyIntegration))

	// Foreign key to users collection
	collection.Fields.Add(&core.RelationField{
		Name:          "user",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  usersCollection.Id,
		CascadeDelete: true,
	})

	// Spotify user ID
	collection.Fields.Add(&core.TextField{
		Name:     "spotify_id",
		Required: true,
	})

	// Access token
	collection.Fields.Add(&core.TextField{
		Name:     "access_token",
		Required: true,
	})

	// Refresh token
	collection.Fields.Add(&core.TextField{
		Name:     "refresh_token",
		Required: true,
	})

	// Token type
	collection.Fields.Add(&core.TextField{
		Name:     "token_type",
		Required: false,
	})

	// Token expiration
	collection.Fields.Add(&core.DateField{
		Name:     "expires_at",
		Required: true,
	})

	// Scopes
	collection.Fields.Add(&core.TextField{
		Name:     "scope",
		Required: false,
	})

	// Display name
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

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}
}

// SetupSyncEventCollection creates the sync_events collection for testing
func SetupSyncEventCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId(string(CollectionSyncEvent))
	if err == nil {
		return // Collection already exists
	}

	collection := core.NewBaseCollection(string(CollectionSyncEvent))

	collection.Fields.Add(&core.TextField{
		Name:     "user_id",
		Required: true,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "base_playlist_id",
		Required: true,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "child_playlist_ids",
		Required: false,
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
		Name:     "completed_at",
		Required: false,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "error_message",
		Required: false,
	})

	collection.Fields.Add(&core.NumberField{
		Name:     "tracks_processed",
		Required: false,
	})

	collection.Fields.Add(&core.NumberField{
		Name:     "total_api_requests",
		Required: false,
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

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create sync_events collection: %v", err)
	}
}

// SetupAllCollections sets up all collections needed for testing
func SetupAllCollections(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()
	SetupBasePlaylistCollection(t, app)
	SetupSpotifyIntegrationsCollection(t, app)
	SetupChildPlaylistCollection(t, app)
	SetupSyncEventCollection(t, app)
}
