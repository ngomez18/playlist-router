package pb

import (
	"context"
	"testing"
	"time"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

func TestSpotifyIntegrationRepositoryPocketbase_CreateOrUpdate_Success(t *testing.T) {
	tests := []struct {
		name         string
		userEmail    string
		integration  *models.SpotifyIntegration
		expectCreate bool // true if creating new, false if updating existing
	}{
		{
			name:      "create new integration",
			userEmail: "user1@example.com",
			integration: &models.SpotifyIntegration{
				SpotifyID:    "spotify_user_123",
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_123",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(1 * time.Hour),
				Scope:        "user-read-email playlist-modify-private",
				DisplayName:  "Test User",
			},
			expectCreate: true,
		},
		{
			name:      "update existing integration",
			userEmail: "user2@example.com",
			integration: &models.SpotifyIntegration{
				SpotifyID:    "spotify_user_456",
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(2 * time.Hour),
				Scope:        "user-read-email playlist-modify-private playlist-modify-public",
				DisplayName:  "Updated User",
			},
			expectCreate: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupSpotifyIntegrationsCollection(t, app)
			repo := NewSpotifyIntegrationRepositoryPocketbase(app)

			// Create test user
			userID := CreateTestUser(t, app, tt.userEmail, "Test User")

			// If testing update, create an existing integration first
			if !tt.expectCreate {
				existingIntegration := &models.SpotifyIntegration{
					UserID: userID,
					SpotifyID:    "old_spotify_id",
					AccessToken:  "old_access_token",
					RefreshToken: "old_refresh_token",
					TokenType:    "Bearer",
					ExpiresAt:    time.Now().Add(30 * time.Minute),
					Scope:        "user-read-email",
					DisplayName:  "Old User",
				}
				_, err := createIntegrationInDB(t, app, existingIntegration)
				assert.NoError(err)
			}

			// Execute test
			ctx := context.Background()
			result, err := repo.CreateOrUpdate(ctx, userID, tt.integration)

			// Verify success
			assert.NoError(err)
			assert.NotNil(result)
			assert.NotEmpty(result.ID)
			assert.Equal(userID, result.UserID)
			assert.Equal(tt.integration.SpotifyID, result.SpotifyID)
			assert.Equal(tt.integration.AccessToken, result.AccessToken)
			assert.Equal(tt.integration.RefreshToken, result.RefreshToken)
			assert.Equal(tt.integration.TokenType, result.TokenType)
			assert.Equal(tt.integration.Scope, result.Scope)
			assert.Equal(tt.integration.DisplayName, result.DisplayName)

			// Verify timestamps
			assert.WithinDuration(tt.integration.ExpiresAt, result.ExpiresAt, 1*time.Second)
			assert.NotZero(result.Created)
			assert.NotZero(result.Updated)

			// Verify the integration was actually saved
			savedIntegration, err := findIntegrationInDB(t, app, result.ID)
			assert.NoError(err)
			assert.Equal(result.ID, savedIntegration.ID)
		})
	}
}

func TestSpotifyIntegrationRepositoryPocketbase_CreateOrUpdate_DatabaseErrors(t *testing.T) {
	t.Run("collection not found", func(t *testing.T) {
		assert := require.New(t)

		// Setup test environment without creating the collection
		app := NewTestApp(t)
		repo := NewSpotifyIntegrationRepositoryPocketbase(app)

		integration := &models.SpotifyIntegration{
			SpotifyID:   "spotify_123",
			AccessToken: "token_123",
		}

		// Execute test
		ctx := context.Background()
		result, err := repo.CreateOrUpdate(ctx, "user_123", integration)

		// Verify error
		assert.Error(err)
		assert.Nil(result)
		assert.ErrorIs(err, repositories.ErrCollectionNotFound)
	})
}

func TestSpotifyIntegrationRepositoryPocketbase_GetByUserID_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	userID := CreateTestUser(t, app, "get@test.com", "GetByUserID Test User")
	integration := &models.SpotifyIntegration{
		UserID:       userID,
		SpotifyID:    "spotify_user_123",
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "user-read-email",
		DisplayName:  "Test User",
	}

	ctx := context.Background()

	// Create integration first
	createdIntegration, err := createIntegrationInDB(t, app, integration)
	assert.NoError(err)

	// Execute test
	result, err := repo.GetByUserID(ctx, userID)

	// Verify success
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(createdIntegration.ID, result.ID)
	assert.Equal(userID, result.UserID)
	assert.Equal(integration.SpotifyID, result.SpotifyID)
	assert.Equal(integration.AccessToken, result.AccessToken)
}

