package models

import "time"

type ChildPlaylist struct {
	ID                string          `json:"id"`
	BasePlaylist      string          `json:"base_playlist" validate:"required"`
	Name              string          `json:"name" validate:"required,min=1,max=100"`
	SpotifyPlaylistID string          `json:"spotify_playlist_id" validate:"required"`
	FilterRules       *FilterRules    `json:"filter_rules,omitempty"`
	IsActive          bool            `json:"is_active"`
	Created           time.Time       `json:"created"`
	Updated           time.Time       `json:"updated"`
}

type FilterRules struct {
	Metadata      *MetadataFilters     `json:"metadata,omitempty"`
}

type MetadataFilters struct {
	Genres     []string     `json:"genres,omitempty"`
	YearRange  *RangeFilter `json:"year_range,omitempty"`
	DurationMS *RangeFilter `json:"duration_ms,omitempty"`
}

type RangeFilter struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type CreateChildPlaylistRequest struct {
	BasePlaylist      string          `json:"base_playlist" validate:"required"`
	Name              string          `json:"name" validate:"required,min=1,max=100"`
	SpotifyPlaylistID string          `json:"spotify_playlist_id" validate:"required"`
	FilterRules       *FilterRules    `json:"filter_rules,omitempty"`
	IsActive          *bool           `json:"is_active,omitempty"`
}

type UpdateChildPlaylistRequest struct {
	Name              *string         `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	SpotifyPlaylistID *string         `json:"spotify_playlist_id,omitempty"`
	FilterRules       *FilterRules    `json:"filter_rules,omitempty"`
	IsActive          *bool           `json:"is_active,omitempty"`
}

type ChildPlaylistResponse struct {
	ID                string          `json:"id"`
	BasePlaylist      string          `json:"base_playlist"`
	Name              string          `json:"name"`
	SpotifyPlaylistID string          `json:"spotify_playlist_id"`
	FilterRules       *FilterRules    `json:"filter_rules"`
	IsActive          bool            `json:"is_active"`
	SyncStatus        string          `json:"sync_status"`
	SongsCount        int             `json:"songs_count"`
	Created           time.Time       `json:"created"`
	Updated           time.Time       `json:"updated"`
}
