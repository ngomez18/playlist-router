package spotifyclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/clients/mocks"
	"github.com/ngomez18/playlist-router/internal/config"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/require"
)

func TestSpotifyClient_GetPlaylist_Success(t *testing.T) {
	tests := []struct {
		name             string
		playlistId       string
		responseBody     *SpotifyPlaylist
		expectedPlaylist *SpotifyPlaylist
		accessToken      string
	}{
		{
			name:       "successful playlist fetch",
			playlistId: "playlist123",
			responseBody: &SpotifyPlaylist{
				ID:            "playlist123",
				Name:          "Test Playlist",
				URI:           "spotify:playlist:playlist123",
				Public:        true,
				Collaborative: false,
				Description:   "A test playlist",
				Href:          "https://api.spotify.com/v1/playlists/playlist123",
				SnapshotID:    "snapshot123",
				Images: []*SpotifyPlaylistImage{
					{URL: "https://image.jpg", Height: 640, Width: 640},
				},
			},
			expectedPlaylist: &SpotifyPlaylist{
				ID:            "playlist123",
				Name:          "Test Playlist",
				URI:           "spotify:playlist:playlist123",
				Public:        true,
				Collaborative: false,
				Description:   "A test playlist",
				Href:          "https://api.spotify.com/v1/playlists/playlist123",
				SnapshotID:    "snapshot123",
				Images: []*SpotifyPlaylistImage{
					{URL: "https://image.jpg", Height: 640, Width: 640},
				},
			},
			accessToken: "valid_token",
		},
		{
			name:       "successful playlist fetch without images",
			playlistId: "playlist456",
			responseBody: &SpotifyPlaylist{
				ID:            "playlist456",
				Name:          "Simple Playlist",
				URI:           "spotify:playlist:playlist456",
				Public:        false,
				Collaborative: true,
				Description:   "",
				Href:          "https://api.spotify.com/v1/playlists/playlist456",
				SnapshotID:    "snapshot456",
				Images:        []*SpotifyPlaylistImage{},
			},
			expectedPlaylist: &SpotifyPlaylist{
				ID:            "playlist456",
				Name:          "Simple Playlist",
				URI:           "spotify:playlist:playlist456",
				Public:        false,
				Collaborative: true,
				Description:   "",
				Href:          "https://api.spotify.com/v1/playlists/playlist456",
				SnapshotID:    "snapshot456",
				Images:        []*SpotifyPlaylistImage{},
			},
			accessToken: "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Create response body
			bodyBytes, _ := json.Marshal(tt.responseBody)
			responseBody := io.NopCloser(bytes.NewReader(bodyBytes))

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       responseBody,
			}

			mockHTTPClient.EXPECT().
				Do(gomock.Any()).
				DoAndReturn(func(req *http.Request) (*http.Response, error) {
					// Validate request
					assert.Equal("GET", req.Method)
					assert.Equal("https://api.spotify.com/v1/playlists/"+tt.playlistId, req.URL.String())
					assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))

					return resp, nil
				}).
				Times(1)

			ctx := contextWithToken(tt.accessToken)
			result, err := client.GetPlaylist(ctx, tt.playlistId)

			assert.NoError(err)
			assert.Equal(tt.expectedPlaylist, result)
		})
	}
}

func TestSpotifyClient_GetPlaylist_Errors(t *testing.T) {
	tests := []struct {
		name           string
		playlistId     string
		responseStatus int
		responseError  error
		accessToken    string
	}{
		{
			name:           "playlist not found",
			playlistId:     "nonexistent",
			responseStatus: http.StatusNotFound,
			accessToken:    "valid_token",
		},
		{
			name:          "http client error",
			playlistId:    "playlist123",
			responseError: errors.New("connection timeout"),
			accessToken:   "valid_token",
		},
		{
			name:           "unauthorized error",
			playlistId:     "playlist123",
			accessToken:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			} 
			if tt.responseStatus > 0 {
				responseBody := io.NopCloser(strings.NewReader(`{"error":"test_error"}`))
				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(resp, nil).
					Times(1)
			}

			ctx := contextWithToken(tt.accessToken)
			result, err := client.GetPlaylist(ctx, tt.playlistId)

			assert.Error(err)
			assert.Nil(result)
		})
	}
}

