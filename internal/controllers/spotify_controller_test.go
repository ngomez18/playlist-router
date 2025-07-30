package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewSpotifyController(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyApiService := mocks.NewMockSpotifyAPIServicer(ctrl)
	controller := NewSpotifyController(mockSpotifyApiService)

	assert.NotNil(controller)
	assert.Equal(mockSpotifyApiService, controller.spotifyApiService)
}

func TestSpotifyController_GetUserPlaylists_Success(t *testing.T) {
	tests := []struct {
		name              string
		serviceResult     []*models.SpotifyPlaylist
		expectedStatus    int
		expectedPlaylists []*models.SpotifyPlaylist
	}{
		{
			name: "successful fetch with playlists",
			serviceResult: []*models.SpotifyPlaylist{
				{ID: "playlist1", Name: "My Rock Playlist", Tracks: 25},
				{ID: "playlist2", Name: "Jazz Favorites", Tracks: 18},
			},
			expectedStatus: http.StatusOK,
			expectedPlaylists: []*models.SpotifyPlaylist{
				{ID: "playlist1", Name: "My Rock Playlist", Tracks: 25},
				{ID: "playlist2", Name: "Jazz Favorites", Tracks: 18},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyApiService := mocks.NewMockSpotifyAPIServicer(ctrl)
			controller := NewSpotifyController(mockSpotifyApiService)

			// Mock the service call
			mockSpotifyApiService.EXPECT().
				GetUserPlaylists(gomock.Any(), "test_user_123").
				Return(tt.serviceResult, nil).
				Times(1)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/spotify/playlists", nil)
			req = addUserToSpotifyContext(req)
			w := httptest.NewRecorder()

			// Execute
			controller.GetUserPlaylists(w, req)

			// Assert response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			// Parse and verify response body
			var responseBody []*models.SpotifyPlaylist
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(err)

			assert.Equal(len(tt.expectedPlaylists), len(responseBody))
			for i, expected := range tt.expectedPlaylists {
				assert.Equal(expected.ID, responseBody[i].ID)
				assert.Equal(expected.Name, responseBody[i].Name)
				assert.Equal(expected.Tracks, responseBody[i].Tracks)
			}
		})
	}
}

func TestSpotifyController_GetUserPlaylists_Errors(t *testing.T) {
	tests := []struct {
		name               string
		serviceError       error
		noUserInContext    bool
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "service error",
			serviceError:       errors.New("some service error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to retrieve spotify playlists",
		},
		{
			name:               "no user in context",
			noUserInContext:    true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "user not found in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyApiService := mocks.NewMockSpotifyAPIServicer(ctrl)
			controller := NewSpotifyController(mockSpotifyApiService)

			if tt.serviceError != nil {
				mockSpotifyApiService.EXPECT().
					GetUserPlaylists(gomock.Any(), "test_user_123").
					Return(nil, tt.serviceError).
					Times(1)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/spotify/playlists", nil)
			if !tt.noUserInContext {
				req = addUserToSpotifyContext(req)
			}

			w := httptest.NewRecorder()
			controller.GetUserPlaylists(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

// Helper function to add user to request context for Spotify controller tests
func addUserToSpotifyContext(req *http.Request) *http.Request {
	user := &models.User{ID: "test_user_123", Email: "test@example.com", Name: "Test User"}
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	return req.WithContext(ctx)
}
