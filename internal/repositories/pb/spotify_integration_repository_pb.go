package pb

import (
	"context"
	"log/slog"
	"time"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type SpotifyIntegrationRepositoryPocketbase struct {
	collection Collection
	app        *pocketbase.PocketBase
	log        *slog.Logger
}

func NewSpotifyIntegrationRepositoryPocketbase(pb *pocketbase.PocketBase) *SpotifyIntegrationRepositoryPocketbase {
	return &SpotifyIntegrationRepositoryPocketbase{
		app:        pb,
		collection: CollectionSpotifyIntegration,
		log:        pb.Logger().With("component", "SpotifyIntegrationRepositoryPocketbase"),
	}
}

func (siRepo *SpotifyIntegrationRepositoryPocketbase) CreateOrUpdate(
	ctx context.Context,
	userId string,
	integration *models.SpotifyIntegration,
) (*models.SpotifyIntegration, error) {
	collection, err := siRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	var record *core.Record
	existing, err := siRepo.app.FindFirstRecordByFilter(
		collection,
		"user = {:user}",
		dbx.Params{"user": userId},
	)
	if err != nil {
		siRepo.log.InfoContext(ctx, "spotify_integration not found", "user", userId)
		record = core.NewRecord(collection)
		record.Set("user", userId)
	} else {
		siRepo.log.InfoContext(ctx, "spotify_integration found", "user", userId, "record", record)
		record = existing
	}

	record.Set("spotify_id", integration.SpotifyID)
	record.Set("access_token", integration.AccessToken)
	record.Set("refresh_token", integration.RefreshToken)
	record.Set("token_type", integration.TokenType)
	record.Set("expires_at", integration.ExpiresAt)
	record.Set("scope", integration.Scope)
	record.Set("display_name", integration.DisplayName)

	if err := siRepo.app.Save(record); err != nil {
		siRepo.log.ErrorContext(ctx, "unable to store spotify_integration record", "error", err)
		return nil, repositories.ErrDatabaseOperation
	}

	siRepo.log.InfoContext(ctx, "spotify_integration stored successfully", "user", userId, "spotify_id", integration.SpotifyID)
	return recordToSpotifyIntegration(record), nil
}

func (siRepo *SpotifyIntegrationRepositoryPocketbase) GetByUserID(ctx context.Context, userId string) (*models.SpotifyIntegration, error) {
	collection, err := siRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record, err := siRepo.app.FindFirstRecordByFilter(
		collection,
		"user = {:user}",
		dbx.Params{"user": userId},
	)
	if err != nil {
		siRepo.log.ErrorContext(ctx, "unable to fetch spotify_integration", "user", userId, "error", err)
		return nil, repositories.ErrSpotifyIntegrationNotFound
	}

	siRepo.log.InfoContext(ctx, "spotify_integration found", "user", userId, "spotify_id", record.Id)
	return recordToSpotifyIntegration(record), nil
}

func (siRepo *SpotifyIntegrationRepositoryPocketbase) GetBySpotifyID(ctx context.Context, spotifyId string) (*models.SpotifyIntegration, error) {
	collection, err := siRepo.getCollection(ctx)
	if err != nil {
		return nil, err
	}

	record, err := siRepo.app.FindFirstRecordByFilter(
		collection,
		"spotify_id = {:spotify_id}",
		dbx.Params{"spotify_id": spotifyId},
	)
	if err != nil {
		siRepo.log.ErrorContext(ctx, "unable to fetch spotify_integration", "spotify_id", spotifyId, "error", err)
		return nil, repositories.ErrSpotifyIntegrationNotFound
	}

	siRepo.log.InfoContext(ctx, "spotify_integration found", "spotify_id", record.Id)
	return recordToSpotifyIntegration(record), nil
}

func (siRepo *SpotifyIntegrationRepositoryPocketbase) UpdateTokens(
	ctx context.Context,
	integrationId string,
	tokens *models.SpotifyIntegrationTokenRefresh,
) error {
	collection, err := siRepo.getCollection(ctx)
	if err != nil {
		return err
	}

	record, err := siRepo.app.FindRecordById(collection, integrationId)
	if err != nil {
		siRepo.log.ErrorContext(ctx, "unable to fetch spotify_integration", "integration_id", integrationId, "error", err)
		return repositories.ErrSpotifyIntegrationNotFound
	}

	record.Set("access_token", tokens.AccessToken)

	if tokens.RefreshToken != "" {
		record.Set("refresh_token", tokens.RefreshToken)
	}

	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	record.Set("expires_at", expiresAt)

	if err := siRepo.app.Save(record); err != nil {
		siRepo.log.ErrorContext(ctx, "unable to update spotify_integration", "integration_id", integrationId, "error", err)
		return repositories.ErrDatabaseOperation
	}

	siRepo.log.InfoContext(ctx, "spotify_integration tokens updated", "spotify_id", record.Id)
	return nil
}

func (siRepo *SpotifyIntegrationRepositoryPocketbase) Delete(ctx context.Context, userId string) error {
	collection, err := siRepo.getCollection(ctx)
	if err != nil {
		return err
	}

	record, err := siRepo.app.FindFirstRecordByFilter(
		collection,
		"user = {:user}",
		map[string]any{"user": userId},
	)
	if err != nil {
		siRepo.log.ErrorContext(ctx, "spotify_integration not found", "user", userId, "error", err)
		return repositories.ErrSpotifyIntegrationNotFound
	}

	if err := siRepo.app.Delete(record); err != nil {
		siRepo.log.ErrorContext(ctx, "unable to delete spotify_integration", "user", userId, "integration_id", record.Id, "error", err)
		return repositories.ErrDatabaseOperation
	}

	siRepo.log.InfoContext(ctx, "spotify_integration deleted", "user", userId, "integration_id", record.Id)
	return nil
}

func (siRepo *SpotifyIntegrationRepositoryPocketbase) getCollection(ctx context.Context) (*core.Collection, error) {
	collection, err := siRepo.app.FindCollectionByNameOrId(string(siRepo.collection))
	if err != nil {
		siRepo.log.ErrorContext(ctx, "unable to find collection", "collection", siRepo.collection, "error", err)
		return nil, repositories.ErrCollectionNotFound
	}

	return collection, nil
}

func recordToSpotifyIntegration(record *core.Record) *models.SpotifyIntegration {
	return &models.SpotifyIntegration{
		ID:           record.Id,
		UserID:       record.GetString("user"),
		SpotifyID:    record.GetString("spotify_id"),
		AccessToken:  record.GetString("access_token"),
		RefreshToken: record.GetString("refresh_token"),
		TokenType:    record.GetString("token_type"),
		ExpiresAt:    record.GetDateTime("expires_at").Time(),
		Scope:        record.GetString("scope"),
		DisplayName:  record.GetString("display_name"),
		Created:      record.GetDateTime("created").Time(),
		Updated:      record.GetDateTime("updated").Time(),
	}
}
