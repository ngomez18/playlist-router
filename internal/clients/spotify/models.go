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

type CreateSpotifyPlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Public      bool   `json:"public"`
}
