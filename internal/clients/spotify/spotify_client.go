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
)

//go:generate mockgen -source=spotify_client.go -destination=mocks/mock_spotify_client.go -package=mocks

type SpotifyAPI interface {
	GenerateAuthURL(state string) string
	ExchangeCodeForTokens(ctx context.Context, code string) (*SpotifyTokenResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*SpotifyTokenResponse, error)
	GetUserProfile(ctx context.Context, accessToken string) (*SpotifyUserProfile, error)
	GetAllUserPlaylists(ctx context.Context, accessToken string) ([]*SpotifyPlaylist, error)
	GetPlaylist(ctx context.Context, accessToken string, playlistId string) (*SpotifyPlaylist, error)
	CreatePlaylist(ctx context.Context, accessToken, userId, name, description string, public bool) (*SpotifyPlaylist, error)
	DeletePlaylist(ctx context.Context, accessToken, userId, playlistId string) error
	UpdatePlaylist(ctx context.Context, accessToken, userId, playlistId, name, description string) error
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

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

	var profile SpotifyUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode profile response", "error", err)
		return nil, fmt.Errorf("failed to decode profile response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully fetched user profile", "user_id", profile.ID, "email", profile.Email)
	return &profile, nil
}

func (c *SpotifyClient) GetUserPlaylists(ctx context.Context, accessToken string, limit, offset int) (*SpotifyPlaylistResponse, error) {
	c.logger.InfoContext(ctx, "fetching user playlists from spotify")

	params := url.Values{
		"limit":  {fmt.Sprint(limit)},
		"offset": {fmt.Sprint(offset)},
	}

	path := "me/playlists"
	url := fmt.Sprintf("%s%s?%s", c.apiBaseUrl, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create user playlists request", "error", err)
		return nil, fmt.Errorf("failed to create user playlists request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get user playlists", "error", err)
		return nil, fmt.Errorf("failed to get user playlists: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify user playlists fetch failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify user playlists fetch failed (status %d): %s", resp.StatusCode, string(body))
	}

	var playlists SpotifyPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode playlists response", "error", err)
		return nil, fmt.Errorf("failed to decode playlists response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully fetched user playlists")
	return &playlists, nil
}

func (c *SpotifyClient) GetAllUserPlaylists(ctx context.Context, accessToken string) ([]*SpotifyPlaylist, error) {
	c.logger.InfoContext(ctx, "fetching all user playlists from spotify")

	allPlaylists := make([]*SpotifyPlaylist, 0)
	limit := 50 // Spotify API max limit
	offset := 0

	for {
		response, err := c.GetUserPlaylists(ctx, accessToken, limit, offset)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to fetch playlists batch", "offset", offset, "error", err)
			return nil, fmt.Errorf("failed to fetch playlists batch at offset %d: %w", offset, err)
		}

		allPlaylists = append(allPlaylists, response.Items...)

		// Break if we have all the items according to the total
		if len(allPlaylists) >= response.Total || len(response.Items) == 0 {
			break
		}

		offset += limit
	}

	c.logger.InfoContext(ctx, "successfully fetched all user playlists", "total_count", len(allPlaylists))
	return allPlaylists, nil
}

func (c *SpotifyClient) GetPlaylist(ctx context.Context, accessToken string, playlistId string) (*SpotifyPlaylist, error) {
	c.logger.InfoContext(ctx, "fetching playlist from spotify")

	path := fmt.Sprintf("playlists/%s", playlistId)
	url := fmt.Sprintf("%s%s", c.apiBaseUrl, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create playlist request", "error", err)
		return nil, fmt.Errorf("failed to create playlist request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get playlist", "error", err)
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify playlist fetch failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify playlist fetch failed (status %d): %s", resp.StatusCode, string(body))
	}

	var playlists SpotifyPlaylist
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode playlists response", "error", err)
		return nil, fmt.Errorf("failed to decode playlists response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully fetched playlist")
	return &playlists, nil
}

func (c *SpotifyClient) CreatePlaylist(ctx context.Context, accessToken, userId, name, description string, public bool) (*SpotifyPlaylist, error) {
	c.logger.InfoContext(ctx, "creating playlist in spotify", "user_id", userId, "name", name)

	path := fmt.Sprintf("users/%s/playlists", userId)
	url := fmt.Sprintf("%s%s", c.apiBaseUrl, path)

	requestBody := SpotifyPlaylistRequest{
		Name:        &name,
		Description: &description,
		Public:      &public,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to marshal playlist request", "error", err)
		return nil, fmt.Errorf("failed to marshal playlist request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create playlist request", "error", err)
		return nil, fmt.Errorf("failed to create playlist request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create playlist", "error", err, "body", string(jsonData))
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify playlist creation failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify playlist creation failed (status %d): %s", resp.StatusCode, string(body))
	}

	var playlist SpotifyPlaylist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode playlist response", "error", err)
		return nil, fmt.Errorf("failed to decode playlist response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully created playlist", "playlist_id", playlist.ID, "name", playlist.Name)
	return &playlist, nil
}

func (c *SpotifyClient) DeletePlaylist(ctx context.Context, accessToken, userId, playlistId string) error {
	c.logger.InfoContext(ctx, "deleting playlist from spotify", "user_id", userId, "playlist_id", playlistId)

	path := fmt.Sprintf("playlists/%s/followers", playlistId)
	url := fmt.Sprintf("%s%s", c.apiBaseUrl, path)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create delete playlist request", "error", err)
		return fmt.Errorf("failed to create delete playlist request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to delete playlist", "error", err)
		return fmt.Errorf("failed to delete playlist: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify playlist deletion failed", "status_code", resp.StatusCode, "response_body", string(body))
		return fmt.Errorf("spotify playlist deletion failed (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.InfoContext(ctx, "successfully deleted playlist", "playlist_id", playlistId)
	return nil
}

func (c *SpotifyClient) UpdatePlaylist(ctx context.Context, accessToken, userId, playlistId, name, description string) error {
	c.logger.InfoContext(ctx, "updating playlist in spotify", "user_id", userId, "playlist_id", playlistId, "name", name)

	path := fmt.Sprintf("playlists/%s", playlistId)
	url := fmt.Sprintf("%s%s", c.apiBaseUrl, path)

	requestBody := SpotifyPlaylistRequest{}

	if name != "" {
		requestBody.Name = &name
	}

	if description != "" {
		requestBody.Description = &description
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to marshal update playlist request", "error", err)
		return fmt.Errorf("failed to marshal update playlist request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create update playlist request", "error", err)
		return fmt.Errorf("failed to create update playlist request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to update playlist", "error", err, "body", string(jsonData))
		return fmt.Errorf("failed to update playlist: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify playlist update failed", "status_code", resp.StatusCode, "response_body", string(body))
		return fmt.Errorf("spotify playlist update failed (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.InfoContext(ctx, "successfully updated playlist", "playlist_id", playlistId)
	return nil
}
