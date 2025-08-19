package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

const (
	MAX_TRACKS  = 50
	MAX_ARTISTS = 50
)

//go:generate mockgen -source=track_aggregator_service.go -destination=mocks/mock_track_aggregator_service.go -package=mocks

type TrackAggregatorServicer interface {
	AggregatePlaylistData(ctx context.Context, userID, basePlaylistID string) (*models.PlaylistTracksInfo, error)
}

type TrackAggregatorService struct {
	spotifyClient    spotifyclient.SpotifyAPI
	basePlaylistRepo repositories.BasePlaylistRepository
	logger           *slog.Logger
}

func NewTrackAggregatorService(spotifyClient spotifyclient.SpotifyAPI, basePlaylistRepo repositories.BasePlaylistRepository, log *slog.Logger) *TrackAggregatorService {
	return &TrackAggregatorService{
		spotifyClient:    spotifyClient,
		basePlaylistRepo: basePlaylistRepo,
		logger:           log,
	}
}

func (taService *TrackAggregatorService) AggregatePlaylistData(ctx context.Context, userID, basePlaylistID string) (*models.PlaylistTracksInfo, error) {
	taService.logger.InfoContext(ctx, "aggregating playlist data", "user", userID, "base_playlist", basePlaylistID)

	basePlaylist, err := taService.basePlaylistRepo.GetByID(ctx, basePlaylistID, userID)
	if err != nil {
		taService.logger.ErrorContext(ctx, "failed to fetch base playlist", "error", err.Error())
		return nil, fmt.Errorf("failed to fetch base playlist: %w", err)
	}

	tracks, err := taService.getAllPlaylistTracks(ctx, basePlaylist.SpotifyPlaylistID)
	if err != nil {
		taService.logger.ErrorContext(ctx, "failed to fetch playlist tracks", "error", err.Error())
		return nil, fmt.Errorf("failed to fetch playlist tracks: %w", err)
	}

	taService.logger.InfoContext(
		ctx,
		"successfully fetched all playlist tracks",
		"user", userID,
		"base_playlist", basePlaylistID,
		"tracks", len(tracks.Tracks),
	)

	artistInfo, apiCallCount, err := taService.getAllPlaylistArtists(ctx, tracks.GetAllArtists())
	if err != nil {
		taService.logger.ErrorContext(ctx, "failed to fetch playlist artists", "error", err.Error())
		return nil, fmt.Errorf("failed to fetch playlist artists: %w", err)
	}

	tracks.Artists = artistInfo
	tracks.APICallCount = tracks.APICallCount + apiCallCount
	tracks.PlaylistID = basePlaylistID
	tracks.UserID = userID

	// Pre-process tracks for efficient filtering
	taService.preprocessTracksForFiltering(tracks)

	taService.logger.InfoContext(
		ctx,
		"successfully fetched all playlist artists",
		"user", userID,
		"base_playlist", basePlaylistID,
		"artists", len(tracks.Artists),
	)

	return tracks, nil
}

func (taService *TrackAggregatorService) getAllPlaylistTracks(ctx context.Context, playlistID string) (*models.PlaylistTracksInfo, error) {
	playlistTracks := models.PlaylistTracksInfo{Tracks: make([]models.TrackInfo, 0)}
	offset := 0

	for {
		tracksResp, err := taService.spotifyClient.GetPlaylistTracks(ctx, playlistID, MAX_TRACKS, offset)
		if err != nil {
			taService.logger.ErrorContext(ctx, "failed to fetch playlist tracks", "error", err.Error())
			return nil, fmt.Errorf("failed to fetch playlist tracks: %w", err)
		}

		playlistTracks.Tracks = append(playlistTracks.Tracks, spotifyclient.ParseManyPlaylistTracks(tracksResp.Items)...)
		playlistTracks.APICallCount++

		if tracksResp.Next == nil {
			break
		}

		offset += MAX_TRACKS
	}

	return &playlistTracks, nil
}

func (taService *TrackAggregatorService) getAllPlaylistArtists(ctx context.Context, artistIDs []string) (map[string]models.ArtistInfo, int, error) {
	artists := make(map[string]models.ArtistInfo, len(artistIDs))
	apiCallCount := 0

	for offset := 0; offset < len(artistIDs); offset += MAX_ARTISTS {
		endIndex := min(offset+MAX_ARTISTS, len(artistIDs))
		artistsResp, err := taService.spotifyClient.GetSeveralArtists(ctx, artistIDs[offset:endIndex])
		if err != nil {
			taService.logger.ErrorContext(ctx, "failed to fetch playlist artists", "error", err.Error())
			return nil, apiCallCount, fmt.Errorf("failed to fetch playlist artists: %w", err)
		}

		for _, artist := range artistsResp {
			artists[artist.ID] = *spotifyclient.ParseArtist(artist)
		}

		apiCallCount++
	}

	return artists, apiCallCount, nil
}

func (taService *TrackAggregatorService) preprocessTracksForFiltering(playlistData *models.PlaylistTracksInfo) {
	for i := range playlistData.Tracks {
		track := &playlistData.Tracks[i]

		// Extract release year from album release date
		track.ReleaseYear = taService.parseReleaseYear(track.Album.ReleaseDate)

		// Collect all genres from track's artists
		genreSet := make(map[string]bool)
		maxArtistPop := 0
		artistNames := make([]string, 0, len(track.Artists))

		for _, artistID := range track.Artists {
			if artist, exists := playlistData.Artists[artistID]; exists {
				// Collect normalized genres
				for _, genre := range artist.Genres {
					genreSet[strings.ToLower(genre)] = true
				}

				// Track max artist popularity
				if artist.Popularity > maxArtistPop {
					maxArtistPop = artist.Popularity
				}

				// Collect artist names
				artistNames = append(artistNames, artist.Name)
			}
		}

		// Convert genre set to slice
		track.AllGenres = make([]string, 0, len(genreSet))
		for genre := range genreSet {
			track.AllGenres = append(track.AllGenres, genre)
		}

		track.MaxArtistPop = maxArtistPop
		track.ArtistNames = artistNames
	}
}

func (taService *TrackAggregatorService) parseReleaseYear(releaseDate string) int {
	if releaseDate == "" {
		return 0
	}

	if len(releaseDate) >= 4 {
		yearStr := releaseDate[:4]
		if year, err := strconv.Atoi(yearStr); err == nil {
			return year
		}
	}

	return 0
}
