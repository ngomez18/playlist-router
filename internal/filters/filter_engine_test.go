package filters

import (
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewFilterEngine(t *testing.T) {
	t.Run("nil filter rules", func(t *testing.T) {
		playlist := &models.ChildPlaylist{FilterRules: nil}
		engine := NewFilterEngine(playlist)

		assert.NotNil(t, engine)
		assert.Empty(t, engine.filters)
	})

	t.Run("with filter rules", func(t *testing.T) {
		playlist := &models.ChildPlaylist{
			FilterRules: &models.MetadataFilters{
				Duration:   &models.RangeFilter{Min: float64Ptr(120000)},
				Popularity: &models.RangeFilter{Max: float64Ptr(80)},
			},
		}
		engine := NewFilterEngine(playlist)

		assert.NotNil(t, engine)
		assert.Len(t, engine.filters, 8) // All filter types are created
	})
}

func TestFilterEngine_MatchTrack(t *testing.T) {
	t.Run("empty filter engine matches all", func(t *testing.T) {
		engine := &FilterEngine{filters: []Filter{}}
		track := models.TrackInfo{DurationMs: 180000}

		assert.True(t, engine.MatchTrack(track))
	})

	t.Run("all filters pass", func(t *testing.T) {
		playlist := &models.ChildPlaylist{
			FilterRules: &models.MetadataFilters{
				Duration:   &models.RangeFilter{Min: float64Ptr(120000), Max: float64Ptr(240000)},
				Popularity: &models.RangeFilter{Min: float64Ptr(50), Max: float64Ptr(90)},
				Explicit:   boolPtr(false),
			},
		}
		engine := NewFilterEngine(playlist)

		track := models.TrackInfo{
			DurationMs: 180000,
			Popularity: 70,
			Explicit:   false,
		}

		assert.True(t, engine.MatchTrack(track))
	})

	t.Run("one filter fails", func(t *testing.T) {
		playlist := &models.ChildPlaylist{
			FilterRules: &models.MetadataFilters{
				Duration:   &models.RangeFilter{Min: float64Ptr(120000), Max: float64Ptr(240000)},
				Popularity: &models.RangeFilter{Min: float64Ptr(80), Max: float64Ptr(90)}, // This will fail
			},
		}
		engine := NewFilterEngine(playlist)

		track := models.TrackInfo{
			DurationMs: 180000,
			Popularity: 50, // Below minimum
		}

		assert.False(t, engine.MatchTrack(track))
	})

	t.Run("complex filter combination", func(t *testing.T) {
		playlist := &models.ChildPlaylist{
			FilterRules: &models.MetadataFilters{
				Duration:    &models.RangeFilter{Min: float64Ptr(180000)},
				Genres:      &models.SetFilter{Include: []string{"rock", "pop"}},
				ReleaseYear: &models.RangeFilter{Min: float64Ptr(2000)},
				Explicit:    boolPtr(false),
			},
		}
		engine := NewFilterEngine(playlist)

		passingTrack := models.TrackInfo{
			DurationMs:  200000,
			AllGenres:   []string{"rock", "alternative"},
			ReleaseYear: 2010,
			Explicit:    false,
		}

		failingTrack := models.TrackInfo{
			DurationMs:  200000,
			AllGenres:   []string{"jazz", "blues"}, // No rock/pop
			ReleaseYear: 2010,
			Explicit:    false,
		}

		assert.True(t, engine.MatchTrack(passingTrack))
		assert.False(t, engine.MatchTrack(failingTrack))
	})
}
