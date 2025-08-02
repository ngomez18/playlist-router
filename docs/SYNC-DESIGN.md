# Sync Engine Design

## Overview

The sync engine is responsible for distributing songs from base playlists to child playlists based on user-defined metadata filters. This document outlines the MVP implementation using a synchronous backend endpoint with manual triggering.

## Core Requirements

### Manual Sync Process
- Users manually trigger sync via frontend button
- Synchronous backend endpoint (blocking operation)
- Complete playlist rebuild (wipe and replace strategy)
- Resource locking during sync operations

### Sync Flow
1. **Track Information Retrieval**: Fetch all tracks from base playlist
2. **Metadata Enrichment**: Fetch additional track metadata (artists, genres, etc.)
3. **Data Transformation**: Create filterable structure with track IDs and metadata
4. **Filter Application**: Apply child playlist filters to determine song matches
5. **Playlist Updates**: Update child playlists with matching tracks
6. **Completion**: Mark sync as complete and unlock resources

## Architecture Components

### 1. SyncEvent Model
Tracks sync operations and provides audit trail.

```go
type SyncEvent struct {
    ID                    string     `json:"id"`
    UserID                string     `json:"user_id" validate:"required"`
    BasePlaylistID        string     `json:"base_playlist_id" validate:"required"`
    ChildPlaylistIDs      []string   `json:"child_playlist_ids"` // IDs processed during sync
    Status                SyncStatus `json:"status"`
    StartedAt             time.Time  `json:"started_at"`
    CompletedAt           *time.Time `json:"completed_at,omitempty"`
    ErrorMessage          *string    `json:"error_message,omitempty"`
    
    // Sync statistics
    TracksProcessed       int `json:"tracks_processed"`
    TotalAPIRequests      int `json:"total_api_requests"`
    
    Created               time.Time `json:"created"`
    Updated               time.Time `json:"updated"`
}

type SyncStatus string
const (
    SyncStatusInProgress SyncStatus = "in_progress"
    SyncStatusCompleted  SyncStatus = "completed" 
    SyncStatusFailed     SyncStatus = "failed"
)
```

### 2. SyncEventRepository
Database operations for sync event tracking.

**Interface:**
```go
type SyncEventRepository interface {
    Create(ctx context.Context, syncEvent *SyncEvent) (*SyncEvent, error)
    Update(ctx context.Context, id string, syncEvent *SyncEvent) (*SyncEvent, error)
    GetByID(ctx context.Context, id string) (*SyncEvent, error)
    HasActiveSyncForBasePlaylist(ctx context.Context, userID, basePlaylistID string) (bool, error)
    GetActiveSyncsByUser(ctx context.Context, userID string) ([]*SyncEvent, error)
    MarkStuckSyncsAsFailed(ctx context.Context, cutoffTime time.Time) error
}
```

### 3. SyncService
Core sync logic and orchestration.

**Interface:**
```go
type SyncService interface {
    SyncBasePlaylist(ctx context.Context, userID, basePlaylistID string) (*SyncEvent, error)
    HasActiveSyncForBasePlaylist(ctx context.Context, userID, basePlaylistID string) (bool, error)
    CleanupOrphanedSyncs(ctx context.Context) error
}
```

**Implementation responsibilities:**
- Orchestrate the 5-step sync process
- Filter matching logic using metadata
- Error handling and sync event updates
- API rate limiting and batch processing

### 4. SyncController
HTTP endpoint for triggering sync operations.

**Endpoints:**
- `POST /api/base-playlists/:id/sync` - Trigger sync for base playlist
- `GET /api/sync-events/:id` - Get sync event details
- `GET /api/base-playlists/:id/sync-status` - Check if sync is active

### 5. Sync Locking Middleware
Prevents resource conflicts during sync operations.

**Locking Strategy:**
- **Child Playlist Operations**: Block create/edit/delete during sync
- **Base Playlist Operations**: Block edit/delete and child playlist creation during sync
- **Scope**: Base playlist level (locks all related child playlists)

