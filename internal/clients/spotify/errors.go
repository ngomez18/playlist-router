package spotifyclient

import "errors"

var (
	ErrSpotifyCredentialsNotFound = errors.New("spotify credentials not found in context")
)
