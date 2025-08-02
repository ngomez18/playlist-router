package pb

import (
	"context"
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/pocketbase"
	"github.com/stretchr/testify/require"
)

func TestChildPlaylistRepositoryPocketbase_Create_Success(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		basePlaylistID    string
		playlistName      string
		description       string
		spotifyPlaylistID string
		filterRules       *models.AudioFeatureFilters
	}{
		{
			name:              "successful creation with minimal data",
			userID:            "user123",
			basePlaylistID:    "base123",
			playlistName:      "My Child Playlist",
			description:       "",
			spotifyPlaylistID: "spotify123",
			filterRules:       nil,
		},
		{
			name:              "successful creation with description",
			userID:            "user456",
			basePlaylistID:    "base456",
			playlistName:      "High Energy",
			description:       "Songs with high energy levels",
			spotifyPlaylistID: "spotify456",
			filterRules:       nil,
		},
		{
			name:              "successful creation with filter rules",
			userID:            "user789",
			basePlaylistID:    "base789",
			playlistName:      "Chill Vibes",
			description:       "Low energy chill songs",
			spotifyPlaylistID: "spotify789",
			filterRules: &models.MetadataFilters{
				Popularity: &models.RangeFilter{Min: ptrFloat64(30), Max: ptrFloat64(70)},
				Genres:     &models.SetFilter{Include: []string{"chill", "ambient"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupChildPlaylistCollection(t, app)
			repo := NewChildPlaylistRepositoryPocketbase(app)

			// Execute test
			ctx := context.Background()
			playlist, err := repo.Create(ctx, tt.userID, tt.basePlaylistID, tt.playlistName, tt.description, tt.spotifyPlaylistID, tt.filterRules)

			// Verify success
			assert.NoError(err)
			assert.NotNil(playlist)

			// Verify playlist fields
			assert.Equal(tt.userID, playlist.UserID)
			assert.Equal(tt.basePlaylistID, playlist.BasePlaylistID)
			assert.Equal(tt.playlistName, playlist.Name)
			assert.Equal(tt.description, playlist.Description)
			assert.Equal(tt.spotifyPlaylistID, playlist.SpotifyPlaylistID)
			assert.True(playlist.IsActive)
			assert.NotEmpty(playlist.ID)
			assert.Equal(tt.filterRules, playlist.FilterRules)

			// Verify the playlist was actually saved to the database
			savedPlaylist, err := findChildPlaylistInDB(t, app, playlist.ID)
			assert.NoError(err)
			assert.Equal(tt.userID, savedPlaylist.UserID)
		})
	}
}

func TestChildPlaylistRepositoryPocketbase_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		basePlaylistID    string
		playlistName      string
		description       string
		spotifyPlaylistID string
		wantErrContains   string
	}{
		{
			name:              "empty user ID",
			userID:            "",
			basePlaylistID:    "base123",
			playlistName:      "My Playlist",
			description:       "",
			spotifyPlaylistID: "spotify123",
			wantErrContains:   "user_id: cannot be blank",
		},
		{
			name:              "empty base playlist ID",
			userID:            "user123",
			basePlaylistID:    "",
			playlistName:      "My Playlist",
			description:       "",
			spotifyPlaylistID: "spotify123",
			wantErrContains:   "base_playlist_id: cannot be blank",
		},
		{
			name:              "empty playlist name",
			userID:            "user123",
			basePlaylistID:    "base123",
			playlistName:      "",
			description:       "",
			spotifyPlaylistID: "spotify123",
			wantErrContains:   "name: cannot be blank",
		},
		{
			name:              "empty spotify ID",
			userID:            "user123",
			basePlaylistID:    "base123",
			playlistName:      "My Playlist",
			description:       "",
			spotifyPlaylistID: "",
			wantErrContains:   "spotify_playlist_id: cannot be blank",
		},
		{
			name:              "playlist name too long",
			userID:            "user123",
			basePlaylistID:    "base123",
			playlistName:      "This is a very long playlist name that exceeds the maximum allowed length of 100 characters which should cause an error",
			description:       "",
			spotifyPlaylistID: "spotify123",
			wantErrContains:   "name: Must be no more than 100 character(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupChildPlaylistCollection(t, app)
			repo := NewChildPlaylistRepositoryPocketbase(app)

			// Execute test
			ctx := context.Background()
			playlist, err := repo.Create(ctx, tt.userID, tt.basePlaylistID, tt.playlistName, tt.description, tt.spotifyPlaylistID, nil)

			// Verify error occurred
			assert.Error(err)
			assert.Nil(playlist)
			assert.Contains(err.Error(), tt.wantErrContains)
		})
	}
}