## Detailed Sync Process

### Step 1: Track Information Retrieval
- Use Spotify "Get Playlist Items" endpoint
- Batch size: 50 tracks per request
- Handle pagination for large playlists
- Extract track IDs and basic metadata

### Step 2: Metadata Enrichment
- Use "Get Several Tracks" (50 tracks/request) for track details
- Use "Get Several Artists" (50 artists/request) for artist info
- Use "Get Several Albums" (20 albums/request) for album/release info
- Build comprehensive metadata structure

### Step 3: Data Transformation
Create filterable structure:
```go
type FilterableTrack struct {
    TrackID           string
    Name              string
    DurationMs        int
    Popularity        int
    Explicit          bool
    Artists           []Artist
    Album             Album
    Genres            []string
    ReleaseYear       int
    ArtistPopularity  int
}
```

### Step 4: Filter Application
For each child playlist:
- Retrieve filter rules
- Apply metadata filters to track list
- Generate list of matching track IDs

### Step 5: Playlist Updates
- Clear existing tracks from child playlist (Spotify API)
- Add matching tracks to child playlist (Spotify API)
- Update sync statistics

## API Usage & Rate Limiting

### Spotify API Calls per Sync
For a playlist with N tracks, M unique artists, P unique albums:
- Playlist items: `ceil(N / 50)` requests
- Track details: `ceil(N / 50)` requests  
- Artist details: `ceil(M / 50)` requests
- Album details: `ceil(P / 20)` requests
- Child playlist updates: `2 * child_count` requests (clear + add per playlist)

**Example (5,000 track playlist):**
- ~100 requests for tracks
- ~100 requests for artists  
- ~250 requests for albums
- ~10 requests for child playlist updates
- **Total: ~460 requests** (4.6 minutes at 100 req/min limit)

### Rate Limit Handling
- **MVP Approach**: Fail fast on rate limit exceeded
- **Future**: Implement exponential backoff and retry logic

## Error Handling Strategy

### Failure Scenarios
1. **API Rate Limiting**: Fail sync, mark as failed
2. **Network Failures**: Fail sync, mark as failed  
3. **Invalid Metadata**: Skip problematic tracks, continue sync
4. **Spotify API Errors**: Fail sync, log error details

### Recovery Strategy
- **MVP**: Full restart required (no partial recovery)
- Mark failed syncs in database with error message
- User must manually retry entire sync

### Orphaned Sync Cleanup
- **Server Startup**: Mark all "in_progress" syncs as failed
- **Timeout Strategy**: Consider syncs older than 30 minutes as stuck

## Resource Locking Implementation

### Middleware Design
```go
func SyncLockMiddleware(syncService SyncService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if requiresSyncCheck(r) {
                userID := getUserFromContext(r.Context())
                basePlaylistID := extractBasePlaylistID(r)
                
                hasActiveSync, err := syncService.HasActiveSyncForBasePlaylist(r.Context(), userID, basePlaylistID)
                if err != nil {
                    http.Error(w, "Error checking sync status", http.StatusInternalServerError)
                    return
                }
                
                if hasActiveSync {
                    http.Error(w, "Cannot perform operation during active sync", http.StatusConflict)
                    return
                }
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Protected Operations
- Child playlist create/edit/delete
- Base playlist edit/delete
- Child playlist creation for base playlist

## Database Schema

### Sync Events Collection
```sql
CREATE TABLE sync_events (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    base_playlist_id TEXT NOT NULL,
    child_playlist_ids TEXT, -- JSON array
    status TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    error_message TEXT,
    tracks_processed INTEGER DEFAULT 0,
    total_api_requests INTEGER DEFAULT 0,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (base_playlist_id) REFERENCES base_playlists(id)
);

