# PlaylistSync API Specification

## Overview

This document outlines the complete API design for PlaylistSync, organized by implementation priority and functionality. The API follows RESTful principles and is designed for a PocketBase backend with Go.

**Base URL:** `https://api.playlistsync.com`  
**Authentication:** JWT tokens via PocketBase  
**Content-Type:** `application/json`

---

## 1. Core CRUD Operations (Phase 1 Priority)

### Base Playlists Management

#### List Base Playlists
```http
GET /api/collections/base_playlists/records
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "page": 1,
  "perPage": 30,
  "totalItems": 5,
  "totalPages": 1,
  "items": [
    {
      "id": "bp_123456",
      "user": "user_789",
      "name": "My Daily Mix",
      "spotify_playlist_id": "37i9dQZF1E4",
      "is_active": true,
      "last_synced": "2025-07-22T10:30:00Z",
      "sync_status": "success",
      "created": "2025-07-20T09:00:00Z",
      "updated": "2025-07-22T10:30:00Z"
    }
  ]
}
```

#### Get Single Base Playlist
```http
GET /api/collections/base_playlists/records/{id}
Authorization: Bearer <jwt_token>
```

#### Create Base Playlist
```http
POST /api/collections/base_playlists/records
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "My Daily Mix",
  "spotify_playlist_id": "37i9dQZF1E4",
  "is_active": true
}
```

**Response:**
```json
{
  "id": "bp_123456",
  "user": "user_789", // Auto-populated from JWT
  "name": "My Daily Mix",
  "spotify_playlist_id": "37i9dQZF1E4",
  "is_active": true,
  "sync_status": "never_synced",
  "created": "2025-07-22T11:00:00Z",
  "updated": "2025-07-22T11:00:00Z"
}
```

#### Update Base Playlist
```http
PATCH /api/collections/base_playlists/records/{id}
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Updated Daily Mix",
  "is_active": false
}
```

#### Delete Base Playlist
```http
DELETE /api/collections/base_playlists/records/{id}
Authorization: Bearer <jwt_token>
```

**Note:** This should cascade delete all associated child playlists.

---

### Child Playlists Management

#### List Child Playlists for Base Playlist
```http
GET /api/collections/child_playlists/records?filter=base_playlist="{base_playlist_id}"
Authorization: Bearer <jwt_token>
```

#### Get Single Child Playlist
```http
GET /api/collections/child_playlists/records/{id}
Authorization: Bearer <jwt_token>
```

#### Create Child Playlist
```http
POST /api/collections/child_playlists/records
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "base_playlist": "bp_123456",
  "name": "High Energy Tracks",
  "spotify_playlist_id": "new_playlist_id",
  "filter_rules": {
    "audio_features": {
      "energy": { "min": 0.7, "max": 1.0 },
      "danceability": { "min": 0.6 }
    },
    "metadata": {
      "year_range": { "min": 2020 }
    }
  },
  "exclusion_rules": {
    "artists": ["Artist Name to Exclude"],
    "spotify_artist_ids": ["spotify_artist_id"]
  },
  "is_active": true
}
```

**Response:**
```json
{
  "id": "cp_789012",
  "user": "user_789",
  "base_playlist": "bp_123456",
  "name": "High Energy Tracks",
  "spotify_playlist_id": "new_playlist_id",
  "filter_rules": {
    "audio_features": {
      "energy": { "min": 0.7, "max": 1.0 },
      "danceability": { "min": 0.6 }
    },
    "metadata": {
      "year_range": { "min": 2020 }
    }
  },
  "exclusion_rules": {
    "artists": ["Artist Name to Exclude"],
    "spotify_artist_ids": ["spotify_artist_id"]
  },
  "is_active": true,
  "sync_status": "never_synced",
  "songs_count": 0,
  "created": "2025-07-22T11:00:00Z",
  "updated": "2025-07-22T11:00:00Z"
}
```

#### Update Child Playlist
```http
PATCH /api/collections/child_playlists/records/{id}
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Updated High Energy",
  "filter_rules": {
    "audio_features": {
      "energy": { "min": 0.8, "max": 1.0 }
    }
  },
  "is_active": false
}
```

#### Delete Child Playlist
```http
DELETE /api/collections/child_playlists/records/{id}
Authorization: Bearer <jwt_token>
```

---

## 2. Authentication & User Management (Phase 2 Priority)

### Spotify OAuth Authentication

#### Initiate Spotify Login
```http
GET /auth/spotify/login
```

**Response:** Redirects to Spotify OAuth authorization URL with required scopes:
- `user-read-email`
- `playlist-read-private` 
- `playlist-modify-public`
- `playlist-modify-private`

#### Spotify OAuth Callback
```http
GET /auth/spotify/callback?code=<auth_code>&state=<state>
```

