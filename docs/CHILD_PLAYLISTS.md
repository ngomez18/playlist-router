# Child Playlist Management

## Overview

Child playlists are filtered playlists automatically populated with songs from their parent base playlist. Users define filtering rules to automatically categorize songs into themed collections (e.g., "High Energy", "Chill Vibes", "2000s Rock").

## Core Requirements

### 1. Child Playlist Viewing
- **Base Playlist Integration**: Click on base playlist card to view associated child playlists
- **Child Playlist List**: Display all child playlists for selected base playlist
- **Status Indicators**: Show sync status, active/inactive state
- **Quick Actions**: Enable/disable, edit, delete operations

### 2. Child Playlist Creation
- **Required Fields**: Name, filtering rules
- **Optional Fields**: Description
- **Spotify Integration**: Automatically create Spotify playlist with naming format: `[Base Playlist Name] > [Child Playlist Name]`
- **Filter Configuration**: Comprehensive rule-based filtering system

### 3. Child Playlist Management
- **Edit Functionality**: Modify name, description, and filtering rules
- **Enable/Disable**: Toggle active sync status
- **Delete Operations**: Remove from PlaylistRouter and delete associated Spotify playlist
- **Sync Control**: Manual sync trigger and status monitoring

## Filtering System Architecture

### ⚠️ IMPORTANT: Spotify API Changes
**Status Update**: Most Spotify audio feature endpoints have been deprecated as of 2024. We have pivoted to metadata-only filtering using available endpoints.

### Current Implementation: Metadata Filters
Based on available Spotify Web API data:

#### Track Information
- **Duration** (seconds): Track length for filtering by song duration
- **Popularity** (0-100): Track popularity score from Spotify
- **Explicit Content** (boolean): Filter for clean/explicit content rating

#### Artist & Album Information
- **Genres** (string array): Include/exclude specific genres
- **Release Year** (year): Filter by album/track release year
- **Artist Popularity** (0-100): Filter by artist popularity score

#### Search-based Filters
- **Track Keywords** (string array): Include/exclude based on track name keywords
- **Artist Keywords** (string array): Include/exclude based on artist name keywords

### Legacy Audio Features (DEPRECATED)
~~The following were planned but are no longer available via Spotify API:~~
- ~~Energy, Danceability, Valence, Tempo~~
- ~~Acousticness, Instrumentalness, Loudness~~
- ~~Key, Mode, Time Signature, Liveness, Speechiness~~

### Exclusion Filters [Will not be implemented yet]
- **Artist Exclusions**: Blacklist specific artists
- **Song Exclusions**: Blacklist specific tracks
- **Album Exclusions**: Blacklist entire albums

### Filter Logic
- **Range Filters**: Min/max values for numerical features
- **Set Filters**: Include/exclude for categorical data
- **Boolean Logic**: AND/OR combinations between filter groups
- **Priority Handling**: Exclusions override inclusions

## Data Model Design

### ✅ IMPLEMENTED: Child Playlist Model
```go
type ChildPlaylist struct {
    ID                string               `json:"id"`
    UserID            string               `json:"user_id" validate:"required"`
    BasePlaylistID    string               `json:"base_playlist_id" validate:"required"`
    Name              string               `json:"name" validate:"required,min=1,max=100"`
    Description       string               `json:"description,omitempty"`
    SpotifyPlaylistID string               `json:"spotify_playlist_id" validate:"required"`
    FilterRules       *AudioFeatureFilters `json:"filter_rules,omitempty"`
    IsActive          bool                 `json:"is_active"`
    Created           time.Time            `json:"created"`
    Updated           time.Time            `json:"updated"`
}

// Current Implementation: MetadataFilters
type MetadataFilters struct {
    // Track Information
    Duration   *RangeFilter `json:"duration_ms,omitempty"`
    Popularity *RangeFilter `json:"popularity,omitempty"`
    Explicit   *bool        `json:"explicit,omitempty"`
    
    // Artist & Album Information
    Genres           *SetFilter   `json:"genres,omitempty"`
    ReleaseYear      *RangeFilter `json:"release_year,omitempty"`
    ArtistPopularity *RangeFilter `json:"artist_popularity,omitempty"`
    
    // Search-based Filters
    TrackKeywords  *SetFilter `json:"track_keywords,omitempty"`
    ArtistKeywords *SetFilter `json:"artist_keywords,omitempty"`
}

// Legacy type alias for backward compatibility
type AudioFeatureFilters = MetadataFilters

type ExclusionFilters struct {
    ExcludedArtists []string `json:"excluded_artists,omitempty"`
    ExcludedTracks  []string `json:"excluded_tracks,omitempty"`
    ExcludedAlbums  []string `json:"excluded_albums,omitempty"`
}

type RangeFilter struct {
    Min *float64 `json:"min,omitempty"`
    Max *float64 `json:"max,omitempty"`
}

type SetFilter struct {
    Include []string `json:"include,omitempty"`
    Exclude []string `json:"exclude,omitempty"`
}
```