func TestSpotifyClient_GetUserPlaylists_Success(t *testing.T) {
	tests := []struct {
		name             string
		limit            int
		offset           int
		responseBody     *SpotifyPlaylistResponse
		expectedResponse *SpotifyPlaylistResponse
		accessToken      string
	}{
		{
			name:   "successful playlist fetch with results",
			limit:  50,
			offset: 0,
			responseBody: &SpotifyPlaylistResponse{
				Total: 2,
				Items: []*SpotifyPlaylist{
					{
						ID:          "playlist1",
						Name:        "My Playlist 1",
						URI:         "spotify:playlist:playlist1",
						Public:      true,
						Description: "Test playlist 1",
						Href:        "https://api.spotify.com/v1/playlists/playlist1",
						SnapshotID:  "snapshot1",
						Images: []*SpotifyPlaylistImage{
							{URL: "https://image1.jpg", Height: 640, Width: 640},
						},
					},
					{
						ID:          "playlist2",
						Name:        "My Playlist 2",
						URI:         "spotify:playlist:playlist2",
						Public:      false,
						Description: "Test playlist 2",
						Href:        "https://api.spotify.com/v1/playlists/playlist2",
						SnapshotID:  "snapshot2",
					},
				},
			},
			expectedResponse: &SpotifyPlaylistResponse{
				Total: 2,
				Items: []*SpotifyPlaylist{
					{
						ID:          "playlist1",
						Name:        "My Playlist 1",
						URI:         "spotify:playlist:playlist1",
						Public:      true,
						Description: "Test playlist 1",
						Href:        "https://api.spotify.com/v1/playlists/playlist1",
						SnapshotID:  "snapshot1",
						Images: []*SpotifyPlaylistImage{
							{URL: "https://image1.jpg", Height: 640, Width: 640},
						},
					},
					{
						ID:          "playlist2",
						Name:        "My Playlist 2",
						URI:         "spotify:playlist:playlist2",
						Public:      false,
						Description: "Test playlist 2",
						Href:        "https://api.spotify.com/v1/playlists/playlist2",
						SnapshotID:  "snapshot2",
					},
				},
			},
			accessToken: "valid_token",
		},
		{
			name:   "successful playlist fetch with empty response",
			limit:  20,
			offset: 100,
			responseBody: &SpotifyPlaylistResponse{
				Total: 0,
				Items: []*SpotifyPlaylist{},
			},
			expectedResponse: &SpotifyPlaylistResponse{
				Total: 0,
				Items: []*SpotifyPlaylist{},
			},
			accessToken: "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Create response body
			bodyBytes, _ := json.Marshal(tt.responseBody)
			responseBody := io.NopCloser(bytes.NewReader(bodyBytes))

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       responseBody,
			}

			mockHTTPClient.EXPECT().
				Do(gomock.Any()).
				DoAndReturn(func(req *http.Request) (*http.Response, error) {
					// Validate request
					assert.Equal("GET", req.Method)
					assert.Contains(req.URL.String(), "https://api.spotify.com/v1/me/playlists")
					assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))

					// Validate query parameters
					params := req.URL.Query()
					assert.Equal(fmt.Sprintf("%d", tt.limit), params.Get("limit"))
					assert.Equal(fmt.Sprintf("%d", tt.offset), params.Get("offset"))

					return resp, nil
				}).
				Times(1)

			ctx := contextWithToken(tt.accessToken)
			result, err := client.GetUserPlaylists(ctx, tt.limit, tt.offset)

			assert.NoError(err)
			assert.Equal(tt.expectedResponse, result)
		})
	}
}

func TestSpotifyClient_GetUserPlaylists_Errors(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseError  error
		accessToken    string
		limit          int
		offset         int
	}{
		{
			name:           "unauthorized error",
			responseStatus: http.StatusUnauthorized,
			accessToken:    "invalid_token",
			limit:          50,
			offset:         0,
		},
		{
			name:          "http client error",
			responseError: errors.New("network error"),
			accessToken:   "valid_token",
			limit:         50,
			offset:        0,
		},
		{
			name:           "missing access token",
			accessToken:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			}
			if tt.responseStatus > 0 {
				responseBody := io.NopCloser(strings.NewReader(`{"error":"test_error"}`))
				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(resp, nil).
					Times(1)
			}

			ctx := contextWithToken(tt.accessToken)
			result, err := client.GetUserPlaylists(ctx, tt.limit, tt.offset)

			assert.Error(err)
			assert.Nil(result)
		})
	}
}


