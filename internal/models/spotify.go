package models

type SpotifyPlaylist struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Tracks int    `json:"tracks"`
}