CREATE INDEX idx_sync_events_user_status ON sync_events(user_id, status);
CREATE INDEX idx_sync_events_base_playlist ON sync_events(base_playlist_id, status);
```

## Frontend Integration

### UI Components
- **Sync Button**: On base playlist cards
- **Sync Status**: Show "Syncing..." state with disabled actions
- **Sync History**: List of recent sync events
- **Error Display**: Show sync failures with retry option

### User Experience
- Clear indication when sync is running
- All playlist modification actions disabled during sync
- Success/failure notifications
- Estimated sync time based on playlist size

## Future Enhancements

### Phase 2 Improvements
1. **Redis Caching System**: Dramatically reduce Spotify API calls
2. **Background Job Processing**: Move to async queue system
3. **Progress Tracking**: Real-time sync progress updates
4. **Partial Recovery**: Resume from failure points
5. **Smart Rate Limiting**: Exponential backoff and retry logic
6. **Incremental Sync**: Only process changed tracks

### Phase 3 Enhancements
1. **Automatic Sync**: Scheduled or trigger-based syncing
2. **Sync Analytics**: Performance metrics and insights
3. **Conflict Resolution**: Handle concurrent modifications
4. **Batch Operations**: Sync multiple base playlists together

## Redis Caching Architecture (Phase 2)

### Overview
Implement Redis caching to reduce Spotify API calls by 70-90% for repeated syncs and overlapping playlists.

### Caching Strategy

#### High-Impact Caching Targets
1. **Artist Metadata** - Very stable, high reuse across tracks/playlists
2. **Album Metadata** - Very stable, high reuse within albums
3. **Track Metadata** - Moderately stable, reused across playlists
4. **Playlist Track Lists** - Lower priority due to frequent changes

#### Performance Impact
**Before Caching (5,000 track playlist):**
- Track API calls: 100 requests
- Artist API calls: 100 requests  
- Album API calls: 250 requests
- **Total: 450 requests** (4.5+ minutes)

**After Caching (80% hit rate):**
- Track API calls: 20 requests (80 cached)
- Artist API calls: 20 requests (80 cached)
- Album API calls: 50 requests (200 cached)
- **Total: 90 requests** (~54 seconds, 80% reduction)

### Implementation Design

#### Option A: Transparent Caching (Recommended)
Implement caching at the Spotify client level for transparent integration:

```go
type CachedSpotifyClient struct {
    client SpotifyClient
    cache  RedisClient
    logger *slog.Logger
}

func (c *CachedSpotifyClient) GetSeveralTracks(ctx context.Context, token string, trackIDs []string) ([]*models.Track, error) {
    var uncachedIDs []string
    var cachedTracks []*models.Track
    
    // Check cache for existing tracks
    for _, id := range trackIDs {
        cacheKey := fmt.Sprintf("spotify:track:%s", id)
        if cached := c.cache.Get(ctx, cacheKey); cached != nil {
            var track models.Track
            if err := json.Unmarshal(cached, &track); err == nil {
                cachedTracks = append(cachedTracks, &track)
                continue
            }
        }
        uncachedIDs = append(uncachedIDs, id)
    }
    
    // Fetch uncached tracks from Spotify API
    if len(uncachedIDs) > 0 {
        newTracks, err := c.client.GetSeveralTracks(ctx, token, uncachedIDs)
        if err != nil {
            return nil, err
        }
        
        // Cache newly fetched tracks
        for _, track := range newTracks {
            cacheKey := fmt.Sprintf("spotify:track:%s", track.ID)
            trackData, _ := json.Marshal(track)
            c.cache.Set(ctx, cacheKey, trackData, 24*time.Hour)
        }
        
        cachedTracks = append(cachedTracks, newTracks...)
    }
    
    return cachedTracks, nil
}

