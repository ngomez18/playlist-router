package requestcontext

import (
	"context"
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/require"
)

func TestContextWithUser(t *testing.T) {
	assert := require.New(t)
	user := &models.User{ID: "123", Email: "test@example.com"}
	ctx := ContextWithUser(context.Background(), user)

	retrievedUser, ok := ctx.Value(UserContextKey).(*models.User)
	assert.True(ok)
	assert.Equal(user, retrievedUser)
}

func TestGetUserFromContext(t *testing.T) {
	assert := require.New(t)
	user := &models.User{ID: "123", Email: "test@example.com"}

	testCases := []struct {
		name         string
		ctx          context.Context
		expectedUser *models.User
		expectedOk   bool
	}{
		{
			name:         "user exists in context",
			ctx:          context.WithValue(context.Background(), UserContextKey, user),
			expectedUser: user,
			expectedOk:   true,
		},
		{
			name:         "user does not exist in context",
			ctx:          context.Background(),
			expectedUser: nil,
			expectedOk:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retrievedUser, ok := GetUserFromContext(tc.ctx)
			assert.Equal(tc.expectedOk, ok)
			assert.Equal(tc.expectedUser, retrievedUser)
		})
	}
}

func TestContextWithSpotifyAuth(t *testing.T) {
	assert := require.New(t)
	spotifyAuth := &models.SpotifyIntegration{AccessToken: "abc", RefreshToken: "def"}
	ctx := ContextWithSpotifyAuth(context.Background(), spotifyAuth)

	retrievedAuth, ok := ctx.Value(SpotifyAuthContextKey).(*models.SpotifyIntegration)
	assert.True(ok)
	assert.Equal(spotifyAuth, retrievedAuth)
}

func TestGetSpotifyAuthFromContext(t *testing.T) {
	assert := require.New(t)
	spotifyAuth := &models.SpotifyIntegration{AccessToken: "abc", RefreshToken: "def"}

	testCases := []struct {
		name                string
		ctx                 context.Context
		expectedSpotifyAuth *models.SpotifyIntegration
		expectedOk          bool
	}{
		{
			name:                "spotify auth exists in context",
			ctx:                 context.WithValue(context.Background(), SpotifyAuthContextKey, spotifyAuth),
			expectedSpotifyAuth: spotifyAuth,
			expectedOk:          true,
		},
		{
			name:                "spotify auth does not exist in context",
			ctx:                 context.Background(),
			expectedSpotifyAuth: nil,
			expectedOk:          false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retrievedAuth, ok := GetSpotifyAuthFromContext(tc.ctx)
			assert.Equal(tc.expectedOk, ok)
			assert.Equal(tc.expectedSpotifyAuth, retrievedAuth)
		})
	}
}

func TestGetUserAndSpotifyAuthFromContext(t *testing.T) {
	assert := require.New(t)
	user := &models.User{ID: "123"}
	spotifyAuth := &models.SpotifyIntegration{AccessToken: "abc"}

	testCases := []struct {
		name                string
		ctx                 context.Context
		expectedUser        *models.User
		expectedSpotifyAuth *models.SpotifyIntegration
		expectedOk          bool
	}{
		{
			name:                "both user and spotify auth exist",
			ctx:                 context.WithValue(context.WithValue(context.Background(), UserContextKey, user), SpotifyAuthContextKey, spotifyAuth),
			expectedUser:        user,
			expectedSpotifyAuth: spotifyAuth,
			expectedOk:          true,
		},
		{
			name:                "only user exists",
			ctx:                 context.WithValue(context.Background(), UserContextKey, user),
			expectedUser:        user,
			expectedSpotifyAuth: nil,
			expectedOk:          false,
		},
		{
			name:                "only spotify auth exists",
			ctx:                 context.WithValue(context.Background(), SpotifyAuthContextKey, spotifyAuth),
			expectedUser:        nil,
			expectedSpotifyAuth: spotifyAuth,
			expectedOk:          false,
		},
		{
			name:                "neither user nor spotify auth exist",
			ctx:                 context.Background(),
			expectedUser:        nil,
			expectedSpotifyAuth: nil,
			expectedOk:          false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retrievedUser, retrievedAuth, ok := GetUserAndSpotifyAuthFromContext(tc.ctx)
			assert.Equal(tc.expectedOk, ok)
			assert.Equal(tc.expectedUser, retrievedUser)
			assert.Equal(tc.expectedSpotifyAuth, retrievedAuth)
		})
	}
}