func TestSpotifyClient_GetAllUserPlaylists_Success(t *testing.T) {
	tests := []struct {
		name              string
		mockResponses     []SpotifyPlaylistResponse
		expectedPlaylists []*SpotifyPlaylist
		accessToken       string
	}{
		{
			name: "single page with results",
			mockResponses: []SpotifyPlaylistResponse{
				{
					Items: []*SpotifyPlaylist{
						{ID: "playlist1", Name: "Rock Classics", Tracks: &SpotifyPlaylistTracks{Total: 25}},
						{ID: "playlist2", Name: "Jazz Favorites", Tracks: &SpotifyPlaylistTracks{Total: 18}},
					},
					Total: 2,
				},
			},
			expectedPlaylists: []*SpotifyPlaylist{
				{ID: "playlist1", Name: "Rock Classics", Tracks: &SpotifyPlaylistTracks{Total: 25}},
				{ID: "playlist2", Name: "Jazz Favorites", Tracks: &SpotifyPlaylistTracks{Total: 18}},
			},
			accessToken: "valid_token",
		},
		{
			name: "multiple pages",
			mockResponses: []SpotifyPlaylistResponse{
				{
					Items: []*SpotifyPlaylist{
						{ID: "playlist1", Name: "Page 1 Playlist 1", Tracks: &SpotifyPlaylistTracks{Total: 10}},
						{ID: "playlist2", Name: "Page 1 Playlist 2", Tracks: &SpotifyPlaylistTracks{Total: 15}},
					},
					Total: 3,
				},
				{
					Items: []*SpotifyPlaylist{
						{ID: "playlist3", Name: "Page 2 Playlist 1", Tracks: &SpotifyPlaylistTracks{Total: 20}},
					},
					Total: 3,
				},
			},
			expectedPlaylists: []*SpotifyPlaylist{
				{ID: "playlist1", Name: "Page 1 Playlist 1", Tracks: &SpotifyPlaylistTracks{Total: 10}},
				{ID: "playlist2", Name: "Page 1 Playlist 2", Tracks: &SpotifyPlaylistTracks{Total: 15}},
				{ID: "playlist3", Name: "Page 2 Playlist 1", Tracks: &SpotifyPlaylistTracks{Total: 20}},
			},
			accessToken: "valid_token",
		},
		{
			name: "empty result",
			mockResponses: []SpotifyPlaylistResponse{
				{
					Items: []*SpotifyPlaylist{},
					Total: 0,
				},
			},
			expectedPlaylists: []*SpotifyPlaylist{},
			accessToken:       "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Set up mock calls for each expected response
			for _, response := range tt.mockResponses {
				responseJSON, _ := json.Marshal(response)
				responseBody := io.NopCloser(bytes.NewReader(responseJSON))
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(resp, nil).
					Times(1)
			}

			ctx := contextWithToken(tt.accessToken)
			result, err := client.GetAllUserPlaylists(ctx)

			assert.NoError(err)
			assert.NotNil(result)
			assert.Equal(len(tt.expectedPlaylists), len(result))

			for i, expected := range tt.expectedPlaylists {
				assert.Equal(expected.ID, result[i].ID)
				assert.Equal(expected.Name, result[i].Name)
				if expected.Tracks != nil && result[i].Tracks != nil {
					assert.Equal(expected.Tracks.Total, result[i].Tracks.Total)
				}
			}
		})
	}
}

func TestSpotifyClient_GetAllUserPlaylists_Error(t *testing.T) {
	tests := []struct {
		name           string
		responseError  error
		responseStatus int
		accessToken    string
	}{
		{
			name:          "http client error",
			responseError: errors.New("network timeout"),
			accessToken:   "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			} else {
				responseBody := io.NopCloser(strings.NewReader(`{"error":"test_error"}`))
				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(resp, nil).
					Times(1)
			}

			ctx := contextWithToken(tt.accessToken)
			result, err := client.GetAllUserPlaylists(ctx)

			assert.Error(err)
			assert.Nil(result)
		})
	}
}

