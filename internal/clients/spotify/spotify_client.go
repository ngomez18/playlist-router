package spotifyclient

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

	"github.com/ngomez18/playlist-router/internal/clients"
	"github.com/ngomez18/playlist-router/internal/config"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
)

const (
	MAX_PLAYLISTS = 50
)

//go:generate mockgen -source=spotify_client.go -destination=mocks/mock_spotify_client.go -package=mocks

type SpotifyAPI interface {
	// Auth
	GenerateAuthURL(state string) string
	ExchangeCodeForTokens(ctx context.Context, code string) (*SpotifyTokenResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*SpotifyTokenResponse, error)
	GetUserProfile(ctx context.Context, accessToken string) (*SpotifyUserProfile, error)

	// Playlists
	GetPlaylist(ctx context.Context, playlistId string) (*SpotifyPlaylist, error)
	GetAllUserPlaylists(ctx context.Context) ([]*SpotifyPlaylist, error)
	CreatePlaylist(ctx context.Context, accessToken, userId, name, description string, public bool) (*SpotifyPlaylist, error)
	DeletePlaylist(ctx context.Context, accessToken, userId, playlistId string) error
	UpdatePlaylist(ctx context.Context, accessToken, userId, playlistId, name, description string) error

	// Tracks
}

type SpotifyClient struct {
	HttpClient clients.HTTPClient
	config     *config.AuthConfig
	logger     *slog.Logger

	// urls
	authBaseUrl string
	apiBaseUrl  string
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
		config:      config,
		logger:      logger.With("component", "SpotifyClient"),
		authBaseUrl: "https://accounts.spotify.com/",
		apiBaseUrl:  "https://api.spotify.com/v1/",
	}
}

func (c *SpotifyClient) GenerateAuthURL(state string) string {
	path := "authorize"
	params := url.Values{
		"client_id":     {c.config.SpotifyClientID},
		"response_type": {"code"},
		"redirect_uri":  {c.config.SpotifyRedirectURI},
		"scope":         {"user-read-email playlist-read-private playlist-modify-public playlist-modify-private"},
		"state":         {state},
	}

	url := fmt.Sprintf("%s%s", c.authBaseUrl, path)
	authURL := fmt.Sprintf("%s?%s", url, params.Encode())
	c.logger.Info("generated spotify auth URL", "state", state)
	return authURL
}

func (c *SpotifyClient) ExchangeCodeForTokens(ctx context.Context, code string) (*SpotifyTokenResponse, error) {
	c.logger.InfoContext(ctx, "exchanging authorization code for tokens")
	path := "api/token"

	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.config.SpotifyRedirectURI},
	}

	url := fmt.Sprintf("%s%s", c.authBaseUrl, path)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
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
	defer c.responseBodyCloser(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify token exchange failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokens SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode token response", "error", err)
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully exchanged code for tokens")
	return &tokens, nil
}

func (c *SpotifyClient) RefreshTokens(ctx context.Context, refreshToken string) (*SpotifyTokenResponse, error) {
	c.logger.InfoContext(ctx, "refreshing spotify access tokens")
	path := "api/token"

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	url := fmt.Sprintf("%s%s", c.authBaseUrl, path)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create token refresh request", "error", err)
		return nil, fmt.Errorf("failed to create token refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.config.SpotifyClientID, c.config.SpotifyClientSecret)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to refresh tokens", "error", err)
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}
	defer c.responseBodyCloser(ctx, resp)


	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify token refresh failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify token refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokens SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode token refresh response", "error", err)
		return nil, fmt.Errorf("failed to decode token refresh response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully refreshed spotify tokens")
	return &tokens, nil
}

func (c *SpotifyClient) GetUserProfile(ctx context.Context, accessToken string) (*SpotifyUserProfile, error) {
	c.logger.InfoContext(ctx, "fetching user profile from spotify")
	path := "me"
	url := fmt.Sprintf("%s%s", c.apiBaseUrl, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
	defer c.responseBodyCloser(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify profile fetch failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify profile fetch failed (status %d): %s", resp.StatusCode, string(body))
	}

	var profile SpotifyUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode profile response", "error", err)
		return nil, fmt.Errorf("failed to decode profile response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully fetched user profile", "user_id", profile.ID, "email", profile.Email)
	return &profile, nil
}

func (c *SpotifyClient) responseBodyCloser(ctx context.Context, resp *http.Response) {
	if closeErr := resp.Body.Close(); closeErr != nil {
		c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
	}
}

func (c *SpotifyClient) getAccessToken(ctx context.Context) (string, error) {
	// Get the user's Spotify integration to get the access token
	integration, ok := requestcontext.GetSpotifyAuthFromContext(ctx)
	if !ok {
		c.logger.ErrorContext(ctx, "failed to get spotify integration")
		return "", ErrSpotifyCredentialsNotFound
	}

	return integration.AccessToken, nil
}