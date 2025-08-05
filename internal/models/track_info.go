package models

// PlaylistTracksInfo contains all aggregated data for a playlist
type PlaylistTracksInfo struct {
	PlaylistID   string
	Tracks       []TrackInfo
	Artists      map[string]ArtistInfo
	APICallCount int
}

// TrackInfo contains all track data needed for routing decisions
type TrackInfo struct {
	ID         string
	Name       string
	URI        string
	DurationMs int
	Popularity int
	Explicit   bool
	Artists    []string
	Album      AlbumInfo
}

type ArtistInfo struct {
	ID         string
	Name       string
	Genres     []string // This is the key data we fetch separately
	Popularity int      // Also enriched from GetSeveralArtists
	URI        string
}

type AlbumInfo struct {
	ID          string
	Name        string
	ReleaseDate string // Already available from playlist tracks
	URI         string
}

func (p *PlaylistTracksInfo) GetAllArtists() []string {
	artists := make(map[string]bool, 0)

	for _, track := range p.Tracks {
		for _, id := range track.Artists {
			artists[id] = true
		}
	}

	uniqueArtistIDs := make([]string, 0, len(artists))
	for key := range artists {
		uniqueArtistIDs = append(uniqueArtistIDs, key)
	}

	return uniqueArtistIDs
}