func TestSpotifyClient_CreatePlaylist_Success(t *testing.T) {
	tests := []struct {
		name             string
		userId           string
		playlistName     string
		description      string
		public           bool
		responseBody     *SpotifyPlaylist
		expectedPlaylist *SpotifyPlaylist
		accessToken      string
	}{
		{
			name:         "successful public playlist creation",
			userId:       "user123",
			playlistName: "My New Playlist",
			description:  "A brand new playlist",
			public:       true,
			responseBody: &SpotifyPlaylist{
				ID:            "new_playlist_123",
				Name:          "My New Playlist",
				URI:           "spotify:playlist:new_playlist_123",
				Public:        true,
				Collaborative: false,
				Description:   "A brand new playlist",
				Href:          "https://api.spotify.com/v1/playlists/new_playlist_123",
				SnapshotID:    "new_snapshot",
				Images:        []*SpotifyPlaylistImage{},
			},
			expectedPlaylist: &SpotifyPlaylist{
				ID:            "new_playlist_123",
				Name:          "My New Playlist",
				URI:           "spotify:playlist:new_playlist_123",
				Public:        true,
				Collaborative: false,
				Description:   "A brand new playlist",
				Href:          "https://api.spotify.com/v1/playlists/new_playlist_123",
				SnapshotID:    "new_snapshot",
				Images:        []*SpotifyPlaylistImage{},
			},
			accessToken: "valid_token",
		},
		{
			name:         "successful private playlist creation",
			userId:       "user456",
			playlistName: "Private Playlist",
			description:  "",
			public:       false,
			responseBody: &SpotifyPlaylist{
				ID:            "private_playlist_456",
				Name:          "Private Playlist",
				URI:           "spotify:playlist:private_playlist_456",
				Public:        false,
				Collaborative: false,
				Description:   "",
				Href:          "https://api.spotify.com/v1/playlists/private_playlist_456",
				SnapshotID:    "private_snapshot",
				Images:        []*SpotifyPlaylistImage{},
			},
			expectedPlaylist: &SpotifyPlaylist{
				ID:            "private_playlist_456",
				Name:          "Private Playlist",
				URI:           "spotify:playlist:private_playlist_456",
				Public:        false,
				Collaborative: false,
				Description:   "",
				Href:          "https://api.spotify.com/v1/playlists/private_playlist_456",
				SnapshotID:    "private_snapshot",
				Images:        []*SpotifyPlaylistImage{},
			},
			accessToken: "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Create response body
			bodyBytes, _ := json.Marshal(tt.responseBody)
			responseBody := io.NopCloser(bytes.NewReader(bodyBytes))

			resp := &http.Response{
				StatusCode: http.StatusCreated,
				Body:       responseBody,
			}

			mockHTTPClient.EXPECT().
				Do(gomock.Any()).
				DoAndReturn(func(req *http.Request) (*http.Response, error) {
					// Validate request
					assert.Equal("POST", req.Method)
					assert.Equal("https://api.spotify.com/v1/users/"+tt.userId+"/playlists", req.URL.String())
					assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))
					assert.Equal("application/json", req.Header.Get("Content-Type"))

					// Validate request body
					body, _ := io.ReadAll(req.Body)
					var requestBody SpotifyPlaylistRequest
					err := json.Unmarshal(body, &requestBody)
					assert.NoError(err)
					assert.NotNil(requestBody.Name)
					assert.Equal(tt.playlistName, *requestBody.Name)
					assert.NotNil(requestBody.Description)
					assert.Equal(tt.description, *requestBody.Description)
					assert.NotNil(requestBody.Public)
					assert.Equal(tt.public, *requestBody.Public)

					return resp, nil
				}).
				Times(1)

			ctx := context.Background()
			result, err := client.CreatePlaylist(ctx, tt.accessToken, tt.userId, tt.playlistName, tt.description, tt.public)

			assert.NoError(err)
			assert.Equal(tt.expectedPlaylist, result)
		})
	}
}

