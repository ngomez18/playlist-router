package models

import "time"

// SyncStatus represents the status of a sync operation
type SyncStatus string

const (
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"
)

// SyncEvent tracks sync operations
type SyncEvent struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id" validate:"required"`
	BasePlaylistID   string     `json:"base_playlist_id" validate:"required"`
	ChildPlaylistIDs []string   `json:"child_playlist_ids"`
	Status           SyncStatus `json:"status"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	// Sync statistics
	TracksProcessed  int `json:"tracks_processed"`
	TotalAPIRequests int `json:"total_api_requests"`
}