func TestChildPlaylistRepositoryPocketbase_Delete_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// First create a child playlist to delete
	playlist, err := repo.Create(ctx, "user123", "base123", "Test Playlist", "Test description", "spotify123", nil)
	assert.NoError(err)
	assert.NotNil(playlist)

	// Verify playlist exists
	foundPlaylist, err := findChildPlaylistInDB(t, app, playlist.ID)
	assert.NoError(err)
	assert.Equal(playlist.ID, foundPlaylist.ID)

	// Execute delete with correct user ID
	err = repo.Delete(ctx, playlist.ID, "user123")
	assert.NoError(err)

	// Verify playlist no longer exists
	_, err = findChildPlaylistInDB(t, app, playlist.ID)
	assert.Error(err)
}

func TestChildPlaylistRepositoryPocketbase_Delete_UnauthorizedError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// First create a playlist owned by user123
	playlist, err := repo.Create(ctx, "user123", "base123", "Test Playlist", "", "spotify123", nil)
	assert.NoError(err)
	assert.NotNil(playlist)

	// Try to delete with different user ID (should fail)
	err = repo.Delete(ctx, playlist.ID, "user456")

	// Verify unauthorized error
	assert.Error(err)
	assert.ErrorIs(err, repositories.ErrUnauthorized)

	// Verify playlist still exists
	foundPlaylist, err := findChildPlaylistInDB(t, app, playlist.ID)
	assert.NoError(err)
	assert.Equal(playlist.ID, foundPlaylist.ID)
}

func TestChildPlaylistRepositoryPocketbase_Delete_NotFoundError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// Execute delete with non-existent ID
	err := repo.Delete(ctx, "nonexistent123", "user123")

	// Verify error
	assert.Error(err)
	assert.ErrorIs(err, repositories.ErrChildPlaylistNotFound)
}

func TestChildPlaylistRepositoryPocketbase_GetByID_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// Create filter rules for testing
	filterRules := &models.AudioFeatureFilters{
		Popularity: &models.RangeFilter{Min: ptrFloat64(70), Max: ptrFloat64(100)},
		Duration: &models.RangeFilter{Min: ptrFloat64(180000), Max: ptrFloat64(300000)},
	}

	// First create a playlist to retrieve
	playlist, err := repo.Create(ctx, "user123", "base123", "Test Playlist", "Test description", "spotify123", filterRules)
	assert.NoError(err)
	assert.NotNil(playlist)

	// Execute GetByID with correct user ID
	retrievedPlaylist, err := repo.GetByID(ctx, playlist.ID, "user123")
	assert.NoError(err)
	assert.NotNil(retrievedPlaylist)

	// Verify the retrieved playlist matches the created one
	assert.Equal(playlist.ID, retrievedPlaylist.ID)
	assert.Equal(playlist.UserID, retrievedPlaylist.UserID)
	assert.Equal(playlist.BasePlaylistID, retrievedPlaylist.BasePlaylistID)
	assert.Equal(playlist.Name, retrievedPlaylist.Name)
	assert.Equal(playlist.Description, retrievedPlaylist.Description)
	assert.Equal(playlist.SpotifyPlaylistID, retrievedPlaylist.SpotifyPlaylistID)
	assert.Equal(playlist.IsActive, retrievedPlaylist.IsActive)
	assert.Equal(filterRules, retrievedPlaylist.FilterRules)
}

func TestChildPlaylistRepositoryPocketbase_GetByID_UnauthorizedError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// First create a playlist owned by user123
	playlist, err := repo.Create(ctx, "user123", "base123", "Test Playlist", "", "spotify123", nil)
	assert.NoError(err)
	assert.NotNil(playlist)

	// Try to retrieve with different user ID (should fail)
	retrievedPlaylist, err := repo.GetByID(ctx, playlist.ID, "user456")

	// Verify unauthorized error
	assert.Error(err)
	assert.Nil(retrievedPlaylist)
	assert.ErrorIs(err, repositories.ErrUnauthorized)
}