func TestSpotifyClient_CreatePlaylist_Errors(t *testing.T) {
	tests := []struct {
		name           string
		userId         string
		playlistName   string
		description    string
		public         bool
		responseStatus int
		responseError  error
		accessToken    string
	}{
		{
			name:           "unauthorized error",
			userId:         "user123",
			playlistName:   "Test Playlist",
			description:    "Test description",
			public:         true,
			responseStatus: http.StatusUnauthorized,
			accessToken:    "invalid_token",
		},
		{
			name:           "forbidden error",
			userId:         "other_user",
			playlistName:   "Test Playlist",
			description:    "Test description",
			public:         true,
			responseStatus: http.StatusForbidden,
			accessToken:    "valid_token",
		},
		{
			name:           "bad request error",
			userId:         "user123",
			playlistName:   "",
			description:    "Test description",
			public:         true,
			responseStatus: http.StatusBadRequest,
			accessToken:    "valid_token",
		},
		{
			name:          "http client error",
			userId:        "user123",
			playlistName:  "Test Playlist",
			description:   "Test description",
			public:        true,
			responseError: errors.New("network timeout"),
			accessToken:   "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			} else {
				responseBody := io.NopCloser(strings.NewReader(`{"error":"test_error"}`))
				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(resp, nil).
					Times(1)
			}

			ctx := context.Background()
			result, err := client.CreatePlaylist(ctx, tt.accessToken, tt.userId, tt.playlistName, tt.description, tt.public)

			assert.Error(err)
			assert.Nil(result)
		})
	}
}

func TestSpotifyClient_DeletePlaylist_Success(t *testing.T) {
	tests := []struct {
		name        string
		userId      string
		playlistId  string
		accessToken string
	}{
		{
			name:        "successful playlist deletion",
			userId:      "user123",
			playlistId:  "playlist456",
			accessToken: "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Create successful response
			responseBody := io.NopCloser(strings.NewReader(""))
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       responseBody,
			}

			mockHTTPClient.EXPECT().
				Do(gomock.Any()).
				DoAndReturn(func(req *http.Request) (*http.Response, error) {
					// Validate request
					assert.Equal("DELETE", req.Method)
					assert.Equal("https://api.spotify.com/v1/playlists/"+tt.playlistId+"/followers", req.URL.String())
					assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))

					return resp, nil
				}).
				Times(1)

			ctx := context.Background()
			err := client.DeletePlaylist(ctx, tt.accessToken, tt.userId, tt.playlistId)

			assert.NoError(err)
		})
	}
}

func TestSpotifyClient_DeletePlaylist_Errors(t *testing.T) {
	tests := []struct {
		name           string
		userId         string
		playlistId     string
		responseStatus int
		responseError  error
		accessToken    string
	}{
		{
			name:           "playlist not found",
			userId:         "user123",
			playlistId:     "nonexistent_playlist",
			responseStatus: http.StatusNotFound,
			accessToken:    "valid_token",
		},
		{
			name:          "http client error",
			userId:        "user123",
			playlistId:    "playlist456",
			responseError: errors.New("connection timeout"),
			accessToken:   "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			} else {
				responseBody := io.NopCloser(strings.NewReader(`{"error":"test_error","error_description":"Test error description"}`))
				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(resp, nil).
					Times(1)
			}

			ctx := context.Background()
			err := client.DeletePlaylist(ctx, tt.accessToken, tt.userId, tt.playlistId)

			assert.Error(err)
		})
	}
}