### Request/Response Models
```go
type CreateChildPlaylistRequest struct {
    Name           string                 `json:"name" validate:"required,min=1,max=100"`
    Description    string                 `json:"description,omitempty"`
    FilterRules    *AudioFeatureFilters   `json:"filter_rules,omitempty"`
    ExclusionRules *ExclusionFilters      `json:"exclusion_rules,omitempty"` // not yet
}

type UpdateChildPlaylistRequest struct {
    Name           *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
    Description    *string                `json:"description,omitempty"`
    FilterRules    *AudioFeatureFilters   `json:"filter_rules,omitempty"`
    ExclusionRules *ExclusionFilters      `json:"exclusion_rules,omitempty"`
    IsActive       *bool                  `json:"is_active,omitempty"`
    SyncEnabled    *bool                  `json:"sync_enabled,omitempty"`
}
```

## API Endpoints

### Child Playlist CRUD
- `GET /api/base-playlists/:base_id/child-playlists` - List child playlists for base
- `POST /api/base-playlists/:base_id/child-playlists` - Create new child playlist
- `GET /api/child-playlists/:id` - Get specific child playlist
- `PUT /api/child-playlists/:id` - Update child playlist
- `DELETE /api/child-playlists/:id` - Delete child playlist
- `PUT /api/child-playlists/:id/sync-toggle` - Enable/disable syncing
- `POST /api/child-playlists/:id/sync` - Manual sync trigger

### Filter Management
- `GET /api/filter-templates` - Get predefined filter templates
- `POST /api/child-playlists/:id/preview` - Preview songs matching current filters

## Frontend Implementation

### UI Components

#### Core Components
- `ChildPlaylistList` - List of child playlists for a base playlist
- `ChildPlaylistCard` - Individual child playlist display with quick actions
- `CreateChildPlaylistModal` - Multi-step creation wizard
- `EditChildPlaylistModal` - Edit existing child playlist
- `FilterRulesEditor` - Comprehensive filter configuration component
- `AudioFeatureSlider` - Range slider for audio features (0.0-1.0)
- `FilterTemplateSelector` - Predefined filter templates
- `ExclusionListEditor` - Artist/song/album exclusion management

#### Advanced Components
- `FilterPreview` - Show matching songs count before saving
- `SyncStatusIndicator` - Visual sync status with progress
- `FilterSummary` - Human-readable filter description

### User Experience Flow

#### Viewing Child Playlists
1. User clicks on base playlist card
2. Navigate to child playlist view for that base
3. Display child playlists in card/list format
4. Show status indicators and quick actions
5. "Create Child Playlist" button prominent

#### Creating Child Playlist
1. Click "Create Child Playlist" button
2. Multi-step modal opens:
   - **Step 1**: Basic info (name, description)
   - **Step 2**: Audio feature filters (collapsible sections)
   - **Step 3**: Exclusion filters (optional)
   - **Step 4**: Preview matching songs
   - **Step 5**: Confirm and create
3. System creates Spotify playlist with format `[Base] > [Child]`
4. Child playlist appears in list with "Creating..." status
5. Success notification with link to Spotify playlist

#### Managing Child Playlists
1. Each child playlist card shows:
   - Name and description
   - Track count and last sync time
   - Sync enabled/disabled toggle
   - Quick actions: edit, delete, manual sync
2. Edit opens similar modal to creation
3. Delete shows confirmation with Spotify warning
4. Toggle sync updates immediately

## Business Logic & Validation

### Subscription Limits
- **Free**: 1 base playlist, 2 child playlists total
- **Basic**: 2 base playlists, 5 child playlists each (10 total)
- **Premium**: Unlimited base and child playlists

### Validation Rules
- Child playlist names must be unique within base playlist scope
- Spotify playlist naming format enforced: `[Base] > [Child]`
- Filter values must be within valid ranges per Spotify spec
- At least one filter rule required (prevent empty filters)

### Error Handling
- **Spotify API failures**: Retry logic with exponential backoff
- **Playlist creation failures**: Clean up partial records
- **Filter validation errors**: Clear field-level error messages
- **Sync failures**: Log errors, show user-friendly messages

## Sync Engine Integration

### Song Matching Logic
1. Fetch new songs from base playlist
2. Get audio features for each song via Spotify API
3. Apply filter rules in order:
   - Check exclusion filters first (early exit)
   - Apply audio feature filters
   - Apply metadata filters
4. Add matching songs to child playlists
5. Update sync timestamps and counts

### Performance Considerations
- **Batch Processing**: Process multiple songs per API call
- **Caching**: Cache audio features to avoid repeated API calls
- **Rate Limiting**: Respect Spotify API limits (100 requests/minute)
- **Async Processing**: Background sync jobs for large playlists