// Similar implementations for GetSeveralArtists, GetSeveralAlbums
```

#### Cache Key Strategy
```go
const (
    TrackCachePrefix    = "spotify:track:"
    ArtistCachePrefix   = "spotify:artist:"  
    AlbumCachePrefix    = "spotify:album:"
    PlaylistCachePrefix = "spotify:playlist:"
    
    // Version for cache invalidation
    CacheVersion = "v1"
)

// Cache key format: prefix + version + id
func buildCacheKey(prefix, id string) string {
    return fmt.Sprintf("%s%s:%s", prefix, CacheVersion, id)
}
```

#### TTL Configuration
```go
var CacheTTLs = map[string]time.Duration{
    "track":    24 * time.Hour,      // Stable metadata, daily refresh
    "artist":   7 * 24 * time.Hour,  // Very stable, weekly refresh
    "album":    7 * 24 * time.Hour,  // Very stable, weekly refresh  
    "playlist": 5 * time.Minute,     // Frequently changing, short TTL
}
```

### Integration Points

#### SyncService Enhancement
```go
type SyncService struct {
    spotifyClient     CachedSpotifyClient // Enhanced with caching
    childPlaylistRepo ChildPlaylistRepository
    syncEventRepo     SyncEventRepository
    cache            RedisClient
    logger           *slog.Logger
}
```

#### Cache Statistics Tracking
Add cache performance metrics to sync events:

```go
type SyncEvent struct {
    // ... existing fields ...
    
    // Cache performance metrics
    CacheHitRate      float64 `json:"cache_hit_rate,omitempty"`
    CachedAPIRequests int     `json:"cached_api_requests,omitempty"`
    ActualAPIRequests int     `json:"actual_api_requests,omitempty"`
}
```

### Cache Management

#### Cache Warming Strategy
```go
// Background job to pre-populate cache for active users
func (s *SyncService) WarmCacheForUser(ctx context.Context, userID string) error {
    // Get user's recent playlists
    // Pre-fetch and cache commonly accessed metadata
    // Schedule during low-traffic periods
}
```

#### Cache Invalidation
```go
// Manual cache invalidation for specific scenarios
func (c *CachedSpotifyClient) InvalidateCache(ctx context.Context, cacheType, id string) error {
    cacheKey := buildCacheKey(cacheType+":", id)
    return c.cache.Del(ctx, cacheKey)
}

