package services

import (
	"context"
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/require"
)

func TestNewTrackRouterService(t *testing.T) {
	require := require.New(t)

	logger := createTestLogger()
	service := NewTrackRouterService(logger)

	require.NotNil(service)
	require.NotNil(service.logger)
}

func TestTrackRouterService_RouteTracksToChildren_Success(t *testing.T) {
	tests := []struct {
		name            string
		tracks          *models.PlaylistTracksInfo
		childPlaylists  []*models.ChildPlaylist
		expectedRouting map[string][]string
		expectedMatches int
	}{
		{
			name: "no active child playlists",
			tracks: &models.PlaylistTracksInfo{
				PlaylistID: "base123",
				Tracks: []models.TrackInfo{
					{URI: "track1", DurationMs: 180000},
				},
			},
			childPlaylists: []*models.ChildPlaylist{
				{
					ID:                "child1",
					SpotifyPlaylistID: "spotify-child1",
					IsActive:          false,
					FilterRules:       nil,
				},
			},
			expectedRouting: map[string][]string{},
			expectedMatches: 0,
		},
		{
			name: "single track matches single child playlist",
			tracks: &models.PlaylistTracksInfo{
				PlaylistID: "base123",
				Tracks: []models.TrackInfo{
					{
						URI:        "track1",
						DurationMs: 180000,
						Popularity: 70,
						Explicit:   false,
					},
				},
			},
			childPlaylists: []*models.ChildPlaylist{
				{
					ID:                "child1",
					SpotifyPlaylistID: "spotify-child1",
					IsActive:          true,
					FilterRules: &models.MetadataFilters{
						Duration: &models.RangeFilter{
							Min: float64ToPointer(120000),
							Max: float64ToPointer(240000),
						},
					},
				},
			},
			expectedRouting: map[string][]string{
				"spotify-child1": {"track1"},
			},
			expectedMatches: 1,
		},
		{
			name: "track matches multiple child playlists",
			tracks: &models.PlaylistTracksInfo{
				PlaylistID: "base123",
				Tracks: []models.TrackInfo{
					{
						URI:        "track1",
						DurationMs: 180000,
						Popularity: 80,
						AllGenres:  []string{"rock", "pop"},
					},
				},
			},
			childPlaylists: []*models.ChildPlaylist{
				{
					ID:                "child1",
					SpotifyPlaylistID: "spotify-child1",
					IsActive:          true,
					FilterRules: &models.MetadataFilters{
						Duration: &models.RangeFilter{Min: float64ToPointer(120000)},
					},
				},
				{
					ID:                "child2",
					SpotifyPlaylistID: "spotify-child2",
					IsActive:          true,
					FilterRules: &models.MetadataFilters{
						Genres: &models.SetFilter{Include: []string{"rock"}},
					},
				},
			},
			expectedRouting: map[string][]string{
				"spotify-child1": {"track1"},
				"spotify-child2": {"track1"},
			},
			expectedMatches: 2,
		},
		{
			name: "track doesn't match filter criteria",
			tracks: &models.PlaylistTracksInfo{
				PlaylistID: "base123",
				Tracks: []models.TrackInfo{
					{
						URI:        "track1",
						DurationMs: 90000, // Too short
						Popularity: 30,    // Too low
					},
				},
			},
			childPlaylists: []*models.ChildPlaylist{
				{
					ID:                "child1",
					SpotifyPlaylistID: "spotify-child1",
					IsActive:          true,
					FilterRules: &models.MetadataFilters{
						Duration:   &models.RangeFilter{Min: float64ToPointer(120000)},
						Popularity: &models.RangeFilter{Min: float64ToPointer(50)},
					},
				},
			},
			expectedRouting: map[string][]string{},
			expectedMatches: 0,
		},
		{
			name: "multiple tracks with mixed results",
			tracks: &models.PlaylistTracksInfo{
				PlaylistID: "base123",
				Tracks: []models.TrackInfo{
					{
						URI:        "track1",
						DurationMs: 180000,
						Popularity: 80,
						AllGenres:  []string{"rock"},
					},
					{
						URI:        "track2",
						DurationMs: 90000, // Too short
						Popularity: 90,
						AllGenres:  []string{"rock"},
					},
					{
						URI:        "track3",
						DurationMs: 200000,
						Popularity: 75,
						AllGenres:  []string{"jazz"}, // Wrong genre
					},
				},
			},
			childPlaylists: []*models.ChildPlaylist{
				{
					ID:                "child1",
					SpotifyPlaylistID: "spotify-child1",
					IsActive:          true,
					FilterRules: &models.MetadataFilters{
						Duration: &models.RangeFilter{Min: float64ToPointer(120000)},
						Genres:   &models.SetFilter{Include: []string{"rock"}},
					},
				},
			},
			expectedRouting: map[string][]string{
				"spotify-child1": {"track1"}, // Only track1 matches both criteria
			},
			expectedMatches: 1,
		},
		{
			name: "nil filter rules matches all tracks",
			tracks: &models.PlaylistTracksInfo{
				PlaylistID: "base123",
				Tracks: []models.TrackInfo{
					{URI: "track1", DurationMs: 180000},
					{URI: "track2", DurationMs: 90000},
				},
			},
			childPlaylists: []*models.ChildPlaylist{
				{
					ID:                "child1",
					SpotifyPlaylistID: "spotify-child1",
					IsActive:          true,
					FilterRules:       nil, // No filters
				},
			},
			expectedRouting: map[string][]string{
				"spotify-child1": {"track1", "track2"},
			},
			expectedMatches: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ctx := context.Background()

			logger := createTestLogger()
			service := NewTrackRouterService(logger)

			routing, err := service.RouteTracksToChildren(ctx, tt.tracks, tt.childPlaylists)

			require.NoError(err)
			require.Equal(tt.expectedRouting, routing)

			// Verify total matches
			totalMatches := 0
			for _, trackIDs := range routing {
				totalMatches += len(trackIDs)
			}
			require.Equal(tt.expectedMatches, totalMatches)
		})
	}
}

