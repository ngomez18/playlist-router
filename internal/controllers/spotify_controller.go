package controllers

import (
	"encoding/json"
	"net/http"

	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/services"
)

type SpotifyController struct {
	spotifyApiService services.SpotifyAPIServicer
}

func NewSpotifyController(spotifyApiService services.SpotifyAPIServicer) *SpotifyController {
	return &SpotifyController{
		spotifyApiService: spotifyApiService,
	}
}

func (c *SpotifyController) GetUserPlaylists(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth context
	user, ok := requestcontext.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	playlists, err := c.spotifyApiService.GetUserPlaylists(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "unable to retrieve spotify playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(playlists)
	if err != nil {
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}