**Response:**
```json
{
  "user": {
    "id": "user_789",
    "email": "user@example.com", 
    "name": "John Doe",
    "spotify_id": "spotify_user_123"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": ""
}
```

#### Refresh Token
```http
POST /api/collections/users/auth-refresh
Authorization: Bearer <jwt_token>
```

#### User Profile
```http
GET /api/collections/users/records/{id}
Authorization: Bearer <jwt_token>
```

---

### Spotify Integration Management

#### Get User's Spotify Integration
```http
GET /api/collections/spotify_integrations/records?filter=user="{user_id}"
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "page": 1,
  "perPage": 30,
  "totalItems": 1,
  "totalPages": 1,
  "items": [
    {
      "id": "si_123456",
      "user": "user_789",
      "spotify_id": "spotify_user_123",
      "display_name": "John Doe",
      "token_type": "Bearer",
      "expires_at": "2025-07-22T12:00:00Z",
      "scope": "user-read-email playlist-read-private playlist-modify-public playlist-modify-private",
      "is_active": true,
      "created": "2025-07-20T09:00:00Z",
      "updated": "2025-07-22T10:00:00Z"
    }
  ]
}
```

**Note:** `access_token` and `refresh_token` are never exposed in API responses for security.

#### Update Spotify Integration Tokens
```http
PATCH /api/collections/spotify_integrations/records/{id}
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "access_token": "new_encrypted_token",
  "expires_at": "2025-07-22T13:00:00Z"
}
```

#### Delete Spotify Integration
```http
DELETE /api/collections/spotify_integrations/records/{id}
Authorization: Bearer <jwt_token>
```

---

## 3. Sync Operations (Phase 2 Priority)

### Manual Sync Triggers

#### Trigger Full Sync
```http
POST /api/sync/trigger
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "type": "full"
}
```

#### Trigger Base Playlist Sync
```http
POST /api/sync/trigger
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "type": "base_playlist",
  "base_playlist_id": "bp_123456"
}
```

#### Trigger Single Child Playlist Sync
```http
POST /api/sync/trigger
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "type": "child_playlist",
  "child_playlist_id": "cp_789012"
}
```

**Response:**
```json
{
  "sync_id": "sync_345678",
  "status": "queued",
  "message": "Sync operation queued successfully",
  "estimated_completion": "2025-07-22T11:05:00Z"
}
```

### Sync Status & History

#### Get Current Sync Status
```http
GET /api/sync/status
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "current_sync": {
    "sync_id": "sync_345678",
    "status": "in_progress",
    "started_at": "2025-07-22T11:00:00Z",
    "progress": {
      "processed": 15,
      "total": 50,
      "current_playlist": "High Energy Tracks"
    }
  },
  "last_full_sync": "2025-07-22T10:30:00Z",
  "next_auto_sync": "2025-07-22T11:15:00Z"
}
```

#### Get Sync History
```http
GET /api/sync/history?limit=10
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "syncs": [
    {
      "sync_id": "sync_345678",
      "type": "full",
      "status": "completed",
      "started_at": "2025-07-22T10:30:00Z",
      "completed_at": "2025-07-22T10:32:00Z",
      "songs_processed": 47,
      "songs_distributed": 38,
      "errors": 0
    }
  ]
}
```

---

## 4. Spotify Integration (Phase 2 Priority)

### Spotify Data Endpoints

#### Get User's Spotify Playlists
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
      "tracks_total": 50,
      "public": false,
      "collaborative": false,
      "owner": "spotify_user_123"
    }
  ]
}
```

#### Get Playlist Tracks
```http
GET /api/spotify/playlists/{playlist_id}/tracks
Authorization: Bearer <jwt_token>
```

#### Get Track Audio Features
```http
GET /api/spotify/tracks/{track_id}/audio-features
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "danceability": 0.735,
  "energy": 0.578,
  "key": 5,
  "loudness": -11.84,
  "mode": 0,
  "speechiness": 0.0461,
  "acousticness": 0.514,
  "instrumentalness": 0.0902,
  "liveness": 0.159,
  "valence": 0.624,
  "tempo": 98.002,
  "duration_ms": 255349,
  "time_signature": 4
}
```

#### Create Spotify Playlist
```http
POST /api/spotify/playlists
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "High Energy Tracks",
  "description": "Auto-generated by PlaylistSync",
  "public": false
}
```

---

## 5. Subscription Management (Phase 3 Priority)

### Subscription Endpoints

#### Get Current Subscription
```http
GET /api/collections/subscriptions/records?filter=user="{user_id}"&filter=status="active"
Authorization: Bearer <jwt_token>
```

#### Get Usage Stats
```http
GET /api/usage/current
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "current_tier": "free",
  "current_period": {
    "start": "2025-07-01T00:00:00Z",
    "end": "2025-07-31T23:59:59Z"
  },
  "usage": {
    "syncs_this_month": 7,
    "sync_limit": 10,
    "base_playlists_count": 0,
    "base_playlists_limit": 0,
    "child_playlists_count": 0
  },
  "limits_by_tier": {
    "free": { "monthly_syncs": 10, "base_playlists": 0, "child_playlists": 0 },
    "basic": { "monthly_syncs": "unlimited", "base_playlists": 2, "child_playlists_per_base": 5 },
    "premium": { "monthly_syncs": "unlimited", "base_playlists": "unlimited", "child_playlists_per_base": "unlimited" }
  }
}
```

### Stripe Integration (Webhook Endpoints)

#### Stripe Webhook Handler
```http
POST /api/webhooks/stripe
Content-Type: application/json
Stripe-Signature: <webhook_signature>

