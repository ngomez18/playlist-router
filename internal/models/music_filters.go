package models

type AudioFeatureFilters struct {
	// Musical Qualities
	Energy       *RangeFilter `json:"energy,omitempty"`
	Danceability *RangeFilter `json:"danceability,omitempty"`
	Valence      *RangeFilter `json:"valence,omitempty"`
	Tempo        *RangeFilter `json:"tempo,omitempty"`
	Liveness     *RangeFilter `json:"liveness,omitempty"`
	Speechiness  *RangeFilter `json:"speechiness,omitempty"`

	// Technical Attributes
	Acousticness     *RangeFilter `json:"acousticness,omitempty"`
	Instrumentalness *RangeFilter `json:"instrumentalness,omitempty"`
	Loudness         *RangeFilter `json:"loudness,omitempty"`
	Key              *SetFilter   `json:"key,omitempty"`
	Mode             *SetFilter   `json:"mode,omitempty"`
	TimeSignature    *SetFilter   `json:"time_signature,omitempty"`

	// Context & Metadata
	Duration    *RangeFilter `json:"duration_ms,omitempty"`
	Popularity  *RangeFilter `json:"popularity,omitempty"`
	Genres      *SetFilter   `json:"genres,omitempty"`
	ReleaseYear *RangeFilter `json:"release_year,omitempty"`
}

type RangeFilter struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type SetFilter struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}
