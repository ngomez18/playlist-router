package filters

import (
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDurationFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   *models.RangeFilter
		duration int
		expected bool
	}{
		{"nil filter", nil, 180000, true},
		{"within range", &models.RangeFilter{Min: float64Ptr(120000), Max: float64Ptr(240000)}, 180000, true},
		{"below min", &models.RangeFilter{Min: float64Ptr(200000), Max: nil}, 180000, false},
		{"above max", &models.RangeFilter{Min: nil, Max: float64Ptr(120000)}, 180000, false},
		{"min only", &models.RangeFilter{Min: float64Ptr(100000), Max: nil}, 180000, true},
		{"max only", &models.RangeFilter{Min: nil, Max: float64Ptr(200000)}, 180000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &DurationFilter{tt.filter}
			track := models.TrackInfo{DurationMs: tt.duration}
			assert.Equal(t, tt.expected, filter.Matches(track))
		})
	}
}

func TestPopularityFilter(t *testing.T) {
	filter := &PopularityFilter{&models.RangeFilter{Min: float64Ptr(50), Max: float64Ptr(80)}}

	assert.True(t, filter.Matches(models.TrackInfo{Popularity: 65}))
	assert.False(t, filter.Matches(models.TrackInfo{Popularity: 30}))
	assert.False(t, filter.Matches(models.TrackInfo{Popularity: 90}))
}

func TestExplicitFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   *bool
		explicit bool
		expected bool
	}{
		{"nil filter", nil, true, true},
		{"require explicit - explicit track", boolPtr(true), true, true},
		{"require explicit - clean track", boolPtr(true), false, false},
		{"require clean - clean track", boolPtr(false), false, true},
		{"require clean - explicit track", boolPtr(false), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &ExplicitFilter{tt.filter}
			track := models.TrackInfo{Explicit: tt.explicit}
			assert.Equal(t, tt.expected, filter.Matches(track))
		})
	}
}

func TestGenresFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   *models.SetFilter
		genres   []string
		expected bool
	}{
		{"nil filter", nil, []string{"rock", "pop"}, true},
		{"include match", &models.SetFilter{Include: []string{"rock", "jazz"}}, []string{"rock", "pop"}, true},
		{"include no match", &models.SetFilter{Include: []string{"jazz", "blues"}}, []string{"rock", "pop"}, false},
		{"exclude match", &models.SetFilter{Exclude: []string{"rock"}}, []string{"rock", "pop"}, false},
		{"exclude no match", &models.SetFilter{Exclude: []string{"jazz"}}, []string{"rock", "pop"}, true},
		{"case insensitive", &models.SetFilter{Include: []string{"ROCK"}}, []string{"rock"}, true},
		{"empty genres", &models.SetFilter{Include: []string{"rock"}}, []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &GenresFilter{tt.filter}
			track := models.TrackInfo{AllGenres: tt.genres}
			assert.Equal(t, tt.expected, filter.Matches(track))
		})
	}
}

func TestReleaseYearFilter(t *testing.T) {
	filter := &ReleaseYearFilter{&models.RangeFilter{Min: float64Ptr(2000), Max: float64Ptr(2020)}}

	assert.True(t, filter.Matches(models.TrackInfo{ReleaseYear: 2010}))
	assert.False(t, filter.Matches(models.TrackInfo{ReleaseYear: 1995}))
	assert.False(t, filter.Matches(models.TrackInfo{ReleaseYear: 2025}))
}

func TestArtistPopularityFilter(t *testing.T) {
	filter := &ArtistPopularityFilter{&models.RangeFilter{Min: float64Ptr(70), Max: nil}}

	assert.True(t, filter.Matches(models.TrackInfo{MaxArtistPop: 80}))
	assert.False(t, filter.Matches(models.TrackInfo{MaxArtistPop: 50}))
}

func TestTrackKeywordsFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    *models.SetFilter
		trackName string
		expected  bool
	}{
		{"nil filter", nil, "Love Song", true},
		{"include match", &models.SetFilter{Include: []string{"love"}}, "Love Song", true},
		{"include no match", &models.SetFilter{Include: []string{"hate"}}, "Love Song", false},
		{"exclude match", &models.SetFilter{Exclude: []string{"love"}}, "Love Song", false},
		{"exclude no match", &models.SetFilter{Exclude: []string{"hate"}}, "Love Song", true},
		{"case insensitive", &models.SetFilter{Include: []string{"LOVE"}}, "love song", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TrackKeywordsFilter{tt.filter}
			track := models.TrackInfo{Name: tt.trackName}
			assert.Equal(t, tt.expected, filter.Matches(track))
		})
	}
}

func TestArtistKeywordsFilter(t *testing.T) {
	filter := &ArtistKeywordsFilter{&models.SetFilter{Include: []string{"beatles"}}}

	track := models.TrackInfo{ArtistNames: []string{"The Beatles", "John Lennon"}}
	assert.True(t, filter.Matches(track))

	track2 := models.TrackInfo{ArtistNames: []string{"Led Zeppelin", "Pink Floyd"}}
	assert.False(t, filter.Matches(track2))
}

// Helper functions
func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}
