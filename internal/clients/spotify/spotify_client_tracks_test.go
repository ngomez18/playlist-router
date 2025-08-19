package spotifyclient

import (
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

func TestSpotifyClient_GetPlaylistTracks_Success(t *testing.T) {
	tests := []struct {
		name             string
		playlistID       string
		limit            int
		offset           int
		responseBody     *SpotifyPlaylistTracksResponse
		expectedResponse *SpotifyPlaylistTracksResponse
		accessToken      string
		responseStatus   int
	}{
		{
			name:           "successful tracks fetch with results",
			playlistID:     "playlist123",
			limit:          10,
			offset:         0,
			accessToken:    "valid_access_token",
			responseStatus: http.StatusOK,
			responseBody: &SpotifyPlaylistTracksResponse{
				Items: []SpotifyPlaylistTrack{
					{
						Track: &SpotifyTrack{
							ID:         "track123",
							Name:       "Test Track",
							URI:        "spotify:track:track123",
							DurationMs: 180000,
							Popularity: 75,
							Explicit:   false,
							Artists: []SpotifyArtist{
								{
									ID:         "artist123",
									Name:       "Test Artist",
									URI:        "spotify:artist:artist123",
									Genres:     []string{"pop", "rock"},
									Popularity: 80,
								},
							},
							Album: SpotifyAlbum{
								ID:          "album123",
								Name:        "Test Album",
								URI:         "spotify:album:album123",
								ReleaseDate: "2023-01-01",
							},
						},
					},
				},
				Total:  1,
				Limit:  10,
				Offset: 0,
				Next:   nil,
			},
			expectedResponse: &SpotifyPlaylistTracksResponse{
				Items: []SpotifyPlaylistTrack{
					{
						Track: &SpotifyTrack{
							ID:         "track123",
							Name:       "Test Track",
							URI:        "spotify:track:track123",
							DurationMs: 180000,
							Popularity: 75,
							Explicit:   false,
							Artists: []SpotifyArtist{
								{
									ID:         "artist123",
									Name:       "Test Artist",
									URI:        "spotify:artist:artist123",
									Genres:     []string{"pop", "rock"},
									Popularity: 80,
								},
							},
							Album: SpotifyAlbum{
								ID:          "album123",
								Name:        "Test Album",
								URI:         "spotify:album:album123",
								ReleaseDate: "2023-01-01",
							},
						},
					},
				},
				Total:  1,
				Limit:  10,
				Offset: 0,
				Next:   nil,
			},
		},
		{
			name:           "successful tracks fetch with pagination",
			playlistID:     "playlist456",
			limit:          5,
			offset:         10,
			accessToken:    "valid_access_token",
			responseStatus: http.StatusOK,
			responseBody: &SpotifyPlaylistTracksResponse{
				Items: []SpotifyPlaylistTrack{
					{
						Track: &SpotifyTrack{
							ID:         "track456",
							Name:       "Second Page Track",
							URI:        "spotify:track:track456",
							DurationMs: 200000,
							Popularity: 60,
							Explicit:   true,
							Artists: []SpotifyArtist{
								{
									ID:   "artist456",
									Name: "Second Artist",
									URI:  "spotify:artist:artist456",
								},
							},
							Album: SpotifyAlbum{
								ID:          "album456",
								Name:        "Second Album",
								URI:         "spotify:album:album456",
								ReleaseDate: "2022-06-15",
							},
						},
					},
				},
				Total:  50,
				Limit:  5,
				Offset: 10,
				Next:   stringPointer("https://api.spotify.com/v1/playlists/playlist456/tracks?offset=15&limit=5"),
			},
			expectedResponse: &SpotifyPlaylistTracksResponse{
				Items: []SpotifyPlaylistTrack{
					{
						Track: &SpotifyTrack{
							ID:         "track456",
							Name:       "Second Page Track",
							URI:        "spotify:track:track456",
							DurationMs: 200000,
							Popularity: 60,
							Explicit:   true,
							Artists: []SpotifyArtist{
								{
									ID:   "artist456",
									Name: "Second Artist",
									URI:  "spotify:artist:artist456",
								},
							},
							Album: SpotifyAlbum{
								ID:          "album456",
								Name:        "Second Album",
								URI:         "spotify:album:album456",
								ReleaseDate: "2022-06-15",
							},
						},
					},
				},
				Total:  50,
				Limit:  5,
				Offset: 10,
				Next:   stringPointer("https://api.spotify.com/v1/playlists/playlist456/tracks?offset=15&limit=5"),
			},
		},
		{
			name:           "successful tracks fetch with empty response",
			playlistID:     "empty_playlist",
			limit:          10,
			offset:         0,
			accessToken:    "valid_access_token",
			responseStatus: http.StatusOK,
			responseBody: &SpotifyPlaylistTracksResponse{
				Items:  []SpotifyPlaylistTrack{},
				Total:  0,
				Limit:  10,
				Offset: 0,
				Next:   nil,
			},
			expectedResponse: &SpotifyPlaylistTracksResponse{
				Items:  []SpotifyPlaylistTrack{},
				Total:  0,
				Limit:  10,
				Offset: 0,
				Next:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			authConfig := &config.AuthConfig{}

			client := NewSpotifyClient(authConfig, logger)
			client.HttpClient = mockHTTPClient

			// Create context with Spotify integration
			spotifyIntegration := &models.SpotifyIntegration{
				AccessToken: tt.accessToken,
				UserID:      "test_user",
			}
			ctx := requestcontext.ContextWithSpotifyAuth(context.Background(), spotifyIntegration)

			if tt.responseStatus > 0 {
				// Create expected URL
				expectedURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?limit=%d&offset=%d", tt.playlistID, tt.limit, tt.offset)

				// Mock response
				responseBody, _ := json.Marshal(tt.responseBody)
				mockResponse := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       io.NopCloser(strings.NewReader(string(responseBody))),
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal("GET", req.Method)
						assert.Equal(expectedURL, req.URL.String())
						assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))
						return mockResponse, nil
					}).
					Times(1)
			}

			// Execute
			result, err := client.GetPlaylistTracks(ctx, tt.playlistID, tt.limit, tt.offset)

			// Assert
			assert.NoError(err)
			assert.NotNil(result)
			assert.Equal(tt.expectedResponse.Total, result.Total)
			assert.Equal(tt.expectedResponse.Limit, result.Limit)
			assert.Equal(tt.expectedResponse.Offset, result.Offset)
			assert.Equal(len(tt.expectedResponse.Items), len(result.Items))

			if len(result.Items) > 0 && result.Items[0].Track != nil {
				firstTrack := result.Items[0].Track
				expectedTrack := tt.expectedResponse.Items[0].Track
				assert.Equal(expectedTrack.ID, firstTrack.ID)
				assert.Equal(expectedTrack.Name, firstTrack.Name)
				assert.Equal(expectedTrack.URI, firstTrack.URI)
				assert.Equal(expectedTrack.DurationMs, firstTrack.DurationMs)
				assert.Equal(expectedTrack.Popularity, firstTrack.Popularity)
				assert.Equal(expectedTrack.Explicit, firstTrack.Explicit)
			}
		})
	}
}

