package filters

import (
	"slices"
	"strings"

	"github.com/ngomez18/playlist-router/internal/models"
)

type Filter interface {
	Matches(track models.TrackInfo) bool
}

type DurationFilter struct {
	*models.RangeFilter
}

func (f *DurationFilter) Matches(track models.TrackInfo) bool {
	return matchesRangeFilter(f.RangeFilter, float64(track.DurationMs))
}

type PopularityFilter struct {
	*models.RangeFilter
}

func (f *PopularityFilter) Matches(track models.TrackInfo) bool {
	return matchesRangeFilter(f.RangeFilter, float64(track.Popularity))
}

type ExplicitFilter struct {
	RequireExplicit *bool
}

func (f *ExplicitFilter) Matches(track models.TrackInfo) bool {
	return matchesBoolFilter(f.RequireExplicit, track.Explicit)
}

type GenresFilter struct {
	*models.SetFilter
}

func (f *GenresFilter) Matches(track models.TrackInfo) bool {
	return matchesSetFilterValues(f.SetFilter, track.AllGenres)
}

type ReleaseYearFilter struct {
	*models.RangeFilter
}

func (f *ReleaseYearFilter) Matches(track models.TrackInfo) bool {
	return matchesRangeFilter(f.RangeFilter, float64(track.ReleaseYear))
}

type ArtistPopularityFilter struct {
	*models.RangeFilter
}

func (f *ArtistPopularityFilter) Matches(track models.TrackInfo) bool {
	return matchesRangeFilter(f.RangeFilter, float64(track.MaxArtistPop))
}

type TrackKeywordsFilter struct {
	*models.SetFilter
}

func (f *TrackKeywordsFilter) Matches(track models.TrackInfo) bool {
	return matchesSetFilterText(f.SetFilter, strings.ToLower(track.Name))
}

type ArtistKeywordsFilter struct {
	*models.SetFilter
}

func (f *ArtistKeywordsFilter) Matches(track models.TrackInfo) bool {
	artistNamesText := strings.ToLower(strings.Join(track.ArtistNames, " "))
	return matchesSetFilterText(f.SetFilter, artistNamesText)
}

// filter matcher functions

func matchesRangeFilter(filter *models.RangeFilter, value float64) bool {
	if filter == nil {
		return true
	}

	minInRange := filter.Min == nil || value >= *filter.Min
	maxInRange := filter.Max == nil || value <= *filter.Max

	return minInRange && maxInRange
}

func matchesBoolFilter(filter *bool, value bool) bool {
	if filter == nil {
		return true
	}

	return *filter == value
}

func matchesSetFilterValues(filter *models.SetFilter, values []string) bool {
	if filter == nil {
		return true
	}

	normalizedValues := make([]string, len(values))
	for i, v := range values {
		normalizedValues[i] = strings.ToLower(v)
	}

	if slices.ContainsFunc(filter.Exclude, func(excludeValue string) bool {
		return slices.Contains(normalizedValues, strings.ToLower(excludeValue))
	}) {
		return false
	}

	if len(filter.Include) > 0 {
		return slices.ContainsFunc(filter.Include, func(includeValue string) bool {
			return slices.Contains(normalizedValues, strings.ToLower(includeValue))
		})
	}

	return true
}

func matchesSetFilterText(filter *models.SetFilter, text string) bool {
	if filter == nil {
		return true
	}

	if slices.ContainsFunc(filter.Exclude, func(excludeKeyword string) bool {
		return strings.Contains(text, strings.ToLower(excludeKeyword))
	}) {
		return false
	}

	if len(filter.Include) > 0 {
		return slices.ContainsFunc(filter.Include, func(includeKeyword string) bool {
			return strings.Contains(text, strings.ToLower(includeKeyword))
		})
	}

	return true
}
