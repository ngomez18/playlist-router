# PlaylistRouter API Specification

## Overview

This document outlines the complete API design for PlaylistRouter, organized by current implementation status and future planned functionality. The API follows RESTful principles and is built with Go, PocketBase, and serves a React frontend.

**Base URL:** `https://playlist-router.fly.dev`  
**Authentication:** JWT tokens via PocketBase  
**Content-Type:** `application/json`  
**Status:** ✅ Core functionality deployed and operational

---

## 1. Authentication (✅ IMPLEMENTED)

### Spotify OAuth Authentication

#### Initiate Spotify Login
```http
GET /auth/spotify/login
```

**Response:** Redirects to Spotify OAuth authorization URL with required scopes:
- `user-read-email` - Read user profile
- `playlist-read-private` - Read private playlists
- `playlist-modify-public` - Modify public playlists
- `playlist-modify-private` - Modify private playlists

#### Spotify OAuth Callback
```http
GET /auth/spotify/callback?code=<auth_code>&state=<state>
```

**Response:** Redirects to frontend with authentication success/failure

#### Validate Token
```http
GET /auth/validate
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "valid": true,
  "user": {
    "id": "user_123",
    "email": "user@example.com"
  }
}
```

---

## 2. Base Playlist Management (✅ IMPLEMENTED)

### List User's Base Playlists
```http
GET /api/base_playlist
Authorization: Bearer <jwt_token>
```

**Response:**
```json
[
  {
    "id": "bp_123456",
    "user_id": "user_789",
    "name": "My Daily Mix",
    "spotify_playlist_id": "37i9dQZF1E4",
    "is_active": true,
    "created": "2025-08-20T09:00:00Z",
    "updated": "2025-08-20T10:30:00Z"
  }
]
```

### Get Single Base Playlist
```http
GET /api/base_playlist/{id}
Authorization: Bearer <jwt_token>
```

### Create Base Playlist
```http
POST /api/base_playlist
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "My Daily Mix",
  "spotify_playlist_id": "37i9dQZF1E4"
}
```

**Response:**
```json
{
  "id": "bp_123456",
  "user_id": "user_789",
  "name": "My Daily Mix", 
  "spotify_playlist_id": "37i9dQZF1E4",
  "is_active": true,
  "created": "2025-08-20T11:00:00Z",
  "updated": "2025-08-20T11:00:00Z"
}
```

### Delete Base Playlist
```http
DELETE /api/base_playlist/{id}
Authorization: Bearer <jwt_token>
```

**Note:** Cascade deletes all associated child playlists from both database and Spotify.

## 3. Child Playlist Management (✅ IMPLEMENTED)

### List Child Playlists for Base Playlist
```http
GET /api/base_playlist/{basePlaylistID}/child_playlist
Authorization: Bearer <jwt_token>
```

### Get Single Child Playlist
```http
GET /api/child_playlist/{id}
Authorization: Bearer <jwt_token>
```

### Create Child Playlist
```http
POST /api/base_playlist/{basePlaylistID}/child_playlist
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "High Energy Tracks",
  "description": "Songs with high energy for workouts",
  "filter_rules": {
    "genres": { "include": ["rock", "indie"] },
    "popularity": { "min": 50 },
    "release_year": { "min": 2020 }
  }
}
```

**Response:**
```json
{
  "id": "cp_789012", 
  "user_id": "user_789",
  "base_playlist_id": "bp_123456",
  "name": "High Energy Tracks",
  "description": "Songs with high energy for workouts",
  "spotify_playlist_id": "3cEYpjA9oz9GiPac4AsH4n",
  "filter_rules": {
    "genres": { "include": ["rock", "indie"] },
    "popularity": { "min": 50 },
    "release_year": { "min": 2020 }
  },
  "is_active": true,
  "created": "2025-08-20T11:00:00Z",
  "updated": "2025-08-20T11:00:00Z"
}
```

### Update Child Playlist
```http
PUT /api/child_playlist/{id}
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Updated High Energy",
  "description": "Updated description",
  "filter_rules": {
    "popularity": { "min": 60 }
  },
  "is_active": false
}
```

### Delete Child Playlist
```http
DELETE /api/child_playlist/{id}
Authorization: Bearer <jwt_token>
```

**Note:** Deletes both from database and Spotify.

## 4. Sync Operations (✅ IMPLEMENTED)

### Trigger Base Playlist Sync
```http
POST /api/base_playlist/{basePlaylistID}/sync
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "id": "sync_345678",
  "user_id": "user_789",
  "base_playlist_id": "bp_123456", 
  "child_playlist_ids": ["cp_789012", "cp_789013"],
  "status": "in_progress",
  "started_at": "2025-08-20T11:00:00Z",
  "tracks_processed": 0,
  "total_api_requests": 0,
  "created": "2025-08-20T11:00:00Z",
  "updated": "2025-08-20T11:00:00Z"
}
```

## 5. Spotify Integration (✅ IMPLEMENTED)

### Get User's Spotify Playlists
```http
GET /api/spotify/playlists
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "playlists": [
    {
      "id": "37i9dQZF1E4",
      "name": "Daily Mix 1", 
      "tracks": {
        "total": 50
      },
      "public": false,
      "collaborative": false,
      "owner": {
        "id": "spotify_user_123"
      },
      "images": [
        {
          "url": "https://i.scdn.co/image/ab67616d0000b273...",
          "height": 640,
          "width": 640
        }
      ]
    }
  ]
}
```

---

## 6. Health Check (✅ IMPLEMENTED)

