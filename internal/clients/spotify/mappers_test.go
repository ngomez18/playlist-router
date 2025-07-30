package spotifyclient

import (
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestParseSpotifyPlaylist(t *testing.T) {
	tests := []struct {
		name     string
		input    *SpotifyPlaylist
		expected *models.SpotifyPlaylist
	}{
		{
			name: "complete playlist with tracks",
			input: &SpotifyPlaylist{
				ID:   "playlist123",
				Name: "My Awesome Playlist",
				Tracks: &SpotifyPlaylistTracks{
					Total: 42,
				},
			},
			expected: &models.SpotifyPlaylist{
				ID:     "playlist123",
				Name:   "My Awesome Playlist",
				Tracks: 42,
			},
		},
		{
			name: "playlist with nil tracks",
			input: &SpotifyPlaylist{
				ID:     "no_tracks_info",
				Name:   "Unknown Track Count",
				Tracks: nil,
			},
			expected: &models.SpotifyPlaylist{
				ID:     "no_tracks_info",
				Name:   "Unknown Track Count",
				Tracks: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			result := ParseSpotifyPlaylist(tt.input)

			assert.Equal(tt.expected.ID, result.ID)
			assert.Equal(tt.expected.Name, result.Name)
			assert.Equal(tt.expected.Tracks, result.Tracks)
		})
	}
}

func TestParseManySpotifyPlaylist(t *testing.T) {
	tests := []struct {
		name     string
		input    []*SpotifyPlaylist
		expected []*models.SpotifyPlaylist
	}{
		{
			name: "multiple playlists",
			input: []*SpotifyPlaylist{
				{
					ID:     "playlist1",
					Name:   "Rock Classics",
					Tracks: &SpotifyPlaylistTracks{Total: 25},
				},
				{
					ID:     "playlist2",
					Name:   "Jazz Favorites",
					Tracks: &SpotifyPlaylistTracks{Total: 18},
				},
				{
					ID:     "playlist3",
					Name:   "Empty List",
					Tracks: &SpotifyPlaylistTracks{Total: 0},
				},
			},
			expected: []*models.SpotifyPlaylist{
				{ID: "playlist1", Name: "Rock Classics", Tracks: 25},
				{ID: "playlist2", Name: "Jazz Favorites", Tracks: 18},
				{ID: "playlist3", Name: "Empty List", Tracks: 0},
			},
		},
		{
			name:     "empty slice",
			input:    []*SpotifyPlaylist{},
			expected: []*models.SpotifyPlaylist{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []*models.SpotifyPlaylist{},
		},
		{
			name: "single playlist",
			input: []*SpotifyPlaylist{
				{
					ID:     "single",
					Name:   "Solo Track",
					Tracks: &SpotifyPlaylistTracks{Total: 1},
				},
			},
			expected: []*models.SpotifyPlaylist{
				{ID: "single", Name: "Solo Track", Tracks: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			result := ParseManySpotifyPlaylist(tt.input)

			assert.Equal(len(tt.expected), len(result))

			for i, expected := range tt.expected {
				assert.Equal(expected.ID, result[i].ID)
				assert.Equal(expected.Name, result[i].Name)
				assert.Equal(expected.Tracks, result[i].Tracks)
			}
		})
	}
}
