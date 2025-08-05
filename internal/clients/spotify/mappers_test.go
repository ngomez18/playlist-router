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

func TestParsePlaylistTrack(t *testing.T) {
	tests := []struct {
		name     string
		input    SpotifyPlaylistTrack
		expected models.TrackInfo
	}{
		{
			name: "track with multiple artists",
			input: SpotifyPlaylistTrack{
				Track: &SpotifyTrack{
					ID:         "track123",
					Name:       "Test Track",
					URI:        "spotify:track:track123",
					DurationMs: 180000,
					Popularity: 75,
					Explicit:   true,
					Artists: []SpotifyArtist{
						{ID: "artist1", Name: "Artist One"},
						{ID: "artist2", Name: "Artist Two"},
					},
					Album: SpotifyAlbum{
						ID:          "album123",
						Name:        "Test Album",
						ReleaseDate: "2023-01-01",
						URI:         "spotify:album:album123",
					},
				},
			},
			expected: models.TrackInfo{
				ID:         "track123",
				Name:       "Test Track",
				URI:        "spotify:track:track123",
				DurationMs: 180000,
				Popularity: 75,
				Explicit:   true,
				Artists:    []string{"artist1", "artist2"},
				Album: models.AlbumInfo{
					ID:          "album123",
					Name:        "Test Album",
					ReleaseDate: "2023-01-01",
					URI:         "spotify:album:album123",
				},
			},
		},
		{
			name: "track with single artist",
			input: SpotifyPlaylistTrack{
				Track: &SpotifyTrack{
					ID:         "track456",
					Name:       "Solo Track",
					URI:        "spotify:track:track456",
					DurationMs: 200000,
					Popularity: 60,
					Explicit:   false,
					Artists: []SpotifyArtist{
						{ID: "artist3", Name: "Solo Artist"},
					},
					Album: SpotifyAlbum{
						ID:          "album456",
						Name:        "Solo Album",
						ReleaseDate: "2022-05-15",
						URI:         "spotify:album:album456",
					},
				},
			},
			expected: models.TrackInfo{
				ID:         "track456",
				Name:       "Solo Track",
				URI:        "spotify:track:track456",
				DurationMs: 200000,
				Popularity: 60,
				Explicit:   false,
				Artists:    []string{"artist3"},
				Album: models.AlbumInfo{
					ID:          "album456",
					Name:        "Solo Album",
					ReleaseDate: "2022-05-15",
					URI:         "spotify:album:album456",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			result := ParsePlaylistTrack(tt.input)
			assert.Equal(tt.expected, result)
		})
	}
}

func TestParseManyPlaylistTracks(t *testing.T) {
	tests := []struct {
		name     string
		input    []SpotifyPlaylistTrack
		expected []models.TrackInfo
	}{
		{
			name: "multiple tracks",
			input: []SpotifyPlaylistTrack{
				{
					Track: &SpotifyTrack{
						ID:         "track1",
						Name:       "Track One",
						URI:        "spotify:track:track1",
						DurationMs: 180000,
						Artists:    []SpotifyArtist{{ID: "artist1"}},
						Album:      SpotifyAlbum{ID: "album1", Name: "Album One"},
					},
				},
				{
					Track: &SpotifyTrack{
						ID:         "track2",
						Name:       "Track Two",
						URI:        "spotify:track:track2",
						DurationMs: 200000,
						Artists:    []SpotifyArtist{{ID: "artist2"}},
						Album:      SpotifyAlbum{ID: "album2", Name: "Album Two"},
					},
				},
			},
			expected: []models.TrackInfo{
				{
					ID:         "track1",
					Name:       "Track One",
					URI:        "spotify:track:track1",
					DurationMs: 180000,
					Artists:    []string{"artist1"},
					Album:      models.AlbumInfo{ID: "album1", Name: "Album One"},
				},
				{
					ID:         "track2",
					Name:       "Track Two",
					URI:        "spotify:track:track2",
					DurationMs: 200000,
					Artists:    []string{"artist2"},
					Album:      models.AlbumInfo{ID: "album2", Name: "Album Two"},
				},
			},
		},
		{
			name:     "empty slice",
			input:    []SpotifyPlaylistTrack{},
			expected: []models.TrackInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			result := ParseManyPlaylistTracks(tt.input)
			assert.Equal(tt.expected, result)
		})
	}
}

func TestParseAlbum(t *testing.T) {
	tests := []struct {
		name     string
		input    *SpotifyAlbum
		expected *models.AlbumInfo
	}{
		{
			name: "complete album",
			input: &SpotifyAlbum{
				ID:          "album123",
				Name:        "Test Album",
				ReleaseDate: "2023-01-01",
				URI:         "spotify:album:album123",
			},
			expected: &models.AlbumInfo{
				ID:          "album123",
				Name:        "Test Album",
				ReleaseDate: "2023-01-01",
				URI:         "spotify:album:album123",
			},
		},
		{
			name: "album with empty fields",
			input: &SpotifyAlbum{
				ID:          "",
				Name:        "",
				ReleaseDate: "",
				URI:         "",
			},
			expected: &models.AlbumInfo{
				ID:          "",
				Name:        "",
				ReleaseDate: "",
				URI:         "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			result := ParseAlbum(tt.input)
			assert.Equal(tt.expected, result)
		})
	}
}

func TestParseArtist(t *testing.T) {
	tests := []struct {
		name     string
		input    *SpotifyArtist
		expected *models.ArtistInfo
	}{
		{
			name: "artist with genres",
			input: &SpotifyArtist{
				ID:         "artist123",
				Name:       "Test Artist",
				Genres:     []string{"pop", "rock", "indie"},
				Popularity: 80,
				URI:        "spotify:artist:artist123",
			},
			expected: &models.ArtistInfo{
				ID:         "artist123",
				Name:       "Test Artist",
				Genres:     []string{"pop", "rock", "indie"},
				Popularity: 80,
				URI:        "spotify:artist:artist123",
			},
		},
		{
			name: "artist without genres",
			input: &SpotifyArtist{
				ID:         "artist456",
				Name:       "No Genre Artist",
				Genres:     []string{},
				Popularity: 45,
				URI:        "spotify:artist:artist456",
			},
			expected: &models.ArtistInfo{
				ID:         "artist456",
				Name:       "No Genre Artist",
				Genres:     []string{},
				Popularity: 45,
				URI:        "spotify:artist:artist456",
			},
		},
		{
			name: "artist with nil genres",
			input: &SpotifyArtist{
				ID:         "artist789",
				Name:       "Nil Genre Artist",
				Genres:     nil,
				Popularity: 30,
				URI:        "spotify:artist:artist789",
			},
			expected: &models.ArtistInfo{
				ID:         "artist789",
				Name:       "Nil Genre Artist",
				Genres:     nil,
				Popularity: 30,
				URI:        "spotify:artist:artist789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			result := ParseArtist(tt.input)
			assert.Equal(tt.expected, result)
		})
	}
}
