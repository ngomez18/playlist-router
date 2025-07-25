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

// SetupChildPlaylistCollection creates the child_playlists collection for testing
// TODO: Implement when child playlist repository is created
func SetupChildPlaylistCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()
	// Implementation will go here when needed
}

// SetupAllCollections sets up all collections needed for testing
func SetupAllCollections(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()
	SetupBasePlaylistCollection(t, app)
	SetupSpotifyIntegrationsCollection(t, app)
	// Add other collections as needed:
	// SetupChildPlaylistCollection(t, app)
}
