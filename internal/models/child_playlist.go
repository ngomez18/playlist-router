package models

import (
	"fmt"
	"strings"
	"time"
)

type ChildPlaylist struct {
	ID                string               `json:"id"`
	UserID            string               `json:"user_id" validate:"required"`
	BasePlaylistID    string               `json:"base_playlist_id" validate:"required"`
	Name              string               `json:"name" validate:"required,min=1,max=100"`
	Description       string               `json:"description,omitempty"`
	SpotifyPlaylistID string               `json:"spotify_playlist_id" validate:"required"`
	FilterRules       *AudioFeatureFilters `json:"filter_rules,omitempty"`
	IsActive          bool                 `json:"is_active"`
	Created           time.Time            `json:"created"`
	Updated           time.Time            `json:"updated"`
}

type CreateChildPlaylistRequest struct {
	Name        string               `json:"name" validate:"required,min=1,max=100"`
	Description string               `json:"description,omitempty"`
	FilterRules *AudioFeatureFilters `json:"filter_rules,omitempty"`
}

type UpdateChildPlaylistRequest struct {
	Name        *string              `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string              `json:"description,omitempty"`
	FilterRules *AudioFeatureFilters `json:"filter_rules,omitempty"`
	IsActive    *bool                `json:"is_active,omitempty"`
	SyncEnabled *bool                `json:"sync_enabled,omitempty"`
}

func BuildChildPlaylistName(basePlaylistName, childPlaylistName string) string {
	return fmt.Sprintf("[%s] > %s", basePlaylistName, childPlaylistName)
}

func BuildChildPlaylistDescription(description string) string {
	descriptionItems := []string{
		"****************************************************",
		"* PLAYLIST GENERATED AND MANAGED BY PlaylistRouter *",
		"*********** PLEASE DO NOT EDIT OR DELETE ***********",
		"****************************************************",
		description,
	}

	return strings.Join(descriptionItems, "\n")
}