func TestSpotifyClient_GetPlaylistTracks_Errors(t *testing.T) {
	tests := []struct {
		name           string
		playlistID     string
		limit          int
		offset         int
		accessToken    string
		responseStatus int
		responseBody   string
		httpError      error
		expectedError  string
	}{
		{
			name:           "playlist not found",
			playlistID:     "nonexistent_playlist",
			limit:          10,
			offset:         0,
			accessToken:    "valid_access_token",
			responseStatus: http.StatusNotFound,
			responseBody:   `{"error":{"status":404,"message":"No such playlist"}}`,
			expectedError:  "spotify playlist tracks fetch failed (status 404)",
		},
		{
			name:          "http client error",
			playlistID:    "playlist123",
			limit:         10,
			offset:        0,
			accessToken:   "valid_access_token",
			httpError:     errors.New("connection timeout"),
			expectedError: "failed to get playlist tracks",
		},
		{
			name:          "missing access token",
			playlistID:    "playlist123",
			limit:         10,
			offset:        0,
			expectedError: "spotify credentials not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			authConfig := &config.AuthConfig{}

			client := NewSpotifyClient(authConfig, logger)
			client.HttpClient = mockHTTPClient

			// Create context
			var ctx context.Context
			if tt.accessToken != "" {
				spotifyIntegration := &models.SpotifyIntegration{
					AccessToken: tt.accessToken,
					UserID:      "test_user",
				}
				ctx = requestcontext.ContextWithSpotifyAuth(context.Background(), spotifyIntegration)
			} else {
				ctx = context.Background()
			}

			if tt.responseStatus > 0 {
				mockResponse := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(mockResponse, tt.httpError).
					Times(1)
			} else if tt.httpError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.httpError).
					Times(1)
			}

			// Execute
			result, err := client.GetPlaylistTracks(ctx, tt.playlistID, tt.limit, tt.offset)

			// Assert
			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedError)
		})
	}
}

