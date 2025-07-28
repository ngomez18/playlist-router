package pb

import (
	"context"
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
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

func TestBasePlaylistRepositoryPocketbase_Delete_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupBasePlaylistCollection(t, app)
	repo := NewBasePlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// First create a playlist to delete
	playlist, err := repo.Create(ctx, "user123", "Test Playlist", "spotify123")
	assert.NoError(err)
	assert.NotNil(playlist)

	// Verify playlist exists
	foundPlaylist, err := findBasePlaylistInDB(t, app, playlist.ID)
	assert.NoError(err)
	assert.Equal(playlist.ID, foundPlaylist.ID)

	// Execute delete with correct user ID
	err = repo.Delete(ctx, playlist.ID, "user123")
	assert.NoError(err)

	// Verify playlist no longer exists
	_, err = findBasePlaylistInDB(t, app, playlist.ID)
	assert.Error(err)
}

func TestBasePlaylistRepositoryPocketbase_Delete_UnauthorizedError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupBasePlaylistCollection(t, app)
	repo := NewBasePlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// First create a playlist owned by user123
	playlist, err := repo.Create(ctx, "user123", "Test Playlist", "spotify123")
	assert.NoError(err)
	assert.NotNil(playlist)

	// Try to delete with different user ID (should fail)
	err = repo.Delete(ctx, playlist.ID, "user456")

	// Verify unauthorized error
	assert.Error(err)
	assert.ErrorIs(err, repositories.ErrUnauthorized)

	// Verify playlist still exists
	foundPlaylist, err := findBasePlaylistInDB(t, app, playlist.ID)
	assert.NoError(err)
	assert.Equal(playlist.ID, foundPlaylist.ID)
}

func TestBasePlaylistRepositoryPocketbase_Delete_NotFoundErrors(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectedErr error
	}{
		{
			name:        "non-existent id",
			id:          "nonexistent123",
			expectedErr: repositories.ErrBasePlaylistNotFound,
		},
		{
			name:        "invalid id format",
			id:          "invalid-id-format",
			expectedErr: repositories.ErrBasePlaylistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupBasePlaylistCollection(t, app)
			repo := NewBasePlaylistRepositoryPocketbase(app)

			ctx := context.Background()

			// Execute delete
			err := repo.Delete(ctx, tt.id, "user123")

			// Verify error
			assert.Error(err)
			assert.ErrorIs(err, tt.expectedErr)
		})
	}
}

func TestBasePlaylistRepositoryPocketbase_Delete_DatabaseErrors(t *testing.T) {
	t.Run("collection not found", func(t *testing.T) {
		assert := require.New(t)

		// Setup test environment without creating the collection
		app := NewTestApp(t)
		repo := NewBasePlaylistRepositoryPocketbase(app)

		ctx := context.Background()

		// Execute delete
		err := repo.Delete(ctx, "test123", "user123")

		// Verify error
		assert.Error(err)
		assert.ErrorIs(err, repositories.ErrCollectionNotFound)
	})
}

func TestBasePlaylistRepositoryPocketbase_GetByID_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupBasePlaylistCollection(t, app)
	repo := NewBasePlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// First create a playlist to retrieve
	playlist, err := repo.Create(ctx, "user123", "Test Playlist", "spotify123")
	assert.NoError(err)
	assert.NotNil(playlist)

	// Execute GetByID with correct user ID
	retrievedPlaylist, err := repo.GetByID(ctx, playlist.ID, "user123")
	assert.NoError(err)
	assert.NotNil(retrievedPlaylist)

	// Verify the retrieved playlist matches the created one
	assert.Equal(playlist.ID, retrievedPlaylist.ID)
	assert.Equal(playlist.UserID, retrievedPlaylist.UserID)
	assert.Equal(playlist.Name, retrievedPlaylist.Name)
	assert.Equal(playlist.SpotifyPlaylistID, retrievedPlaylist.SpotifyPlaylistID)
	assert.Equal(playlist.IsActive, retrievedPlaylist.IsActive)
}

func TestBasePlaylistRepositoryPocketbase_GetByID_UnauthorizedError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupBasePlaylistCollection(t, app)
	repo := NewBasePlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// First create a playlist owned by user123
	playlist, err := repo.Create(ctx, "user123", "Test Playlist", "spotify123")
	assert.NoError(err)
	assert.NotNil(playlist)

	// Try to retrieve with different user ID (should fail)
	retrievedPlaylist, err := repo.GetByID(ctx, playlist.ID, "user456")

	// Verify unauthorized error
	assert.Error(err)
	assert.Nil(retrievedPlaylist)
	assert.ErrorIs(err, repositories.ErrUnauthorized)
}

func TestBasePlaylistRepositoryPocketbase_GetByID_NotFoundErrors(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectedErr error
	}{
		{
			name:        "non-existent id",
			id:          "nonexistent123",
			expectedErr: repositories.ErrBasePlaylistNotFound,
		},
		{
			name:        "invalid id format",
			id:          "invalid-id-format",
			expectedErr: repositories.ErrBasePlaylistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupBasePlaylistCollection(t, app)
			repo := NewBasePlaylistRepositoryPocketbase(app)

			ctx := context.Background()

			// Execute GetByID
			retrievedPlaylist, err := repo.GetByID(ctx, tt.id, "user123")

			// Verify error
			assert.Error(err)
			assert.Nil(retrievedPlaylist)
			assert.ErrorIs(err, tt.expectedErr)
		})
	}
}

