package spotifyclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/clients/mocks"
	"github.com/ngomez18/playlist-router/internal/config"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/require"
)

func TestSpotifyClient_GetSeveralArtists_Success(t *testing.T) {
	tests := []struct {
		name           string
		artistIDs      []string
		responseBody   struct{ Artists []*SpotifyArtist }
		expectedResult []*SpotifyArtist
		accessToken    string
		responseStatus int
	}{
		{
			name:           "successful artists fetch with multiple artists",
			artistIDs:      []string{"artist123", "artist456"},
			accessToken:    "valid_access_token",
			responseStatus: http.StatusOK,
			responseBody: struct{ Artists []*SpotifyArtist }{
				Artists: []*SpotifyArtist{
					{
						ID:         "artist123",
						Name:       "Test Artist 1",
						URI:        "spotify:artist:artist123",
						Genres:     []string{"pop", "rock"},
						Popularity: 80,
					},
					{
						ID:         "artist456",
						Name:       "Test Artist 2",
						URI:        "spotify:artist:artist456",
						Genres:     []string{"jazz", "blues"},
						Popularity: 65,
					},
				},
			},
			expectedResult: []*SpotifyArtist{
				{
					ID:         "artist123",
					Name:       "Test Artist 1",
					URI:        "spotify:artist:artist123",
					Genres:     []string{"pop", "rock"},
					Popularity: 80,
				},
				{
					ID:         "artist456",
					Name:       "Test Artist 2",
					URI:        "spotify:artist:artist456",
					Genres:     []string{"jazz", "blues"},
					Popularity: 65,
				},
			},
		},
		{
			name:           "successful artists fetch with single artist",
			artistIDs:      []string{"artist789"},
			accessToken:    "valid_access_token",
			responseStatus: http.StatusOK,
			responseBody: struct{ Artists []*SpotifyArtist }{
				Artists: []*SpotifyArtist{
					{
						ID:         "artist789",
						Name:       "Solo Artist",
						URI:        "spotify:artist:artist789",
						Genres:     []string{"electronic", "ambient"},
						Popularity: 45,
					},
				},
			},
			expectedResult: []*SpotifyArtist{
				{
					ID:         "artist789",
					Name:       "Solo Artist",
					URI:        "spotify:artist:artist789",
					Genres:     []string{"electronic", "ambient"},
					Popularity: 45,
				},
			},
		},
		{
			name:           "empty artist IDs returns empty slice",
			artistIDs:      []string{},
			accessToken:    "valid_access_token",
			expectedResult: []*SpotifyArtist{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := setupMockController(t)

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			logger := createTestLogger()
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
				// Create expected URL (note: URL encoding will encode commas as %2C)
				expectedArtistIDs := strings.Join(tt.artistIDs, "%2C")
				expectedURL := fmt.Sprintf("https://api.spotify.com/v1/artists?ids=%s", expectedArtistIDs)

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
			result, err := client.GetSeveralArtists(ctx, tt.artistIDs)

			// Assert
			assert.NoError(err)
			assert.NotNil(result)
			assert.Equal(len(tt.expectedResult), len(result))

			for i, expectedArtist := range tt.expectedResult {
				if i < len(result) {
					assert.Equal(expectedArtist.ID, result[i].ID)
					assert.Equal(expectedArtist.Name, result[i].Name)
					assert.Equal(expectedArtist.URI, result[i].URI)
					assert.Equal(expectedArtist.Genres, result[i].Genres)
					assert.Equal(expectedArtist.Popularity, result[i].Popularity)
				}
			}
		})
	}
}

func TestSpotifyClient_GetSeveralArtists_Errors(t *testing.T) {
	tests := []struct {
		name           string
		artistIDs      []string
		accessToken    string
		responseStatus int
		responseBody   string
		httpError      error
		expectedError  string
	}{
		{
			name:           "artists not found",
			artistIDs:      []string{"nonexistent_artist"},
			accessToken:    "valid_access_token",
			responseStatus: http.StatusNotFound,
			responseBody:   `{"error":{"status":404,"message":"No such artist"}}`,
			expectedError:  "spotify artists fetch failed (status 404)",
		},
		{
			name:          "missing access token",
			artistIDs:     []string{"artist123"},
			expectedError: "spotify credentials not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := setupMockController(t)

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			logger := createTestLogger()
			authConfig := &config.AuthConfig{}

			client := NewSpotifyClient(authConfig, logger)
			client.HttpClient = mockHTTPClient

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
			}

			if tt.httpError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.httpError).
					Times(1)
			}

			// Execute
			result, err := client.GetSeveralArtists(ctx, tt.artistIDs)

			// Assert
			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedError)
		})
	}
}
