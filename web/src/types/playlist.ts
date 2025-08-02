export interface BasePlaylist {
  id: string
  user_id: string
  name: string
  spotify_playlist_id: string
  is_active: boolean
  created: string
  updated: string
}

export interface CreateBasePlaylistRequest {
  name: string
  spotify_playlist_id?: string
}

// Audio Feature Filter Types
export interface RangeFilter {
  min?: number
  max?: number
}

export interface SetFilter {
  include?: string[]
  exclude?: string[]
}

export interface MetadataFilters {
  // Track Information
  duration_ms?: RangeFilter
  popularity?: RangeFilter
  explicit?: boolean // true = explicit only, false = clean only, undefined = both
  
  // Artist & Album Information
  genres?: SetFilter
  release_year?: RangeFilter
  artist_popularity?: RangeFilter
  
  // Search-based Filters
  track_keywords?: SetFilter // Keywords to search for in track names
  artist_keywords?: SetFilter // Keywords to search for in artist names
}

// Child Playlist Types
export interface ChildPlaylist {
  id: string
  user_id: string
  base_playlist_id: string
  name: string
  description?: string
  spotify_playlist_id: string
  filter_rules?: MetadataFilters
  is_active: boolean
  created: string
  updated: string
}

export interface CreateChildPlaylistRequest {
  name: string
  description?: string
  filter_rules?: MetadataFilters
}

export interface UpdateChildPlaylistRequest {
  name?: string
  description?: string
  filter_rules?: MetadataFilters
  is_active?: boolean
}