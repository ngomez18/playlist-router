package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/ngomez18/playlist-router/internal/middleware"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
)

type BasePlaylistController struct {
	basePlaylistService services.BasePlaylistServicer
	validator           *validator.Validate
}

func NewBasePlaylistController(bpService services.BasePlaylistServicer) *BasePlaylistController {
	return &BasePlaylistController{
		basePlaylistService: bpService,
		validator:           validator.New(),
	}
}

func (c *BasePlaylistController) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBasePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract user ID from auth context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	newBasePlaylist, err := c.basePlaylistService.CreateBasePlaylist(r.Context(), user.ID, &req)
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

func (c *BasePlaylistController) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	basePlaylistId := r.PathValue("id")

	if basePlaylistId == "" {
		http.Error(w, "playlist id is required", http.StatusBadRequest)
		return
	}

	// Extract user ID from auth context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	err := c.basePlaylistService.DeleteBasePlaylist(r.Context(), basePlaylistId, user.ID)
	if err != nil {
		http.Error(w, "unable to delete base playlist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *BasePlaylistController) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	basePlaylistId := r.PathValue("id")

	if basePlaylistId == "" {
		http.Error(w, "playlist id is required", http.StatusBadRequest)
		return
	}

	// Extract user ID from auth context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	basePlaylist, err := c.basePlaylistService.GetBasePlaylist(r.Context(), basePlaylistId, user.ID)
	if err != nil {
		http.Error(w, "unable to retrieve base playlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(basePlaylist)
	if err != nil {
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}

func (c *BasePlaylistController) GetByUserID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	basePlaylists, err := c.basePlaylistService.GetBasePlaylistsByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "unable to retrieve base playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(basePlaylists)
	if err != nil {
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}
