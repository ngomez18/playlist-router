package repositories

import "errors"

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrInvalidBasePlaylist = errors.New("invalid base playlist")
)