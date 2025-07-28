package config

type AuthConfig struct {
	SpotifyClientID     string `env:"SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret string `env:"SPOTIFY_CLIENT_SECRET"`
	SpotifyRedirectURI  string `env:"SPOTIFY_REDIRECT_URI"`
	EncryptionKey       string `env:"ENCRYPTION_KEY"`
	FrontendURL         string `env:"FRONTEND_URL" envDefault:"http://localhost:5173"`
}

func (c *AuthConfig) Validate() error {
	if c.SpotifyClientID == "" {
		return ErrMissingSpotifyClientID
	}
	if c.SpotifyClientSecret == "" {
		return ErrMissingSpotifyClientSecret
	}
	if c.SpotifyRedirectURI == "" {
		return ErrMissingSpotifyRedirectURI
	}
	if c.EncryptionKey == "" {
		return ErrMissingEncryptionKey
	}
	return nil
}
