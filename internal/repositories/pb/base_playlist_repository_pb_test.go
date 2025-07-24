package pb

import (
	"context"
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/pocketbase/pocketbase"
	"github.com/stretchr/testify/require"
)

func TestBasePlaylistRepositoryPocketbase_Create_Success(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		playlistName      string
		spotifyPlaylistID string
	}{
		{
			name:              "successful creation with valid data",
			userID:            "user123",
			playlistName:      "My Test Playlist",
			spotifyPlaylistID: "spotify123",
		},
		{
			name:              "successful creation with minimum valid name",
			userID:            "user456",
			playlistName:      "A",
			spotifyPlaylistID: "spotify456",
		},
		{
			name:              "successful creation with maximum valid name",
			userID:            "user789",
			playlistName:      "This is a valid playlist name that is exactly 100 characters long and should pass validation",
			spotifyPlaylistID: "spotify789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupBasePlaylistCollection(t, app)
			repo := NewBasePlaylistRepositoryPocketbase(app)

			// Execute test
			ctx := context.Background()
			playlist, err := repo.Create(ctx, tt.userID, tt.playlistName, tt.spotifyPlaylistID)

			// Verify success
			assert.NoError(err)
			assert.NotNil(playlist)

			// Verify playlist fields
			assert.Equal(tt.userID, playlist.UserID)
			assert.Equal(tt.playlistName, playlist.Name)
			assert.Equal(tt.spotifyPlaylistID, playlist.SpotifyPlaylistID)
			assert.True(playlist.IsActive)
			assert.NotEmpty(playlist.ID)

			// Verify the playlist was actually saved to the database
			savedPlaylist, err := findBasePlaylistInDB(t, app, playlist.ID)
			assert.NoError(err)
			assert.Equal(tt.userID, savedPlaylist.UserID)
		})
	}
}

func TestBasePlaylistRepositoryPocketbase_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		playlistName      string
		spotifyPlaylistID string
		wantErrContains   string
	}{
		{
			name:              "empty user ID",
			userID:            "",
			playlistName:      "My Test Playlist",
			spotifyPlaylistID: "spotify123",
			wantErrContains:   "user_id: cannot be blank",
		},
		{
			name:              "empty playlist name",
			userID:            "user123",
			playlistName:      "",
			spotifyPlaylistID: "spotify123",
			wantErrContains:   "name: cannot be blank",
		},
		{
			name:              "empty spotify ID",
			userID:            "user123",
			playlistName:      "My Test Playlist",
			spotifyPlaylistID: "",
			wantErrContains:   "spotify_playlist_id: cannot be blank",
		},
		{
			name:              "playlist name too long",
			userID:            "user123",
			playlistName:      "This is a very long playlist name that exceeds the maximum allowed length of 100 characters which should cause an error",
			spotifyPlaylistID: "spotify123",
			wantErrContains:   "name: Must be no more than 100 character(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupBasePlaylistCollection(t, app)
			repo := NewBasePlaylistRepositoryPocketbase(app)

			// Execute test
			ctx := context.Background()
			playlist, err := repo.Create(ctx, tt.userID, tt.playlistName, tt.spotifyPlaylistID)

			// Verify error occurred
			assert.Error(err)
			assert.Nil(playlist)

			// Verify specific error message contains expected text
			if tt.wantErrContains != "" {
				assert.Contains(err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestBasePlaylistRepositoryPocketbase_Create_DatabaseErrors(t *testing.T) {
	t.Run("collection not found", func(t *testing.T) {
		assert := require.New(t)

		// Setup test environment without creating the collection
		app := NewTestApp(t)
		repo := NewBasePlaylistRepositoryPocketbase(app)

		// Execute test
		ctx := context.Background()
		playlist, err := repo.Create(ctx, "user123", "Test Playlist", "spotify123")

		// Verify error occurred
		assert.Error(err)
		assert.Nil(playlist)
	})
}

// findBasePlaylistInDB is a helper function to verify a playlist exists in the database
func findBasePlaylistInDB(t *testing.T, app *pocketbase.PocketBase, id string) (*models.BasePlaylist, error) {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("base_playlists")
	if err != nil {
		return nil, err
	}

	record, err := app.FindRecordById(collection, id)
	if err != nil {
		return nil, err
	}

	return recordToBasePlaylist(record), nil
}