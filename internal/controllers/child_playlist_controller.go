package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
)

type ChildPlaylistController struct {
	childPlaylistService services.ChildPlaylistServicer
	validator            *validator.Validate
}

func NewChildPlaylistController(cpService services.ChildPlaylistServicer) *ChildPlaylistController {
	return &ChildPlaylistController{
		childPlaylistService: cpService,
		validator:            validator.New(),
	}
}

func (c *ChildPlaylistController) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateChildPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract user ID from auth context
	user, ok := requestcontext.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	// Extract base playlist ID from URL path
	basePlaylistID := r.PathValue("basePlaylistID")
	if basePlaylistID == "" {
		http.Error(w, "base playlist ID is required", http.StatusBadRequest)
		return
	}

	newChildPlaylist, err := c.childPlaylistService.CreateChildPlaylist(r.Context(), user.ID, basePlaylistID, &req)
	if err != nil {
		http.Error(w, "unable to create child playlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newChildPlaylist); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (c *ChildPlaylistController) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth context
	user, ok := requestcontext.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	// Extract child playlist ID from URL path
	childPlaylistID := r.PathValue("id")
	if childPlaylistID == "" {
		http.Error(w, "child playlist ID is required", http.StatusBadRequest)
		return
	}

	childPlaylist, err := c.childPlaylistService.GetChildPlaylist(r.Context(), childPlaylistID, user.ID)
	if err != nil {
		http.Error(w, "child playlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(childPlaylist); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (c *ChildPlaylistController) GetByBasePlaylistID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth context
	user, ok := requestcontext.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	// Extract base playlist ID from URL path
	basePlaylistID := r.PathValue("basePlaylistID")
	if basePlaylistID == "" {
		http.Error(w, "base playlist ID is required", http.StatusBadRequest)
		return
	}

	childPlaylists, err := c.childPlaylistService.GetChildPlaylistsByBasePlaylistID(r.Context(), basePlaylistID, user.ID)
	if err != nil {
		http.Error(w, "unable to retrieve child playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(childPlaylists); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (c *ChildPlaylistController) Update(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateChildPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract user ID from auth context
	user, ok := requestcontext.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	// Extract child playlist ID from URL path
	childPlaylistID := r.PathValue("id")
	if childPlaylistID == "" {
		http.Error(w, "child playlist ID is required", http.StatusBadRequest)
		return
	}

	updatedChildPlaylist, err := c.childPlaylistService.UpdateChildPlaylist(r.Context(), childPlaylistID, user.ID, &req)
	if err != nil {
		http.Error(w, "unable to update child playlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedChildPlaylist); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (c *ChildPlaylistController) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth context
	user, ok := requestcontext.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	// Extract child playlist ID from URL path
	childPlaylistID := r.PathValue("id")
	if childPlaylistID == "" {
		http.Error(w, "child playlist ID is required", http.StatusBadRequest)
		return
	}

	err := c.childPlaylistService.DeleteChildPlaylist(r.Context(), childPlaylistID, user.ID)
	if err != nil {
		http.Error(w, "unable to delete child playlist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
