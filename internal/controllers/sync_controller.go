package controllers

import (
	"encoding/json"
	"net/http"

	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/orchestrators"
)

type SyncController struct {
	syncOrchestrator orchestrators.SyncOrchestrator
}

func NewSyncController(syncOrchestrator orchestrators.SyncOrchestrator) *SyncController {
	return &SyncController{
		syncOrchestrator: syncOrchestrator,
	}
}

func (c *SyncController) SyncBasePlaylist(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth context
	user, ok := requestcontext.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	basePlaylistID := r.PathValue("basePlaylistID")
	if basePlaylistID == "" {
		http.Error(w, "base playlist ID is required", http.StatusBadRequest)
		return
	}

	syncEvent, err := c.syncOrchestrator.SyncBasePlaylist(r.Context(), user.ID, basePlaylistID)
	if err != nil {
		// Check if it's a sync already in progress error
		if err.Error() == "sync already in progress for base playlist "+basePlaylistID {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, "failed to sync base playlist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(syncEvent); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