func TestSpotifyIntegrationRepositoryPocketbase_GetByUserID_NotFound(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	ctx := context.Background()

	// Execute test with non-existent user
	result, err := repo.GetByUserID(ctx, "nonexistent_user")

	// Verify error
	assert.Error(err)
	assert.Nil(result)
	assert.ErrorIs(err, repositories.ErrSpotifyIntegrationNotFound)
}

func TestSpotifyIntegrationRepositoryPocketbase_GetBySpotifyID_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	userID := CreateTestUser(t, app, "get@test.com", "Get By SpotifyID Test User")
	spotifyID := "spotify_user_123"
	integration := &models.SpotifyIntegration{
		UserID:       userID,
		SpotifyID:    spotifyID,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "user-read-email",
		DisplayName:  "Test User",
	}

	ctx := context.Background()

	// Create integration first
	createdIntegration, err := createIntegrationInDB(t, app, integration)
	assert.NoError(err)

	// Execute test
	result, err := repo.GetBySpotifyID(ctx, spotifyID)

	// Verify success
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(createdIntegration.ID, result.ID)
	assert.Equal(userID, result.UserID)
	assert.Equal(spotifyID, result.SpotifyID)
}

func TestSpotifyIntegrationRepositoryPocketbase_GetBySpotifyID_NotFound(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	ctx := context.Background()

	// Execute test with non-existent Spotify ID
	result, err := repo.GetBySpotifyID(ctx, "nonexistent_spotify_id")

	// Verify error
	assert.Error(err)
	assert.Nil(result)
	assert.ErrorIs(err, repositories.ErrSpotifyIntegrationNotFound)
}

func TestSpotifyIntegrationRepositoryPocketbase_UpdateTokens_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	userID := CreateTestUser(t, app, "updatetokens@test.com", "Update Tokens Test User")
	integration := &models.SpotifyIntegration{
		UserID:       userID,
		SpotifyID:    "spotify_user_123",
		AccessToken:  "old_access_token",
		RefreshToken: "old_refresh_token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		Scope:        "user-read-email",
		DisplayName:  "Test User",
	}

	ctx := context.Background()

	// Create integration first
	createdIntegration, err := createIntegrationInDB(t, app, integration)
	assert.NoError(err)

	// Prepare new tokens
	newTokens := &models.SpotifyTokenResponse{
		AccessToken:  "new_access_token",
		TokenType:    "Bearer",
		Scope:        "user-read-email playlist-modify-private",
		ExpiresIn:    7200, // 2 hours
		RefreshToken: "new_refresh_token",
	}

	// Execute test
	err = repo.UpdateTokens(ctx, createdIntegration.ID, newTokens)

	// Verify success
	assert.NoError(err)

	// Verify tokens were updated
	updatedIntegration, err := findIntegrationInDB(t, app, createdIntegration.ID)
	assert.NoError(err)
	assert.Equal("new_access_token", updatedIntegration.AccessToken)
	assert.Equal("new_refresh_token", updatedIntegration.RefreshToken)
	assert.Equal("Bearer", updatedIntegration.TokenType)
	assert.Equal("user-read-email playlist-modify-private", updatedIntegration.Scope)

	// Verify expiration time was updated (should be ~2 hours from now)
	expectedExpiration := time.Now().Add(2 * time.Hour)
	assert.WithinDuration(expectedExpiration, updatedIntegration.ExpiresAt, 10*time.Second)
}

func TestSpotifyIntegrationRepositoryPocketbase_UpdateTokens_WithoutRefreshToken(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	userID := CreateTestUser(t, app, "updatetokens@test.com", "Update Tokens Test User")
	originalRefreshToken := "original_refresh_token"
	integration := &models.SpotifyIntegration{
		UserID:       userID,
		SpotifyID:    "spotify_user_123",
		AccessToken:  "old_access_token",
		RefreshToken: originalRefreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		Scope:        "user-read-email",
		DisplayName:  "Test User",
	}

	ctx := context.Background()

	// Create integration first
	createdIntegration, err := createIntegrationInDB(t, app, integration)
	assert.NoError(err)

	// Prepare new tokens without refresh token
	newTokens := &models.SpotifyTokenResponse{
		AccessToken: "new_access_token",
		TokenType:   "Bearer",
		Scope:       "user-read-email playlist-modify-private",
		ExpiresIn:   3600,
		// RefreshToken is empty - should preserve existing one
	}

	// Execute test
	err = repo.UpdateTokens(ctx, createdIntegration.ID, newTokens)

	// Verify success
	assert.NoError(err)

	// Verify tokens were updated but refresh token preserved
	updatedIntegration, err := findIntegrationInDB(t, app, createdIntegration.ID)
	assert.NoError(err)
	assert.Equal("new_access_token", updatedIntegration.AccessToken)
	assert.Equal(originalRefreshToken, updatedIntegration.RefreshToken) // Should be preserved
	assert.Equal("Bearer", updatedIntegration.TokenType)
	assert.Equal("user-read-email playlist-modify-private", updatedIntegration.Scope)
}

