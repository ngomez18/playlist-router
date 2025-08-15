package filters

import "github.com/ngomez18/playlist-router/internal/models"

type FilterEngine struct {
	filters []Filter
}

func NewFilterEngine(playlist *models.ChildPlaylist) *FilterEngine {
	if playlist.FilterRules == nil {
		return &FilterEngine{filters: []Filter{}}
	}

	filters := []Filter{
		&DurationFilter{playlist.FilterRules.Duration},
		&PopularityFilter{playlist.FilterRules.Popularity},
		&ExplicitFilter{playlist.FilterRules.Explicit},
		&GenresFilter{playlist.FilterRules.Genres},
		&ReleaseYearFilter{playlist.FilterRules.ReleaseYear},
		&ArtistPopularityFilter{playlist.FilterRules.ArtistPopularity},
		&TrackKeywordsFilter{playlist.FilterRules.TrackKeywords},
		&ArtistKeywordsFilter{playlist.FilterRules.ArtistKeywords},
	}

	return &FilterEngine{filters: filters}
}

func (eng *FilterEngine) MatchTrack(track models.TrackInfo) bool {
	for _, filter := range eng.filters {
		if ok := filter.Matches(track); !ok {
			return false
		}
	}

	return true
}
