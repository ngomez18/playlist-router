package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	clientmocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	repomocks "github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewTrackAggregatorService(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := clientmocks.NewMockSpotifyAPI(ctrl)
	mockBasePlaylistRepo := repomocks.NewMockBasePlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	service := NewTrackAggregatorService(mockSpotifyClient, mockBasePlaylistRepo, logger)

	assert.NotNil(service)
	assert.Equal(mockSpotifyClient, service.spotifyClient)
	assert.Equal(mockBasePlaylistRepo, service.basePlaylistRepo)
	assert.Equal(logger, service.logger)
}

func TestTrackAggregatorService_AggregatePlaylistData_Success(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		basePlaylistID    string
		basePlaylist      *models.BasePlaylist
		tracksResponse    *spotifyclient.SpotifyPlaylistTracksResponse
		artistsResponse   []*spotifyclient.SpotifyArtist
		expectedAPICount  int
		expectedTrackCount int
		expectedArtistCount int
	}{
		{
			name:           "single page with two tracks and artists",
			userID:         "user123",
			basePlaylistID: "base456",
			basePlaylist: &models.BasePlaylist{
				ID:                "base456",
				UserID:            "user123",
				SpotifyPlaylistID: "spotify789",
				Name:              "Test Playlist",
			},
			tracksResponse: &spotifyclient.SpotifyPlaylistTracksResponse{
				Items: []spotifyclient.SpotifyPlaylistTrack{
					{
						Track: &spotifyclient.SpotifyTrack{
							ID:         "track1",
							Name:       "Track One",
							URI:        "spotify:track:track1",
							DurationMs: 180000,
							Artists: []spotifyclient.SpotifyArtist{
								{ID: "artist1", Name: "Artist One"},
								{ID: "artist2", Name: "Artist Two"},
							},
							Album: spotifyclient.SpotifyAlbum{
								ID:   "album1",
								Name: "Album One",
							},
						},
					},
					{
						Track: &spotifyclient.SpotifyTrack{
							ID:         "track2",
							Name:       "Track Two",
							URI:        "spotify:track:track2",
							DurationMs: 200000,
							Artists: []spotifyclient.SpotifyArtist{
								{ID: "artist2", Name: "Artist Two"}, // Duplicate
								{ID: "artist3", Name: "Artist Three"},
							},
							Album: spotifyclient.SpotifyAlbum{
								ID:   "album2",
								Name: "Album Two",
							},
						},
					},
				},
				Next: nil, // Single page
			},
			artistsResponse: []*spotifyclient.SpotifyArtist{
				{
					ID:         "artist1",
					Name:       "Artist One",
					Genres:     []string{"rock", "pop"},
					Popularity: 80,
					URI:        "spotify:artist:artist1",
				},
				{
					ID:         "artist2",
					Name:       "Artist Two",
					Genres:     []string{"jazz"},
					Popularity: 70,
					URI:        "spotify:artist:artist2",
				},
				{
					ID:         "artist3",
					Name:       "Artist Three",
					Genres:     []string{},
					Popularity: 60,
					URI:        "spotify:artist:artist3",
				},
			},
			expectedAPICount:    2, // 1 for tracks + 1 for artists
			expectedTrackCount:  2,
			expectedArtistCount: 3, // artist1, artist2, artist3 (deduplicated)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			ctx := context.Background()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Setup mocks
			mockSpotifyClient := clientmocks.NewMockSpotifyAPI(ctrl)
			mockBasePlaylistRepo := repomocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

			// Setup expectations
			mockBasePlaylistRepo.EXPECT().
				GetByID(ctx, tt.basePlaylistID, tt.userID).
				Return(tt.basePlaylist, nil).
				Times(1)

			mockSpotifyClient.EXPECT().
				GetPlaylistTracks(ctx, tt.basePlaylist.SpotifyPlaylistID, MAX_TRACKS, 0).
				Return(tt.tracksResponse, nil).
				Times(1)

			mockSpotifyClient.EXPECT().
				GetSeveralArtists(ctx, gomock.Any()).
				Return(tt.artistsResponse, nil).
				Times(1)

			// Execute
			service := NewTrackAggregatorService(mockSpotifyClient, mockBasePlaylistRepo, logger)
			result, err := service.AggregatePlaylistData(ctx, tt.userID, tt.basePlaylistID)

			// Assert
			assert.NoError(err)
			assert.NotNil(result)
			assert.Equal(tt.basePlaylist.SpotifyPlaylistID, result.PlaylistID)
			assert.Equal(tt.expectedTrackCount, len(result.Tracks))
			assert.Equal(tt.expectedArtistCount, len(result.Artists))
			assert.Equal(tt.expectedAPICount, result.APICallCount)

			// Verify track data
			assert.Equal("track1", result.Tracks[0].ID)
			assert.Equal("Track One", result.Tracks[0].Name)
			assert.Equal([]string{"artist1", "artist2"}, result.Tracks[0].Artists)

			// Verify artist data
			assert.Contains(result.Artists, "artist1")
			assert.Equal("Artist One", result.Artists["artist1"].Name)
			assert.Equal([]string{"rock", "pop"}, result.Artists["artist1"].Genres)
		})
	}
}

