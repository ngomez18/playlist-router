package repositories

import "errors"

var (
	// General DB errors
	ErrDatabaseOperation  = errors.New("unable to complete db operation")
	ErrCollectionNotFound = errors.New("collection not found")
	ErrUnauthorized       = errors.New("user can not access this resource")

	// User errors
	ErrUseNotFound = errors.New("user not found")

	// Base playlist errors
	ErrBasePlaylistNotFound = errors.New("base playlist not found")

	// Child playlist errors
	ErrChildPlaylistNotFound = errors.New("child playlist not found")

	// Spotify integration errors
	ErrSpotifyIntegrationNotFound = errors.New("spotify integration not found")

	// Sync event errors
	ErrSyncEventNotFound = errors.New("sync event not found")
)
