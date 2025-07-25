package config

import "errors"

var (
	ErrMissingSpotifyClientID     = errors.New("SPOTIFY_CLIENT_ID environment variable is required")
	ErrMissingSpotifyClientSecret = errors.New("SPOTIFY_CLIENT_SECRET environment variable is required")
	ErrMissingSpotifyRedirectURI  = errors.New("SPOTIFY_REDIRECT_URI environment variable is required")
	ErrMissingEncryptionKey       = errors.New("ENCRYPTION_KEY environment variable is required")
)