func TestChildPlaylistRepositoryPocketbase_GetByID_NotFoundError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// Execute GetByID with non-existent ID
	retrievedPlaylist, err := repo.GetByID(ctx, "nonexistent123", "user123")

	// Verify error
	assert.Error(err)
	assert.Nil(retrievedPlaylist)
	assert.ErrorIs(err, repositories.ErrChildPlaylistNotFound)
}

func TestChildPlaylistRepositoryPocketbase_GetByBasePlaylistID_Success(t *testing.T) {
	tests := []struct {
		name                   string
		userID                 string
		basePlaylistID         string
		childPlaylistsToCreate []struct{ name, description, spotifyID string }
		expectedCount          int
	}{
		{
			name:           "base playlist with multiple children",
			userID:         "user123",
			basePlaylistID: "base123",
			childPlaylistsToCreate: []struct{ name, description, spotifyID string }{
				{"High Energy", "Energetic songs", "spotify1"},
				{"Low Energy", "Chill songs", "spotify2"},
				{"Dance", "Danceable tracks", "spotify3"},
			},
			expectedCount: 3,
		},
		{
			name:           "base playlist with single child",
			userID:         "user456",
			basePlaylistID: "base456",
			childPlaylistsToCreate: []struct{ name, description, spotifyID string }{
				{"Only Child", "The only one", "spotify4"},
			},
			expectedCount: 1,
		},
		{
			name:                   "base playlist with no children",
			userID:                 "user789",
			basePlaylistID:         "base789",
			childPlaylistsToCreate: []struct{ name, description, spotifyID string }{},
			expectedCount:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupChildPlaylistCollection(t, app)
			repo := NewChildPlaylistRepositoryPocketbase(app)

			ctx := context.Background()

			// Create child playlists for this base playlist
			createdPlaylists := make([]*models.ChildPlaylist, 0, len(tt.childPlaylistsToCreate))
			for _, childData := range tt.childPlaylistsToCreate {
				created, err := repo.Create(ctx, tt.userID, tt.basePlaylistID, childData.name, childData.description, childData.spotifyID, nil)
				assert.NoError(err)
				createdPlaylists = append(createdPlaylists, created)
			}

			// Create some child playlists for a different base playlist to ensure isolation
			_, err := repo.Create(ctx, tt.userID, "other_base", "Other Child", "", "spotify999", nil)
			assert.NoError(err)

			// Create some child playlists for a different user to ensure user isolation
			_, err = repo.Create(ctx, "other_user", tt.basePlaylistID, "Other User Child", "", "spotify888", nil)
			assert.NoError(err)

			// Execute GetByBasePlaylistID
			retrievedPlaylists, err := repo.GetByBasePlaylistID(ctx, tt.basePlaylistID, tt.userID)

			// Verify success
			assert.NoError(err)
			assert.NotNil(retrievedPlaylists)
			assert.Len(retrievedPlaylists, tt.expectedCount)

			// If we have playlists, verify they match what we created
			if tt.expectedCount > 0 {
				// Verify all retrieved playlists belong to the correct user and base playlist
				for _, playlist := range retrievedPlaylists {
					assert.Equal(tt.userID, playlist.UserID)
					assert.Equal(tt.basePlaylistID, playlist.BasePlaylistID)
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
					assert.True(playlistNames[created.Name], "Child playlist %s should be in results", created.Name)
				}
			}
		})
	}
}

