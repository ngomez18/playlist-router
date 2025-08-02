package models

type MetadataFilters struct {
	// Track Information
	Duration   *RangeFilter `json:"duration_ms,omitempty"`
	Popularity *RangeFilter `json:"popularity,omitempty"`
	Explicit   *bool        `json:"explicit,omitempty"` // true = explicit only, false = clean only, nil = both

	// Artist & Album Information
	Genres           *SetFilter   `json:"genres,omitempty"`
	ReleaseYear      *RangeFilter `json:"release_year,omitempty"`
	ArtistPopularity *RangeFilter `json:"artist_popularity,omitempty"`

	// Search-based Filters
	TrackKeywords  *SetFilter `json:"track_keywords,omitempty"`  // Keywords to search for in track names
	ArtistKeywords *SetFilter `json:"artist_keywords,omitempty"` // Keywords to search for in artist names
}

// Legacy type alias for backward compatibility during transition
type AudioFeatureFilters = MetadataFilters

type RangeFilter struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type SetFilter struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}