func TestSpotifyIntegrationRepositoryPocketbase_UpdateTokens_NotFound(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	newTokens := &models.SpotifyTokenResponse{
		AccessToken: "new_access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	ctx := context.Background()

	// Execute test with non-existent integration ID
	err := repo.UpdateTokens(ctx, "nonexistent_id", newTokens)

	// Verify error
	assert.Error(err)
	assert.ErrorIs(err, repositories.ErrSpotifyIntegrationNotFound)
}

func TestSpotifyIntegrationRepositoryPocketbase_Delete_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	userID := CreateTestUser(t, app, "delete@test.com", "Delete Spotify Integration User")
	integration := &models.SpotifyIntegration{
		UserID:       userID,
		SpotifyID:    "spotify_user_123",
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "user-read-email",
		DisplayName:  "Test User",
	}

	ctx := context.Background()

	// Create integration first
	integration, err := createIntegrationInDB(t, app, integration)
	assert.NoError(err)

	// Verify integration exists
	foundIntegration, err := findIntegrationInDB(t, app, integration.ID)
	assert.NoError(err)
	assert.NotNil(foundIntegration)

	// Execute delete
	err = repo.Delete(ctx, userID)
	assert.NoError(err)

	// Verify integration no longer exists
	foundIntegration, err = findIntegrationInDB(t, app, integration.ID)
	assert.Nil(foundIntegration)
	assert.Error(err)
}

func TestSpotifyIntegrationRepositoryPocketbase_Delete_NotFound(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSpotifyIntegrationsCollection(t, app)
	repo := NewSpotifyIntegrationRepositoryPocketbase(app)

	ctx := context.Background()

	// Execute delete with non-existent user
	err := repo.Delete(ctx, "nonexistent_user")

	// Verify error
	assert.Error(err)
	assert.ErrorIs(err, repositories.ErrSpotifyIntegrationNotFound)
}

func TestSpotifyIntegrationRepositoryPocketbase_Delete_DatabaseErrors(t *testing.T) {
	t.Run("collection not found", func(t *testing.T) {
		assert := require.New(t)

		// Setup test environment without creating the collection
		app := NewTestApp(t)
		repo := NewSpotifyIntegrationRepositoryPocketbase(app)

		ctx := context.Background()

		// Execute delete
		err := repo.Delete(ctx, "user_123")

		// Verify error
		assert.Error(err)
		assert.ErrorIs(err, repositories.ErrCollectionNotFound)
	})
}

// findIntegrationInDB is a helper function to verify an integration exists in the database
func findIntegrationInDB(t *testing.T, app *pocketbase.PocketBase, id string) (*models.SpotifyIntegration, error) {
	t.Helper()
	assert := require.New(t)

	collection, err := app.FindCollectionByNameOrId(string(CollectionSpotifyIntegration))
	assert.NoError(err)

	record, err := app.FindRecordById(collection, id)
	if err != nil {
		return nil, err
	}

	return recordToSpotifyIntegration(record), nil
}

// createIntegrationInDB is a helper function to insert an integration in the database
func createIntegrationInDB(t *testing.T, app *pocketbase.PocketBase, integration *models.SpotifyIntegration) (*models.SpotifyIntegration, error) {
	t.Helper()
	assert := require.New(t)

	collection, err := app.FindCollectionByNameOrId(string(CollectionSpotifyIntegration))
	assert.NoError(err)

	record := core.NewRecord(collection)
	record.Set("user", integration.UserID)
	record.Set("spotify_id", integration.SpotifyID)
	record.Set("access_token", integration.AccessToken)
	record.Set("refresh_token", integration.RefreshToken)
	record.Set("token_type", integration.TokenType)
	record.Set("expires_at", integration.ExpiresAt)
	record.Set("scope", integration.Scope)
	record.Set("display_name", integration.DisplayName)

	err = app.Save(record)
	if err != nil {
		return nil, err
	}

	return recordToSpotifyIntegration(record), nil
}