func TestChildPlaylistRepositoryPocketbase_Update_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// Create initial child playlist
	initialFilterRules := &models.AudioFeatureFilters{
		Popularity: &models.RangeFilter{Min: ptrFloat64(50), Max: ptrFloat64(80)},
	}
	playlist, err := repo.Create(ctx, "user123", "base123", "Original Name", "Original description", "spotify123", initialFilterRules)
	assert.NoError(err)
	assert.NotNil(playlist)

	// Prepare update request
	newName := "Updated Name"
	newDescription := "Updated description"
	newIsActive := false
	newFilterRules := &models.AudioFeatureFilters{
		Popularity: &models.RangeFilter{Min: ptrFloat64(70), Max: ptrFloat64(100)},
		Duration: &models.RangeFilter{Min: ptrFloat64(200000), Max: ptrFloat64(350000)},
	}

	updateReq := &models.UpdateChildPlaylistRequest{
		Name:        &newName,
		Description: &newDescription,
		IsActive:    &newIsActive,
		FilterRules: newFilterRules,
	}

	// Execute update
	updatedPlaylist, err := repo.Update(ctx, playlist.ID, "user123", updateReq)
	assert.NoError(err)
	assert.NotNil(updatedPlaylist)

	// Verify updated fields
	assert.Equal(newName, updatedPlaylist.Name)
	assert.Equal(newDescription, updatedPlaylist.Description)
	assert.Equal(newIsActive, updatedPlaylist.IsActive)
	assert.NotNil(updatedPlaylist.FilterRules)
	assert.Equal(newFilterRules, updatedPlaylist.FilterRules)

	// Verify unchanged fields
	assert.Equal(playlist.ID, updatedPlaylist.ID)
	assert.Equal(playlist.UserID, updatedPlaylist.UserID)
	assert.Equal(playlist.BasePlaylistID, updatedPlaylist.BasePlaylistID)
	assert.Equal(playlist.SpotifyPlaylistID, updatedPlaylist.SpotifyPlaylistID)
}

func TestChildPlaylistRepositoryPocketbase_Update_PartialUpdate(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// Create initial child playlist
	playlist, err := repo.Create(ctx, "user123", "base123", "Original Name", "Original description", "spotify123", nil)
	assert.NoError(err)
	assert.NotNil(playlist)

	// Update only the name
	newName := "Updated Name Only"
	updateReq := &models.UpdateChildPlaylistRequest{
		Name: &newName,
	}

	// Execute update
	updatedPlaylist, err := repo.Update(ctx, playlist.ID, "user123", updateReq)
	assert.NoError(err)
	assert.NotNil(updatedPlaylist)

	// Verify only name was updated
	assert.Equal(newName, updatedPlaylist.Name)
	assert.Equal(playlist.Description, updatedPlaylist.Description) // Unchanged
	assert.Equal(playlist.IsActive, updatedPlaylist.IsActive)       // Unchanged
}

func TestChildPlaylistRepositoryPocketbase_Update_UnauthorizedError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// Create a playlist owned by user123
	playlist, err := repo.Create(ctx, "user123", "base123", "Test Playlist", "", "spotify123", nil)
	assert.NoError(err)
	assert.NotNil(playlist)

	// Try to update with different user ID
	newName := "Hacked Name"
	updateReq := &models.UpdateChildPlaylistRequest{
		Name: &newName,
	}

	updatedPlaylist, err := repo.Update(ctx, playlist.ID, "user456", updateReq)

	// Verify unauthorized error
	assert.Error(err)
	assert.Nil(updatedPlaylist)
	assert.ErrorIs(err, repositories.ErrUnauthorized)
}

func TestChildPlaylistRepositoryPocketbase_Update_NotFoundError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupChildPlaylistCollection(t, app)
	repo := NewChildPlaylistRepositoryPocketbase(app)

	ctx := context.Background()

	// Try to update non-existent playlist
	newName := "New Name"
	updateReq := &models.UpdateChildPlaylistRequest{
		Name: &newName,
	}

	updatedPlaylist, err := repo.Update(ctx, "nonexistent123", "user123", updateReq)

	// Verify error
	assert.Error(err)
	assert.Nil(updatedPlaylist)
	assert.ErrorIs(err, repositories.ErrChildPlaylistNotFound)
}

// ptrFloat64 returns a pointer to a float64 value
func ptrFloat64(f float64) *float64 {
	return &f
}

// findChildPlaylistInDB is a helper function to verify a child playlist exists in the database
func findChildPlaylistInDB(t *testing.T, app *pocketbase.PocketBase, id string) (*models.ChildPlaylist, error) {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId(string(CollectionChildPlaylist))
	if err != nil {
		return nil, err
	}

	record, err := app.FindRecordById(collection, id)
	if err != nil {
		return nil, err
	}

	return recordToChildPlaylist(record), nil
}
