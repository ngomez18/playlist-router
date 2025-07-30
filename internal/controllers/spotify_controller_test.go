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

func TestSpotifyController_GetUserPlaylists_AuthenticationError(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "user not found in context",
			setupContext: func(req *http.Request) *http.Request {
				// Return request without user context
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not found in context\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyApiService := mocks.NewMockSpotifyAPIServicer(ctrl)
			controller := NewSpotifyController(mockSpotifyApiService)

			// Service should not be called when auth fails
			mockSpotifyApiService.EXPECT().
				GetUserPlaylists(gomock.Any(), gomock.Any()).
				Times(0)

			// Create request with specific context setup
			req := httptest.NewRequest(http.MethodGet, "/api/spotify/playlists", nil)
			req = tt.setupContext(req)
			w := httptest.NewRecorder()

			// Execute
			controller.GetUserPlaylists(w, req)

			// Assert response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal(tt.expectedBody, w.Body.String())
		})
	}
}

func TestSpotifyController_GetUserPlaylists_ServiceError(t *testing.T) {
	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "spotify integration not found",
			serviceError:   errors.New("failed to get spotify integration: spotify integration not found"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "unable to retrieve spotify playlists\n",
		},
		{
			name:           "spotify API unauthorized",
			serviceError:   errors.New("failed to fetch user playlists: spotify API unauthorized (401)"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "unable to retrieve spotify playlists\n",
		},
		{
			name:           "spotify API network error",
			serviceError:   errors.New("failed to fetch user playlists: network timeout"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "unable to retrieve spotify playlists\n",
		},
		{
			name:           "database connection error",
			serviceError:   errors.New("failed to get spotify integration: database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "unable to retrieve spotify playlists\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyApiService := mocks.NewMockSpotifyAPIServicer(ctrl)
			controller := NewSpotifyController(mockSpotifyApiService)

			// Mock service error
			mockSpotifyApiService.EXPECT().
				GetUserPlaylists(gomock.Any(), "test_user_123").
				Return(nil, tt.serviceError).
				Times(1)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/spotify/playlists", nil)
			req = addUserToSpotifyContext(req)
			w := httptest.NewRecorder()

			// Execute
			controller.GetUserPlaylists(w, req)

			// Assert response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal(tt.expectedBody, w.Body.String())
		})
	}
}

// Helper function to add user to request context for Spotify controller tests
func addUserToSpotifyContext(req *http.Request) *http.Request {
	user := &models.User{ID: "test_user_123", Email: "test@example.com", Name: "Test User"}
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	return req.WithContext(ctx)
}
