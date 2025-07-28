package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ngomez18/playlist-router/internal/config"
	"github.com/ngomez18/playlist-router/internal/middleware"
	"github.com/ngomez18/playlist-router/internal/services"
)

type AuthController struct {
	authService services.AuthServicer
	config      *config.Config
}

func NewAuthController(authService services.AuthServicer, config *config.Config) *AuthController {
	return &AuthController{
		authService: authService,
		config:      config,
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

	// Redirect to frontend with token as URL parameter
	redirectURL := fmt.Sprintf("%s/?token=%s", c.config.Auth.FrontendURL, result.Token)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (c *AuthController) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// This endpoint is protected by auth middleware, so user is already validated
	// and available in context. Just return the user.
	user, found := middleware.GetUserFromContext(r.Context())
	if !found {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func generateState() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
