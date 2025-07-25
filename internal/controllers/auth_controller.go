package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/ngomez18/playlist-router/internal/services"
)

type AuthController struct {
	authService services.AuthServicer
}

func NewAuthController(authService services.AuthServicer) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) SpotifyLogin(w http.ResponseWriter, r *http.Request) {
	// Generate random state for CSRF protection
	state := generateState()

	// Store state in session/cookie for validation (TODO: implement proper state storage)

	authURL := c.authService.GenerateSpotifyAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (c *AuthController) SpotifyCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "authorization code is required", http.StatusBadRequest)
		return
	}

	// TODO: Validate state parameter against stored value

	// Handle OAuth callback
	result, err := c.authService.HandleSpotifyCallback(r.Context(), code, state)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func generateState() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
