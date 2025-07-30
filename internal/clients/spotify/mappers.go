package spotifyclient

import "github.com/ngomez18/playlist-router/internal/models"

func ParseSpotifyPlaylist(p *SpotifyPlaylist) *models.SpotifyPlaylist {
	tracks := 0
	if p.Tracks != nil {
		tracks = p.Tracks.Total
	}
	
	return &models.SpotifyPlaylist{
		ID: p.ID,
		Name: p.Name,
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