### Health Check Endpoint
```http
GET /health
```

**Response:**
```
OK
```

## 7. Filter Types Reference

### Metadata Filters
The Spotify API has deprecated access to detailed audio features (energy, danceability, etc.). PlaylistRouter now uses metadata-based filtering.

```typescript
interface MetadataFilters {
  // Track Information
  duration_ms?: RangeFilter;   // Track duration in milliseconds
  popularity?: RangeFilter;    // 0-100 (Spotify popularity score)
  explicit?: boolean;          // true = explicit only, false = clean only, nil = both

  // Artist & Album Information
  genres?: SetFilter;          // List of genres (e.g., "rock", "pop")
  release_year?: RangeFilter;  // Year of release (e.g., 2023)
  artist_popularity?: RangeFilter; // 0-100 (Artist popularity score)

  // Search-based Filters
  track_keywords?: SetFilter;  // Keywords to match in track name
  artist_keywords?: SetFilter; // Keywords to match in artist name
}

interface RangeFilter {
  min?: number;
  max?: number;
}

interface SetFilter {
  include?: string[];
  exclude?: string[];
}
```

## 8. Future Planned Features (🔮 NOT YET IMPLEMENTED)

### Advanced Sync Operations

#### Trigger Full Sync (All Base Playlists)
```http
POST /api/sync/trigger
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "type": "full"
}
```

#### Get Sync History
```http
GET /api/sync/history?limit=10
Authorization: Bearer <jwt_token>
```

#### Automated Sync Configuration
```http
POST /api/sync/schedule
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "interval": "15m",  // 15 minutes
  "enabled": true
}
```

### Advanced Spotify Integration





### Subscription Management (Planned)

#### Get Current Subscription Status
```http
GET /api/subscription/current
Authorization: Bearer <jwt_token>
```

#### Get Usage Statistics
```http
GET /api/usage/current
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "current_tier": "free",
  "usage": {
    "syncs_this_month": 7,
    "sync_limit": 10,
    "base_playlists_count": 2,
    "base_playlists_limit": 0
  },
  "limits_by_tier": {
    "free": { "monthly_syncs": 10, "base_playlists": 0 },
    "basic": { "monthly_syncs": "unlimited", "base_playlists": 2 },
    "premium": { "monthly_syncs": "unlimited", "base_playlists": "unlimited" }
  }
}
```

### Analytics & Insights (Planned)

#### Dashboard Analytics
```http
GET /api/analytics/dashboard
Authorization: Bearer <jwt_token>
```

#### Playlist Performance Metrics
```http
GET /api/analytics/playlist/{id}/stats
Authorization: Bearer <jwt_token>
```

## 9. Error Handling

### Standard Error Response Format
```json
{
  "error": "Error message description",
  "details": "Additional error details if available"
}
```

### Common HTTP Status Codes
- `200` - Success
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid/missing auth token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `500` - Internal Server Error

## 10. Implementation Status

### ✅ Completed Features (Deployed)
- **Authentication**: Spotify OAuth flow with JWT tokens
- **User Management**: User creation and management with PocketBase
- **Base Playlists**: Full CRUD operations (create, read, delete)
- **Child Playlists**: Full CRUD operations with metadata filtering (genres, popularity, etc.)
- **Manual Sync**: Trigger sync operations for base playlists
- **Spotify Integration**: List user playlists, create/delete playlists
- **Frontend**: React app with Chakra UI served as static assets
- **Deployment**: Production deployment on fly.io with persistent database
- **Health Monitoring**: Health check endpoint for monitoring

### 🔮 Future Features (Planned)
- **Advanced Filtering**: Genres, release years, popularity, artist/track exclusions
- **Automated Sync**: Scheduled automatic sync operations
- **Sync History**: Detailed sync operation tracking and history
- **Usage Analytics**: Dashboard with playlist performance metrics
- **Subscription Tiers**: Free/Basic/Premium with usage limits
- **Billing Integration**: Stripe integration for subscription management
- **Advanced Spotify Features**: Batch operations
- **Performance Optimizations**: Caching, pagination, rate limiting
- **Mobile Optimizations**: Enhanced mobile experience
- **Social Features**: Playlist sharing and templates

### 🚧 Known Limitations (Current MVP)
- **Manual Sync Only**: No automated sync scheduling
- **Metadata Filtering**: Filtering limited to available metadata (no audio analysis like BPM/energy)
- **No Usage Tracking**: Unlimited usage during MVP phase
- **Limited Analytics**: No detailed performance metrics
- **Basic Error Handling**: Simple error responses
- **No Batch Operations**: Single playlist operations only

---

## Development Notes

### Current Architecture
- **Single Go Binary**: Serves both API and React frontend
- **PocketBase Backend**: SQLite database with built-in auth and CRUD
- **Spotify Web API**: Direct integration for playlist operations
- **JWT Authentication**: Token-based auth with automatic refresh
- **Docker Deployment**: Multi-stage build deployed on fly.io

### Security
- All API endpoints except health check require authentication
- User isolation enforced through middleware and database relations
- Spotify tokens securely stored and auto-refreshed
- CORS configured for production domain only
- HTTPS enforced with automatic redirects

### Performance
- Endpoints typically respond in < 500ms
- Spotify API rate limits respected (100 requests/minute)
- Database queries optimized with proper indexing
- Frontend assets served from embedded storage
- Health checks for monitoring and reliability

This API serves as the foundation for the PlaylistRouter MVP and will be extended with additional features in future releases based on user feedback and usage patterns.