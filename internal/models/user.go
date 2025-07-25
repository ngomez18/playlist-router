package models

import "time"

// User represents a user in the PocketBase users collection
type User struct {
	ID       string    `json:"id" db:"id"`
	Username string    `json:"username" db:"username"`
	Email    string    `json:"email" db:"email"`
	Name     string    `json:"name" db:"name"`
	Avatar   string    `json:"avatar" db:"avatar"`
	Created  time.Time `json:"created" db:"created"`
	Updated  time.Time `json:"updated" db:"updated"`
}

// ToAuthUser converts a User to an AuthUser for API responses
func (u *User) ToAuthUser(spotifyIntegration *SpotifyIntegration) *AuthUser {
	authUser := &AuthUser{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	}
	
	if spotifyIntegration != nil {
		authUser.SpotifyID = spotifyIntegration.SpotifyID
	}
	
	return authUser
}

// AuthUser represents user data returned in authentication responses
// This excludes sensitive fields like tokens
type AuthUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	SpotifyID string `json:"spotify_id"`
}
