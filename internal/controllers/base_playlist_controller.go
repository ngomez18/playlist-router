package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
)

type BasePlaylistController struct {
	basePlaylistService services.BasePlaylistServicer
}

func NewBasePlaylistController(bpService services.BasePlaylistServicer) *BasePlaylistController {
	return &BasePlaylistController{
		basePlaylistService: bpService,
	}
}

func (c *BasePlaylistController) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBasePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	newBasePlaylist, err := c.basePlaylistService.CreateBasePlaylist(r.Context(), &req)
	if err != nil {
		http.Error(w, "unable to create base playlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newBasePlaylist)
	if err != nil {
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}