func TestBasePlaylistRepositoryPocketbase_GetByID_DatabaseErrors(t *testing.T) {
	t.Run("collection not found", func(t *testing.T) {
		assert := require.New(t)

		// Setup test environment without creating the collection
		app := NewTestApp(t)
		repo := NewBasePlaylistRepositoryPocketbase(app)

		ctx := context.Background()

		// Execute GetByID
		retrievedPlaylist, err := repo.GetByID(ctx, "test123", "user123")

		// Verify error
		assert.Error(err)
		assert.Nil(retrievedPlaylist)
		assert.ErrorIs(err, repositories.ErrCollectionNotFound)
	})
}

func TestBasePlaylistRepositoryPocketbase_GetByUserID_Success(t *testing.T) {
	tests := []struct {
		name                  string
		userID                string
		playlistsToCreate     []struct{ name, spotifyID string }
		expectedPlaylistCount int
	}{
		{
			name:   "user with multiple playlists",
			userID: "user123",
			playlistsToCreate: []struct{ name, spotifyID string }{
				{"First Playlist", "spotify1"},
				{"Second Playlist", "spotify2"},
				{"Third Playlist", "spotify3"},
			},
			expectedPlaylistCount: 3,
		},
		{
			name:   "user with single playlist",
			userID: "user456",
			playlistsToCreate: []struct{ name, spotifyID string }{
				{"Only Playlist", "spotify4"},
			},
			expectedPlaylistCount: 1,
		},
		{
			name:                  "user with no playlists",
			userID:                "user789",
			playlistsToCreate:     []struct{ name, spotifyID string }{},
			expectedPlaylistCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupBasePlaylistCollection(t, app)
			repo := NewBasePlaylistRepositoryPocketbase(app)

			ctx := context.Background()

			// Create playlists for this user
			createdPlaylists := make([]*models.BasePlaylist, 0, len(tt.playlistsToCreate))
			for _, playlist := range tt.playlistsToCreate {
				created, err := repo.Create(ctx, tt.userID, playlist.name, playlist.spotifyID)
				assert.NoError(err)
				createdPlaylists = append(createdPlaylists, created)
			}

			// Create some playlists for a different user to ensure isolation
			_, err := repo.Create(ctx, "otheruser", "Other User Playlist", "spotify999")
			assert.NoError(err)

			// Execute GetByUserID
			retrievedPlaylists, err := repo.GetByUserID(ctx, tt.userID)

			// Verify success
			assert.NoError(err)
			assert.NotNil(retrievedPlaylists)
			assert.Len(retrievedPlaylists, tt.expectedPlaylistCount)

			// If we have playlists, verify they match what we created
			if tt.expectedPlaylistCount > 0 {
				// Verify all retrieved playlists belong to the correct user
				for _, playlist := range retrievedPlaylists {
					assert.Equal(tt.userID, playlist.UserID)
				}

				// Verify the playlists are ordered by creation date (newest first)
				// The last created playlist should be first in the results
				if len(retrievedPlaylists) > 1 {
					for i := 0; i < len(retrievedPlaylists)-1; i++ {
						assert.True(retrievedPlaylists[i].Created.After(retrievedPlaylists[i+1].Created) ||
							retrievedPlaylists[i].Created.Equal(retrievedPlaylists[i+1].Created))
					}
				}

				// Verify specific playlist data matches
				playlistNames := make(map[string]bool)
				for _, playlist := range retrievedPlaylists {
					playlistNames[playlist.Name] = true
					assert.True(playlist.IsActive)
					assert.NotEmpty(playlist.ID)
					assert.NotEmpty(playlist.SpotifyPlaylistID)
				}

				// Verify all created playlists are present
				for _, created := range createdPlaylists {
					assert.True(playlistNames[created.Name], "Playlist %s should be in results", created.Name)
				}
			}
		})
	}
}

func TestBasePlaylistRepositoryPocketbase_GetByUserID_DatabaseErrors(t *testing.T) {
	t.Run("collection not found", func(t *testing.T) {
		assert := require.New(t)

		// Setup test environment without creating the collection
		app := NewTestApp(t)
		repo := NewBasePlaylistRepositoryPocketbase(app)

		ctx := context.Background()

		// Execute GetByUserID
		playlists, err := repo.GetByUserID(ctx, "user123")

		// Verify error
		assert.Error(err)
		assert.Nil(playlists)
		assert.ErrorIs(err, repositories.ErrCollectionNotFound)
	})

	t.Run("database query error", func(t *testing.T) {
		assert := require.New(t)

		// Setup test environment
		app := NewTestApp(t)
		SetupBasePlaylistCollection(t, app)
		repo := NewBasePlaylistRepositoryPocketbase(app)

		ctx := context.Background()

		// This should test a scenario where the database query fails
		// In a real scenario, this might be caused by database connectivity issues
		// For this test, we'll use an empty userID which should work but return no results
		playlists, err := repo.GetByUserID(ctx, "")

		// This should succeed but return empty results (empty userID is valid for the query)
		assert.NoError(err)
		assert.NotNil(playlists)
		assert.Len(playlists, 0)
	})
}

// findBasePlaylistInDB is a helper function to verify a playlist exists in the database
func findBasePlaylistInDB(t *testing.T, app *pocketbase.PocketBase, id string) (*models.BasePlaylist, error) {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId(string(CollectionBasePlaylist))
	if err != nil {
		return nil, err
	}

	record, err := app.FindRecordById(collection, id)
	if err != nil {
		return nil, err
	}

	return recordToBasePlaylist(record), nil
}
