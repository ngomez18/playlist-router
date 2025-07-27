package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/clients"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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
	user, err := s.createOrUpdateUser(ctx, profile, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}

	return &AuthResult{
		User:         user,
		Token:        "",
		RefreshToken: "", // PocketBase handles its own refresh
	}, nil
}

func (s *AuthService) createOrUpdateUser(ctx context.Context, profile *models.SpotifyUserProfile, tokens *models.SpotifyTokenResponse) (*models.AuthUser, error) {
	user, spotifyIntegration, err := s.findUserBySpotifyID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	if user == nil && spotifyIntegration == nil {
		return s.createNewUser(ctx, profile, tokens)
	}

	return s.updateExistingUser(ctx, user, spotifyIntegration, profile, tokens)
}

func (s *AuthService) findUserBySpotifyID(ctx context.Context, spotifyID string) (*models.User, *models.SpotifyIntegration, error) {
	return nil, nil, nil
}

func (s *AuthService) createNewUser(
	ctx context.Context, 
	profile *models.SpotifyUserProfile, 
	tokens *models.SpotifyTokenResponse,
) (*models.AuthUser, error) {
	return nil, nil
}

func (s *AuthService) updateExistingUser(
	ctx context.Context, 
	user *models.User, 
	integration *models.SpotifyIntegration, 
	profile *models.SpotifyUserProfile,
	tokens *models.SpotifyTokenResponse,
) (*models.AuthUser, error) {
	return nil, nil
}

func (s *AuthService) generateAuthToken(userRecord *core.Record) (string, error) {
	// Generate PocketBase auth token for the user
	// Note: Using a simple approach - in production should use proper JWT with expiration
	token := userRecord.Id + "_auth_token"

	s.logger.Info("generated auth token", "user_id", userRecord.Id)
	return token, nil
}