func TestSpotifyClient_AddTracksToPlaylist_Success(t *testing.T) {
	tests := []struct {
		name           string
		playlistID     string
		trackURIs      []string
		accessToken    string
		responseStatus int
	}{
		{
			name:       "successful tracks addition",
			playlistID: "playlist123",
			trackURIs: []string{
				"spotify:track:track1",
				"spotify:track:track2",
				"spotify:track:track3",
			},
			accessToken:    "valid_access_token",
			responseStatus: http.StatusCreated,
		},
		{
			name:           "successful addition with single track",
			playlistID:     "playlist456",
			trackURIs:      []string{"spotify:track:single_track"},
			accessToken:    "valid_access_token",
			responseStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			authConfig := &config.AuthConfig{}

			client := NewSpotifyClient(authConfig, logger)
			client.HttpClient = mockHTTPClient

			// Create context with Spotify integration
			spotifyIntegration := &models.SpotifyIntegration{
				AccessToken: tt.accessToken,
				UserID:      "test_user",
			}
			ctx := requestcontext.ContextWithSpotifyAuth(context.Background(), spotifyIntegration)

			if tt.responseStatus > 0 {
				// Create expected URL and request body
				expectedURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", tt.playlistID)
				expectedRequestBody := map[string][]string{
					"uris": tt.trackURIs,
				}

				// Mock response
				mockResponse := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       io.NopCloser(strings.NewReader(`{"snapshot_id": "new_snapshot"}`)),
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal("POST", req.Method)
						assert.Equal(expectedURL, req.URL.String())
						assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))
						assert.Equal("application/json", req.Header.Get("Content-Type"))

						// Verify request body
						var requestBody map[string][]string
						bodyBytes, _ := io.ReadAll(req.Body)
						err := json.Unmarshal(bodyBytes, &requestBody)
						assert.NoError(err)
						assert.Equal(expectedRequestBody, requestBody)

						return mockResponse, nil
					}).
					Times(1)
			}

			// Execute
			err := client.AddTracksToPlaylist(ctx, tt.playlistID, tt.trackURIs)

			// Assert
			assert.NoError(err)
		})
	}
}

func TestSpotifyClient_AddTracksToPlaylist_Errors(t *testing.T) {
	tests := []struct {
		name           string
		playlistID     string
		trackURIs      []string
		accessToken    string
		responseStatus int
		responseBody   string
		httpError      error
		expectedError  string
	}{
		{
			name:       "playlist not found",
			playlistID: "nonexistent_playlist",
			trackURIs: []string{
				"spotify:track:track1",
				"spotify:track:track2",
			},
			accessToken:    "valid_access_token",
			responseStatus: http.StatusNotFound,
			responseBody:   `{"error":{"status":404,"message":"No such playlist"}}`,
			expectedError:  "spotify add tracks failed (status 404)",
		},
		{
			name:       "http client error",
			playlistID: "playlist123",
			trackURIs: []string{
				"spotify:track:track1",
			},
			accessToken:   "valid_access_token",
			httpError:     errors.New("connection timeout"),
			expectedError: "failed to add tracks to playlist",
		},
		{
			name:       "missing access token",
			playlistID: "playlist123",
			trackURIs: []string{
				"spotify:track:track1",
			},
			expectedError: "spotify credentials not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			authConfig := &config.AuthConfig{}

			client := NewSpotifyClient(authConfig, logger)
			client.HttpClient = mockHTTPClient

			// Create context
			var ctx context.Context
			if tt.accessToken != "" {
				spotifyIntegration := &models.SpotifyIntegration{
					AccessToken: tt.accessToken,
					UserID:      "test_user",
				}
				ctx = requestcontext.ContextWithSpotifyAuth(context.Background(), spotifyIntegration)
			} else {
				ctx = context.Background()
			}

			if tt.responseStatus > 0 {
				mockResponse := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(mockResponse, tt.httpError).
					Times(1)
			} else if tt.httpError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.httpError).
					Times(1)
			}

			// Execute
			err := client.AddTracksToPlaylist(ctx, tt.playlistID, tt.trackURIs)

			// Assert
			assert.Error(err)
			assert.Contains(err.Error(), tt.expectedError)
		})
	}
}

// Helper function to create string pointer
func stringPointer(s string) *string {
	return &s
}
