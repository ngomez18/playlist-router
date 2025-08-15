package spotifyclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *SpotifyClient) GetSeveralArtists(ctx context.Context, artistIDs []string) ([]*SpotifyArtist, error) {
	if len(artistIDs) == 0 {
		return []*SpotifyArtist{}, nil
	}

	c.logger.InfoContext(ctx, "fetching artists from spotify", "artist_count", len(artistIDs))

	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Join artist IDs with commas
	artistIDsParam := strings.Join(artistIDs, ",")
	params := url.Values{
		"ids": {artistIDsParam},
	}

	path := "artists"
	url := fmt.Sprintf("%s%s?%s", c.apiBaseUrl, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create artists request", "error", err)
		return nil, fmt.Errorf("failed to create artists request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get artists", "error", err)
		return nil, fmt.Errorf("failed to get artists: %w", err)
	}
	defer c.responseBodyCloser(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify artists fetch failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify artists fetch failed (status %d): %s", resp.StatusCode, string(body))
	}

	var artistsResponse struct {
		Artists []*SpotifyArtist `json:"artists"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&artistsResponse); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode artists response", "error", err)
		return nil, fmt.Errorf("failed to decode artists response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully fetched artists", "artists_count", len(artistsResponse.Artists))
	return artistsResponse.Artists, nil
}