// Bulk invalidation for maintenance
func (c *CachedSpotifyClient) InvalidateAll(ctx context.Context, cacheType string) error {
    pattern := fmt.Sprintf("%s%s:*", cacheType, CacheVersion)
    return c.cache.DelByPattern(ctx, pattern)
}
```

### Monitoring & Observability

#### Cache Metrics
- Hit/miss ratios per cache type
- Cache size and memory usage
- API call reduction percentages
- Cache refresh frequencies

#### Dashboard Integration
- Real-time cache performance
- Sync time improvements
- Cost savings from reduced API usage
- Cache health and expiration monitoring

### Implementation Considerations

#### Error Handling
- **Cache Miss Fallback**: Always fallback to API if cache fails
- **Partial Cache Failures**: Continue with available cached data
- **Cache Connectivity**: Degrade gracefully if Redis unavailable

#### Data Consistency
- **Cache Versioning**: Handle schema changes with version prefixes
- **Atomic Operations**: Ensure cache updates are atomic
- **Race Conditions**: Handle concurrent cache updates safely

#### Memory Management
- **Memory Limits**: Configure Redis max memory and eviction policies
- **Data Compression**: Consider compressing cached JSON for larger objects
- **TTL Optimization**: Balance freshness vs. API call reduction

### Migration Strategy

#### Phase 2A: Basic Caching
1. Implement CachedSpotifyClient wrapper
2. Add Redis dependency and configuration
3. Cache track, artist, and album metadata
4. Monitor performance improvements

#### Phase 2B: Advanced Caching
1. Add cache warming for active users
2. Implement cache analytics and monitoring
3. Optimize TTL values based on usage patterns
4. Add cache management endpoints

#### Phase 2C: Cache Optimization
1. Implement intelligent cache prefetching
2. Add compressed storage for large datasets
3. Implement cache clustering for scale
4. Advanced invalidation strategies

## Implementation Status

### ✅ **Completed Components:**

#### 1. SyncEvent Model & Repository Layer
- **SyncEvent Model** (`internal/models/sync_event.go`) - Tracks sync operations with all required fields
- **SyncEventRepository Interface** (`internal/repositories/sync_event_repository.go`) - Clean 5-method interface for CRUD operations
- **SyncEventRepository PocketBase Implementation** (`internal/repositories/pb/sync_event_repository_pb.go`) - Full implementation with JSON serialization, error handling, and logging
- **Comprehensive Unit Tests** (`internal/repositories/pb/sync_event_repository_pb_test.go`) - Table-driven tests with 100% coverage
- **Database Schema Support** - Collection setup and helper functions for testing

#### 2. SyncEvent Service Layer  
- **SyncEventService Interface & Implementation** (`internal/services/sync_event_service.go`) - CRUD operations and active sync checks
- **Unit Tests** (`internal/services/sync_event_service_test.go`) - Full test coverage following existing patterns
- **Purpose**: Used by middleware for sync locking and controllers for sync status/history

#### 3. Track Routing Service Interface
- **TrackRouterService Interface** (`internal/services/track_router_service.go`) - Single-method interface for core routing logic
- **Purpose**: Will orchestrate the complete track routing process from base playlists to child playlists

### 🔄 **Next Steps (MVP):**

#### 4. TrackRouterService Implementation
- Core track routing orchestration logic
- Integration with existing services (SyncEventService, ChildPlaylistService, BasePlaylistService)
- Spotify API integration for track fetching and metadata enrichment
- Filter application logic for child playlists

#### 5. SyncController - HTTP endpoints for sync operations
- `POST /api/base-playlists/:id/sync` - Trigger sync using TrackRouterService
- `GET /api/sync-events/:id` - Get sync event details using SyncEventService
- `GET /api/base-playlists/:id/sync-status` - Check active sync status using SyncEventService

#### 6. Sync Locking Middleware - Prevent conflicts during sync operations
- Uses SyncEventService.HasActiveSyncForBasePlaylist() for locking logic
- Block child playlist create/edit/delete during sync
- Block base playlist edit/delete during sync
- Base playlist level locking (locks all related child playlists)

#### 7. Server Startup Cleanup - Handle orphaned syncs on restart
- Mark stuck "in_progress" syncs as failed on server startup
- Can be implemented as startup function using SyncEventService

#### 8. Frontend Integration - Update UI components
- Sync button on base playlist cards
- Sync status indicators and loading states  
- Disabled actions during sync

## Revised Architecture

Based on implementation experience, the architecture has been refined:

### Service Layer Separation
- **SyncEventService**: Pure CRUD operations for sync events, used by middleware and controllers
- **TrackRouterService**: Core track routing business logic, focused solely on the routing process
- **Clear Dependencies**: TrackRouterService uses SyncEventService for sync event management

### Benefits of This Architecture
- **Single Responsibility**: Each service has one clear purpose
- **Reusability**: SyncEventService perfect for middleware sync locking
- **Testability**: Clean interfaces with focused responsibilities
- **Maintainability**: Clear separation between data management and business logic

### Success Criteria
- Users can manually trigger sync for base playlists
- Child playlists are correctly updated based on filters  
- System prevents conflicts during sync operations
- Failed syncs are properly tracked and reported
- Orphaned syncs are cleaned up on server restart

## Testing Strategy

### Unit Tests
- SyncService filter matching logic
- SyncEventRepository database operations
- Middleware locking behavior

### Integration Tests  
- End-to-end sync process with mock Spotify API
- Error scenarios and failure handling
- Concurrent operation blocking

### Performance Tests
- Large playlist sync times
- API rate limit handling
- Memory usage during sync operations