func TestSpotifyClient_UpdatePlaylist_Success(t *testing.T) {
	tests := []struct {
		name                  string
		userId                string
		playlistId            string
		playlistName          string
		description           string
		accessToken           string
		expectedRequestFields map[string]any
	}{
		{
			name:         "update both name and description",
			userId:       "user123",
			playlistId:   "playlist456",
			playlistName: "Updated Playlist Name",
			description:  "Updated description",
			accessToken:  "valid_token",
			expectedRequestFields: map[string]any{
				"name":        "Updated Playlist Name",
				"description": "Updated description",
			},
		},
		{
			name:         "update only name (empty description)",
			userId:       "user123",
			playlistId:   "playlist456",
			playlistName: "New Name Only",
			description:  "",
			accessToken:  "valid_token",
			expectedRequestFields: map[string]any{
				"name": "New Name Only",
			},
		},
		{
			name:         "update only description (empty name)",
			userId:       "user123",
			playlistId:   "playlist456",
			playlistName: "",
			description:  "New description only",
			accessToken:  "valid_token",
			expectedRequestFields: map[string]any{
				"description": "New description only",
			},
		},
		{
			name:                  "no updates (both empty)",
			userId:                "user123",
			playlistId:            "playlist456",
			playlistName:          "",
			description:           "",
			accessToken:           "valid_token",
			expectedRequestFields: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Create successful response
			responseBody := io.NopCloser(strings.NewReader(""))
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       responseBody,
			}

			mockHTTPClient.EXPECT().
				Do(gomock.Any()).
				DoAndReturn(func(req *http.Request) (*http.Response, error) {
					// Validate request
					assert.Equal("PUT", req.Method)
					assert.Equal("https://api.spotify.com/v1/playlists/"+tt.playlistId, req.URL.String())
					assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))
					assert.Equal("application/json", req.Header.Get("Content-Type"))

					// Validate request body
					body, _ := io.ReadAll(req.Body)
					var requestBody map[string]any
					err := json.Unmarshal(body, &requestBody)
					assert.NoError(err)

					// Check that only expected fields are present
					assert.Equal(len(tt.expectedRequestFields), len(requestBody))
					for key, expectedValue := range tt.expectedRequestFields {
						assert.Equal(expectedValue, requestBody[key], "Field %s should match expected value", key)
					}

					// Ensure fields that should be omitted are not present
					if tt.playlistName == "" {
						_, hasName := requestBody["name"]
						assert.False(hasName, "Name field should be omitted when empty")
					}
					if tt.description == "" {
						_, hasDescription := requestBody["description"]
						assert.False(hasDescription, "Description field should be omitted when empty")
					}

					// Public field should never be present in update requests
					_, hasPublic := requestBody["public"]
					assert.False(hasPublic, "Public field should not be present in update requests")

					return resp, nil
				}).
				Times(1)

			ctx := context.Background()
			err := client.UpdatePlaylist(ctx, tt.accessToken, tt.userId, tt.playlistId, tt.playlistName, tt.description)

			assert.NoError(err)
		})
	}
}

func TestSpotifyClient_UpdatePlaylist_Errors(t *testing.T) {
	tests := []struct {
		name           string
		userId         string
		playlistId     string
		playlistName   string
		description    string
		responseStatus int
		responseError  error
		accessToken    string
	}{
		{
			name:           "playlist not found",
			userId:         "user123",
			playlistId:     "nonexistent_playlist",
			playlistName:   "New Name",
			description:    "New Description",
			responseStatus: http.StatusNotFound,
			accessToken:    "valid_token",
		},
		{
			name:           "unauthorized error",
			userId:         "user123",
			playlistId:     "playlist456",
			playlistName:   "New Name",
			description:    "New Description",
			responseStatus: http.StatusUnauthorized,
			accessToken:    "invalid_token",
		},
		{
			name:           "forbidden error - not owner",
			userId:         "user123",
			playlistId:     "other_users_playlist",
			playlistName:   "New Name",
			description:    "New Description",
			responseStatus: http.StatusForbidden,
			accessToken:    "valid_token",
		},
		{
			name:           "bad request error",
			userId:         "user123",
			playlistId:     "playlist456",
			playlistName:   "Name with invalid characters that are too long for Spotify's limits and exceed the maximum allowed length",
			description:    "Description",
			responseStatus: http.StatusBadRequest,
			accessToken:    "valid_token",
		},
		{
			name:          "http client error",
			userId:        "user123",
			playlistId:    "playlist456",
			playlistName:  "New Name",
			description:   "New Description",
			responseError: errors.New("network timeout"),
			accessToken:   "valid_token",
		},
		{
			name:           "internal server error",
			userId:         "user123",
			playlistId:     "playlist456",
			playlistName:   "New Name",
			description:    "New Description",
			responseStatus: http.StatusInternalServerError,
			accessToken:    "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			} else {
				responseBody := io.NopCloser(strings.NewReader(`{"error":"test_error","error_description":"Test error description"}`))
				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(resp, nil).
					Times(1)
			}

			ctx := context.Background()
			err := client.UpdatePlaylist(ctx, tt.accessToken, tt.userId, tt.playlistId, tt.playlistName, tt.description)

			assert.Error(err)
		})
	}
}

func contextWithToken(token string) context.Context {
	ctx := context.Background()
	
	if token == "" {
		return ctx
	}

	return requestcontext.ContextWithSpotifyAuth(context.Background(), &models.SpotifyIntegration{
		AccessToken: token,
	})
}