package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ngomez18/playlist-router/internal/clients"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
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
	userService               UserServicer
	spotifyIntegrationService SpotifyIntegrationServicer
	spotifyClient             clients.SpotifyAPI
	logger                    *slog.Logger
}

func NewAuthService(
	userService UserServicer,
	spotifyIntegrationService SpotifyIntegrationServicer,
	spotifyClient clients.SpotifyAPI,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userService:               userService,
		spotifyIntegrationService: spotifyIntegrationService,
		spotifyClient:             spotifyClient,
		logger:                    logger.With("component", "AuthService"),
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

	// Generate PocketBase JWT token for this user
	token, err := s.userService.GenerateAuthToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	return &AuthResult{
		User:         user,
		Token:        token,
		RefreshToken: "", // PocketBase handles its own refresh
	}, nil
}

func (s *AuthService) createOrUpdateUser(ctx context.Context, profile *models.SpotifyUserProfile, tokens *models.SpotifyTokenResponse) (*models.AuthUser, error) {
	user, err := s.findUserBySpotifyID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return s.createNewUser(ctx, profile, tokens)
	}

	return s.updateExistingUser(ctx, user, profile, tokens)
}

func (s *AuthService) findUserBySpotifyID(ctx context.Context, spotifyID string) (*models.User, error) {
	s.logger.InfoContext(ctx, "finding user by spotify ID", "spotify_id", spotifyID)

	// First, try to find the Spotify integration
	integration, err := s.spotifyIntegrationService.GetIntegrationBySpotifyID(ctx, spotifyID)
	if err == repositories.ErrSpotifyIntegrationNotFound {
		// Check if it's a "not found" error
		s.logger.InfoContext(ctx, "spotify integration not found", "spotify_id", spotifyID)
		return nil, nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch spotify integration", "spotify_id", spotifyID, "error", err.Error())
		return nil, err
	}

	// If integration found, get the associated user
	user, err := s.userService.GetUserByID(ctx, integration.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch user for spotify integration", "user_id", integration.UserID, "spotify_id", spotifyID, "error", err.Error())
		return nil, err
	}

	s.logger.InfoContext(ctx, "found user by spotify ID", "user_id", user.ID, "spotify_id", spotifyID)
	return user, nil
}

func (s *AuthService) createNewUser(
	ctx context.Context,
	profile *models.SpotifyUserProfile,
	tokens *models.SpotifyTokenResponse,
) (*models.AuthUser, error) {
	s.logger.InfoContext(ctx, "creating new user from spotify profile", "spotify_id", profile.ID, "email", profile.Email)

	// Create user from Spotify profile
	user := &models.User{
		Email: profile.Email,
		Name:  profile.Name,
	}

	// Create the user in the database
	createdUser, err := s.userService.CreateUser(ctx, user)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create user", "spotify_id", profile.ID, "error", err.Error())
		return nil, err
	}

	// Calculate expiration time from ExpiresIn seconds
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	// Create Spotify integration with tokens
	integration := &models.SpotifyIntegration{
		SpotifyID:    profile.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    expiresAt,
		Scope:        tokens.Scope,
		DisplayName:  profile.Name,
	}

	createdIntegration, err := s.spotifyIntegrationService.CreateOrUpdateIntegration(ctx, createdUser.ID, integration)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create spotify integration", "user_id", createdUser.ID, "spotify_id", profile.ID, "error", err.Error())
		return nil, err
	}

	// Create AuthUser response
	authUser := createdUser.ToAuthUser(createdIntegration)
	s.logger.InfoContext(ctx, "new user created successfully", "user_id", createdUser.ID, "spotify_id", profile.ID)

	return authUser, nil
}

func (s *AuthService) updateExistingUser(
	ctx context.Context,
	user *models.User,
	profile *models.SpotifyUserProfile,
	tokens *models.SpotifyTokenResponse,
) (*models.AuthUser, error) {
	s.logger.InfoContext(ctx, "updating existing user from spotify profile", "user_id", user.ID, "spotify_id", profile.ID, "email", profile.Email)

	// Update user profile information if it has changed
	var updatedUser *models.User
	if user.Email != profile.Email || user.Name != profile.Name {
		s.logger.InfoContext(ctx, "user profile data changed, updating user", "user_id", user.ID, "old_email", user.Email, "new_email", profile.Email)

		userToUpdate := &models.User{
			ID:    user.ID,
			Email: profile.Email,
			Name:  profile.Name,
		}

		var err error
		updatedUser, err = s.userService.UpdateUser(ctx, userToUpdate)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to update user profile", "user_id", user.ID, "error", err.Error())
			return nil, fmt.Errorf("failed to update user profile: %w", err)
		}
	} else {
		// No changes needed, use existing user
		updatedUser = user
	}

	// Calculate new expiration time from tokens
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	// Update Spotify integration with new tokens and profile info
	integrationToUpdate := &models.SpotifyIntegration{
		SpotifyID:    profile.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    expiresAt,
		Scope:        tokens.Scope,
		DisplayName:  profile.Name,
	}

	updatedIntegration, err := s.spotifyIntegrationService.CreateOrUpdateIntegration(ctx, updatedUser.ID, integrationToUpdate)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update spotify integration", "user_id", updatedUser.ID, "spotify_id", profile.ID, "error", err.Error())
		return nil, fmt.Errorf("failed to update spotify integration: %w", err)
	}

	// Create AuthUser response
	authUser := updatedUser.ToAuthUser(updatedIntegration)

	s.logger.InfoContext(ctx, "existing user updated successfully", "user_id", updatedUser.ID, "spotify_id", profile.ID)

	return authUser, nil
}
