# PlaylistSync Database Schema - PocketBase Collections

## Overview

This document defines the complete database schema for PlaylistSync using PocketBase collections. The design leverages PocketBase's built-in features while maintaining clean separation of concerns across authentication, billing, integrations, and playlist management.

## Collection Architecture

```
┌─────────────────┐
│ users (built-in)│
└─────────┬───────┘
          │
          ├── subscriptions (1:many)
          ├── spotify_integrations (1:1)
          ├── base_playlists (1:many)
          └── child_playlists (1:many)
                    │
                    └── base_playlists (many:1)
```

---

## 1. Users Collection (Built-in)

**Collection Name:** `users`  
**Type:** Built-in PocketBase authentication collection  
**Purpose:** Core user authentication and profile management

### Schema
```typescript
interface User {
  id: string;              // Auto-generated UUID
  email: string;           // Unique, required
  emailVisibility: boolean; // PocketBase built-in
  verified: boolean;       // Email verification status
  username?: string;       // Optional, unique if provided
  avatar?: string;         // File upload, optional
  created: Date;           // Auto-generated
  updated: Date;           // Auto-updated
}
```

### Access Rules
```javascript
listRule: ""      // Admin only
viewRule: "id = @request.auth.id"
createRule: ""    // Public registration handled by PocketBase
updateRule: "id = @request.auth.id"
deleteRule: "id = @request.auth.id"
```

### Indexes
- `email` (unique, built-in)
- `username` (unique, built-in)

---

## 2. Subscriptions Collection

**Collection Name:** `subscriptions`  
**Purpose:** Manage user subscription tiers and billing integration with Stripe

### Schema
```typescript
interface Subscription {
  id: string;                    // Auto-generated UUID
  user: string;                  // Relation to users.id (required)
  
  // Stripe Integration
  stripe_subscription_id?: string; // Unique Stripe subscription ID
  stripe_customer_id?: string;     // Stripe customer ID
  
  // Subscription Details
  tier: 'basic' | 'premium';                              // Required
  status: 'active' | 'past_due' | 'canceled' | 'incomplete' | 'unpaid'; // Required
  current_period_start: Date;     // Billing period start (required)
  current_period_end: Date;       // Billing period end (required)
  cancel_at_period_end: boolean;  // Default: false
  
  // Timestamps
  created: Date;                  // Auto-generated
  updated: Date;                  // Auto-updated
}
```

### Field Validations
- `stripe_subscription_id`: Unique when provided
- `stripe_customer_id`: Required when `stripe_subscription_id` exists
- `current_period_end` must be after `current_period_start`
- Only one `active` subscription per user (enforced at application level)

### Access Rules
```javascript
listRule: "user = @request.auth.id"
viewRule: "user = @request.auth.id"
createRule: ""    // Backend only (via Stripe webhooks)
updateRule: ""    // Backend only (via Stripe webhooks)
deleteRule: ""    // Backend only
```

### Indexes
- `user` (for user subscription lookup)
- `stripe_subscription_id` (unique, for webhook processing)
- `status` (for billing operations)
- `current_period_end` (for expiration checks)

---

## 3. Spotify Integrations Collection

**Collection Name:** `spotify_integrations`  
**Purpose:** Store Spotify OAuth tokens and user linking (one-to-one relationship with users)

### Schema
```typescript
interface SpotifyIntegration {
  id: string;                  // Auto-generated UUID
  user: string;                // Relation to users.id (required, unique)
  
  // Spotify Account Details
  spotify_id: string;          // Spotify user ID (required, unique)
  display_name?: string;       // Spotify display name
  
  // OAuth Tokens (secured/encrypted by PocketBase)
  access_token: string;        // Required
  refresh_token: string;       // Required
  token_type: string;          // Default: "Bearer"
  expires_at: Date;            // Token expiration timestamp
  scope?: string;              // Granted permissions (space-separated)
  
  // Status
  is_active: boolean;          // Default: true
  
  // Timestamps
  created: Date;               // Auto-generated
  updated: Date;               // Auto-updated
}
```

### Field Validations
- `user`: Unique (one Spotify integration per user)
- `spotify_id`: Unique globally (one Spotify account per integration)
- `access_token` and `refresh_token`: Required for active integrations
- `expires_at`: Must be future date when creating/updating tokens

### Access Rules
```javascript
listRule: "user = @request.auth.id"
viewRule: "user = @request.auth.id"
createRule: "user = @request.auth.id"
updateRule: "user = @request.auth.id"
deleteRule: "user = @request.auth.id"
```

### Indexes
- `user` (unique, for user integration lookup)
- `spotify_id` (unique, for Spotify account lookup)
- `is_active` (for active integrations)
- `expires_at` (for token refresh operations)

---

## 4. Base Playlists Collection

**Collection Name:** `base_playlists`  
**Purpose:** Track user's primary playlists that serve as sources for distribution

### Schema
```typescript
interface BasePlaylist {
  id: string;                  // Auto-generated UUID
  user: string;                // Relation to users.id (required)
  
  // Playlist Details
  name: string;                // User-friendly name (required)
  spotify_playlist_id: string; // Spotify playlist ID (required)
  
  // Status & Sync
  is_active: boolean;          // Default: true
  last_synced?: Date;          // Last successful sync timestamp
  sync_status: 'never_synced' | 'syncing' | 'success' | 'error'; // Default: never_synced
  
  // Timestamps
  created: Date;               // Auto-generated
  updated: Date;               // Auto-updated
}
```