func TestTrackRouterService_RouteTracksToChildren_EmptyInputs(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	logger := createTestLogger()
	service := NewTrackRouterService(logger)

	t.Run("empty tracks", func(t *testing.T) {
		tracks := &models.PlaylistTracksInfo{
			PlaylistID: "base123",
			Tracks:     []models.TrackInfo{},
		}
		childPlaylists := []*models.ChildPlaylist{
			{
				ID:                "child1",
				SpotifyPlaylistID: "spotify-child1",
				IsActive:          true,
				FilterRules:       nil,
			},
		}

		routing, err := service.RouteTracksToChildren(ctx, tracks, childPlaylists)

		require.NoError(err)
		require.Empty(routing)
	})

	t.Run("empty child playlists", func(t *testing.T) {
		tracks := &models.PlaylistTracksInfo{
			PlaylistID: "base123",
			Tracks: []models.TrackInfo{
				{URI: "track1", DurationMs: 180000},
			},
		}
		childPlaylists := []*models.ChildPlaylist{}

		routing, err := service.RouteTracksToChildren(ctx, tracks, childPlaylists)

		require.NoError(err)
		require.Empty(routing)
	})
}

func TestTrackRouterService_RouteTracksToChildren_ComplexFilters(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	logger := createTestLogger()
	service := NewTrackRouterService(logger)

	tracks := &models.PlaylistTracksInfo{
		PlaylistID: "base123",
		Tracks: []models.TrackInfo{
			{
				URI:          "track1",
				Name:         "Love Song",
				DurationMs:   180000,
				Popularity:   80,
				Explicit:     false,
				AllGenres:    []string{"rock", "pop"},
				ReleaseYear:  2010,
				MaxArtistPop: 85,
				ArtistNames:  []string{"The Beatles", "John Lennon"},
			},
		},
	}

	childPlaylists := []*models.ChildPlaylist{
		{
			ID:                "child1",
			SpotifyPlaylistID: "spotify-child1",
			IsActive:          true,
			FilterRules: &models.MetadataFilters{
				Duration:         &models.RangeFilter{Min: float64ToPointer(120000), Max: float64ToPointer(240000)},
				Popularity:       &models.RangeFilter{Min: float64ToPointer(70)},
				Explicit:         boolToPointer(false),
				Genres:           &models.SetFilter{Include: []string{"rock"}},
				ReleaseYear:      &models.RangeFilter{Min: float64ToPointer(2000), Max: float64ToPointer(2020)},
				ArtistPopularity: &models.RangeFilter{Min: float64ToPointer(80)},
				TrackKeywords:    &models.SetFilter{Include: []string{"love"}},
				ArtistKeywords:   &models.SetFilter{Include: []string{"beatles"}},
			},
		},
	}

	routing, err := service.RouteTracksToChildren(ctx, tracks, childPlaylists)

	require.NoError(err)
	require.Equal(map[string][]string{
		"spotify-child1": {"track1"},
	}, routing)
}
