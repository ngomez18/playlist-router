package spotifyclient

import "github.com/ngomez18/playlist-router/internal/models"

func ParseSpotifyPlaylist(p *SpotifyPlaylist) *models.SpotifyPlaylist {
	tracks := 0
	if p.Tracks != nil {
		tracks = p.Tracks.Total
	}

	return &models.SpotifyPlaylist{
		ID:     p.ID,
		Name:   p.Name,
		Tracks: tracks,
	}
}

func ParseManySpotifyPlaylist(ps []*SpotifyPlaylist) []*models.SpotifyPlaylist {
	parsed := make([]*models.SpotifyPlaylist, 0, len(ps))
	for _, p := range ps {
		parsed = append(parsed, ParseSpotifyPlaylist(p))
	}

	return parsed
}

func ParsePlaylistTrack(t SpotifyPlaylistTrack) models.TrackInfo {
	artists := make([]string, 0, len(t.Track.Artists))
	for _, a := range t.Track.Artists {
		artists = append(artists, a.ID)
	}

	return models.TrackInfo{
		ID:         t.Track.ID,
		Name:       t.Track.Name,
		URI:        t.Track.URI,
		DurationMs: t.Track.DurationMs,
		Popularity: t.Track.Popularity,
		Explicit:   t.Track.Explicit,
		Album:      *ParseAlbum(&t.Track.Album),
		Artists:    artists,
	}
}

func ParseManyPlaylistTracks(ts []SpotifyPlaylistTrack) []models.TrackInfo {
	parsed := make([]models.TrackInfo, 0, len(ts))
	for _, t := range ts {
		parsed = append(parsed, ParsePlaylistTrack(t))
	}

	return parsed
}

func ParseAlbum(a *SpotifyAlbum) *models.AlbumInfo {
	return &models.AlbumInfo{
		ID:          a.ID,
		Name:        a.Name,
		ReleaseDate: a.ReleaseDate,
		URI:         a.URI,
	}
}

func ParseArtist(a *SpotifyArtist) *models.ArtistInfo {
	return &models.ArtistInfo{
		ID:         a.ID,
		Name:       a.Name,
		Genres:     a.Genres,
		Popularity: a.Popularity,
		URI:        a.URI,
	}
}