{
  "type": "customer.subscription.created",
  "data": {
    "object": {
      "id": "sub_1234567890",
      "customer": "cus_1234567890",
      "status": "active",
      "current_period_start": 1690000000,
      "current_period_end": 1692678400
    }
  }
}
```

---

## 6. Analytics & Metrics (Phase 4 Priority)

### User Analytics

#### Get Dashboard Stats
```http
GET /api/analytics/dashboard
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "overview": {
    "total_base_playlists": 3,
    "total_child_playlists": 8,
    "total_songs_distributed": 247,
    "last_sync": "2025-07-22T10:30:00Z"
  },
  "sync_activity": {
    "syncs_this_week": 12,
    "syncs_this_month": 45,
    "average_songs_per_sync": 15.2
  },
  "playlist_performance": [
    {
      "child_playlist_name": "High Energy Tracks",
      "songs_added_this_week": 8,
      "total_songs": 42
    }
  ]
}
```

#### Get Detailed Analytics
```http
GET /api/analytics/detailed?period=30d
Authorization: Bearer <jwt_token>
```

---

## Error Handling

### Standard Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "The provided playlist ID is invalid",
    "details": {
      "field": "spotify_playlist_id",
      "value": "invalid_id"
    },
    "timestamp": "2025-07-22T11:00:00Z"
  }
}
```

### Common Error Codes
- `VALIDATION_ERROR` (400) - Invalid input data
- `UNAUTHORIZED` (401) - Missing or invalid authentication
- `FORBIDDEN` (403) - Insufficient permissions or tier limits
- `NOT_FOUND` (404) - Resource doesn't exist
- `RATE_LIMIT_EXCEEDED` (429) - Too many requests
- `SPOTIFY_API_ERROR` (502) - Spotify integration issues
- `SYNC_IN_PROGRESS` (409) - Another sync operation is running

---

## Implementation Priority

### Phase 1: Core CRUD ✅ COMPLETED
- ✅ Base playlist CRUD operations (Create, Read, Delete)
- ✅ Framework-agnostic HTTP abstraction
- ✅ Comprehensive unit and integration tests
- ✅ PocketBase collection setup and repository pattern
- ⏳ Child playlist CRUD operations (TODO)

### Phase 2: Authentication & Integration ⏳ IN PROGRESS  
- ✅ Spotify OAuth flow (login/callback handlers)
- ✅ User and SpotifyIntegration models
- ✅ SpotifyClient with HTTP mocking
- ⏳ JWT authentication middleware (TODO)
- ⏳ Complete createOrUpdateUser implementation (TODO)
- ⏳ Spotify API integration endpoints (TODO)

### Phase 3: Sync Operations (TODO)
- Manual sync triggers
- Automated sync operations  
- Sync status and history
- Background processing

### Phase 4: Business Logic (TODO)
- Subscription management
- Usage tracking and limits
- Stripe webhook integration
- Tier enforcement

### Phase 5: Analytics & Polish (TODO)
- User analytics endpoints
- Performance monitoring
- Advanced error handling
- Rate limiting implementation

---

## Development Notes

### PocketBase Integration
- Leverage PocketBase's built-in CRUD operations for basic endpoints
- Use PocketBase relations for foreign keys (users ↔ spotify_integrations, base_playlists)
- Implement custom authentication handlers for Spotify OAuth flow
- Use repository pattern with PocketBase as the data layer
- PocketBase automatically handles field encryption for sensitive token data

### Security Considerations
- All endpoints require JWT authentication except Spotify OAuth flow
- Implement user isolation through PocketBase access rules  
- PocketBase automatically encrypts sensitive fields (tokens)
- One-to-one user/Spotify account relationship prevents conflicts
- Validate Spotify playlist ownership before operations
- Rate limit per-user API calls

### Performance Optimizations
- Implement pagination for all list endpoints (default 30 items)
- Cache Spotify API responses where appropriate
- Use database indexes for frequent query patterns
- Batch operations for sync processes
- Implement request queuing for sync operations