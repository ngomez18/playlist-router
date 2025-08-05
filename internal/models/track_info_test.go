package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlaylistTracksInfo_GetAllArtists(t *testing.T) {
	tests := []struct {
		name           string
		playlistTracks PlaylistTracksInfo
		expectedCount  int
		expectedArtists []string
	}{
		{
			name: "single track with one artist",
			playlistTracks: PlaylistTracksInfo{
				PlaylistID: "playlist1",
				Tracks: []TrackInfo{
					{
						ID:      "track1",
						Name:    "Test Track",
						Artists: []string{"artist1"},
					},
				},
			},
			expectedCount:   1,
			expectedArtists: []string{"artist1"},
		},
		{
			name: "single track with multiple artists",
			playlistTracks: PlaylistTracksInfo{
				PlaylistID: "playlist1",
				Tracks: []TrackInfo{
					{
						ID:      "track1",
						Name:    "Test Track",
						Artists: []string{"artist1", "artist2", "artist3"},
					},
				},
			},
			expectedCount:   3,
			expectedArtists: []string{"artist1", "artist2", "artist3"},
		},
		{
			name: "multiple tracks with duplicate artists",
			playlistTracks: PlaylistTracksInfo{
				PlaylistID: "playlist1",
				Tracks: []TrackInfo{
					{
						ID:      "track1",
						Name:    "Test Track 1",
						Artists: []string{"artist1", "artist2"},
					},
					{
						ID:      "track2",
						Name:    "Test Track 2",
						Artists: []string{"artist2", "artist3"},
					},
					{
						ID:      "track3",
						Name:    "Test Track 3",
						Artists: []string{"artist1", "artist3"},
					},
				},
			},
			expectedCount:   3,
			expectedArtists: []string{"artist1", "artist2", "artist3"},
		},
		{
			name: "empty tracks list",
			playlistTracks: PlaylistTracksInfo{
				PlaylistID: "playlist1",
				Tracks:     []TrackInfo{},
			},
			expectedCount:   0,
			expectedArtists: []string{},
		},
		{
			name: "tracks with no artists",
			playlistTracks: PlaylistTracksInfo{
				PlaylistID: "playlist1",
				Tracks: []TrackInfo{
					{
						ID:      "track1",
						Name:    "Test Track",
						Artists: []string{},
					},
				},
			},
			expectedCount:   0,
			expectedArtists: []string{},
		},
		{
			name: "large playlist with many duplicate artists",
			playlistTracks: PlaylistTracksInfo{
				PlaylistID: "playlist1",
				Tracks: []TrackInfo{
					{ID: "track1", Artists: []string{"artist1", "artist2"}},
					{ID: "track2", Artists: []string{"artist1", "artist3"}},
					{ID: "track3", Artists: []string{"artist2", "artist4"}},
					{ID: "track4", Artists: []string{"artist3", "artist4"}},
					{ID: "track5", Artists: []string{"artist1", "artist5"}},
					{ID: "track6", Artists: []string{"artist2", "artist5"}},
				},
			},
			expectedCount:   5,
			expectedArtists: []string{"artist1", "artist2", "artist3", "artist4", "artist5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			result := tt.playlistTracks.GetAllArtists()

			// Check count
			assert.Equal(tt.expectedCount, len(result))

			// Check that all expected artists are present (order may vary due to map iteration)
			if tt.expectedCount > 0 {
				resultMap := make(map[string]bool)
				for _, artist := range result {
					resultMap[artist] = true
				}

				for _, expectedArtist := range tt.expectedArtists {
					assert.True(resultMap[expectedArtist], "Expected artist %s not found in result", expectedArtist)
				}
			}
		})
	}
}
