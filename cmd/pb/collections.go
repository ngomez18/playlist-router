package main

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// initCollections creates all required collections if they don't exist
func initCollections(app *pocketbase.PocketBase) error {
	if err := createBasePlaylistCollection(app); err != nil {
		return err
	}

	// Add other collection initialization functions here as needed
	// if err := createChildPlaylistCollection(app); err != nil {
	//     return err
	// }

	return nil
}

// createBasePlaylistCollection creates the base_playlists collection
func createBasePlaylistCollection(app *pocketbase.PocketBase) error {
	// Check if base_playlists collection exists
	_, err := app.FindCollectionByNameOrId("base_playlists")
	if err == nil {
		// Collection already exists
		return nil
	}

	// Create base_playlists collection
	collection := core.NewBaseCollection("base_playlists")

	// Add fields
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
