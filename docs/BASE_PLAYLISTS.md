# Base Playlist Management

## Overview

Base playlists are the source playlists that users add songs to. PlaylistRouter monitors these playlists and automatically distributes new songs to child playlists based on configured filtering rules.

## Core Requirements

### 1. Base Playlist Creation
Users can create base playlists in two ways:

#### Option A: Link Existing Spotify Playlist
- Display user's existing Spotify playlists
- Allow selection of playlist to link
- Validate playlist ownership and permissions
- Import playlist metadata (name, description, track count)

#### Option B: Create New Spotify Playlist
- User provides playlist name and optional description
- PlaylistRouter creates new playlist via Spotify API
- Automatically links the newly created playlist
- Set appropriate playlist visibility (public/private)

### 2. Base Playlist Management
- **Enable/Disable Syncing**: Toggle to control whether playlist is actively monitored
- **Edit Details**: Modify name and description (updates both PlaylistRouter and Spotify)
- **View Stats**: Show track count, last sync time, connected child playlists
- **Delete**: Remove from PlaylistRouter only (preserves Spotify playlist)

### 3. Listing & Overview
- Dashboard showing all user's base playlists
- Status indicators (active/inactive, sync enabled/disabled)
- Quick actions (edit, toggle sync, delete)
- Search and filtering capabilities

## Technical Implementation

### Data Model Extensions
```go
type BasePlaylist struct {
    ID                string    `json:"id"`
    UserID            string    `json:"user_id"`
    Name              string    `json:"name"`
    Description       string    `json:"description"`
    SpotifyPlaylistID string    `json:"spotify_playlist_id"`
    IsActive          bool      `json:"is_active"`          // User can disable
    SyncEnabled       bool      `json:"sync_enabled"`       // Monitoring toggle
    TrackCount        int       `json:"track_count"`        // Cached from Spotify
    LastSyncedAt      *time.Time `json:"last_synced_at"`
    Created           time.Time `json:"created"`
    Updated           time.Time `json:"updated"`
}
```

### API Endpoints Needed

#### Base Playlist Management
- `GET /api/base-playlists` - List user's base playlists
- `POST /api/base-playlists` - Create new base playlist
- `GET /api/base-playlists/:id` - Get specific base playlist
- `PUT /api/base-playlists/:id` - Update base playlist
- `DELETE /api/base-playlists/:id` - Delete base playlist
- `PUT /api/base-playlists/:id/sync-toggle` - Enable/disable syncing

#### Spotify Integration
- `GET /api/spotify/playlists` - Fetch user's Spotify playlists
- `POST /api/spotify/playlists` - Create new Spotify playlist
- `GET /api/spotify/playlists/:id/tracks` - Get playlist tracks (for validation)

### Frontend Components

#### Core Components
- `BasePlaylistList` - Dashboard overview of all base playlists
- `BasePlaylistCard` - Individual playlist display with actions
- `CreateBasePlaylistModal` - Creation wizard with both options
- `EditBasePlaylistModal` - Edit playlist details
- `SpotifyPlaylistSelector` - Component to select existing playlists
- `BasePlaylistForm` - Form for new playlist creation

#### UI Flow
1. **Dashboard**: Show existing base playlists with status
2. **Create Flow**: Modal with tabs for "Link Existing" vs "Create New"
3. **Management**: Inline editing, toggle switches, delete confirmations

### Spotify Integration Requirements

#### Required Scopes
- `playlist-read-private` - Read user's playlists
- `playlist-modify-private` - Create/modify private playlists
- `playlist-modify-public` - Create/modify public playlists

#### API Operations
- Fetch user playlists with pagination
- Create playlist with metadata
- Update playlist details
- Get playlist track count (for caching)

## User Experience Flow

### Creating Base Playlist - Link Existing
1. User clicks "Add Base Playlist" 
2. Modal opens with "Link Existing Playlist" tab active
3. System fetches user's Spotify playlists
4. User selects playlist from list
5. System validates playlist (ownership, track access)
6. User confirms and playlist is linked
7. Dashboard updates with new base playlist

### Creating Base Playlist - New Playlist
1. User clicks "Add Base Playlist"
2. Modal opens, user switches to "Create New Playlist" tab
3. User enters name, description, visibility settings
4. System creates Spotify playlist via API
5. System automatically links the new playlist
6. Dashboard updates with new base playlist

### Managing Base Playlists
1. User sees base playlists on dashboard
2. Each playlist shows: name, track count, sync status, last sync
3. Quick actions: edit (pencil), sync toggle (switch), delete (trash)
4. Edit opens modal with current details
5. Toggle immediately updates sync status
6. Delete shows confirmation dialog

## Validation & Error Handling

### Validation Rules
- Playlist names: 1-100 characters, no special characters
- User can only link playlists they own
- Maximum base playlists per subscription tier
- Prevent duplicate links to same Spotify playlist

### Error Scenarios
- **Spotify API failures**: Show user-friendly error, retry options
- **Permission issues**: Clear messaging about required scopes
- **Playlist not found**: Handle deleted/unavailable playlists
- **Network issues**: Offline state handling
- **Rate limiting**: Queue operations, show progress

## Business Rules

### Subscription Limits
- **Free**: 1 base playlist
- **Basic**: 2 base playlists
- **Premium**: Unlimited base playlists

### Sync Behavior
- Only active base playlists with sync enabled are monitored
- Disabled playlists retain their configuration but aren't processed
- Deleted base playlists stop all associated syncing immediately

## Success Metrics

### Technical Metrics
- Base playlist creation success rate > 95%
- Average creation time < 5 seconds
- Spotify API error rate < 5%

### User Engagement
- Average base playlists per user
- Percentage of users who create vs link existing
- Sync enable/disable usage patterns

## Future Considerations

### Potential Enhancements
- **Bulk Import**: Link multiple existing playlists at once
- **Playlist Templates**: Pre-configured base playlist setups
- **Collaborative Playlists**: Handle shared Spotify playlists
- **Backup/Export**: Export playlist configurations
- **Advanced Stats**: Detailed analytics per base playlist

### Technical Debt Prevention
- Implement proper caching for Spotify data
- Design for eventual real-time sync capabilities
- Plan for webhook-based playlist change detection
- Consider offline-first design for mobile users

## Implementation Priority

### Phase 1 (MVP)
1. Basic CRUD operations for base playlists
2. Link existing Spotify playlist flow
3. Simple dashboard with list view
4. Enable/disable sync toggle

### Phase 2 (Enhanced)
1. Create new Spotify playlist flow
2. Enhanced UI with cards and quick actions
3. Edit playlist details functionality
4. Comprehensive error handling

### Phase 3 (Polish)
1. Advanced search and filtering
2. Bulk operations
3. Detailed statistics and analytics
4. Mobile-optimized interface