package models

import "time"

type BasePlaylist struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id" validate:"required"`
	Name              string     `json:"name" validate:"required,min=1,max=100"`
	SpotifyPlaylistID string     `json:"spotify_playlist_id" validate:"required"`
	IsActive          bool       `json:"is_active"`
	Created           time.Time  `json:"created"`
	Updated           time.Time  `json:"updated"`
}

type CreateBasePlaylistRequest struct {
	Name              string `json:"name" validate:"required,min=1,max=100"`
	SpotifyPlaylistID string `json:"spotify_playlist_id" validate:"required"`
	IsActive          *bool  `json:"is_active,omitempty"`
}