func TestTrackAggregatorService_AggregatePlaylistData_Errors(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		basePlaylistID    string
		basePlaylistError error
		tracksError       error
		artistsError      error
		expectedError     string
	}{
		{
			name:              "base playlist not found",
			userID:            "user123",
			basePlaylistID:    "nonexistent",
			basePlaylistError: errors.New("playlist not found"),
			expectedError:     "failed to fetch base playlist",
		},
		{
			name:           "spotify tracks fetch error",
			userID:         "user123",
			basePlaylistID: "base456",
			tracksError:    errors.New("spotify api error"),
			expectedError:  "failed to fetch playlist tracks",
		},
		{
			name:           "spotify artists fetch error",
			userID:         "user123",
			basePlaylistID: "base456",
			artistsError:   errors.New("artists api error"),
			expectedError:  "failed to fetch playlist artists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			ctx := context.Background()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyClient := clientmocks.NewMockSpotifyAPI(ctrl)
			mockBasePlaylistRepo := repomocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

			// Setup expectations based on error type
			if tt.basePlaylistError != nil {
				mockBasePlaylistRepo.EXPECT().
					GetByID(ctx, tt.basePlaylistID, tt.userID).
					Return(nil, tt.basePlaylistError).
					Times(1)
			} else {
				basePlaylist := &models.BasePlaylist{
					ID:                tt.basePlaylistID,
					UserID:            tt.userID,
					SpotifyPlaylistID: "spotify789",
					Name:              "Test Playlist",
				}
				mockBasePlaylistRepo.EXPECT().
					GetByID(ctx, tt.basePlaylistID, tt.userID).
					Return(basePlaylist, nil).
					Times(1)

				if tt.tracksError != nil {
					mockSpotifyClient.EXPECT().
						GetPlaylistTracks(ctx, "spotify789", MAX_TRACKS, 0).
						Return(nil, tt.tracksError).
						Times(1)
				} else if tt.artistsError != nil {
					tracksResponse := &spotifyclient.SpotifyPlaylistTracksResponse{
						Items: []spotifyclient.SpotifyPlaylistTrack{
							{
								Track: &spotifyclient.SpotifyTrack{
									ID:      "track1",
									Artists: []spotifyclient.SpotifyArtist{{ID: "artist1"}},
									Album:   spotifyclient.SpotifyAlbum{ID: "album1"},
								},
							},
						},
						Next: nil,
					}
					mockSpotifyClient.EXPECT().
						GetPlaylistTracks(ctx, "spotify789", MAX_TRACKS, 0).
						Return(tracksResponse, nil).
						Times(1)

					mockSpotifyClient.EXPECT().
						GetSeveralArtists(ctx, []string{"artist1"}).
						Return(nil, tt.artistsError).
						Times(1)
				}
			}

			// Execute
			service := NewTrackAggregatorService(mockSpotifyClient, mockBasePlaylistRepo, logger)
			result, err := service.AggregatePlaylistData(ctx, tt.userID, tt.basePlaylistID)

			// Assert
			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedError)
		})
	}
}

func TestTrackAggregatorService_EmptyPlaylist(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := clientmocks.NewMockSpotifyAPI(ctrl)
	mockBasePlaylistRepo := repomocks.NewMockBasePlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	basePlaylist := &models.BasePlaylist{
		ID:                "base123",
		UserID:            "user123",
		SpotifyPlaylistID: "spotify456",
		Name:              "Empty Playlist",
	}

	emptyTracksResponse := &spotifyclient.SpotifyPlaylistTracksResponse{
		Items: []spotifyclient.SpotifyPlaylistTrack{},
		Next:  nil,
	}

	// Setup expectations
	mockBasePlaylistRepo.EXPECT().
		GetByID(ctx, "base123", "user123").
		Return(basePlaylist, nil).
		Times(1)

	mockSpotifyClient.EXPECT().
		GetPlaylistTracks(ctx, "spotify456", MAX_TRACKS, 0).
		Return(emptyTracksResponse, nil).
		Times(1)

	// No artists call expected since artistIDs will be empty

	// Execute
	service := NewTrackAggregatorService(mockSpotifyClient, mockBasePlaylistRepo, logger)
	result, err := service.AggregatePlaylistData(ctx, "user123", "base123")

	// Assert
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal("spotify456", result.PlaylistID)
	assert.Equal(0, len(result.Tracks))
	assert.Equal(0, len(result.Artists))
	assert.Equal(1, result.APICallCount) // Only tracks call
}
