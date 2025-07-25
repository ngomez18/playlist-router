package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ngomez18/playlist-router/internal/config"
	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=spotify_client.go -destination=mocks/mock_spotify_client.go -package=mocks

type SpotifyAPI interface {
	GenerateAuthURL(state string) string
	ExchangeCodeForTokens(ctx context.Context, code string) (*models.SpotifyTokenResponse, error)
	GetUserProfile(ctx context.Context, accessToken string) (*models.SpotifyUserProfile, error)
}

type SpotifyClient struct {
	HttpClient HTTPClient
	config     *config.AuthConfig
	logger     *slog.Logger
}

func NewSpotifyClient(config *config.AuthConfig, logger *slog.Logger) *SpotifyClient {
	return &SpotifyClient{
		HttpClient: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
		config: config,
		logger: logger.With("component", "SpotifyClient"),
	}
}

func (c *SpotifyClient) GenerateAuthURL(state string) string {
	baseURL := "https://accounts.spotify.com/authorize"
	params := url.Values{
		"client_id":     {c.config.SpotifyClientID},
		"response_type": {"code"},
		"redirect_uri":  {c.config.SpotifyRedirectURI},
		"scope":         {"user-read-email playlist-read-private playlist-modify-public playlist-modify-private"},
		"state":         {state},
	}

	authURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	c.logger.Info("generated spotify auth URL", "state", state)
	return authURL
}

func (c *SpotifyClient) ExchangeCodeForTokens(ctx context.Context, code string) (*models.SpotifyTokenResponse, error) {
	c.logger.InfoContext(ctx, "exchanging authorization code for tokens")
	tokenURL := "https://accounts.spotify.com/api/token"

	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.config.SpotifyRedirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create token request", "error", err)
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.config.SpotifyClientID, c.config.SpotifyClientSecret)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to exchange code", "error", err)
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify token exchange failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokens models.SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode token response", "error", err)
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully exchanged code for tokens")
	return &tokens, nil
}

func (c *SpotifyClient) GetUserProfile(ctx context.Context, accessToken string) (*models.SpotifyUserProfile, error) {
	c.logger.InfoContext(ctx, "fetching user profile from spotify")
	profileURL := "https://api.spotify.com/v1/me"

	req, err := http.NewRequestWithContext(ctx, "GET", profileURL, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create profile request", "error", err)
		return nil, fmt.Errorf("failed to create profile request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get user profile", "error", err)
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify profile fetch failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify profile fetch failed (status %d): %s", resp.StatusCode, string(body))
	}

	var profile models.SpotifyUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode profile response", "error", err)
		return nil, fmt.Errorf("failed to decode profile response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully fetched user profile", "user_id", profile.ID, "email", profile.Email)
	return &profile, nil
}
