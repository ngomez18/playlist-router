package models

import "time"

// SpotifyIntegration represents a user's Spotify account integration
type SpotifyIntegration struct {
	ID      string    `json:"id" db:"id"`
	Created time.Time `json:"created" db:"created"`
	Updated time.Time `json:"updated" db:"updated"`

	// Foreign key to users collection
	UserID string `json:"user_id" db:"user"`

	// Spotify account details
	SpotifyID string `json:"spotify_id" db:"spotify_id"`

	// Authentication tokens (hidden from JSON responses)
	AccessToken  string    `json:"-" db:"access_token"`
	RefreshToken string    `json:"-" db:"refresh_token"`
	TokenType    string    `json:"-" db:"token_type"`
	ExpiresAt    time.Time `json:"-" db:"expires_at"`
	Scope        string    `json:"-" db:"scope"`

	// Additional Spotify profile info
	DisplayName string `json:"display_name" db:"display_name"`
}

type SpotifyIntegrationTokenRefresh struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}
