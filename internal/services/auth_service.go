package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/clients"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/pocketbase/pocketbase"
)

//go:generate mockgen -source=auth_service.go -destination=mocks/mock_auth_service.go -package=mocks

type AuthServicer interface {
	GenerateSpotifyAuthURL(state string) string
	HandleSpotifyCallback(ctx context.Context, code, state string) (*AuthResult, error)
}

type AuthResult struct {
	User         *models.AuthUser `json:"user"`
	Token        string           `json:"token"`
	RefreshToken string           `json:"refresh_token"`
}

type AuthService struct {
	app           *pocketbase.PocketBase
	spotifyClient clients.SpotifyAPI
	logger        *slog.Logger
}

func NewAuthService(app *pocketbase.PocketBase, spotifyClient clients.SpotifyAPI, logger *slog.Logger) *AuthService {
	return &AuthService{
		app:           app,
		spotifyClient: spotifyClient,
		logger:        logger.With("component", "AuthService"),
	}
}

func (s *AuthService) GenerateSpotifyAuthURL(state string) string {
	authURL := s.spotifyClient.GenerateAuthURL(state)
	s.logger.Info("generated spotify auth url", "state", state)
	return authURL
}

func (s *AuthService) HandleSpotifyCallback(ctx context.Context, code, state string) (*AuthResult, error) {
	s.logger.InfoContext(ctx, "handling spotify callback", "code", code, "state", state)

	// Exchange code for tokens
	tokens, err := s.spotifyClient.ExchangeCodeForTokens(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}

	// Get user profile from Spotify
	profile, err := s.spotifyClient.GetUserProfile(ctx, tokens.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Create or update user in PocketBase
	user, pbToken, err := s.createOrUpdateUser(ctx, profile, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}

	return &AuthResult{
		User:         user,
		Token:        pbToken,
		RefreshToken: "", // PocketBase handles its own refresh
	}, nil
}

func (s *AuthService) createOrUpdateUser(ctx context.Context, profile *models.SpotifyUserProfile, tokens *models.SpotifyTokenResponse) (*models.AuthUser, string, error) {
	// TODO: Implement user creation/update in PocketBase
	// TODO: Encrypt and store Spotify tokens
	// TODO: Generate PocketBase auth token

	user := &models.AuthUser{
		ID:        "temp_id", // Will be PocketBase record ID
		Email:     profile.Email,
		Name:      profile.Name,
		SpotifyID: profile.ID,
	}

	pbToken := "temp_token" // Will be actual PocketBase token

	s.logger.InfoContext(ctx, "created/updated user", "spotify_id", profile.ID, "email", profile.Email)

	return user, pbToken, nil
}
