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

func (c *SpotifyClient) GetPlaylistTracks(ctx context.Context, playlistID string, limit, offset int) (*SpotifyPlaylistTracksResponse, error) {
	c.logger.InfoContext(ctx, "fetching playlist tracks from spotify", "playlist_id", playlistID, "limit", limit, "offset", offset)

	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"limit":  {fmt.Sprint(limit)},
		"offset": {fmt.Sprint(offset)},
		// "fields": {"items(track(id,name,duration_ms,popularity,explicit,uri,artists(id,name,genres,popularity,uri),album(id,name,release_date,uri))),total,limit,offset,next"},
	}

	path := fmt.Sprintf("playlists/%s/tracks", playlistID)
	url := fmt.Sprintf("%s%s?%s", c.apiBaseUrl, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create playlist tracks request", "error", err)
		return nil, fmt.Errorf("failed to create playlist tracks request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get playlist tracks", "error", err)
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}
	defer c.responseBodyCloser(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify playlist tracks fetch failed", "status_code", resp.StatusCode, "response_body", string(body))
		return nil, fmt.Errorf("spotify playlist tracks fetch failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tracksResponse SpotifyPlaylistTracksResponse
	if err := json.NewDecoder(resp.Body).Decode(&tracksResponse); err != nil {
		c.logger.ErrorContext(ctx, "failed to decode playlist tracks response", "error", err)
		return nil, fmt.Errorf("failed to decode playlist tracks response: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully fetched playlist tracks", "playlist_id", playlistID, "tracks_count", len(tracksResponse.Items), "total", tracksResponse.Total)
	return &tracksResponse, nil
}

func (c *SpotifyClient) AddTracksToPlaylist(ctx context.Context, playlistID string, trackURIs []string) error {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}

	c.logger.InfoContext(ctx, "adding tracks to playlist", 
		"playlist_id", playlistID, 
		"track_count", len(trackURIs),
	)

	path := fmt.Sprintf("playlists/%s/tracks", playlistID)
	url := fmt.Sprintf("%s%s", c.apiBaseUrl, path)

	requestBody := map[string][]string{
		"uris": trackURIs,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to marshal add tracks request", "error", err)
		return fmt.Errorf("failed to marshal add tracks request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create add tracks request", "error", err)
		return fmt.Errorf("failed to create add tracks request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to add tracks to playlist", "error", err)
		return fmt.Errorf("failed to add tracks to playlist: %w", err)
	}
	defer c.responseBodyCloser(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "spotify add tracks failed", 
			"status_code", resp.StatusCode, 
			"response_body", string(body),
			"playlist_id", playlistID,
		)
		return fmt.Errorf("spotify add tracks failed (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.InfoContext(ctx, "successfully added tracks to playlist", 
		"playlist_id", playlistID, 
		"tracks_added", len(trackURIs),
	)
	return nil
}