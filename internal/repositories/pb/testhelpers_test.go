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

// SetupBasePlaylistCollection creates the base_playlists collection for testing
func SetupBasePlaylistCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("base_playlists")
	if err == nil {
		return // Collection already exists
	}

	collection := core.NewBaseCollection("base_playlists")

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

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create base_playlists collection: %v", err)
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
	// Add other collections as needed:
	// SetupChildPlaylistCollection(t, app)
}