### Field Validations
- `spotify_playlist_id`: Unique per user (user can't add same playlist twice)
- `name`: 1-100 characters

### Access Rules
```javascript
listRule: "user = @request.auth.id"
viewRule: "user = @request.auth.id"
createRule: "user = @request.auth.id"
updateRule: "user = @request.auth.id"
deleteRule: "user = @request.auth.id"
```

### Indexes
- `user` (for user playlist lookup)
- `user + spotify_playlist_id` (unique combination)
- `is_active` (for active playlists)
- `sync_status` (for sync operations)

---

## 5. Child Playlists Collection

**Collection Name:** `child_playlists`  
**Purpose:** Store filtered playlists with rules for automatic song distribution

### Schema
```typescript
interface ChildPlaylist {
  id: string;                  // Auto-generated UUID
  user: string;                // Relation to users.id (required)
  base_playlist: string;       // Relation to base_playlists.id (required)
  
  // Playlist Details
  name: string;                // User-friendly name (required)
  spotify_playlist_id: string; // Spotify playlist ID (required)
  
  // Filtering Rules
  filter_rules: FilterRules;   // JSON object with comprehensive filtering
  exclusion_rules: ExclusionRules; // JSON object for artist/song exclusions
  
  // Status & Sync
  is_active: boolean;          // Default: true
  last_synced?: Date;          // Last successful sync timestamp
  sync_status: 'never_synced' | 'syncing' | 'success' | 'error'; // Default: never_synced
  songs_count: number;         // Current number of songs, default: 0
  
  // Timestamps
  created: Date;               // Auto-generated
  updated: Date;               // Auto-updated
}

// Supporting Types
interface FilterRules {
  audio_features?: {
    energy?: { min?: number; max?: number };
    danceability?: { min?: number; max?: number };
    valence?: { min?: number; max?: number };
    tempo?: { min?: number; max?: number };
    acousticness?: { min?: number; max?: number };
    instrumentalness?: { min?: number; max?: number };
    liveness?: { min?: number; max?: number };
    speechiness?: { min?: number; max?: number };
    loudness?: { min?: number; max?: number };
    key?: number[];
    mode?: (0 | 1)[];
    time_signature?: number[];
  };
  metadata?: {
    genres?: string[];
    year_range?: { min?: number; max?: number };
    popularity?: { min?: number; max?: number };
    duration_ms?: { min?: number; max?: number };
  };
}

interface ExclusionRules {
  artists?: string[];          // Artist names to exclude
  songs?: string[];            // Song titles to exclude
  spotify_artist_ids?: string[]; // Spotify artist IDs to exclude
  spotify_track_ids?: string[];  // Spotify track IDs to exclude
}
```

### Field Validations
- `spotify_playlist_id`: Unique per user
- `name`: 1-100 characters
- `filter_rules`: Valid JSON conforming to FilterRules interface
- `base_playlist.user` must equal `user` (enforced via access rules)

### Access Rules
```javascript
listRule: "user = @request.auth.id"
viewRule: "user = @request.auth.id"
createRule: "user = @request.auth.id && base_playlist.user = @request.auth.id"
updateRule: "user = @request.auth.id"
deleteRule: "user = @request.auth.id"
```

### Indexes
- `user` (for user playlist lookup)
- `base_playlist` (for child playlist lookup)
- `user + spotify_playlist_id` (unique combination)
- `is_active` (for active playlists)
- `sync_status` (for sync operations)

---

## Business Logic & Computed Fields

### Current Subscription Tier
```typescript
function getCurrentTier(userId: string): 'free' | 'basic' | 'premium' {
  const activeSubscription = subscriptions.find(
    s => s.user === userId && s.status === 'active'
  );
  return activeSubscription ? activeSubscription.tier : 'free';
}
```

### Usage Limits by Tier
```typescript
const TIER_LIMITS = {
  free: { monthly_syncs: 10, base_playlists: 0, child_playlists: 0 },
  basic: { monthly_syncs: Infinity, base_playlists: 2, child_playlists_per_base: 5 },
  premium: { monthly_syncs: Infinity, base_playlists: Infinity, child_playlists_per_base: Infinity }
};
```

---

## Data Relationships

### One-to-Many Relationships
- `users` → `subscriptions` (user can have subscription history)
- `users` → `base_playlists` (user can have multiple base playlists)
- `users` → `child_playlists` (user can have multiple child playlists)
- `base_playlists` → `child_playlists` (base playlist can have multiple children)

### One-to-One Relationships
- `users` → `spotify_integrations` (user can have one Spotify integration)

### Constraints
- User can have only one active subscription at a time
- User can have only one Spotify integration (enforced by unique user field)
- Each Spotify account can only be linked to one user (enforced by unique spotify_id)
- User cannot add the same Spotify playlist twice (as base or child)
- Child playlist must belong to same user as its base playlist

---

## Migration & Scaling Considerations

### Future Schema Evolution
- **Playlist Templates**: Add `playlist_templates` collection for reusable filter configurations
- **Usage Tracking**: Add `monthly_usage` collection if detailed analytics needed
- **Sync Logs**: Add `sync_logs` collection for debugging and audit trails
- **Multi-provider**: Add additional integration collections for Apple Music, YouTube Music (e.g., `apple_music_integrations`)

### Performance Optimizations
- Consider materialized views for complex playlist relationship queries
- Cache frequently accessed data (user tier, playlist counts)
- Implement pagination for users with many playlists
- Background sync processing to avoid blocking user operations

### Security Considerations
- OAuth tokens are automatically secured by PocketBase field encryption
- Implement rate limiting per user for API operations
- Audit trail for subscription changes and billing events
- Secure webhook endpoints for Stripe integration
- Unique constraints prevent account linking conflicts