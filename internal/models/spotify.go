package models

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