## Templates & Presets

### Predefined Filter Templates
- **High Energy**: Energy > 0.7, Danceability > 0.6
- **Chill Vibes**: Energy < 0.5, Valence 0.3-0.7
- **Workout**: Energy > 0.8, Tempo > 120 BPM
- **Focus Music**: Instrumentalness > 0.8, Energy 0.3-0.7
- **Party Mix**: Danceability > 0.7, Valence > 0.6
- **Acoustic**: Acousticness > 0.7
- **Recent Releases**: Release year >= current year - 2

### Custom Template Creation
- Users can save their filter combinations as personal templates
- Templates can be shared between base playlists
- Import/export template functionality

## Analytics & Insights

### Child Playlist Statistics
- Total songs distributed
- Most/least active child playlists
- Filter effectiveness (songs matched per filter)
- Sync frequency and success rates

### User Insights
- Filter usage patterns
- Most popular template combinations
- Playlist growth trends
- Feature adoption metrics

## Missing Features & Future Enhancements

### Immediate Implementation Needs
1. **Complete filtering system** - Implement all Spotify audio features
2. **Spotify playlist creation** - Auto-create with naming format
3. **Sync engine integration** - Connect to song distribution logic
4. **Frontend filter editor** - Complex but user-friendly filter UI

### Future Considerations
- **Collaborative filtering** - Share filter templates between users
- **Machine learning** - Suggest optimal filters based on user behavior
- **Advanced sync options** - Scheduled syncs, conditional rules
- **Mobile optimization** - Touch-friendly filter controls
- **Bulk operations** - Multi-select and batch actions
- **Export/import** - Backup and restore child playlist configurations

## 🎯 CURRENT IMPLEMENTATION STATUS

### ✅ COMPLETED (Phase 1)
1. **Backend Implementation**
   - ✅ Complete child playlist data models with metadata filtering
   - ✅ Full CRUD operations (Create, Read, Update, Delete)
   - ✅ Repository layer with PocketBase integration
   - ✅ Service layer with business logic
   - ✅ Controller layer with validation
   - ✅ Complete test coverage (100+ tests passing)
   - ✅ API endpoints for all child playlist operations

2. **Frontend Implementation**
   - ✅ Child playlist viewing and listing
   - ✅ Create child playlist form with comprehensive filter UI
   - ✅ Edit child playlist functionality
   - ✅ Delete child playlist with confirmation
   - ✅ Collapsible filter categories (Track Info, Artist & Album, Search-based)
   - ✅ User-friendly input validation and error handling
   - ✅ Integration with base playlist management

3. **Filter System**
   - ✅ Complete metadata filtering implementation
   - ✅ Range filters for numerical values (duration, popularity, year)
   - ✅ Set filters for categorical data (genres, keywords)
   - ✅ Boolean filters for explicit content
   - ✅ User-friendly duration conversion (seconds vs milliseconds)
   - ✅ Form validation requiring at least one filter

### ✅ ALSO COMPLETED 
1. **Spotify Integration**
   - ✅ Spotify playlist creation with naming format `[Base] > [Child]`
   - ✅ Integration with Spotify Web API for playlist creation
   - ✅ Automatic playlist description with PlaylistRouter branding

### 🚧 REMAINING
1. **Sync Engine**
   - ⏳ Song distribution logic based on metadata filters
   - ⏳ Metadata retrieval and matching for existing songs
   - ⏳ Batch processing for large playlists

### ❌ NOT IMPLEMENTED (Future)
1. **Advanced Features**
   - ❌ Manual sync triggers
   - ❌ Sync status indicators and progress tracking
   - ❌ Filter templates and presets
   - ❌ Filter preview functionality
   - ❌ Exclusion filters for specific artists/songs
   - ❌ Analytics and usage insights
   - ❌ Bulk operations and batch actions

### 🔄 Architecture Adaptations Made
1. **API Compatibility**: Switched from deprecated Spotify audio features to metadata-only filtering
2. **Type System**: Implemented type alias system for backward compatibility
3. **User Experience**: Created comprehensive filter UI with collapsible categories
4. **Validation**: Required at least one filter to ensure meaningful child playlists
5. **Integration**: Seamlessly integrated with existing base playlist management

## Implementation Priority (Updated)

### Phase 1 (COMPLETED) ✅
1. ✅ Enhanced data models with metadata filter support
2. ✅ Complete CRUD operations for child playlists
3. ✅ Advanced filter editor UI with user-friendly controls
4. ✅ Full frontend integration with base playlist management

### Phase 2 (NEXT) 🎯
1. **Sync Engine Implementation**
   - Song distribution logic based on metadata filters
   - Metadata retrieval from Spotify Web API
   - Batch processing and sync orchestration

### Phase 3 (FUTURE) 🔮
1. Advanced sync features and status tracking
2. Filter templates and presets
3. Analytics and insights
4. Performance optimizations and caching