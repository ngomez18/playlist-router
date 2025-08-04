package spotifyclient

type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type SpotifyUserProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"display_name"`
}

type SpotifyPlaylistResponse struct {
	Total int                `json:"total"`
	Items []*SpotifyPlaylist `json:"items"`
}

type SpotifyPlaylist struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	URI           string                  `json:"uri"`
	Public        bool                    `json:"public"`
	Collaborative bool                    `json:"collaborative"`
	Description   string                  `json:"description"`
	Href          string                  `json:"href"`
	Images        []*SpotifyPlaylistImage `json:"images"`
	Tracks        *SpotifyPlaylistTracks  `json:"tracks"`
	SnapshotID    string                  `json:"snapshot_id"`
}

type SpotifyPlaylistImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type SpotifyPlaylistTracks struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

type SpotifyPlaylistRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Public      *bool   `json:"public,omitempty"`
}

type SpotifyPlaylistTracksResponse struct {
	Items  []SpotifyPlaylistTrack `json:"items"`
	Total  int                    `json:"total"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
	Next   *string                `json:"next"`
}

type SpotifyPlaylistTrack struct {
	Track *SpotifyTrack `json:"track"`
}

type SpotifyTrack struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	DurationMs int             `json:"duration_ms"`
	Popularity int             `json:"popularity"`
	Explicit   bool            `json:"explicit"`
	Artists    []SpotifyArtist `json:"artists"`
	Album      SpotifyAlbum    `json:"album"`
	URI        string          `json:"uri"`
}

type SpotifyArtist struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Genres     []string `json:"genres"`
	Popularity int      `json:"popularity"`
	URI        string   `json:"uri"`
}

type SpotifyAlbum struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ReleaseDate string `json:"release_date"`
	URI         string `json:"uri"`
}
